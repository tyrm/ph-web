package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seePhotoSize(apiPhotoSize *tgbotapi.PhotoSize) (tgPhotoSize *models.TGPhotoSize, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// Get TGMessage entry, return if exists
	tgps, err2 := models.ReadTGPhotoSizeByFileID(apiPhotoSize.FileID)
	if err2 == nil {
		tgPhotoSize = tgps
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeMessage: error reading message from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tgps, err2 = models.CreateTGPhotoSizeFromAPI(apiPhotoSize)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tgps)
	tgPhotoSize = tgps
	return
}