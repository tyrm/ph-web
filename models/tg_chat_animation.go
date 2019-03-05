package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGChatAnimation struct {
	ID              int
	FileID          string
	Width           int
	Height          int
	Duration        int
	ThumbnailID     sql.NullInt64
	FileName        sql.NullString
	MimeType        sql.NullString
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGChatAnimation) GetFileID() string {
	return myself.FileID
}

func (myself *TGChatAnimation) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGChatAnimation) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGChatAnimationUpdateFileRetrieved = `
UPDATE tg_chat_animations
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGChatAnimation) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGChatAnimationUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGChatAnimationLastSeen = `
UPDATE tg_chat_animations
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGChatAnimation) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGChatAnimationLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGChatAnimation = `
INSERT INTO "public"."tg_chat_animations" (file_id, width, height, duration, thumbnail_id, file_name, mime_type, file_size, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGChatAnimation(fileID string, width int, height int, duration int, thumb *TGPhotoSize,
	fileName sql.NullString, mimeType sql.NullString, fileSize sql.NullInt64) (tgani *TGChatAnimation, err error) {
	createdAt := time.Now()

	thumbnailID := sql.NullInt64{Valid: false}
	if thumb != nil {
		thumbnailID = sql.NullInt64{
			Int64: int64(thumb.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGChatAnimation, fileID, width, height, duration, thumbnailID, fileName, mimeType, fileSize, createdAt,
		createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGAnimation error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newAnimation := &TGChatAnimation{
		ID:          newID,
		FileID:      fileID,
		Width:       width,
		Height:      height,
		Duration:    duration,
		ThumbnailID: thumbnailID,
		FileName:    fileName,
		MimeType:    mimeType,
		FileSize:    fileSize,
		CreatedAt:   createdAt,
		LastSeen:    createdAt,
	}
	tgani = newAnimation
	return
}

func CreateTGChatAnimationFromAPI(apiChatAnimation *tgbotapi.ChatAnimation, thumbnail *TGPhotoSize) (*TGChatAnimation, error) {
	fileName := sql.NullString{Valid: false}
	if apiChatAnimation.FileName != "" {
		fileName = sql.NullString{
			String: apiChatAnimation.FileName,
			Valid:  true,
		}
	}

	mimeType := sql.NullString{Valid: false}
	if apiChatAnimation.MimeType != "" {
		mimeType = sql.NullString{
			String: apiChatAnimation.MimeType,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiChatAnimation.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiChatAnimation.FileSize),
			Valid: true,
		}
	}

	return CreateTGChatAnimation(apiChatAnimation.FileID, apiChatAnimation.Width, apiChatAnimation.Height, apiChatAnimation.Duration, thumbnail, fileName, mimeType, fileSize)
}

const sqlReadTGChatAnimationByFileID = `
SELECT id, file_id, width, height, duration, thumbnail_id, file_name, mime_type, file_size, file_location, file_suffix, file_retrieved_at, created_at, last_seen
FROM tg_chat_animations
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGChatAnimationByFileID(fileID string) (tgca *TGChatAnimation, err error) {
	var newID int
	var newFileID string
	var newWidth int
	var newHeight int
	var newDuration int
	var newThumbnailID sql.NullInt64
	var newFileName sql.NullString
	var newMimeType sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGChatAnimationByFileID, fileID).Scan(&newID, &newFileID, &newWidth, &newHeight,
		&newDuration, &newThumbnailID, &newFileName, &newMimeType, &newFileSize, &newFileLocation, &newFileSuffix,
		&newFileRetrievedAt, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newChatAnimation := &TGChatAnimation{
		ID:              newID,
		FileID:          newFileID,
		Width:           newWidth,
		Height:          newHeight,
		Duration:        newDuration,
		ThumbnailID:     newThumbnailID,
		FileName:        newFileName,
		MimeType:        newMimeType,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tgca = newChatAnimation
	return
}
