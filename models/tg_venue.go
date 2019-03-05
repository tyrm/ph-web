package models

import (
	"database/sql"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

type TGVenue struct {
	ID             int
	LocationID     int
	Title          string
	Address        string
	FoursquareID   sql.NullString
	FoursquareType sql.NullString
	CreatedAt      time.Time
	LastSeen       time.Time
}

const sqlUpdateTGVenueLastSeen = `
UPDATE tg_venues
SET last_seen = now()
WHERE id = $1
RETURNING last_seen;`

// UpdateLastSeen updates the LastSeen field in the database to now.
func (tg *TGVenue) UpdateLastSeen() error {
	var newLastSeen time.Time

	err := db.QueryRow(sqlUpdateTGVenueLastSeen, tg.ID).Scan(&newLastSeen)
	if err != nil {
		return err
	}

	tg.LastSeen = newLastSeen
	return nil
}

const sqlCreateTGVenue = `
INSERT INTO "public"."tg_venues" (location_id, title, address, foursquare_id, foursquare_type, created_at, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;`

// CreateTGUserMeta creates a new instance of a telegram user in the database.
func CreateTGVenue(location *TGLocation, title string, address string, foursquareID sql.NullString,
	foursquareType sql.NullString) (tgv *TGVenue, err error) {
	createdAt := time.Now()

	var newID int
	err = db.QueryRow(sqlCreateTGVenue, location.ID, title, address, foursquareID, foursquareType, createdAt,createdAt).
		Scan(&newID)
	if sqlErr, ok := err.(*pq.Error); ok {
		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
		logger.Errorf("CreateTGUserMeta error %d: %s", sqlErr.Code, sqlErr.Code.Name())
		return
	}

	newVenue := &TGVenue{
		ID:             newID,
		LocationID:      location.ID,
		Title:          title,
		Address:        address,
		FoursquareID:   foursquareID,
		FoursquareType: foursquareType,
		CreatedAt:      createdAt,
		LastSeen:       createdAt,
	}
	tgv = newVenue
	return
}

func CreateTGVenueFromAPI(apiVenue *tgbotapi.Venue, location *TGLocation) (*TGVenue, error) {
	foursquareID := sql.NullString{Valid: false}
	if apiVenue.FoursquareID != "" {
		foursquareID = sql.NullString{
			String: apiVenue.FoursquareID,
			Valid:  true,
		}
	}

	foursquareType := sql.NullString{Valid: false}

	return CreateTGVenue(location, apiVenue.Title, apiVenue.Address, foursquareID, foursquareType)
}

const sqlReadTGVenueByMetaData = `
SELECT id, location_id, title, address, foursquare_id, foursquare_type, created_at, last_seen
FROM tg_venues
WHERE location_id = $1 AND title = $2 AND address = $3 AND foursquare_id = $4;`

// ReadTGPhotoSizeByFileID returns an instance of a telegram user by api_id from the database.
func ReadTGVenueByMetaData(location *TGLocation, title string, address string, foursquareID sql.NullString) (
	tgv *TGVenue, err error) {

	var newID int
	var newLocationID int
	var newTitle string
	var newAddress string
	var newFoursquareID sql.NullString
	var newFoursquareType sql.NullString
	var newCreatedAt time.Time
	var newLastSeen time.Time

	err = db.QueryRow(sqlReadTGVenueByMetaData, location.ID, title, address, foursquareID).Scan(&newID,	&newLocationID,
		&newTitle, &newAddress, &newFoursquareID, &newFoursquareType, &newCreatedAt, &newLastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrDoesNotExist
		}
		return
	}

	newVenue := &TGVenue{
		ID:             newID,
		LocationID:     newLocationID,
		Title:          newTitle,
		Address:        newAddress,
		FoursquareID:   newFoursquareID,
		FoursquareType: newFoursquareType,
		CreatedAt:      newCreatedAt,
		LastSeen:       newLastSeen,
	}
	tgv = newVenue
	return
}
