package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/lib/pq"
)

// TGChatMeta represents a telegram chat
type TGChatMeta struct {
	ID        int
	APIID     int64
	CreatedAt time.Time
}

func (tgu *TGChatMeta) ReadLatestHistory() (*TGChatHistory, error) {
	return readLatestTGChatHistoryByTgcID(tgu.ID)
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
func CreateTGChat(apiID int64) (tgu *TGChatMeta, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGChat, apiID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser := &TGChatMeta{
		ID:        newID,
		APIID:     apiID,
		CreatedAt: createdAt,
	}
	tgu = newUser
	return
}

const sqlReadTGChatByAPIID = `
SELECT id, api_id, created_at
FROM tg_chats
WHERE api_id = $1;`

// ReadTGChatByAPIID returns an instance of a telegram chat by api_id from the database.
func ReadTGChatByAPIID(apiID int64) (tgu *TGChatMeta, err error) {
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

	newChat := &TGChatMeta{
		ID:        newID,
		APIID:     newAPIID,
		CreatedAt: newCreatedAt,
	}
	tgu = newChat
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
