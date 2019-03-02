package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGPhotoSize struct {
	ID        int
	FileID    string
	Width     int
	Height    int
	FileSize  sql.NullInt64
	CreatedAt time.Time
	LastSeen  time.Time
}

const sqlUpdateTGPhotoSizeLastSeen = `
UPDATE tg_photo_sizes
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGPhotoSize) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGPhotoSizeLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGPhotoSize = `
INSERT INTO "public"."tg_photo_sizes" (file_id, width, height, file_size, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;`

// CreateTGPhotoSize creates a new instance of a telegram user in the database.
func CreateTGPhotoSize(fileID string, width int, height int, fileSize sql.NullInt64) (tgps *TGPhotoSize, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGPhotoSize, fileID, width, height, fileSize, createdAt, createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUser error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newPhotoSize := &TGPhotoSize{
		ID:        newID,
		FileID:    fileID,
		Width:     width,
		Height:    height,
		FileSize:  fileSize,
		CreatedAt: createdAt,
		LastSeen:  createdAt,
	}
	tgps = newPhotoSize
	return
}

func CreateTGPhotoSizeFromAPI(apiPhotoSize *tgbotapi.PhotoSize) (*TGPhotoSize, error) {
	fileSize := sql.NullInt64{Valid: false}
	if apiPhotoSize.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiPhotoSize.FileSize),
			Valid: true,
		}
	}

	return CreateTGPhotoSize(apiPhotoSize.FileID, apiPhotoSize.Width, apiPhotoSize.Height, fileSize)
}

const sqlReadTGPhotoSizeByFileID = `
SELECT id, file_id, width, height, file_size, created_at, last_seen
FROM tg_photo_sizes
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGPhotoSizeByFileID(fileID string) (tgps *TGPhotoSize, err error) {
	var newID int
	var newFileID string
	var newWidth int
	var newHeight int
	var newFileSize sql.NullInt64
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGPhotoSizeByFileID, fileID).Scan(&newID, &newFileID, &newWidth, &newHeight, &newFileSize,
		&newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newPhotoSize := &TGPhotoSize{
		ID:        newID,
		FileID:    newFileID,
		Width:     newWidth,
		Height:    newHeight,
		FileSize:  newFileSize,
		CreatedAt: newCreatedAt,
		LastSeen:  newLastSeen,
	}
	tgps = newPhotoSize
	return
}
