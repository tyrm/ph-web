package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGMessage struct {
	ID                     int
	MessageID              int
	FromID                 int
	Date                   time.Time
	ChatID                 int
	ForwardedFromID        sql.NullInt64
	ForwardedFromChatID    sql.NullInt64
	ForwardedFromMessageID sql.NullInt64
	ForwardSignature       sql.NullString
	ForwardDate            pq.NullTime
	ReplyToMessage         sql.NullInt64
	EditDate               pq.NullTime
	Text                   sql.NullString
	EntityIDs              []int
	PhotoIDs               []int
	StickerID              sql.NullInt64
	CreatedAt              time.Time
}

const sqlCreateMessageEntity = `
INSERT INTO "public"."tg_message_entities" (tgm_id, "type", "offset", "length", url, "user", created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

func (m *TGMessage) CreateMessageEntity(nType string, offset sql.NullInt64, length sql.NullInt64,
	url sql.NullString, user *TGUser) (tgme *TGMessageEntity, err error) {

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


func (m *TGMessage) CreateMessageEntityFromAPI(apiMessageEntity *tgbotapi.MessageEntity, user *TGUser) (tgme *TGMessageEntity, err error) {

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


const sqlCreatePhoto = `
INSERT INTO "public"."tg_message_photos" (tgm_id, tgps_id, created_at)
VALUES ($1, $2, $3)
RETURNING id;`

func (m *TGMessage) CreatePhoto(photo *TGPhotoSize) (err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreatePhoto, m.ID, photo.ID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreatePhoto error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	return
}

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

const sqlCreateTGMessage = `
INSERT INTO "public"."tg_messages" (message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, 
	forwarded_from_message_id, forward_date, reply_to_message, edit_date, text, sticker_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id;`

// CreateTGMessage
func CreateTGMessage(messageID int, from *TGUser, date time.Time, chat *TGChatMeta, forwardedFrom *TGUser,
	forwardedFromChat *TGChatMeta, forwardedFromMessageID sql.NullInt64, forwardDate pq.NullTime, replyToMessage *TGMessage,
	editDate pq.NullTime, text sql.NullString, sticker *TGSticker) (tgMessage *TGMessage, err error) {

	createdAt := time.Now()

	forwardedFromID := sql.NullInt64{Valid: false}
	if forwardedFrom != nil {
		forwardedFromID = sql.NullInt64{
			Int64: int64(forwardedFrom.ID),
			Valid: true,
		}
	}

	forwardedFromChatID := sql.NullInt64{Valid: false}
	if forwardedFromChat != nil {
		forwardedFromChatID = sql.NullInt64{
			Int64: int64(forwardedFromChat.ID),
			Valid: true,
		}
	}

	replyToMessageID := sql.NullInt64{Valid: false}
	if replyToMessage != nil {
		replyToMessageID = sql.NullInt64{
			Int64: int64(replyToMessage.ID),
			Valid: true,
		}
	}

	stickerID := sql.NullInt64{Valid: false}
	if sticker != nil {
		stickerID = sql.NullInt64{
			Int64: int64(sticker.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGMessage, messageID, from.ID, date, chat.ID, forwardedFromID, forwardedFromChatID,
		forwardedFromMessageID, forwardDate, replyToMessageID, editDate, text, stickerID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	tgMessage = &TGMessage{
		ID:                     newID,
		MessageID:              messageID,
		FromID:                 from.ID,
		Date:                   date,
		ChatID:                 chat.ID,
		ForwardedFromID:        forwardedFromID,
		ForwardedFromChatID:    forwardedFromChatID,
		ForwardedFromMessageID: forwardedFromMessageID,
		ForwardDate:            forwardDate,
		ReplyToMessage:         replyToMessageID,
		EditDate:               editDate,
		Text:                   text,
		StickerID:              stickerID,
		CreatedAt:              createdAt,
	}
	return
}

const sqlReadTGMessageByAPIID = `
SELECT id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, text, sticker_id, created_at
FROM tg_messages
WHERE message_id = $1
LIMIT 1;`

// ReadTGMessageByAPIID returns an instance of a telegram chat by api_id from the database.
func ReadTGMessageByAPIID(apiID int) (tgMessage *TGMessage, err error) {
	var id int
	var messageID int
	var fromID int
	var date time.Time
	var chatID int
	var forwardedFromID sql.NullInt64
	var forwardedFromChatID sql.NullInt64
	var forwardedFromMessageID sql.NullInt64
	var forwardDate pq.NullTime
	var replyToMessage sql.NullInt64
	var editDate pq.NullTime
	var text sql.NullString
	var stickerID sql.NullInt64
	var createdAt time.Time

	err = db.QueryRow(sqlReadTGMessageByAPIID, apiID).Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID,
		&forwardedFromChatID, &forwardedFromMessageID, &forwardDate, &replyToMessage, &editDate, &text, &stickerID,
		&createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	tgMessage = &TGMessage{
		ID:                     id,
		MessageID:              messageID,
		FromID:                 fromID,
		Date:                   date,
		ChatID:                 chatID,
		ForwardedFromID:        forwardedFromID,
		ForwardedFromChatID:    forwardedFromChatID,
		ForwardedFromMessageID: forwardedFromMessageID,
		ForwardDate:            forwardDate,
		ReplyToMessage:         replyToMessage,
		EditDate:               editDate,
		Text:                   text,
		StickerID:              stickerID,
		CreatedAt:              createdAt,
	}
	return
}
