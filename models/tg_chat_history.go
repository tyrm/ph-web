package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

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

const sqlCreateTGChatHistory = `
INSERT INTO "public"."tg_chats_history" (tgc_id, "type", title, username, first_name, last_name, 
                                         all_members_are_admins, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;`

// CreateTGChatHistory creates a new instance of telegram user history in the database.
func CreateTGChatHistory(tgc *TGChatMeta, newType string, title sql.NullString, username sql.NullString,
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

func CreateTGChatHistoryFromAPI(tgChat *TGChatMeta, apiChat *tgbotapi.Chat) (*TGChatHistory, error) {
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