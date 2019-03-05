package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	)

func seeChatAnimation(apiChatAnimation *tgbotapi.ChatAnimation) (tgAnimation *models.TGChatAnimation, err error) {
	// Get TGMessage entry, return if exists
	tgani, err2 := models.ReadTGChatAnimationByFileID(apiChatAnimation.FileID)
	if err2 == nil {
		tgani.UpdateLastSeen()
		tgAnimation = tgani
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeSticker: error reading message from db: %s", err2)
		err = err2
		return
	}

	var thumbnail *models.TGPhotoSize
	if &apiChatAnimation.Thumbnail != nil {
		thumbnail, err2 = seePhotoSize(apiChatAnimation.Thumbnail)
		if err2 != nil {
			logger.Errorf("seeSticker: error seeing thumbnail: %s", err2)
			err = err2
			return
		}
	}

	err2 = nil
	tgani, err2 = models.CreateTGChatAnimationFromAPI(apiChatAnimation, thumbnail)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tgani)
	tgAnimation = tgani
	return
}
