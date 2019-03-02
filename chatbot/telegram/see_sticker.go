package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeSticker(apiSticker *tgbotapi.Sticker) (tgSticker *models.TGSticker, err error) {
	// Get TGMessage entry, return if exists
	tgps, err2 := models.ReadTGStickerByFileID(apiSticker.FileID)
	if err2 == nil {
		tgSticker = tgps
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeSticker: error reading message from db: %s", err2)
		err = err2
		return
	}

	var thumbnail *models.TGPhotoSize
	if apiSticker.Thumbnail != nil {
		thumbnail, err2 = seePhotoSize(apiSticker.Thumbnail)
		if err2 != nil {
			logger.Errorf("seeSticker: error seeing thumbnail: %s", err2)
			err = err2
			return
		}
	}

	err2 = nil
	tgps, err2 = models.CreateTGStickerFromAPI(apiSticker, thumbnail)
	if err2 != nil {
		err = err2
		return
	}
	tgSticker = tgps
	return
}
