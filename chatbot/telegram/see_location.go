package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeLocation(apiLocation *tgbotapi.Location) (tgLocation *models.TGLocation, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// Get TGMessage entry, return if exists
	tgl, err2 := models.ReadTGLocationByLongLat(apiLocation.Longitude, apiLocation.Latitude)
	if err2 == nil {
		tgl.UpdateLastSeen()
		tgLocation = tgl
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeMessage: error reading message from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tgl, err2 = models.CreateTGLocationFromAPI(apiLocation)
	if err2 != nil {
		err = err2
		return
	}
	tgLocation = tgl
	return
}