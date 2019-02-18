package models

import (
	"fmt"
	"time"

	"github.com/eefret/gravatar"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Token string

	Username  string
	Password  string
	Email     string

	LoginCount int
	LastLogin pq.NullTime

	CreatedAt pq.NullTime
	UpdatedAt pq.NullTime
	DeletedAt pq.NullTime
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) GetCreatedAt() string {
	timeStr := ""

	if u.CreatedAt.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.CreatedAt.Time.Year(), u.CreatedAt.Time.Month(), u.CreatedAt.Time.Day(),
			u.CreatedAt.Time.Hour(), u.CreatedAt.Time.Minute(), u.CreatedAt.Time.Second())
	}
	return timeStr
}

func (u *User) GetGravatar(size int) string {
	g, err := gravatar.New()
	if err != nil {
		logger.Errorf("Error making gravatar client: %s", err.Error())
		return ""
	}
	g.SetSize(uint(size))
	return g.URLParse(u.Email)
}

func (u *User) GetLastLogin() string {
	timeStr := ""

	if u.LastLogin.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.LastLogin.Time.Year(), u.LastLogin.Time.Month(), u.LastLogin.Time.Day(),
			u.LastLogin.Time.Hour(), u.LastLogin.Time.Minute(), u.LastLogin.Time.Second())
	}

	return timeStr
}

func (u *User) GetUpdatedAt() string {
	timeStr := ""

	if u.UpdatedAt.Valid == true {
		timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			u.UpdatedAt.Time.Year(), u.UpdatedAt.Time.Month(), u.UpdatedAt.Time.Day(),
			u.UpdatedAt.Time.Hour(), u.UpdatedAt.Time.Minute(), u.UpdatedAt.Time.Second())
	}
	return timeStr
}

func (u *User) UpdateLastLogin() (err error) {
	var loginCount int
	var newTime time.Time
	err = DB.QueryRow(sqlUserUpdateLastLogin, u.Token).Scan(&loginCount, &newTime)
	if err != nil {
		logger.Errorf("Error estimating user count: %s", err.Error())
		return
	}
	u.LoginCount = loginCount
	u.LastLogin.Time = newTime
	logger.Tracef("UpdateLastLogin(%d) (%v)[%d, %s]", u.Token, err, loginCount, newTime)
	return
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func EstimateCountUsers() (count uint, err error) {
	err = DB.QueryRow(sqlUserEstimateCount).Scan(&count)
	if err != nil {
		logger.Errorf("Error estimating user count: %s", err.Error())
	}
	logger.Tracef("EstimateCountUsers() (%d, %v)", count, err)
	return
}

func GetUserCount() (count uint, err error) {
	err = DB.QueryRow(sqlUserCount).Scan(&count)
	if err != nil {
		logger.Errorf("Error estimating user count: %s", err.Error())
	}
	logger.Tracef("GetUserCount() (%d, %v)", count, err)
	return
}

func GetUser(sid string) (user *User, err error) {
	var token string
	var username string
	var password string
	var email string

	var loginCount int
	var lastLogin pq.NullTime

	var createdAt pq.NullTime
	var updatedAt pq.NullTime

	err = DB.QueryRow(sqlUserGet, sid).Scan(&token, &username, &password, &email, &loginCount, &lastLogin, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	user = &User{
		Token:      token,
		Username:   username,
		Password:   password,
		Email:      email,
		LoginCount: loginCount,
		LastLogin:  lastLogin,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	logger.Tracef("GetUserByUsername(%s) (%v, %v)", sid, user, err)
	return
}

func GetUserByUsername(usernameStr string) (user User, err error) {
	var token string
	var username string
	var password string
	var email string

	var loginCount int
	var lastLogin pq.NullTime

	var createdAt pq.NullTime
	var updatedAt pq.NullTime

	err = DB.QueryRow(sqlUserGetByUsername, usernameStr).Scan(&token, &username, &password, &email, &loginCount, &lastLogin, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	user = User{
		Token:      token,
		Username:   username,
		Password:   password,
		Email:      email,
		LoginCount: loginCount,
		LastLogin:  lastLogin,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	logger.Tracef("GetUserByUsername(%s) (%v, %v)", usernameStr, user, err)
	return
}

func GetUserIdExists(id string)(exists bool, err error) {
	var newExists bool
	err = DB.QueryRow(sqlUserIdExists, id).Scan(&newExists)
	if err != nil {
		logger.Errorf("Error checking is user id exists: %s", err.Error())
		return
	}
	exists = newExists
	logger.Tracef("GetUserIdExists(%s) (%v, %v)", id, newExists, err)
	return
}

func GetUsernameExists(id string)(exists bool, err error) {
	var newExists bool

	err = DB.QueryRow(sqlUsernameExists, id).Scan(&newExists)
	if err != nil {
		logger.Errorf("Error checking is user id exists: %s", err.Error())
		return
	}
	exists = newExists
	logger.Tracef("GetUsernameExists(%s) (%v, %v)", id, newExists, err)
	return
}

func GetUsersPage(limit uint, page uint) (userList []*User, err error) {
	offset := limit * page
	var newUserList []*User

	rows, err := DB.Query(sqlUsersGetPage, limit, offset)
	if err != nil {
		logger.Tracef("GetUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
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
			logger.Tracef("GetUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
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
	logger.Tracef("GetUsersPage(%d, %d) ([%d]User, %v)", limit, page, len(userList), nil)

	return
}

func GetUsernameByID(uid string) string {
	var username string

	if u, found := cUsernameByID.Get(uid); found {
		username = u.(string)
		logger.Tracef("GetUsernameByID(%s) (%s) [HIT]", uid, username)
		return username
	}

	err := DB.QueryRow(sqlUserGetUsernameByID, uid).Scan(&username)
	if err != nil {
		logger.Errorf(err.Error())
		return uid
	}

	cUsernameByID.Set(uid, username, cache.DefaultExpiration)
	logger.Tracef("GetUsernameByID(%s) (%s) [MISS]", uid, username)
	return username
}

func NewUser(username string, password string, email string) (user User, err error) {
	createdAt := pq.NullTime{
		Time: time.Now(),
		Valid: true,
	}
	passHash, err := hashPassword(password)
	if err != nil {
		logger.Errorf("Error hashing password: %s", err.Error())
		return
	}

	newUser := User{
		Username: username,
		Password: passHash,
		Email: email,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	newUser.Token = RandString(8)

	var newID string
	err = DB.QueryRow(sqlUserInsert, newUser.Token, newUser.Username, newUser.Password, newUser.Email, newUser.CreatedAt, newUser.UpdatedAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("pq error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	logger.Debugf("New user created: %s", newUser.Token)
	user = newUser
	return
}

func getValidID() (id string, err error){
	var exists bool

	err = DB.QueryRow(sqlUserIdExists, id).Scan(&exists)
	if err != nil {
		logger.Errorf("Error inserting record into db: %s", err.Error())
		return
	}
	return
}