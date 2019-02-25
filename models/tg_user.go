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

const sqlReadTGUserByAPIID = `
SELECT id, api_id, is_bot, created_at
FROM tg_users
WHERE api_id = $1;`

// ReadTGUserByAPIID returns an instance of a telegram user by api_id from the database.
func ReadTGUserByAPIID(apiID int) (tgu *TGUser, err error) {

	var newID int
	var newAPIID int
	var newIsBot bool
	var newCreatedAt time.Time

	err = db.QueryRow(sqlReadTGUserByAPIID, apiID).Scan(&newID, &newAPIID, &newIsBot, &newCreatedAt)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("ReadTGUserByAPIID error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser := &TGUser{
		ID:        newID,
		APIID:     newAPIID,
		IsBot:     newIsBot,
		CreatedAt: newCreatedAt,
	}
	tgu = newUser
	return
}

const sqlReadLatestTGUserHistoryByTguID = `
SELECT id, tgu_id, first_name, last_name, username, language_code, created_at, last_seen
FROM tg_users_history
WHERE tgu_id = $1;`

// ReadTGUserHistory returns an instance of a telegram user by all fields from the database.
func ReadLatestTGUserHistoryByTguID(tguID int) (tguh *TGUserHistory, err error) {

	var newID              int
	var newTGUserID        int
	var newFirstName       string
	var newLastName        sql.NullString
	var newUsername        sql.NullString
	var newLanguageCode    sql.NullString
	var newCreatedAt       time.Time
	var newCreatedLastSeen time.Time

	err = db.QueryRow(sqlReadLatestTGUserHistoryByTguID, tguID).
		Scan(&newID, &newTGUserID, &newFirstName, &newLastName, &newUsername, &newLanguageCode, &newCreatedAt, &newCreatedLastSeen)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
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
		CreatedLastSeen: newCreatedLastSeen,
	}
	tguh = TGUserHistory
	return
}