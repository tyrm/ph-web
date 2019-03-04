package models

import (
	"database/sql"
	"time"

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

const sqlCreateTGChatMeta = `
INSERT INTO "public"."tg_chats" (api_id, created_at)
VALUES ($1, $2)
RETURNING id;`

// CreateTGChatMeta creates a new instance of a telegram chat in the database.
func CreateTGChatMeta(apiID int64) (tgu *TGChatMeta, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGChatMeta, apiID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
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

const sqlReadTGChatMetaByAPIID = `
SELECT id, api_id, created_at
FROM tg_chats
WHERE api_id = $1;`

// ReadTGChatMetaByAPIID returns an instance of a telegram chat by api_id from the database.
func ReadTGChatMetaByAPIID(apiID int64) (tgu *TGChatMeta, err error) {
	var newID int
	var newAPIID int64
	var newCreatedAt time.Time

	err = db.QueryRow(sqlReadTGChatMetaByAPIID, apiID).Scan(&newID, &newAPIID, &newCreatedAt)
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

