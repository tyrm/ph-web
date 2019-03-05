package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeVideoNote(apiVideoNote *tgbotapi.VideoNote) (tgVideoNote *models.TGVideoNote, err error) {
	// Get TGMessage entry, return if exists
	tgps, err2 := models.ReadTGVideoNoteByFileID(apiVideoNote.FileID)
	if err2 == nil {
		tgps.UpdateLastSeen()
		tgVideoNote = tgps
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeVideoNote: error reading message from db: %s", err2)
		err = err2
		return
	}

	var thumbnail *models.TGPhotoSize
	if apiVideoNote.Thumbnail != nil {
		thumbnail, err2 = seePhotoSize(apiVideoNote.Thumbnail)
		if err2 != nil {
			logger.Errorf("seeVideoNote: error seeing thumbnail: %s", err2)
			err = err2
			return
		}
	}

	err2 = nil
	tgps, err2 = models.CreateTGVideoNoteFromAPI(apiVideoNote, thumbnail)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tgps)
	tgVideoNote = tgps
	return
}
