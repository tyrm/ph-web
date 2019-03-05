package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGVideo struct {
	ID              int
	FileID          string
	Width           int
	Height          int
	Duration        int
	ThumbnailID     sql.NullInt64
	MimeType        sql.NullString
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGVideo) GetFileID() string {
	return myself.FileID
}

func (myself *TGVideo) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGVideo) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGVideoUpdateFileRetrieved = `
UPDATE tg_videos
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVideo) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGVideoUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGVideoLastSeen = `
UPDATE tg_videos
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVideo) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGVideoLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGVideo = `
INSERT INTO "public"."tg_videos" (file_id, width, height, duration, thumbnail_id, mime_type, file_size, created_at, 
	last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGVideo(fileID string, width int, height int, duration int, thumb *TGPhotoSize,
	mimeType sql.NullString, fileSize sql.NullInt64) (tgani *TGVideo, err error) {
	createdAt := time.Now()

	thumbnailID := sql.NullInt64{Valid: false}
	if thumb != nil {
		thumbnailID = sql.NullInt64{
			Int64: int64(thumb.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGVideo, fileID, width, height, duration, thumbnailID, mimeType, fileSize, createdAt,
		createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGAnimation error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newAnimation := &TGVideo{
		ID:          newID,
		FileID:      fileID,
		Width:       width,
		Height:      height,
		Duration:    duration,
		ThumbnailID: thumbnailID,
		MimeType:    mimeType,
		FileSize:    fileSize,
		CreatedAt:   createdAt,
		LastSeen:    createdAt,
	}
	tgani = newAnimation
	return
}

func CreateTGVideoFromAPI(apiVideo *tgbotapi.Video, thumbnail *TGPhotoSize) (*TGVideo, error) {

	mimeType := sql.NullString{Valid: false}
	if apiVideo.MimeType != "" {
		mimeType = sql.NullString{
			String: apiVideo.MimeType,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiVideo.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiVideo.FileSize),
			Valid: true,
		}
	}

	return CreateTGVideo(apiVideo.FileID, apiVideo.Width, apiVideo.Height, apiVideo.Duration, thumbnail, mimeType, fileSize)
}

const sqlReadTGVideoByFileID = `
SELECT id, file_id, width, height, duration, thumbnail_id, mime_type, file_size, file_location, file_suffix, 
       file_retrieved_at, created_at, last_seen
FROM tg_videos
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGVideoByFileID(fileID string) (tgca *TGVideo, err error) {
	var newID int
	var newFileID string
	var newWidth int
	var newHeight int
	var newDuration int
	var newThumbnailID sql.NullInt64
	var newMimeType sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGVideoByFileID, fileID).Scan(&newID, &newFileID, &newWidth, &newHeight,
		&newDuration, &newThumbnailID, &newMimeType, &newFileSize, &newFileLocation, &newFileSuffix,
		&newFileRetrievedAt, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newVideo := &TGVideo{
		ID:              newID,
		FileID:          newFileID,
		Width:           newWidth,
		Height:          newHeight,
		Duration:        newDuration,
		ThumbnailID:     newThumbnailID,
		MimeType:        newMimeType,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tgca = newVideo
	return
}
