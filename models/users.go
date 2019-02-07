package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint

	Username  string
	Password  string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

const sqlUserGet = `
SELECT id, username, password, created_at, updated_at
FROM users
WHERE ;`

const sqlUserInsert = `
INSERT INTO "public"."users" (id, username, path, method, content_length, "from", body)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func GetUserByUsername(usernameStr string) (user User, err error) {
	var id int
	var username string
	var password string

	var createdAt time.Time
	var updatedAt time.Time

	err = DB.QueryRow(sqlUserGet, usernameStr).Scan(&id, &username, &password, &createdAt, &updatedAt)

	return
}