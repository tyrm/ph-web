package models

import (
	"database/sql"
	"time"

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
	AudioID                sql.NullInt64
	DocumentID             sql.NullInt64
	AnimationID            sql.NullInt64
	StickerID              sql.NullInt64
	VideoID                sql.NullInt64
	VideoNoteID            sql.NullInt64
	VoiceID                sql.NullInt64
	Caption                sql.NullString
	ContactID              sql.NullInt64
	LocationID             sql.NullInt64
	VenueID                sql.NullInt64
	LeftChatMember         sql.NullInt64
	NewChatTitle           sql.NullString
	CreatedAt              time.Time
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

const sqlCreateTGMessage = `
INSERT INTO "public"."tg_messages" (message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, 
	forwarded_from_message_id, forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, 
    sticker_id, video_id, video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, 
    new_chat_title, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, 
    $25)
RETURNING id;`

// CreateTGMessage
func CreateTGMessage(messageID int, from *TGUserMeta, date time.Time, chat *TGChatMeta, forwardedFrom *TGUserMeta,
	forwardedFromChat *TGChatMeta, forwardedFromMessageID sql.NullInt64, forwardDate pq.NullTime,
	replyToMessage *TGMessage, editDate pq.NullTime, text sql.NullString, audio *TGAudio, document *TGDocument,
	animation *TGChatAnimation, sticker *TGSticker, video *TGVideo, videoNote *TGVideoNote, voice *TGVoice,
	caption sql.NullString, contact *TGContact, location *TGLocation, venue *TGVenue, leftChatMember *TGUserMeta,
	newChatTitle sql.NullString) (tgMessage *TGMessage, err error) {

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

	audioID := sql.NullInt64{Valid: false}
	if audio != nil {
		audioID = sql.NullInt64{
			Int64: int64(audio.ID),
			Valid: true,
		}
	}

	documentID := sql.NullInt64{Valid: false}
	if document != nil {
		documentID = sql.NullInt64{
			Int64: int64(document.ID),
			Valid: true,
		}
	}

	animationID := sql.NullInt64{Valid: false}
	if animation != nil {
		animationID = sql.NullInt64{
			Int64: int64(animation.ID),
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

	videoID := sql.NullInt64{Valid: false}
	if video != nil {
		videoID = sql.NullInt64{
			Int64: int64(video.ID),
			Valid: true,
		}
	}

	videoNoteID := sql.NullInt64{Valid: false}
	if video != nil {
		videoNoteID = sql.NullInt64{
			Int64: int64(videoNote.ID),
			Valid: true,
		}
	}

	voiceID := sql.NullInt64{Valid: false}
	if voice != nil {
		voiceID = sql.NullInt64{
			Int64: int64(voice.ID),
			Valid: true,
		}
	}

	contactID := sql.NullInt64{Valid: false}
	if contact != nil {
		contactID = sql.NullInt64{
			Int64: int64(contact.ID),
			Valid: true,
		}
	}

	locationID := sql.NullInt64{Valid: false}
	if location != nil {
		locationID = sql.NullInt64{
			Int64: int64(location.ID),
			Valid: true,
		}
	}

	venueID := sql.NullInt64{Valid: false}
	if venue != nil {
		venueID = sql.NullInt64{
			Int64: int64(venue.ID),
			Valid: true,
		}
	}

	leftChatMemberID := sql.NullInt64{Valid: false}
	if leftChatMember != nil {
		leftChatMemberID = sql.NullInt64{
			Int64: int64(leftChatMember.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGMessage, messageID, from.ID, date, chat.ID, forwardedFromID, forwardedFromChatID,
		forwardedFromMessageID, forwardDate, replyToMessageID, editDate, text, audioID, documentID, animationID,
		stickerID, videoID, videoNoteID, voiceID, caption, contactID, locationID, venueID, leftChatMemberID,
		newChatTitle, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %s: %s", sqlErr.Code, sqlErr.Code.Name())
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
		AudioID:                audioID,
		AnimationID:            animationID,
		StickerID:              stickerID,
		VideoID:                videoID,
		VideoNoteID:            videoNoteID,
		Caption:                caption,
		ContactID:              contactID,
		LocationID:             locationID,
		VenueID:                venueID,
		LeftChatMember:         leftChatMemberID,
		NewChatTitle:           newChatTitle,
		CreatedAt:              createdAt,
	}
	return
}

const sqlReadTGMessageByAPIIDChat = `
SELECT id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, sticker_id, video_id, 
    video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, new_chat_title, created_at
FROM tg_messages
WHERE message_id = $1 AND chat_id = $2 AND edit_date IS NULL
LIMIT 1;`
const sqlReadTGMessageByAPIIDChatEditDate = `
SELECT id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, sticker_id, video_id, 
    video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, new_chat_title, created_at
FROM tg_messages
WHERE message_id = $1 AND chat_id = $2 AND edit_date = $3
LIMIT 1;`

// ReadTGMessageByAPIIDChat returns an instance of a telegram chat by api_id from the database.
func ReadTGMessageByAPIIDChat(apiID int, chat *TGChatMeta, editDateInt int) (tgMessage *TGMessage, err error) {
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
	var newEditDate pq.NullTime
	var text sql.NullString
	var audioID sql.NullInt64
	var documentID sql.NullInt64
	var animationID sql.NullInt64
	var stickerID sql.NullInt64
	var videoID sql.NullInt64
	var videoNoteID sql.NullInt64
	var voiceID sql.NullInt64
	var caption sql.NullString
	var contactID sql.NullInt64
	var locationID sql.NullInt64
	var venueID sql.NullInt64
	var leftChatMemberID sql.NullInt64
	var newChatTitle sql.NullString
	var createdAt time.Time

	logger.Tracef("ReadTGMessageByAPIIDChat: %d, %d, %d", apiID, chat.ID, editDateInt)

	if editDateInt == 0 {
		err = db.QueryRow(sqlReadTGMessageByAPIIDChat, apiID, chat.ID).
			Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID, &forwardedFromChatID,
				&forwardedFromMessageID, &forwardDate, &replyToMessage, &newEditDate, &text, &audioID, &documentID,
				&animationID, &stickerID, &videoID, &videoNoteID, &voiceID, &caption, &contactID, &locationID, &venueID,
				&leftChatMemberID, &newChatTitle, &createdAt)
	} else {
		err = db.QueryRow(sqlReadTGMessageByAPIIDChatEditDate, apiID, chat.ID, time.Unix(int64(editDateInt), 0)).
			Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID, &forwardedFromChatID,
				&forwardedFromMessageID, &forwardDate, &replyToMessage, &newEditDate, &text, &audioID, &documentID,
				&animationID, &stickerID, &videoID, &videoNoteID, &voiceID, &caption, &contactID, &locationID, &venueID,
			&leftChatMemberID, &newChatTitle, &createdAt)
	}

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
		EditDate:               newEditDate,
		Text:                   text,
		AudioID:                audioID,
		DocumentID:             documentID,
		AnimationID:            animationID,
		StickerID:              stickerID,
		VideoID:                videoID,
		VideoNoteID:            videoNoteID,
		VoiceID:                voiceID,
		Caption:                caption,
		ContactID:              contactID,
		LocationID:             locationID,
		VenueID:                venueID,
		LeftChatMember:         leftChatMemberID,
		NewChatTitle:           newChatTitle,
		CreatedAt:              createdAt,
	}
	return
}
