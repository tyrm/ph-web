package models

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
)

type TGSticker struct {
	ID              int
	FileID          string
	Width           int
	Height          int
	ThumbnailID     sql.NullInt64
	Emoji           sql.NullString
	FileSize        sql.NullInt64
	FileLocation    sql.NullString
	FileSuffix      sql.NullString
	FileRetrievedAt pq.NullTime
	SetName         sql.NullString
	CreatedAt       time.Time
	LastSeen        time.Time
}

func (myself *TGSticker) GetFileID() string {
	return myself.FileID
}

func (myself *TGSticker) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGSticker) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGStickerUpdateFileRetrieved = `
UPDATE tg_stickers
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGSticker) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGStickerUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGStickerLastSeen = `
UPDATE tg_stickers
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGSticker) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGStickerLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGSticker = `
INSERT INTO "public"."tg_stickers" (file_id, width, height, thumbnail_id, emoji, file_size, set_name, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGSticker(fileID string, width int, height int, thumbnail *TGPhotoSize, emoji sql.NullString,
	fileSize sql.NullInt64, setName sql.NullString) (tgs *TGSticker, err error) {
	createdAt := time.Now()

	thumbnailID := sql.NullInt64{Valid: false}
	if thumbnail != nil {
		thumbnailID = sql.NullInt64{
			Int64: int64(thumbnail.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGSticker, fileID, width, height, thumbnailID, emoji, fileSize, setName, createdAt,
		createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newSticker := &TGSticker{
		ID:          newID,
		FileID:      fileID,
		Width:       width,
		Height:      height,
		ThumbnailID: thumbnailID,
		Emoji:       emoji,
		FileSize:    fileSize,
		SetName:     setName,
		CreatedAt:   createdAt,
		LastSeen:    createdAt,
	}
	tgs = newSticker
	return
}

func CreateTGStickerFromAPI(apiSticker *tgbotapi.Sticker, thumbnail *TGPhotoSize) (*TGSticker, error) {
	emoji := sql.NullString{Valid: false}
	if apiSticker.Emoji != "" {
		emoji = sql.NullString{
			String: apiSticker.Emoji,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiSticker.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiSticker.FileSize),
			Valid: true,
		}
	}

	setName := sql.NullString{Valid: false}
	if apiSticker.SetName != "" {
		setName = sql.NullString{
			String: apiSticker.SetName,
			Valid:  true,
		}
	}

	return CreateTGSticker(apiSticker.FileID, apiSticker.Width, apiSticker.Height, thumbnail, emoji, fileSize, setName)
}

const sqlReadTGSticker = `
SELECT id, file_id, width, height, thumbnail_id, emoji, file_size, file_location, file_suffix, file_retrieved_at, set_name, created_at, last_seen
FROM tg_stickers
WHERE id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGSticker(id int) (tgs *TGSticker, err error) {
	idStr := strconv.Itoa(id)
	if u, found := cTGStickerByID.Get(idStr); found {
		tgs = u.(*TGSticker)
		logger.Tracef("ReadTGSticker(%d) (%s) [HIT]", id, tgs.FileID)
		return
	}

	var newID int
	var newFileID string
	var newWidth int
	var newHeight int
	var newThumbnailID sql.NullInt64
	var newEmoji sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newSetName sql.NullString
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGSticker, id).Scan(&newID, &newFileID, &newWidth, &newHeight,
		&newThumbnailID, &newEmoji, &newFileSize, &newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newSetName,
		&newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	tgs = &TGSticker{
		ID:              newID,
		FileID:          newFileID,
		Width:           newWidth,
		Height:          newHeight,
		ThumbnailID:     newThumbnailID,
		Emoji:           newEmoji,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		SetName:         newSetName,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}

	logger.Tracef("ReadTGSticker(%d) (%s) [MISS]", id, tgs.FileID)
	cTGStickerByID.Set(idStr, tgs, cache.DefaultExpiration)
	return
}

const sqlReadTGStickerByFileID = `
SELECT id, file_id, width, height, thumbnail_id, emoji, file_size, file_location, file_suffix, file_retrieved_at, set_name, created_at, last_seen
FROM tg_stickers
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGStickerByFileID(fileID string) (tgs *TGSticker, err error) {
	var newID int
	var newFileID string
	var newWidth int
	var newHeight int
	var newThumbnailID sql.NullInt64
	var newEmoji sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newSetName sql.NullString
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGStickerByFileID, fileID).Scan(&newID, &newFileID, &newWidth, &newHeight,
		&newThumbnailID, &newEmoji, &newFileSize, &newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newSetName,
		&newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newSticker := &TGSticker{
		ID:              newID,
		FileID:          newFileID,
		Width:           newWidth,
		Height:          newHeight,
		ThumbnailID:     newThumbnailID,
		Emoji:           newEmoji,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		SetName:         newSetName,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tgs = newSticker
	return
}
