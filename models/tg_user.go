package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// TGUser represents a telegram user
type TGUser struct {
	ID        int
	APIID     int
	IsBot     bool
	CreatedAt time.Time
}

// TGUserHistory represents the varying data of a telegram user
type TGUserHistory struct {
	ID              int
	TGUserID        int
	FirstName       string
	LastName        sql.NullString
	Username        sql.NullString
	LanguageCode    sql.NullString
	CreatedAt       time.Time
	CreatedLastSeen time.Time
}

const sqlCreateTGUser = `
INSERT INTO "public"."tg_users" (api_id, is_bot, created_at)
VALUES ($1, $2, $3)
RETURNING id;`

// CreateTGUser creates a new instance of a telegram user in the database.
func CreateTGUser(apiID int, isBot bool) (tgu *TGUser, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGUser, apiID, isBot, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser := &TGUser{
		ID:        newID,
		APIID:     apiID,
		IsBot:     isBot,
		CreatedAt: createdAt,
	}
	tgu = newUser
	return
}

const sqlCreateTGUserHistory = `
INSERT INTO "public"."tg_users_history" (tgu_id, first_name, last_name, username, language_code, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

// CreateTGUserHistory creates a new instance of telegram user history in the database.
func CreateTGUserHistory(tguID int, firstName string, lastName sql.NullString, username sql.NullString, languageCode sql.NullString) (tguh *TGUserHistory, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGUser, tguID, firstName, lastName, username, languageCode, createdAt, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	TGUserHistory := &TGUserHistory{
		ID:              newID,
		TGUserID:        tguID,
		FirstName:       firstName,
		LastName:        lastName,
		Username:        username,
		LanguageCode:    languageCode,
		CreatedAt:       createdAt,
		CreatedLastSeen: createdAt,
	}
	tguh = TGUserHistory
	return
}
