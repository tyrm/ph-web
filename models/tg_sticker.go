package models

import (
	"database/sql"
	"time"
)

type TGSticker struct {
	ID          int
	FileID      string
	Width       int
	Height      int
	ThumbnailID sql.NullInt64
	Emoji       sql.NullString
	FileSize    sql.NullInt64
	SetName     sql.NullString
	CreatedAt   time.Time
}
