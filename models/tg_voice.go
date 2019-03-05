package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGVoice struct {
	ID              int
	FileID          string
	Duration        int
	MimeType        sql.NullString
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGVoice) GetFileID() string {
	return myself.FileID
}

func (myself *TGVoice) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGVoice) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGVoiceUpdateFileRetrieved = `
UPDATE tg_voices
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVoice) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGVoiceUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGVoiceLastSeen = `
UPDATE tg_voices
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVoice) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGVoiceLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGVoice = `
INSERT INTO "public"."tg_voices" (file_id, duration, mime_type, file_size, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;`

// CreateTGPhotoSize creates a new instance of a telegram user in the database.
func CreateTGVoice(fileID string, duration int, mimeType sql.NullString, fileSize sql.NullInt64) (tga *TGVoice,
	err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGVoice, fileID, duration, mimeType, fileSize, createdAt, createdAt).
		Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newVoice := &TGVoice{
		ID:        newID,
		FileID:    fileID,
		Duration:  duration,
		MimeType:  mimeType,
		FileSize:  fileSize,
		CreatedAt: createdAt,
		LastSeen:  createdAt,
	}
	tga = newVoice
	return
}

func CreateTGVoiceFromAPI(apiVoice *tgbotapi.Voice) (*TGVoice, error) {

	mimeType := sql.NullString{Valid: false}
	if apiVoice.MimeType != "" {
		mimeType = sql.NullString{
			String: apiVoice.MimeType,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiVoice.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiVoice.FileSize),
			Valid: true,
		}
	}

	return CreateTGVoice(apiVoice.FileID, apiVoice.Duration, mimeType, fileSize)
}

const sqlReadTGVoiceByFileID = `
SELECT id, file_id, duration, mime_type, file_size, file_location, file_suffix, file_retrieved_at, created_at, last_seen
FROM tg_voices
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGVoiceByFileID(fileID string) (tga *TGVoice, err error) {
	var newID int
	var newFileID string
	var newDuration int
	var newMimeType sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGVoiceByFileID, fileID).Scan(&newID, &newFileID, &newDuration, &newMimeType, &newFileSize,
		&newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newVoice := &TGVoice{
		ID:              newID,
		FileID:          newFileID,
		Duration:        newDuration,
		MimeType:        newMimeType,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tga = newVoice
	return
}
