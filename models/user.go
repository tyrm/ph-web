package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/eefret/gravatar"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
)

// User represents a pup haus user
type User struct {
	ID    int
	Token string

	Username string
	Password string
	Email    string

	LoginCount int
	LastLogin  pq.NullTime

	CreatedAt pq.NullTime
	UpdatedAt pq.NullTime
	DeletedAt pq.NullTime
}

// CheckPassword returns true if entered value matches the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// GetCreatedAt returns formatted string of CreatedAt
func (u *User) GetCreatedAt() string {
	timeStr := ""

	if u.CreatedAt.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.CreatedAt.Time.Year(), u.CreatedAt.Time.Month(), u.CreatedAt.Time.Day(),
			u.CreatedAt.Time.Hour(), u.CreatedAt.Time.Minute(), u.CreatedAt.Time.Second())
	}
	return timeStr
}

// GetGravatar returns url of Gravatar icon
func (u *User) GetGravatar(size int) string {
	g, err := gravatar.New()
	if err != nil {
		logger.Errorf("Error making gravatar client: %s", err.Error())
		return ""
	}

	g.SetSize(uint(size))
	return g.URLParse(u.Email)
}

// GetLastLogin returns formatted string of LastLogin
func (u *User) GetLastLogin() string {
	timeStr := ""

	if u.LastLogin.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.LastLogin.Time.Year(), u.LastLogin.Time.Month(), u.LastLogin.Time.Day(),
			u.LastLogin.Time.Hour(), u.LastLogin.Time.Minute(), u.LastLogin.Time.Second())
	}

	return timeStr
}

// GetUpdatedAt returns formatted string of UpdatedAt
func (u *User) GetUpdatedAt() string {
	timeStr := ""

	if u.UpdatedAt.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.UpdatedAt.Time.Year(), u.UpdatedAt.Time.Month(), u.UpdatedAt.Time.Day(),
			u.UpdatedAt.Time.Hour(), u.UpdatedAt.Time.Minute(), u.UpdatedAt.Time.Second())
	}
	return timeStr
}
const sqlUserUpdateLastLogin = `
UPDATE users
SET login_count = login_count + 1, last_login = now()
WHERE token = $1
RETURNING login_count, last_login;`

// UpdateLastLogin updates the LastLogin field in the database to now.
func (u *User) UpdateLastLogin() (err error) {
	var loginCount int
	var newTime time.Time
	err = db.QueryRow(sqlUserUpdateLastLogin, u.Token).Scan(&loginCount, &newTime)
	if err != nil {
		logger.Errorf("Error estimating user count: %s", err.Error())
		return
	}
	u.LoginCount = loginCount
	u.LastLogin.Time = newTime
	logger.Tracef("UpdateLastLogin(%d) (%v)[%d, %s]", u.Token, err, loginCount, newTime)
	return
}

// publics
const sqlUserInsert = `
INSERT INTO "public"."users" (token, username, password, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;`

// CreateUser creates a new instance of a user in the database.
func CreateUser(username string, password string, email string) (user User, err error) {
	createdAt := pq.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	passHash, err := hashPassword(password)
	if err != nil {
		logger.Errorf("Error hashing password: %s", err.Error())
		return
	}

	newUser := User{
		Username:  username,
		Password:  passHash,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	newUser.Token = RandString(8)

	var newID int
	err = db.QueryRow(sqlUserInsert, newUser.Token, newUser.Username, newUser.Password, newUser.Email, newUser.CreatedAt, newUser.UpdatedAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("pq error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser.ID = newID

	logger.Debugf("New user created: %s", newUser.Token)
	user = newUser
	return
}

const sqlUserCount = `
SELECT count(*)
FROM users
WHERE deleted_at IS NULL;`

// GetUserCount returns number of users in the database.
func GetUserCount() (count uint, err error) {
	err = db.QueryRow(sqlUserCount).Scan(&count)
	if err != nil {
		logger.Errorf("Error estimating user count: %s", err.Error())
	}
	logger.Tracef("GetUserCount() (%d, %v)", count, err)
	return
}

const sqlUserGetUsernameByID = `
SELECT username
FROM users
WHERE id = $1 AND deleted_at IS NULL;`

// GetUsernameByID returns the username of the provided id
func GetUsernameByID(uid int) string {
	var username string

	uisStr := strconv.Itoa(uid)
	if u, found := cUsernameByID.Get(uisStr); found {
		username = u.(string)
		logger.Tracef("GetUsernameByID(%d) (%s) [HIT]", uid, username)
		return username
	}

	err := db.QueryRow(sqlUserGetUsernameByID, uid).Scan(&username)
	if err != nil {
		logger.Errorf(err.Error())
		return uisStr
	}

	cUsernameByID.Set(uisStr, username, cache.DefaultExpiration)
	logger.Tracef("GetUsernameByID(%d) (%s) [MISS]", uid, username)
	return username
}

const sqlUsernameExists = `
SELECT exists(SELECT 1 FROM users WHERE lower(username)=lower($1) AND deleted_at IS NULL);`

// GetUsernameExists returns true if username exists in the database
func GetUsernameExists(username string) (exists bool, err error) {
	var newExists bool

	err = db.QueryRow(sqlUsernameExists, username).Scan(&newExists)
	if err != nil {
		logger.Errorf("Error checking is user username exists: %s", err.Error())
		return
	}
	exists = newExists
	logger.Tracef("GetUsernameExists(%s) (%v, %v)", username, newExists, err)
	return
}

const sqlUserGet = `
SELECT id, token, username, password, email, login_count, last_login, created_at, updated_at
FROM users
WHERE token = $1 AND deleted_at IS NULL;`

// ReadUser returns a user by is from the database
func ReadUser(sid string) (user *User, err error) {
	var id int
	var token string
	var username string
	var password string
	var email string

	var loginCount int
	var lastLogin pq.NullTime

	var createdAt pq.NullTime
	var updatedAt pq.NullTime

	err = db.QueryRow(sqlUserGet, sid).Scan(&id, &token, &username, &password, &email, &loginCount, &lastLogin, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	user = &User{
		ID:         id,
		Token:      token,
		Username:   username,
		Password:   password,
		Email:      email,
		LoginCount: loginCount,
		LastLogin:  lastLogin,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	logger.Tracef("ReadUserByUsername(%s) (%v, %v)", sid, user.ID, err)
	return
}

const sqlUserGetByUsername = `
SELECT id, token, username, password, email, login_count, last_login, created_at, updated_at
FROM users
WHERE lower(username) = lower($1) AND deleted_at IS NULL;`

// ReadUserByUsername returns a user by username from the database
func ReadUserByUsername(usernameStr string) (user User, err error) {
	var id int
	var token string
	var username string
	var password string
	var email string

	var loginCount int
	var lastLogin pq.NullTime

	var createdAt pq.NullTime
	var updatedAt pq.NullTime

	err = db.QueryRow(sqlUserGetByUsername, usernameStr).Scan(&id, &token, &username, &password, &email, &loginCount, &lastLogin, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	user = User{
		ID:         id,
		Token:      token,
		Username:   username,
		Password:   password,
		Email:      email,
		LoginCount: loginCount,
		LastLogin:  lastLogin,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	logger.Tracef("ReadUserByUsername(%s) (%v, %v)", usernameStr, user, err)
	return
}

const sqlUsersGetPage = `
SELECT token, username, email, login_count, last_login, created_at, updated_at
FROM users WHERE deleted_at IS NULL
ORDER BY created_at asc LIMIT $1 OFFSET $2;`

// ReadUsersPage returns a paginated group of users
func ReadUsersPage(limit uint, page uint) (userList []*User, err error) {
	offset := limit * page
	var newUserList []*User

	rows, err := db.Query(sqlUsersGetPage, limit, offset)
	if err != nil {
		logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
		return
	}
	for rows.Next() {
		var token string
		var username string
		var email string

		var loginCount int
		var lastLogin pq.NullTime

		var createdAt pq.NullTime
		var updatedAt pq.NullTime

		err = rows.Scan(&token, &username, &email, &loginCount, &lastLogin, &createdAt, &updatedAt)
		if err != nil {
			logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
			return
		}

		newUser := User{
			Token:      token,
			Username:   username,
			Email:      email,
			LoginCount: loginCount,
			LastLogin:  lastLogin,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}

		newUserList = append(newUserList, &newUser)
	}

	userList = newUserList
	logger.Tracef("ReadUsersPage(%d, %d) ([%d]User, %v)", limit, page, len(userList), nil)

	return
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}