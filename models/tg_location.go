package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

// TGChatMeta represents a telegram chat
type TGLocation struct {
	ID        int
	Longitude float64
	Latitude  float64
	CreatedAt time.Time
	LastSeen  time.Time
}

const sqlUpdateTGLocationLastSeen = `
UPDATE tg_locations
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tgc *TGLocation) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGLocationLastSeen, tgc.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tgc.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGLocation = `
INSERT INTO "public"."tg_locations" (longitude, latitude, created_at, last_seen)
VALUES ($1, $2, $3, $4)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGLocation(longitude float64, latitude float64) (tgl *TGLocation, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGLocation, longitude, latitude, createdAt,
		createdAt).Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newLocation := &TGLocation{
		ID:        newID,
		Longitude: longitude,
		Latitude:  latitude,
		CreatedAt: createdAt,
		LastSeen:  createdAt,
	}
	tgl = newLocation
	return
}

func CreateTGLocationFromAPI(apiLocation *tgbotapi.Location) (*TGLocation, error) {
	return CreateTGLocation(apiLocation.Longitude, apiLocation.Latitude)
}

const sqlReadTGLocationByLongLat = `
SELECT id, longitude, latitude, created_at, last_seen
FROM tg_locations
WHERE longitude = $1 AND latitude = $2;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGLocationByLongLat(longitude float64, latitude float64) (tgl *TGLocation, err error) {
	var newID int
	var newLongitude float64
	var newLatitude float64
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGLocationByLongLat, longitude, latitude).Scan(&newID, &newLongitude, &newLatitude,
		&newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newLocation := &TGLocation{
		ID:        newID,
		Longitude: newLongitude,
		Latitude:  newLatitude,
		CreatedAt: newCreatedAt,
		LastSeen:  newLastSeen,
	}
	tgl = newLocation
	return
}
