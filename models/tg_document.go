package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGDocument struct {
	ID              int
	FileID          string
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

func (myself *TGDocument) GetFileID() string {
	return myself.FileID
}

func (myself *TGDocument) GetFileLocation() string {
	return myself.FileLocation.String
}

func (myself *TGDocument) IsFileLocationValid() bool {
	return myself.FileLocation.Valid
}

const sqlTGDocumentUpdateFileRetrieved = `
UPDATE tg_documents
SET file_location = $2, file_suffix = $3, file_retrieved_at = now()
WHERE id = $1
RETURNING file_retrieved_at;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGDocument) UpdateFileRetrieved(fileLocation string, fileSuffix string) error {
	var newRetrievedAt pq.NullTime

	err := db.QueryRow(sqlTGDocumentUpdateFileRetrieved, tgc.ID, fileLocation, fileSuffix).Scan(&newRetrievedAt)
	if err != nil {
		return err
	}

	tgc.FileRetrievedAt = newRetrievedAt
	return nil
}

const sqlUpdateTGDocumentLastSeen = `
UPDATE tg_documents
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGDocument) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGDocumentLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGDocument = `
INSERT INTO "public"."tg_documents" (file_id, thumbnail_id, file_name, mime_type, file_size, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

// CreateTGPhotoSize creates a new instance of a telegram user in the database.
func CreateTGDocument(fileID string, thumbnail *TGPhotoSize, fileName sql.NullString, mimeType sql.NullString,
	fileSize sql.NullInt64) (tga *TGDocument, err error) {

	createdAt := time.Now()

	thumbnailID := sql.NullInt64{Valid: false}
	if thumbnail != nil {
		thumbnailID = sql.NullInt64{
			Int64: int64(thumbnail.ID),
			Valid: true,
		}
	}

	var newID int
	err = db.QueryRow(sqlCreateTGDocument, fileID, thumbnailID, fileName, mimeType, fileSize, createdAt, createdAt).
		Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newAudio := &TGDocument{
		ID:        newID,
		FileID:    fileID,
		ThumbnailID:        thumbnailID,
		FileName:       fileName,
		MimeType:  mimeType,
		FileSize:  fileSize,
		CreatedAt: createdAt,
		LastSeen:  createdAt,
	}
	tga = newAudio
	return
}

func CreateTGDocumentFromAPI(apiDocument *tgbotapi.Document, thumbnail *TGPhotoSize) (*TGDocument, error) {

	fileName := sql.NullString{Valid: false}
	if apiDocument.FileName != "" {
		fileName = sql.NullString{
			String: apiDocument.FileName,
			Valid:  true,
		}
	}

	mimeType := sql.NullString{Valid: false}
	if apiDocument.MimeType != "" {
		mimeType = sql.NullString{
			String: apiDocument.MimeType,
			Valid:  true,
		}
	}

	fileSize := sql.NullInt64{Valid: false}
	if apiDocument.FileSize > 0 {
		fileSize = sql.NullInt64{
			Int64: int64(apiDocument.FileSize),
			Valid: true,
		}
	}

	return CreateTGDocument(apiDocument.FileID, thumbnail, fileName, mimeType, fileSize)
}

const sqlReadTGDocumentByFileID = `
SELECT id, file_id, thumbnail_id, file_name, mime_type, file_size, file_location, file_suffix, file_retrieved_at, 
       created_at, last_seen
FROM tg_documents
WHERE file_id = $1;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGDocumentByFileID(fileID string) (tgd *TGDocument, err error) {
	var newID int
	var newFileID string
	var newThumbnailID sql.NullInt64
	var newFileName sql.NullString
	var newMimeType sql.NullString
	var newFileSize sql.NullInt64
	var newFileLocation sql.NullString
	var newFileSuffix sql.NullString
	var newFileRetrievedAt pq.NullTime
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGDocumentByFileID, fileID).Scan(&newID, &newFileID, &newThumbnailID, &newFileName,
		&newMimeType, &newFileSize, &newFileLocation, &newFileSuffix, &newFileRetrievedAt, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newDocument := &TGDocument{
		ID:              newID,
		FileID:          newFileID,
		ThumbnailID:        newThumbnailID,
		FileName:       newFileName,
		MimeType:        newMimeType,
		FileSize:        newFileSize,
		FileLocation:    newFileLocation,
		FileSuffix:      newFileSuffix,
		FileRetrievedAt: newFileRetrievedAt,
		CreatedAt:       newCreatedAt,
		LastSeen:        newLastSeen,
	}
	tgd = newDocument
	return
}
