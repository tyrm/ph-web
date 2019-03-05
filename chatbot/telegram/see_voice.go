package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeVoice(apiVoice *tgbotapi.Voice) (tgVoice *models.TGVoice, err error) {
	// Get TGVoice entry, return if exists
	tga, err2 := models.ReadTGVoiceByFileID(apiVoice.FileID)
	if err2 == nil {
		tgVoice = tga
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeVoice: error reading audio from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tga, err2 = models.CreateTGVoiceFromAPI(apiVoice)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tga)
	tgVoice = tga
	return
}

