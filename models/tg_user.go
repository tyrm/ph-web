package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/patrickmn/go-cache"
)

type TGUser struct {
	ID           int
	APIID        int
	IsBot        bool
	FirstName    string
	LastName     sql.NullString
	Username     sql.NullString
	LanguageCode sql.NullString
	CreatedAt    time.Time
	LastSeen     time.Time

	// NonDB
	ProfilePhotoURL string
}

// GetCreatedAtHuman returns humanized string of CreatedAt
func (u *TGUser) GetCreatedAtHuman() string {
	return humanize.Time(u.CreatedAt)
}

// GetCreatedAtFormatted returns formatted string of CreatedAt
func (u *TGUser) GetCreatedAtFormatted() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		u.CreatedAt.Year(), u.CreatedAt.Month(), u.CreatedAt.Day(),
		u.CreatedAt.Hour(), u.CreatedAt.Minute(), u.CreatedAt.Second())

	return timeStr
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGUser) GetLastSeenHuman() string {
	return humanize.Time(u.LastSeen)
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGUser) GetLastSeenFormatted() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		u.LastSeen.Year(), u.LastSeen.Month(), u.LastSeen.Day(),
		u.LastSeen.Hour(), u.LastSeen.Minute(), u.LastSeen.Second())

	return timeStr
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGUser) GetLongFormattedName() string {

	var nameStr []string
	nameStr = append(nameStr, u.FirstName)

	if u.LastName.Valid {
		nameStr = append(nameStr, u.LastName.String)
	}
	if u.Username.Valid {
		nameStr = append(nameStr, fmt.Sprintf("(@%s)", u.Username.String))
	}

	return strings.Join(nameStr, " ")
}

// GetName returns long formatted
func (u *TGUser) GetName() string {
	var nameStr []string
	nameStr = append(nameStr, u.FirstName)

	if u.LastName.Valid {
		nameStr = append(nameStr, u.LastName.String)
	}

	return strings.Join(nameStr, " ")
}

const sqlTGUserCount = `
SELECT count(*)
FROM tg_users;`

// GetUserCount returns number of users in the database.
func GetTGUserCount() (count uint, err error) {
	err = db.QueryRow(sqlTGUserCount).Scan(&count)
	if err != nil {
		logger.Errorf("Error getting tg_user count: %s", err.Error())
	}
	logger.Tracef("GetTGUserCount() (%d, %v)", count, err)
	return
}

const sqlReadTGUser = `
SELECT DISTINCT ON (tg_users.id) tg_users.id, tg_users.api_id, tg_users.is_bot, tg_users_history.first_name, 
    tg_users_history.last_name, tg_users_history.username, tg_users_history.language_code, tg_users.created_at, tg_users_history.last_seen
FROM tg_users LEFT JOIN tg_users_history
ON tg_users."id" = tg_users_history.tgu_id
WHERE tg_users.id = $1
ORDER BY tg_users.id ASC, tg_users_history.last_seen DESC;`

func ReadTGUser(id int) (user *TGUser, err error) {
	idStr := strconv.Itoa(id)
	if u, found := cTGUserByID.Get(idStr); found {
		user = u.(*TGUser)
		logger.Tracef("ReadTGUser(%d) (%s) [HIT]", id, user.APIID)
		return
	}

	var newID int
	var newAPIID int
	var newIsBot bool
	var newFirstName string
	var newLastName sql.NullString
	var newUsername sql.NullString
	var newLanguageCode sql.NullString
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGUser, id).Scan(&newID, &newAPIID, &newIsBot, &newFirstName, &newLastName,
		&newUsername, &newLanguageCode, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		logger.Tracef("ReadTGChatByAPIID(%d) (%v, %v)", id, nil, err)
		return
	}

	user = &TGUser{
		ID:           newID,
		APIID:        newAPIID,
		IsBot:        newIsBot,
		FirstName:    newFirstName,
		LastName:     newLastName,
		Username:     newUsername,
		LanguageCode: newLanguageCode,
		CreatedAt:    newCreatedAt,
		LastSeen:     newLastSeen,
	}

	logger.Tracef("ReadTGUser(%d) (%s) [MISS]", id, user.APIID)
	cTGUserByID.Set(idStr, user, cache.DefaultExpiration)
	return
}

const sqlReadTGUserPage = `
SELECT DISTINCT ON (tg_users.id) tg_users.id, tg_users.api_id, tg_users.is_bot, tg_users_history.first_name, 
    tg_users_history.last_name, tg_users_history.username, tg_users_history.language_code, tg_users.created_at, tg_users_history.last_seen
FROM tg_users LEFT JOIN tg_users_history
ON tg_users."id" = tg_users_history.tgu_id
ORDER BY tg_users.id ASC, tg_users_history.last_seen DESC LIMIT $1 OFFSET $2;`

func ReadTGUserPage(limit uint, page uint) (userList *[]TGUser, err error) {
	offset := limit * page
	var newUserList []TGUser

	rows, err := db.Query(sqlReadTGUserPage, limit, offset)
	if err != nil {
		logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
		return
	}
	for rows.Next() {
		var newID int
		var newAPIID int
		var newIsBot bool
		var newFirstName string
		var newLastName sql.NullString
		var newUsername sql.NullString
		var newLanguageCode sql.NullString
		var newCreatedAt time.Time
		var newLastSeen time.Time

		err = rows.Scan(&newID, &newAPIID, &newIsBot, &newFirstName, &newLastName,
			&newUsername, &newLanguageCode, &newCreatedAt, &newLastSeen)
		if err != nil {
			logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
			return
		}

		user := TGUser{
			ID:           newID,
			APIID:        newAPIID,
			IsBot:        newIsBot,
			FirstName:    newFirstName,
			LastName:     newLastName,
			Username:     newUsername,
			LanguageCode: newLanguageCode,
			CreatedAt:    newCreatedAt,
			LastSeen:     newLastSeen,
		}

		newUserList = append(newUserList, user)
	}

	userList = &newUserList
	return

}
