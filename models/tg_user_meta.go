package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// TGUserMeta represents a telegram user
type TGUserMeta struct {
	ID        int
	APIID     int
	IsBot     bool
	CreatedAt time.Time
}

func (tgu *TGUserMeta) ReadLatestHistory() (*TGUserHistory, error) {
	return readLatestTGUserHistoryByTguID(tgu.ID)
}


const sqlCreateTGUserMeta = `
INSERT INTO "public"."tg_users" (api_id, is_bot, created_at)
VALUES ($1, $2, $3)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGUserMeta(apiID int, isBot bool) (tgu *TGUserMeta, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGUserMeta, apiID, isBot, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newUser := &TGUserMeta{
		ID:        newID,
		APIID:     apiID,
		IsBot:     isBot,
		CreatedAt: createdAt,
	}
	tgu = newUser
	return
}

const sqlReadTGUserMetaByAPIID = `
SELECT id, api_id, is_bot, created_at
FROM tg_users
WHERE api_id = $1;`

// ReadTGUserMetaByAPIID returns an instance of a telegram user by api_id from the database.
func ReadTGUserMetaByAPIID(apiID int) (tgu *TGUserMeta, err error) {
	var newID int
	var newAPIID int
	var newIsBot bool
	var newCreatedAt time.Time

	err = db.QueryRow(sqlReadTGUserMetaByAPIID, apiID).Scan(&newID, &newAPIID, &newIsBot, &newCreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newUser := &TGUserMeta{
		ID:        newID,
		APIID:     newAPIID,
		IsBot:     newIsBot,
		CreatedAt: newCreatedAt,
	}
	tgu = newUser
	return
}
