package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeDocument(apiDocument *tgbotapi.Document) (tgDocument *models.TGDocument, err error) {
	// Get TGMessage entry, return if exists
	tgps, err2 := models.ReadTGDocumentByFileID(apiDocument.FileID)
	if err2 == nil {
		tgps.UpdateLastSeen()
		tgDocument = tgps
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeDocument: error reading message from db: %s", err2)
		err = err2
		return
	}

	var thumbnail *models.TGPhotoSize
	if apiDocument.Thumbnail != nil {
		thumbnail, err2 = seePhotoSize(apiDocument.Thumbnail)
		if err2 != nil {
			logger.Errorf("seeDocument: error seeing thumbnail: %s", err2)
			err = err2
			return
		}
	}

	err2 = nil
	tgps, err2 = models.CreateTGDocumentFromAPI(apiDocument, thumbnail)
	if err2 != nil {
		err = err2
		return
	}
	go GetFile(tgps)
	tgDocument = tgps
	return
}
