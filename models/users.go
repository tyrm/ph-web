package models

import (
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
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

const sqlUserCount = `
SELECT count(*)
FROM users
WHERE deleted_at IS NULL;`

const sqlUserEstimateCount = `
SELECT n_live_tup
FROM pg_stat_all_tables
WHERE relname = 'users';`

const sqlUserGet = `
SELECT id, username, password, created_at, updated_at
FROM users
WHERE lower(username) = lower($1) AND deleted_at IS NULL;`

const sqlUserGetUsernameByID = `
SELECT username
FROM users
WHERE id = $1 AND deleted_at IS NULL;`

const sqlUserInsert = `
INSERT INTO "public"."users" (username, password, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING id;`

const sqlUsersGetPage = `
SELECT id, username, created_at, updated_at
FROM users WHERE deleted_at IS NULL
ORDER BY id asc LIMIT $1 OFFSET $2;`

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
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

func GetUserByUsername(usernameStr string) (user User, err error) {
	var id uint
	var username string
	var password string

	var createdAt time.Time
	var updatedAt time.Time

	err = DB.QueryRow(sqlUserGet, usernameStr).Scan(&id, &username, &password, &createdAt, &updatedAt)
	if err != nil {
		logger.Errorf(err.Error())
		return
	}

	user = User{
		ID: id,
		Username: username,
		Password: password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	logger.Tracef("GetUserByUsername(%s) (%v, %v)", usernameStr, user, err)
	return
}

func GetUsersPage(limit uint, page uint) (userList []*User, err error) {
	offset := limit * page
	var newUserList []*User

	rows, err := DB.Query(sqlUsersGetPage, limit, offset)
	if err != nil {
		logger.Tracef("GetUsernameByID(%d, %d) (%v, %v)", limit, page, nil, err)
		return
	}
	for rows.Next() {
		var id uint
		var username string
		var createdAt time.Time
		var updatedAt time.Time

		err = rows.Scan(&id, &username, &createdAt, &updatedAt)
		if err != nil {
			logger.Tracef("GetUsernameByID(%d, %d) (%v, %v)", limit, page, nil, err)
			return
		}

		newUser := User{
			ID: id,
			Username: username,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		newUserList = append(newUserList, &newUser)
	}

	userList = newUserList
	logger.Tracef("GetUsernameByID(%d, %d) ([%d]User, %v)", limit, page, len(userList), nil)

	return
}

func GetUsernameByID(uid uint) string {
	var username string

	if u, found := cUsernameByID.Get(strconv.FormatUint(uint64(uid), 10)); found {
		username = u.(string)
		logger.Tracef("GetUsernameByID(%d) (%s) [HIT]", uid, username)
		return username
	}

	err := DB.QueryRow(sqlUserGetUsernameByID, uid).Scan(&username)
	if err != nil {
		logger.Errorf(err.Error())
		return strconv.FormatUint(uint64(uid), 10)
	}

	cUsernameByID.Set(strconv.FormatUint(uint64(uid), 10), username, cache.DefaultExpiration)
	logger.Tracef("GetUsernameByID(%d) (%s) [MISS]", uid, username)
	return username
}

func NewUser(username string, password string) (user User, err error) {
	createdAt := time.Now()
	passHash, err := hashPassword(password)
	if err != nil {
		logger.Errorf("Error hashing password: %s", err.Error())
		return
	}

	newUser := User{
		Username: username,
		Password: passHash,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	var newID uint
	err = DB.QueryRow(sqlUserInsert, newUser.Username, newUser.Password, newUser.CreatedAt, newUser.UpdatedAt).Scan(&newID)
	if err != nil {
		logger.Errorf("Error inserting record into db: %s", err.Error())
		return
	}

	newUser.ID = newID
	logger.Tracef("New user created: %v", newUser)
	user = newUser
	return
}