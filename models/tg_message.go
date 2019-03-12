package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/lib/pq"
)

type TGMessage struct {
	ID                     int
	MessageID              int
	FromID                 sql.NullInt64
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
	DeleteChatPhoto        bool
	GroupChatCreated       bool
	SuperGroupChatCreated  bool
	ChannelChatCreated     bool
	MigrateToChatId        sql.NullInt64
	MigrateFromChatId      sql.NullInt64
	PinnedMessage          sql.NullInt64
	CreatedAt              time.Time

	fromUser *TGUserMeta
}

const sqlCreateNewChatMembers = `
INSERT INTO "public"."tg_message_new_chat_members" (tgm_id, tgu_id, created_at)
VALUES ($1, $2, $3)
RETURNING id;`

func (m *TGMessage) CreateNewChatMember(user *TGUserMeta) (err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateNewChatMembers, m.ID, user.ID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreatePhoto error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	return
}

const sqlCreateNewChatPhoto = `
INSERT INTO "public"."tg_message_new_chat_photos" (tgm_id, tgps_id, created_at)
VALUES ($1, $2, $3)
RETURNING id;`

func (m *TGMessage) CreateNewChatPhoto(photo *TGPhotoSize) (err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateNewChatPhoto, m.ID, photo.ID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreatePhoto error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	return
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

// GetDateHuman returns formatted string of Date
func (m *TGMessage) GetDateHuman() string {
	return humanize.Time(m.Date)
}

// GetDateFormatted returns formatted string of Date
func (m *TGMessage) GetDateFormatted() string {
	timeStr := ""

	timeStr = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		m.Date.Year(), m.Date.Month(), m.Date.Day(),
		m.Date.Hour(), m.Date.Minute(), m.Date.Second())

	return timeStr
}

func (m *TGMessage) GetFromName() string {
	from, err := m.GetFromUser()
	if err != nil {
		logger.Errorf("(%d) GetFromName(): error: %s", err)
		return strconv.Itoa(int(m.FromID.Int64))
	}

	return from.GetName()
}

func (m *TGMessage) GetFromUser() (*TGUser, error) {
	from, err := ReadTGUserByAPIID(int(m.FromID.Int64))
	if err != nil {
		logger.Errorf("(%d) GetFromName(): error: %s", err)
		return nil, err
	}

	return from, nil
}

const sqlTGMessageGetPhotos = `
SELECT mp.tgps_id, ps.file_id, ps.width, ps.height, ps.file_size, ps.created_at, ps.last_seen
FROM tg_message_photos as mp
LEFT JOIN tg_photo_sizes as ps ON mp.tgps_id = ps.id
WHERE mp.tgm_id = $1
;`

func (m *TGMessage) GetPhotos() ([]*TGPhotoSize, error) {
	var newPhotoList []*TGPhotoSize

	rows, err := db.Query(sqlTGMessageGetPhotos, m.ID)
	if err != nil {
		logger.Tracef("GetPhotos() (nil, %v)", err)
		return nil, err
	}
	for rows.Next() {

		var newID int
		var newFileID string
		var newWidth int
		var newHeight int
		var newFileSize sql.NullInt64
		var newFileLocation sql.NullString
		var newFileSuffix sql.NullString
		var newFileRetrievedAt pq.NullTime
		var newCreatedAt time.Time
		var newLastSeen time.Time

		err = rows.Scan(&newID, &newFileID, &newWidth, &newHeight, &newFileSize, &newCreatedAt, &newLastSeen)
		if err != nil {
			if err == sql.ErrNoRows {
				err = ErrDoesNotExist
			}
			return nil, err
		}

		newPhotoSize := &TGPhotoSize{
			ID:              newID,
			FileID:          newFileID,
			Width:           newWidth,
			Height:          newHeight,
			FileSize:        newFileSize,
			FileLocation:    newFileLocation,
			FileSuffix:      newFileSuffix,
			FileRetrievedAt: newFileRetrievedAt,
			CreatedAt:       newCreatedAt,
			LastSeen:        newLastSeen,
		}

		newPhotoList = append(newPhotoList, newPhotoSize)
	}

	return newPhotoList, nil
}

func (m *TGMessage) GetPhotoURL(size int) string {
	photos, err := m.GetPhotos()
	if err != nil {
		logger.Errorf("(%d) GetFromName(): couldn't get photos: %s", err)
		return ""
	}
	// Get Smallest image larger than request
	var fileID string
	var foundW, foundH int

	for _, photo := range photos {
		if photo.Width > size && photo.Height > size {
			// Init if zero
			if foundW == 0 || foundH == 0 {
				foundW = photo.Width
				foundH = photo.Height
				fileID = photo.FileID
			}

			if photo.Width < foundW || photo.Height < foundH {
				foundW = photo.Width
				foundH = photo.Height
				fileID = photo.FileID
			}
		}
	}

	// If Empty return the largest image
	if fileID == "" {
		foundW = 0
		foundH = 0
		for _, photo := range photos {
			if foundW == 0 || foundH == 0 {
				foundW = photo.Width
				foundH = photo.Height
				fileID = photo.FileID
			}

			if photo.Width > foundW || photo.Height > foundH {
				foundW = photo.Width
				foundH = photo.Height
				fileID = photo.FileID
			}
		}

	}

	logger.Tracef("/web/chatbot/tg/photos/%s/file", fileID)
	return fmt.Sprintf("/web/chatbot/tg/photos/%s/file", fileID)
}

func (m *TGMessage) GetStickerURL() string {
	sticker, err := ReadTGSticker(int(m.StickerID.Int64))
	if err != nil {
		logger.Errorf("(%d) GetFromName(): error: %s", err)
		return ""
	}

	return fmt.Sprintf("/web/chatbot/tg/stickers/%s/file", sticker.FileID)
}

const sqlTGMessageHasPhotos = `
SELECT exists(SELECT 1 FROM tg_message_photos WHERE tgm_id = $1);`

// GetUsernameExists returns true if username exists in the database
func (m *TGMessage) HasPhotos() (exists bool, err error) {
	var newExists bool

	err = db.QueryRow(sqlTGMessageHasPhotos, m.ID).Scan(&newExists)
	if err != nil {
		logger.Errorf("Error checking if user has photos: %s", err.Error())
		return
	}
	exists = newExists
	return
}

// publics

const sqlCreateTGMessage = `
INSERT INTO "public"."tg_messages" (message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, 
	forwarded_from_message_id, forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, 
    sticker_id, video_id, video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, 
    new_chat_title, delete_chat_photo, group_chat_created, supergroup_chat_created, channel_chat_created, 
    migrate_to_chat_id, migrate_from_chat_id, pinned_message_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, 
    $25, $26, $27, $28, $29, $30, $31, $32)
RETURNING id;`

// CreateTGMessage
func CreateTGMessage(messageID int, from *TGUserMeta, date time.Time, chat *TGChatMeta, forwardedFrom *TGUserMeta,
	forwardedFromChat *TGChatMeta, forwardedFromMessageID sql.NullInt64, forwardDate pq.NullTime,
	replyToMessage *TGMessage, editDate pq.NullTime, text sql.NullString, audio *TGAudio, document *TGDocument,
	animation *TGChatAnimation, sticker *TGSticker, video *TGVideo, videoNote *TGVideoNote, voice *TGVoice,
	caption sql.NullString, contact *TGContact, location *TGLocation, venue *TGVenue, leftChatMember *TGUserMeta,
	newChatTitle sql.NullString, deleteChatPhoto bool, groupChatCreated bool, superGroupChatCreated bool,
	channelChatCreated bool, migrateToChatId sql.NullInt64, migrateFromChatId sql.NullInt64, pinnedMessage *TGMessage) (
	tgMessage *TGMessage, err error) {

	createdAt := time.Now()

	fromID := sql.NullInt64{Valid: false}
	if from != nil {
		fromID = sql.NullInt64{
			Int64: int64(from.ID),
			Valid: true,
		}
	}

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

	pinnedMessageID := sql.NullInt64{Valid: false}
	if pinnedMessage != nil {
		pinnedMessageID = sql.NullInt64{
			Int64: int64(pinnedMessage.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGMessage, messageID, fromID, date, chat.ID, forwardedFromID, forwardedFromChatID,
		forwardedFromMessageID, forwardDate, replyToMessageID, editDate, text, audioID, documentID, animationID,
		stickerID, videoID, videoNoteID, voiceID, caption, contactID, locationID, venueID, leftChatMemberID,
		newChatTitle, deleteChatPhoto, groupChatCreated, superGroupChatCreated, channelChatCreated, migrateToChatId,
		migrateFromChatId, pinnedMessageID, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %s: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	tgMessage = &TGMessage{
		ID:                     newID,
		MessageID:              messageID,
		FromID:                 fromID,
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
		DeleteChatPhoto:        deleteChatPhoto,
		GroupChatCreated:       groupChatCreated,
		SuperGroupChatCreated:  superGroupChatCreated,
		ChannelChatCreated:     channelChatCreated,
		MigrateToChatId:        migrateToChatId,
		MigrateFromChatId:      migrateFromChatId,
		PinnedMessage:          pinnedMessageID,
		CreatedAt:              createdAt,
	}
	return
}

const sqlReadTGMessageByAPIIDChat = `
SELECT id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, sticker_id, video_id, 
    video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, new_chat_title,
    delete_chat_photo, group_chat_created, supergroup_chat_created, channel_chat_created, migrate_to_chat_id,
    migrate_from_chat_id, pinned_message_id, created_at
FROM tg_messages
WHERE message_id = $1 AND chat_id = $2 AND edit_date IS NULL
LIMIT 1;`
const sqlReadTGMessageByAPIIDChatEditDate = `
SELECT id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, sticker_id, video_id, 
    video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, new_chat_title,
    delete_chat_photo, group_chat_created, supergroup_chat_created, channel_chat_created, migrate_to_chat_id, 
    migrate_from_chat_id, pinned_message_id, created_at
FROM tg_messages
WHERE message_id = $1 AND chat_id = $2 AND edit_date = $3
LIMIT 1;`

// ReadTGMessageByAPIIDChat returns an instance of a telegram chat by api_id from the database.
func ReadTGMessageByAPIIDChat(apiID int, chat *TGChatMeta, editDateInt int) (tgMessage *TGMessage, err error) {
	var id int
	var messageID int
	var fromID sql.NullInt64
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
	var deleteChatPhoto bool
	var groupChatCreated bool
	var superGroupChatCreated bool
	var channelChatCreated bool
	var migrateToChatId sql.NullInt64
	var migrateFromChatId sql.NullInt64
	var pinnedMessageID sql.NullInt64
	var createdAt time.Time

	logger.Tracef("ReadTGMessageByAPIIDChat: %d, %d, %d", apiID, chat.ID, editDateInt)

	if editDateInt == 0 {
		err = db.QueryRow(sqlReadTGMessageByAPIIDChat, apiID, chat.ID).
			Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID, &forwardedFromChatID,
				&forwardedFromMessageID, &forwardDate, &replyToMessage, &newEditDate, &text, &audioID, &documentID,
				&animationID, &stickerID, &videoID, &videoNoteID, &voiceID, &caption, &contactID, &locationID, &venueID,
				&leftChatMemberID, &newChatTitle, &deleteChatPhoto, &groupChatCreated, &superGroupChatCreated,
				&channelChatCreated, &migrateToChatId, &migrateFromChatId, &pinnedMessageID, &createdAt)
	} else {
		err = db.QueryRow(sqlReadTGMessageByAPIIDChatEditDate, apiID, chat.ID, time.Unix(int64(editDateInt), 0)).
			Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID, &forwardedFromChatID,
				&forwardedFromMessageID, &forwardDate, &replyToMessage, &newEditDate, &text, &audioID, &documentID,
				&animationID, &stickerID, &videoID, &videoNoteID, &voiceID, &caption, &contactID, &locationID, &venueID,
				&leftChatMemberID, &newChatTitle, &deleteChatPhoto, &groupChatCreated, &superGroupChatCreated,
				&channelChatCreated, &migrateToChatId, &migrateFromChatId, &pinnedMessageID, &createdAt)
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
		DeleteChatPhoto:        deleteChatPhoto,
		GroupChatCreated:       groupChatCreated,
		SuperGroupChatCreated:  superGroupChatCreated,
		ChannelChatCreated:     channelChatCreated,
		MigrateToChatId:        migrateToChatId,
		MigrateFromChatId:      migrateFromChatId,
		PinnedMessage:          pinnedMessageID,
		CreatedAt:              createdAt,
	}
	return
}

const sqlReadTGMessageChatPage = `
SELECT DISTINCT ON (message_id) id, message_id, from_id, date, chat_id, forwarded_from_id, forwarded_from_chat_id, forwarded_from_message_id, 
	forward_date, reply_to_message, edit_date, text, audio_id, document_id, animation_id, sticker_id, video_id, 
    video_note_id, voice_id, caption, contact_id, location_id, venue_id, left_chat_member_id, new_chat_title,
    delete_chat_photo, group_chat_created, supergroup_chat_created, channel_chat_created, migrate_to_chat_id, 
    migrate_from_chat_id, pinned_message_id, created_at
FROM tg_messages
WHERE chat_id = $1
ORDER BY message_id DESC, edit_date DESC NULLS LAST
LIMIT $2 OFFSET $3;`

// ReadTGMessageByAPIIDChat returns an instance of a telegram chat by api_id from the database.
func ReadTGMessageChatPage(chat *TGChat, limit uint, page uint) ([]*TGMessage, error) {
	start := time.Now()

	offset := limit * page
	var newMessageList []*TGMessage

	rows, err := db.Query(sqlReadTGMessageChatPage, chat.ID, limit, offset)
	if err != nil {
		logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
		return nil, err
	}
	for rows.Next() {
		var id int
		var messageID int
		var fromID sql.NullInt64
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
		var deleteChatPhoto bool
		var groupChatCreated bool
		var superGroupChatCreated bool
		var channelChatCreated bool
		var migrateToChatId sql.NullInt64
		var migrateFromChatId sql.NullInt64
		var pinnedMessageID sql.NullInt64
		var createdAt time.Time

		err = rows.Scan(&id, &messageID, &fromID, &date, &chatID, &forwardedFromID, &forwardedFromChatID,
			&forwardedFromMessageID, &forwardDate, &replyToMessage, &newEditDate, &text, &audioID, &documentID,
			&animationID, &stickerID, &videoID, &videoNoteID, &voiceID, &caption, &contactID, &locationID, &venueID,
			&leftChatMemberID, &newChatTitle, &deleteChatPhoto, &groupChatCreated, &superGroupChatCreated,
			&channelChatCreated, &migrateToChatId, &migrateFromChatId, &pinnedMessageID, &createdAt)

		if err != nil {
			if err == sql.ErrNoRows {
				err = ErrDoesNotExist
			}
			return nil, err
		}

		tgMessage := &TGMessage{
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
			DeleteChatPhoto:        deleteChatPhoto,
			GroupChatCreated:       groupChatCreated,
			SuperGroupChatCreated:  superGroupChatCreated,
			ChannelChatCreated:     channelChatCreated,
			MigrateToChatId:        migrateToChatId,
			MigrateFromChatId:      migrateFromChatId,
			PinnedMessage:          pinnedMessageID,
			CreatedAt:              createdAt,
		}

		newMessageList = append(newMessageList, tgMessage)
	}

	elapsed := time.Since(start)
	logger.Tracef("ReadTGMessageChatPage(%d, %d, %d) (%d, %v) [%s]", chat.ID, limit, offset, len(newMessageList),
		nil, elapsed)
	return newMessageList, nil
}
