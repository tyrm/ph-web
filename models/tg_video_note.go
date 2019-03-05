package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGVideoNote struct {
	ID              int
	FileID          string
	Length          int
	Duration        int
	ThumbnailID     sql.NullInt64
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGVideoNote) GetFileID() string {
	return myself.FileID
}

func (myself *TGVideoNote) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGVideoNote) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGVideoNoteUpdateFileRetrieved = `
UPDATE tg_video_notes
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVideoNote) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGVideoNoteUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGVideoNoteLastSeen = `
UPDATE tg_video_notes
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGVideoNote) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGVideoNoteLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGVideoNote = `
INSERT INTO "public"."tg_video_notes" (file_id, length, duration, thumbnail_id, file_size, created_at, 
	last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGVideoNote(fileID string, length int, duration int, thumb *TGPhotoSize, fileSize sql.NullInt64) (
	tgani *TGVideoNote, err error) {

	createdAt := time.Now()

	thumbnailID := sql.NullInt64{Valid: false}
	if thumb != nil {
		thumbnailID = sql.NullInt64{
			Int64: int64(thumb.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGVideoNote, fileID, length, duration, thumbnailID, fileSize, createdAt,
		createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGAnimation error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newAnimation := &TGVideoNote{
		ID:          newID,
		FileID:      fileID,
		Duration:    duration,
		ThumbnailID: thumbnailID,
		FileSize:    fileSize,
		CreatedAt:   createdAt,
		LastSeen:    createdAt,
	}
	tgani = newAnimation
	return
}

func CreateTGVideoNoteFromAPI(apiVideoNote *tgbotapi.VideoNote, thumbnail *TGPhotoSize) (*TGVideoNote, error) {

	fileSize := sql.NullInt64{Valid: false}
	if apiVideoNote.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiVideoNote.FileSize),
			Valid: true,
		}
	}

	return CreateTGVideoNote(apiVideoNote.FileID, apiVideoNote.Length, apiVideoNote.Duration, thumbnail, fileSize)
}

const sqlReadTGVideoNoteByFileID = `
SELECT id, file_id, length, duration, thumbnail_id, file_size, file_location, file_suffix, file_retrieved_at, 
       created_at, last_seen
FROM tg_video_notes
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGVideoNoteByFileID(fileID string) (tgvn *TGVideoNote, err error) {
	var newID int
	var newFileID string
	var newLength int
	var newDuration int
	var newThumbnailID sql.NullInt64
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGVideoNoteByFileID, fileID).Scan(&newID, &newFileID, &newLength, &newDuration,
		&newThumbnailID, &newFileSize, &newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newCreatedAt,
		&newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newVideoNote := &TGVideoNote{
		ID:              newID,
		FileID:          newFileID,
		Length:          newLength,
		Duration:        newDuration,
		ThumbnailID:     newThumbnailID,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tgvn = newVideoNote
	return
}
