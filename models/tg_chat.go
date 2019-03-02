package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

// TGChat represents a telegram chat
type TGChat struct {
	ID        int
	APIID     int64
	CreatedAt time.Time
}

func (tgu *TGChat) ReadLatestHistory() (*TGChatHistory, error) {
	return readLatestTGChatHistoryByTgcID(tgu.ID)
}

// TGChatHistory represents the varying data of a telegram chat
type TGChatHistory struct {
	ID                  int
	TGChatID            int
	Type                string
	Title               sql.NullString
	Username            sql.NullString
	FirstName           sql.NullString
	LastName            sql.NullString
	AllMembersAreAdmins bool
	CreatedAt           time.Time
	LastSeen            time.Time
}

func (tgc *TGChatHistory) Matches(apiUser *tgbotapi.Chat) bool {
	if tgc.Type != apiUser.Type {
		logger.Tracef("Matches() false [Type]")
		return false
	}

	if apiUser.Title != "" || tgc.Title.Valid != false {
		if apiUser.Title != tgc.Title.String {
			logger.Tracef("Matches() false [Title]")
			return false
		}
	}

	if apiUser.UserName != "" || tgc.Username.Valid != false {
		if apiUser.UserName != tgc.Username.String {
			logger.Tracef("Matches() false [Username]")
			return false
		}
	}

	if apiUser.FirstName != "" || tgc.FirstName.Valid != false {
		if apiUser.FirstName != tgc.FirstName.String {
			logger.Tracef("Matches() false [FirstName]")
			return false
		}
	}

	if apiUser.LastName != "" || tgc.LastName.Valid != false {
		if apiUser.LastName != tgc.LastName.String {
			logger.Tracef("Matches() false [LastName]")
			return false
		}
	}

	if tgc.AllMembersAreAdmins != apiUser.AllMembersAreAdmins {
		logger.Tracef("Matches() false [AllMembersAreAdmins]")
		return false
	}

	return true
}

const sqlUpdateTGChatHistoryLastSeen = `
UPDATE tg_chats_history
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGChatHistory) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGChatHistoryLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

// TGChatHistory represents the varying data of a telegram chat
type TGChatPage struct {
	ID        int
	APIID     int64
	Type      string
	Title     sql.NullString
	Username  sql.NullString
	FirstName sql.NullString
	LastName  sql.NullString
	CreatedAt time.Time
	LastSeen  time.Time
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChatPage) GetFormatedName() string {
	if u.Title.Valid {
		return u.Title.String
	}

	var nameStr []string
	nameStr = append(nameStr, u.FirstName.String)

	if u.LastName.Valid {
		nameStr = append(nameStr, u.LastName.String)
	}
	if u.Username.Valid {
		nameStr = append(nameStr, fmt.Sprintf("(@%s)", u.Username.String))
	}

	return strings.Join(nameStr, " ")
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChatPage) GetLastSeenHuman() string {
	return humanize.Time(u.LastSeen)
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChatPage) GetLastSeenFormated() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		u.LastSeen.Year(), u.LastSeen.Month(), u.LastSeen.Day(),
		u.LastSeen.Hour(), u.LastSeen.Minute(), u.LastSeen.Second())

	return timeStr
}

const sqlCreateTGChat = `
INSERT INTO "public"."tg_chats" (api_id, created_at)
VALUES ($1, $2)
RETURNING id;`

// CreateTGChat creates a new instance of a telegram chat in the database.
func CreateTGChat(apiID int64) (tgu *TGChat, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGChat, apiID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser := &TGChat{
		ID:        newID,
		APIID:     apiID,
		CreatedAt: createdAt,
	}
	tgu = newUser
	return
}

const sqlCreateTGChatHistory = `
INSERT INTO "public"."tg_chats_history" (tgc_id, "type", title, username, first_name, last_name, 
                                         all_members_are_admins, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;`

// CreateTGChatHistory creates a new instance of telegram user history in the database.
func CreateTGChatHistory(tgc *TGChat, newType string, title sql.NullString, username sql.NullString,
	firstName sql.NullString, lastName sql.NullString, allMembersAreAdmin bool) (tgch *TGChatHistory, err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGChatHistory, tgc.ID, newType, title, username, firstName, lastName, allMembersAreAdmin,
		createdAt, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	TGUserHistory := &TGChatHistory{
		ID:                  newID,
		TGChatID:            tgc.ID,
		Type:                newType,
		Title:               title,
		Username:            username,
		FirstName:           firstName,
		LastName:            lastName,
		AllMembersAreAdmins: allMembersAreAdmin,
		CreatedAt:           createdAt,
		LastSeen:            createdAt,
	}
	tgch = TGUserHistory
	return
}

func CreateTGChatHistoryFromAPI(tgChat *TGChat, apiChat *tgbotapi.Chat) (*TGChatHistory, error) {
	title := &sql.NullString{Valid: false}
	if apiChat.Title != "" {
		title = &sql.NullString{
			String: apiChat.Title,
			Valid:  true,
		}
	}

	username := &sql.NullString{Valid: false}
	if apiChat.UserName != "" {
		username = &sql.NullString{
			String: apiChat.UserName,
			Valid:  true,
		}
	}

	firstName := &sql.NullString{Valid: false}
	if apiChat.FirstName != "" {
		firstName = &sql.NullString{
			String: apiChat.FirstName,
			Valid:  true,
		}
	}

	lastName := &sql.NullString{Valid: false}
	if apiChat.LastName != "" {
		lastName = &sql.NullString{
			String: apiChat.LastName,
			Valid:  true,
		}
	}

	return CreateTGChatHistory(tgChat, apiChat.Type, *title, *username, *firstName, *lastName, apiChat.AllMembersAreAdmins)
}

const sqlReadTGChatByAPIID = `
SELECT id, api_id, created_at
FROM tg_chats
WHERE api_id = $1;`

// ReadTGChatByAPIID returns an instance of a telegram chat by api_id from the database.
func ReadTGChatByAPIID(apiID int64) (tgu *TGChat, err error) {
	var newID int
	var newAPIID int64
	var newCreatedAt time.Time

	err = db.QueryRow(sqlReadTGChatByAPIID, apiID).Scan(&newID, &newAPIID, &newCreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newChat := &TGChat{
		ID:        newID,
		APIID:     newAPIID,
		CreatedAt: newCreatedAt,
	}
	tgu = newChat
	return
}

const sqlReadLatestTGChatHistoryByTguID = `
SELECT id, tgc_id, "type", title, username, first_name, last_name, all_members_are_admins, created_at, last_seen
FROM tg_chats_history
WHERE tgc_id = $1
ORDER BY created_at DESC
LIMIT 1;`

// ReadTGUserHistory returns an instance of a telegram user by all fields from the database.
func readLatestTGChatHistoryByTgcID(tgcID int) (tgch *TGChatHistory, err error) {
	var newID int
	var newTGChatID int
	var newType string
	var newTitle sql.NullString
	var newUsername sql.NullString
	var newFirstName sql.NullString
	var newLastName sql.NullString
	var newAllMembersAreAdmins bool
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadLatestTGChatHistoryByTguID, tgcID).Scan(&newID, &newTGChatID, &newType, &newTitle,
		&newUsername, &newFirstName, &newLastName, &newAllMembersAreAdmins, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	TGChatHistory := &TGChatHistory{
		ID:                  newID,
		TGChatID:            newTGChatID,
		Type:                newType,
		Title:               newTitle,
		Username:            newUsername,
		FirstName:           newFirstName,
		LastName:            newLastName,
		AllMembersAreAdmins: newAllMembersAreAdmins,
		CreatedAt:           newCreatedAt,
		LastSeen:            newLastSeen,
	}
	tgch = TGChatHistory
	return
}

const sqlReadTGChatPage = `
SELECT DISTINCT ON (tg_chats.id) tg_chats.id, tg_chats.api_id, tg_chats_history."type", tg_chats_history.title, 
	tg_chats_history.username, tg_chats_history.first_name, tg_chats_history.last_name, tg_chats.created_at, 
    tg_chats_history.last_seen
FROM tg_chats LEFT JOIN tg_chats_history
ON tg_chats."id" = tg_chats_history.tgc_id
ORDER BY tg_chats.id ASC, tg_chats_history.last_seen DESC LIMIT $1 OFFSET $2;`

func ReadTGChatPage(limit uint, page uint) (chatList []*TGChatPage, err error) {
	offset := limit * page
	var newChatList []*TGChatPage

	rows, err := db.Query(sqlReadTGChatPage, limit, offset)
	if err != nil {
		logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
		return
	}
	for rows.Next() {
		var newID int
		var newAPIID int64
		var newType string
		var newTitle sql.NullString
		var newUsername sql.NullString
		var newFirstName sql.NullString
		var newLastName sql.NullString
		var newCreatedAt time.Time
		var newLastSeen time.Time

		err = rows.Scan(&newID, &newAPIID, &newType, &newTitle, &newUsername, &newFirstName, &newLastName,
			&newCreatedAt, &newLastSeen)
		if err != nil {
			logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
			return
		}

		tgChatHistory := &TGChatPage{
			ID:        newID,
			APIID:     newAPIID,
			Type:      newType,
			Title:     newTitle,
			Username:  newUsername,
			FirstName: newFirstName,
			LastName:  newLastName,
			CreatedAt: newCreatedAt,
			LastSeen:  newLastSeen,
		}

		newChatList = append(newChatList, tgChatHistory)
	}

	chatList = newChatList
	return

}
