package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeAudio(apiAudio *tgbotapi.Audio) (tgAudio *models.TGAudio, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// Get TGAudio entry, return if exists
	tga, err2 := models.ReadTGAudioByFileID(apiAudio.FileID)
	if err2 == nil {
		tgAudio = tga
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeAudio: error reading audio from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tga, err2 = models.CreateTGAudioFromAPI(apiAudio)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tga)
	tgAudio = tga
	return
}

