package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeVideo(apiVideo *tgbotapi.Video) (tgVideo *models.TGVideo, err error) {
	// Get TGMessage entry, return if exists
	tgps, err2 := models.ReadTGVideoByFileID(apiVideo.FileID)
	if err2 == nil {
		tgps.UpdateLastSeen()
		tgVideo = tgps
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeVideo: error reading message from db: %s", err2)
		err = err2
		return
	}

	var thumbnail *models.TGPhotoSize
	if apiVideo.Thumbnail != nil {
		thumbnail, err2 = seePhotoSize(apiVideo.Thumbnail)
		if err2 != nil {
			logger.Errorf("seeVideo: error seeing thumbnail: %s", err2)
			err = err2
			return
		}
	}

	err2 = nil
	tgps, err2 = models.CreateTGVideoFromAPI(apiVideo, thumbnail)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tgps)
	tgVideo = tgps
	return
}
