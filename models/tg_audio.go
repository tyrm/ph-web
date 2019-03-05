package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGAudio struct {
	ID              int
	FileID          string
	Duration        int
	Performer       sql.NullString
	Title           sql.NullString
	MimeType        sql.NullString
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGAudio) GetFileID() string {
	return myself.FileID
}

func (myself *TGAudio) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGAudio) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGAudioUpdateFileRetrieved = `
UPDATE tg_audios
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGAudio) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGAudioUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGAudioLastSeen = `
UPDATE tg_audios
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGAudio) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGAudioLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGAudio = `
INSERT INTO "public"."tg_audios" (file_id, duration, performer, title, mime_type, file_size, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;`

// CreateTGPhotoSize creates a new instance of a telegram user in the database.
func CreateTGAudio(fileID string, duration int, performer sql.NullString, title sql.NullString, mimeType sql.NullString,
	fileSize sql.NullInt64) (tga *TGAudio, err error) {

	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGAudio, fileID, duration, performer, title, mimeType, fileSize, createdAt, createdAt).
		Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newAudio := &TGAudio{
		ID:        newID,
		FileID:    fileID,
		Duration:  duration,
		Performer: performer,
		Title:     title,
		MimeType:  mimeType,
		FileSize:  fileSize,
		CreatedAt: createdAt,
		LastSeen:  createdAt,
	}
	tga = newAudio
	return
}

func CreateTGAudioFromAPI(apiAudio *tgbotapi.Audio) (*TGAudio, error) {
	performer := sql.NullString{Valid: false}
	if apiAudio.Performer != "" {
		performer = sql.NullString{
			String: apiAudio.Performer,
			Valid:  true,
		}
	}

	title := sql.NullString{Valid: false}
	if apiAudio.Title != "" {
		title = sql.NullString{
			String: apiAudio.Title,
			Valid:  true,
		}
	}

	mimeType := sql.NullString{Valid: false}
	if apiAudio.MimeType != "" {
		mimeType = sql.NullString{
			String: apiAudio.MimeType,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiAudio.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiAudio.FileSize),
			Valid: true,
		}
	}

	return CreateTGAudio(apiAudio.FileID, apiAudio.Duration, performer, title, mimeType, fileSize)
}

const sqlReadTGAudioByFileID = `
SELECT id, file_id, duration, performer, title, mime_type, file_size, file_location, file_suffix, file_retrieved_at, created_at, last_seen
FROM tg_audios
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGAudioByFileID(fileID string) (tga *TGAudio, err error) {
	var newID int
	var newFileID string
	var newDuration int
	var newPerformer sql.NullString
	var newTitle sql.NullString
	var newMimeType sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGAudioByFileID, fileID).Scan(&newID, &newFileID, &newDuration, &newPerformer,
		&newTitle, &newMimeType, &newFileSize, &newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newCreatedAt,
		&newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newAudio := &TGAudio{
		ID:              newID,
		FileID:          newFileID,
		Duration:        newDuration,
		Performer:       newPerformer,
		Title:           newTitle,
		MimeType:        newMimeType,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tga = newAudio
	return
}
