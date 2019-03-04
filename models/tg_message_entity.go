package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TGMessageEntity struct {
	ID          int
	TGMessageID int
	Type        string
	Offset      sql.NullInt64
	Length      sql.NullInt64
	URL         sql.NullString
	UserID      sql.NullInt64
	CreatedAt   time.Time
}

const sqlCreateMessageEntity = `
INSERT INTO "public"."tg_message_entities" (tgm_id, "type", "offset", "length", url, "user", created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

func (m *TGMessage) CreateMessageEntity(nType string, offset sql.NullInt64, length sql.NullInt64,
	url sql.NullString, user *TGUserMeta) (tgme *TGMessageEntity, err error) {

	createdAt := time.Now()

	userID := sql.NullInt64{Valid: false}
	if user != nil {
		userID = sql.NullInt64{
			Int64: int64(user.ID),
			Valid: true,
		}
	}

	var newID int
	err2 := db.QueryRow(sqlCreateMessageEntity, m.ID, nType, offset, length, url, userID, createdAt).Scan(&newID)
	if err2 != nil {
		err = err2
		return
	}

	tgMessageEntity := &TGMessageEntity{
		ID:          newID,
		TGMessageID: m.ID,
		Type:        nType,
		Offset:      offset,
		Length:      length,
		URL:         url,
		UserID:      userID,
		CreatedAt:   createdAt,
	}
	tgme = tgMessageEntity
	return

}

func (m *TGMessage) CreateMessageEntityFromAPI(apiMessageEntity *tgbotapi.MessageEntity, user *TGUserMeta) (tgme *TGMessageEntity, err error) {

	offset := sql.NullInt64{
		Int64: int64(apiMessageEntity.Offset),
		Valid: true,
	}

	length := sql.NullInt64{
		Int64: int64(apiMessageEntity.Length),
		Valid: true,
	}

	url := sql.NullString{Valid: false}
	if apiMessageEntity.URL != "" {
		url = sql.NullString{
			String: apiMessageEntity.URL,
			Valid: true,
		}
	}

	return m.CreateMessageEntity(apiMessageEntity.Type, offset, length, url, user)
}
