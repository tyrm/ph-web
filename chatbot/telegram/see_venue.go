package telegram

import (
	"database/sql"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeVenue(apiVenue *tgbotapi.Venue) (tgVenue *models.TGVenue, err error) {
	location, err2 := seeLocation(&apiVenue.Location)
	if err2 != nil {
		logger.Errorf("seeSticker: error seeing thumbnail: %s", err2)
		err = err2
		return
	}

	foursquareID := sql.NullString{Valid: false}
	if apiVenue.FoursquareID != "" {
		foursquareID = sql.NullString{
			String: apiVenue.FoursquareID,
			Valid:  true,
		}
	}

	// Get TGVenue entry, return if exists
	tgv, err2 := models.ReadTGVenueByMetaData(location, apiVenue.Title, apiVenue.Address, foursquareID)
	if err2 == nil {
		tgv.UpdateLastSeen()
		tgVenue = tgv
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeVenue: error reading message from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tgv, err2 = models.CreateTGVenueFromAPI(apiVenue, location)
	if err2 != nil {
		err = err2
		return
	}
	tgVenue = tgv
	return
}
