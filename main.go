package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"rest_api/config"
	"rest_api/model"
	"rest_api/recaptcha"
	"rest_api/session"
	"strconv"
	"time"
)

type App struct {
	Router  *mux.Router
	Manager *session.Manager
	DB      *sql.DB
}

type Req struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Response string `json:"g-recaptcha-Response"`
}

const (
	cookieName  = "SID"
	sessionsTTL = 24 * time.Hour
)

func (a *App) InitManager() {
	var err error

	manager, err := session.NewManager("memory", cookieName, sessionsTTL)
	if err != nil {
		os.Exit(1)
	}
	a.Manager = manager
}

func (a *App) InitializeDb(host, user, password, dbname string) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&multiStatements=true",
		user, password,
		host, dbname)
	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) InitializeRoute() {
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func main() {
	a := App{}
	conf := config.Get()

	a.InitManager()
	a.InitializeDb(conf.MySqlHost, conf.MySqlUser, conf.MySqlPassword, conf.MySqlDB)
	a.InitializeRoute()
	a.Run(conf.Bind)
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/login", a.login).Methods("POST")
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/user", a.createUser).Methods("POST")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {

	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	sess := a.Manager.SessionGet(w, r, a.DB)
	if perm := sess.GetPerm(); perm == session.PermAdmin || perm == session.PermUser {

		var user model.User
		users, err := user.GetUsers(a.DB, start, count)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, users)
	} else {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("invalid user"))
	}
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Email != "" && req.Password != "" {

		if err := recaptcha.Verify(recaptcha.Request{
			Response: req.Response,
			RemoteIp: r.RemoteAddr,
		}); err == nil {
			var user = model.User{
				Name:     req.Email,
				Password: req.Password,
			}
			if err := user.GetUser(a.DB); err == nil {
				a.Manager.SessionStart(w, r, a.DB, &user)
				return

			} else {
				fmt.Println("user not found")
			}
		} else {
			fmt.Println("recaptcha verify error")
		}
	}
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request p")
		return
	}

	sess := a.Manager.SessionGet(w, r, a.DB)
	if perm := sess.GetPerm(); perm == session.PermAdmin || perm == session.PermUser {
		if err := user.CreateUser(a.DB); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, user)
	} else {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("invalid user"))
	}
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user Id")
		return
	}

	sess := a.Manager.SessionGet(w, r, a.DB)
	if perm := sess.GetPerm(); perm == session.PermAdmin {
		user := model.User{Id: uint32(id)}
		if err := user.DeleteUser(a.DB); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, user)
	} else {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("invalid user"))
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
