package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

// TGUserHistory represents the varying data of a telegram user
type TGUserHistory struct {
	ID           int
	TGUserID     int
	FirstName    string
	LastName     sql.NullString
	Username     sql.NullString
	LanguageCode sql.NullString
	CreatedAt    time.Time
	LastSeen     time.Time
}

func (tgu *TGUserHistory) Matches(apiUser *tgbotapi.User) bool {
	if tgu.FirstName != apiUser.FirstName {
		logger.Tracef("Matches() false [FirstName]")
		return false
	}

	if apiUser.LastName != "" || tgu.LastName.Valid != false {
		if apiUser.LastName != tgu.LastName.String {
			logger.Tracef("Matches() false [LastName]")
			return false
		}
	}

	if apiUser.UserName != "" || tgu.Username.Valid != false {
		if apiUser.UserName != tgu.Username.String {
			logger.Tracef("Matches() false [Username]")
			return false
		}
	}

	if apiUser.LanguageCode != "" || tgu.LanguageCode.Valid != false {
		if apiUser.LanguageCode != tgu.LanguageCode.String {
			logger.Tracef("Matches() false [LanguageCode]")
			return false
		}
	}

	return true
}

const sqlUpdateTGUserHistoryLastSeen = `
UPDATE tg_users_history
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastLogin updates the LastSeen field in the database to now.
func (tgu *TGUserHistory) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGUserHistoryLastSeen, tgu.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgu.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGUserHistory = `
INSERT INTO "public"."tg_users_history" (tgu_id, first_name, last_name, username, language_code, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

// CreateTGUserHistory creates a new instance of telegram user history in the database.
func CreateTGUserHistory(tgu *TGUserMeta, firstName string, lastName sql.NullString, username sql.NullString,
	languageCode sql.NullString) (tguh *TGUserHistory, err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGUserHistory, tgu.ID, firstName, lastName, username, languageCode, createdAt, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	TGUserHistory := &TGUserHistory{
		ID:              newID,
		TGUserID:        tgu.ID,
		FirstName:       firstName,
		LastName:        lastName,
		Username:        username,
		LanguageCode:    languageCode,
		CreatedAt:       createdAt,
		LastSeen: createdAt,
	}
	tguh = TGUserHistory
	return
}

func CreateTGUserHistoryFromAPI (tgUser *TGUserMeta, apiUser *tgbotapi.User) (*TGUserHistory, error) {
	firstName := apiUser.FirstName

	lastName := &sql.NullString{Valid: false}
	if apiUser.LastName != "" {
		lastName = &sql.NullString{
			String: apiUser.LastName,
			Valid: true,
		}
	}

	username := &sql.NullString{Valid: false}
	if apiUser.UserName != "" {
		username = &sql.NullString{
			String: apiUser.UserName,
			Valid: true,
		}
	}

	languageCode := &sql.NullString{Valid: false}
	if apiUser.LanguageCode != "" {
		languageCode = &sql.NullString{
			String: apiUser.LanguageCode,
			Valid: true,
		}
	}

	return CreateTGUserHistory(tgUser, firstName, *lastName, *username, *languageCode)
}

// privates
const sqlReadLatestTGUserHistoryByTguID = `
SELECT id, tgu_id, first_name, last_name, username, language_code, created_at, last_seen
FROM tg_users_history
WHERE tgu_id = $1
ORDER BY created_at DESC
LIMIT 1;`

// ReadTGUserHistory returns an instance of a telegram user by all fields from the database.
func readLatestTGUserHistoryByTguID(tguID int) (tguh *TGUserHistory, err error) {
	var newID int
	var newTGUserID int
	var newFirstName string
	var newLastName sql.NullString
	var newUsername sql.NullString
	var newLanguageCode sql.NullString
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadLatestTGUserHistoryByTguID, tguID).
		Scan(&newID, &newTGUserID, &newFirstName, &newLastName, &newUsername, &newLanguageCode, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	TGUserHistory := &TGUserHistory{
		ID:              newID,
		TGUserID:        newTGUserID,
		FirstName:       newFirstName,
		LastName:        newLastName,
		Username:        newUsername,
		LanguageCode:    newLanguageCode,
		CreatedAt:       newCreatedAt,
		LastSeen: newLastSeen,
	}
	tguh = TGUserHistory
	return
}
