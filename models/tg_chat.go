package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// TGChatHistory represents the varying data of a telegram chat
type TGChat struct {
	ID                  int
	APIID               int64
	Type                string
	Title               sql.NullString
	Username            sql.NullString
	FirstName           sql.NullString
	LastName            sql.NullString
	AllMembersAreAdmins bool
	CreatedAt           time.Time
	LastSeen            time.Time
}

// GetCreatedAtHuman returns humanized string of CreatedAt
func (u *TGChat) GetCreatedAtHuman() string {
	return humanize.Time(u.CreatedAt)
}

// GetCreatedAtFormatted returns formatted string of CreatedAt
func (u *TGChat) GetCreatedAtFormatted() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		u.CreatedAt.Year(), u.CreatedAt.Month(), u.CreatedAt.Day(),
		u.CreatedAt.Hour(), u.CreatedAt.Minute(), u.CreatedAt.Second())

	return timeStr
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChat) GetLastSeenHuman() string {
	return humanize.Time(u.LastSeen)
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChat) GetLastSeenFormatted() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		u.LastSeen.Year(), u.LastSeen.Month(), u.LastSeen.Day(),
		u.LastSeen.Hour(), u.LastSeen.Minute(), u.LastSeen.Second())

	return timeStr
}

// GetLastSeen returns formatted string of LastSeen
func (u *TGChat) GetLongFormattedName() string {
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

// GetName returns long formatted
func (u *TGChat) GetName() string {
	var nameStr []string
	nameStr = append(nameStr, u.FirstName.String)

	if u.LastName.Valid {
		nameStr = append(nameStr, u.LastName.String)
	}

	return strings.Join(nameStr, " ")
}

const sqlReadTGChat = `
SELECT DISTINCT ON (tg_chats.id) tg_chats.id, tg_chats.api_id, tg_chats_history."type", tg_chats_history.title, 
	tg_chats_history.username, tg_chats_history.first_name, tg_chats_history.last_name, 
    tg_chats_history.all_members_are_admins, tg_chats.created_at, tg_chats_history.last_seen
FROM tg_chats LEFT JOIN tg_chats_history
ON tg_chats."id" = tg_chats_history.tgc_id
WHERE tg_chats.api_id = $1
ORDER BY tg_chats.id ASC, tg_chats_history.last_seen DESC;`

func ReadTGChatByAPIID(apiID int64) (chat *TGChat, err error) {
	var newID int
	var newAPIID int64
	var newType string
	var newTitle sql.NullString
	var newUsername sql.NullString
	var newFirstName sql.NullString
	var newLastName sql.NullString
	var newAllMembersAreAdmins bool
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGChat, apiID).Scan(&newID, &newAPIID, &newType, &newTitle, &newUsername, &newFirstName,
		&newLastName, &newAllMembersAreAdmins, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		logger.Tracef("ReadTGChatByAPIID(%d) (%v, %v)", apiID, nil, err)
		return
	}

	chat = &TGChat{
		ID:                  newID,
		APIID:               newAPIID,
		Type:                newType,
		Title:               newTitle,
		Username:            newUsername,
		FirstName:           newFirstName,
		LastName:            newLastName,
		AllMembersAreAdmins: newAllMembersAreAdmins,
		CreatedAt:           newCreatedAt,
		LastSeen:            newLastSeen,
	}
	return
}

const sqlReadTGChatPage = `
SELECT DISTINCT ON (tg_chats.id) tg_chats.id, tg_chats.api_id, tg_chats_history."type", tg_chats_history.title, 
	tg_chats_history.username, tg_chats_history.first_name, tg_chats_history.last_name, tg_chats.created_at, 
    tg_chats_history.last_seen
FROM tg_chats LEFT JOIN tg_chats_history
ON tg_chats."id" = tg_chats_history.tgc_id
ORDER BY tg_chats.id ASC, tg_chats_history.last_seen DESC LIMIT $1 OFFSET $2;`

func ReadTGChatPage(limit uint, page uint) (chatList []*TGChat, err error) {
	offset := limit * page
	var newChatList []*TGChat

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

		tgChatHistory := &TGChat{
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
