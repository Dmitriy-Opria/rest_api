package session

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"rest_api/model"
	"strings"
	"sync"
	"time"
)

type (
	Session interface {
		Key() string             // current sessionID
		Set(string, interface{}) // set session value
		Get(string) interface{}  // get session value
		Delete(string)           // delete session value
		Clear()                  //
		Time() time.Time         //
		SetTime(time.Time)       //
		SetIp(uint32)            //
		GetIp() uint32           //
		SetUID(uint32)           //
		GetUID() uint32          //
		SetPerm(uint32)          //
		GetPerm() uint32         //
		SetName(string)          //
		GetName() string         //
		SetEmail(string)         //
		GetEmail() string        //
	}

	Provider interface {
		GC(lifeTime time.Duration)
		Len() int
		Init()
		SessionInit(key string) Session
		SessionRead(key string) Session
	}

	Manager struct {
		cookieName string        // private cookiename
		domain     string        //
		provider   Provider      //
		lifetime   time.Duration //
		lock       sync.Mutex    // protects session
	}
)

var (
	provides = map[string]Provider{}
)

var (
	ErrProvideNil     = errors.New("session: Register provide is nil")
	ErrProvideDup     = errors.New("session: Provider already registar")
	ErrUnknownProvide = errors.New("session: unknown provide")
)

const (
	PermUser  uint32 = 0
	PermAdmin uint32 = 3
)

// ==

func init() {

	Register("memory", &MemoryProvider{})
}

func Register(name string, provider Provider) error {

	if provider == nil {
		return ErrProvideNil
	}

	if _, dup := provides[name]; dup {
		return ErrProvideDup
	}

	provider.Init()
	provides[name] = provider
	return nil
}

func NewManager(provideName, cookieName string, lifetime time.Duration) (manager *Manager, err error) {

	provider, ok := provides[provideName]
	if !ok {
		err = ErrUnknownProvide
		return
	}

	manager = &Manager{
		cookieName: cookieName,
		provider:   provider,
		lifetime:   lifetime,
	}

	go manager.gc()

	return
}

// ==

func (manager *Manager) gc() {
	for {
		time.Sleep(5 * time.Minute)
		manager.provider.GC(manager.lifetime)
	}
}

func (manager *Manager) sessionKey() string {

	dst := [30]byte{}

	rand.Read(dst[8:])
	binary.BigEndian.PutUint64(dst[:8], uint64(time.Now().UnixNano()))

	buf := new(bytes.Buffer)

	encoder := base64.NewEncoder(base64.RawURLEncoding, buf)
	encoder.Write(dst[:])
	encoder.Close()

	return buf.String()
}

func (manager *Manager) SessionGet(w http.ResponseWriter, r *http.Request, db *sql.DB) (session Session) {

	manager.lock.Lock()
	defer manager.lock.Unlock()

	if cookie, err := r.Cookie(manager.cookieName); err == nil && cookie.Value != "" {

		key := cookie.Value
		session = manager.provider.SessionRead(key)

		if session == nil {

			var user model.User
			if dbSession, ok := user.GetSessionByKey(db, key); ok {

				if user, ok := user.GetUserById(db, dbSession.UserID); ok {

					fmt.Println("load from DB")

					session = manager.provider.SessionInit(key)

					session.SetIp(dbSession.IP)
					session.SetUID(dbSession.UserID)
					session.SetPerm(user.Perm)
					session.SetName(user.Name)
					session.SetEmail(user.Email)
				}
			}

		} else {
			session.SetTime(time.Now())
		}
	}
	return
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request, db *sql.DB, user *model.User) (session Session) {

	manager.lock.Lock()
	defer manager.lock.Unlock()

	loadSession := func(key string) bool {

		if dbSession, ok := user.GetSessionByKey(db, key); ok {

			if user, ok := user.GetUserById(db, dbSession.UserID); ok {

				fmt.Println("load from DB")

				session = manager.provider.SessionInit(key)

				session.SetIp(dbSession.IP)
				session.SetUID(dbSession.UserID)
				session.SetPerm(user.Perm)
				session.SetName(user.Name)
				session.SetEmail(user.Email)

				return true
			}
		}

		return false
	}

	createSession := func() {

		key := manager.sessionKey()
		session = manager.provider.SessionInit(key)

		cookie := http.Cookie{
			Name:     manager.cookieName,
			Domain:   manager.domain,
			Value:    key,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(manager.lifetime),
		}

		http.SetCookie(w, &cookie)

		// ==

		var ip uint32
		var ipStr = r.RemoteAddr

		if n := strings.Index(ipStr, ":"); n != -1 {
			ipStr = ipStr[:n]
		}

		if ipByte := net.ParseIP(ipStr).To4(); ipByte != nil {
			binary.Read(bytes.NewReader([]byte(ipByte)), binary.BigEndian, &ip)
		}

		dbSession := model.Session{UserID: user.Id, IP: ip}
		copy(dbSession.Key[:], key)

		user.SetSession(db, dbSession) // TODO add to DB

		fmt.Println("save to DB")

		session.SetIp(ip)
		session.SetUID(user.Id)
		session.SetPerm(user.Perm)
		session.SetName(user.Name)
		session.SetEmail(user.Email)
	}

	if cookie, err := r.Cookie(manager.cookieName); err != nil || cookie.Value == "" {

		createSession()

	} else {

		key := cookie.Value
		session = manager.provider.SessionRead(key)

		if session == nil {

			if !loadSession(key) {

				createSession()
			}
		}
	}
	return
}
