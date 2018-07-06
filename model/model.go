package model

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	NameLength = 12
)

type Session struct {
	Key        [40]byte // `key_id`
	UserID     uint32   // `user_id`
	IP         uint32   // `ip`
	AccessTime uint32   // `access_time`
	Flags      uint8    // `flags`
}
type User struct {
	Id            uint32 `json:"id"`
	Name          string `json:"name"`
	Email         string
	Password      string `json:"password"`
	Perm          uint32
	GroupId       uint32
	Active        uint8
	LastLoginTime time.Time
}

func (u *User) DeleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE `id`=?", u.Id)
	return err
}

func (u *User) GetUser(db *sql.DB) error {
	switch {
	case !u.validateName():
		fmt.Errorf("invalid user name")
	case !u.validatePassword():
		fmt.Errorf("invalid user password")
	}

	jwtPassword, err := u.convertPassword()
	if err != nil {
		return err
	}
	row := db.QueryRow("SELECT id, perm, name, email, active, groupId FROM users WHERE `password`=?", jwtPassword)

	err = row.Scan(&u.Id, &u.Perm, &u.Name, &u.Email, &u.Active, &u.GroupId)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) CreateUser(db *sql.DB) error {
	switch {
	case !u.validateName():
		fmt.Errorf("invalid user name")
	case !u.validatePassword():
		fmt.Errorf("invalid user password")
	}

	jwtPassword, err := u.convertPassword()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (password, email, name, perm, groupId, active, create_time) VALUES (?, ?, ?, ?, ?, ?, NOW())",
		jwtPassword, u.Email, u.Name, u.Perm, u.GroupId, u.Active)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.Id)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUsers(db *sql.DB, start, count int) ([]User, error) {
	rows, err := db.Query("SELECT id, name FROM users LIMIT ? OFFSET ?", count, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (u *User) GetSessionByKey(db *sql.DB, keyStr string) (session Session, ok bool) {

	key := [40]byte{}
	copy(key[:], keyStr)

	row := db.QueryRow("SELECT user_id, ip, access_time, flags FROM login WHERE key_id=?", key[:])

	err := row.Scan(&session.UserID, &session.IP, &session.AccessTime, &session.Flags)

	if err != nil {
		return
	}

	ok = true
	return
}

func (u *User) GetUserById(db *sql.DB, userId uint32) (user User, ok bool) {

	row := db.QueryRow("SELECT id, perm, name, email, active, groupId FROM users WHERE `id`=?", userId)

	err := row.Scan(&user.Id, &user.Perm, &user.Name, &user.Email, &user.Active, &user.GroupId)

	if err != nil {
		return
	}

	ok = true
	return
}

func (u *User) SetSession(db *sql.DB, session Session) bool {

	_, err := db.Exec("INSERT INTO login (user_id, ip, access_time, key_id, flags) VALUES (?, ?, UNIX_TIMESTAMP(), ?, ?)",
		session.UserID, session.IP, session.Key[:], session.Flags)

	if err != nil {
		return false
	}

	return true
}

func (u *User) validateName() (ok bool) {
	if len(u.Name) != NameLength {
		return false
	}

	for _, c := range u.Name {
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || c == '@' {
			continue
		}
		return false
	}
	return true
}

func (u *User) validatePassword() (ok bool) {
	if len(u.Password) < 100 || len(u.Password) > 200 {
		return false
	}

	for _, c := range u.Password {
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || c == '.' {
			continue
		}
		return false
	}
	return true
}

func (u *User) convertPassword() (jwtPassword string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Email,
		"password": u.Password,
	})
	jwtPassword, err = token.SignedString([]byte("secret"))
	return
}
