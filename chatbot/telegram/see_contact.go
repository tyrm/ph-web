package telegram

import (
	"database/sql"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeContact(apiContact *tgbotapi.Contact) (tgContact *models.TGContact, err error) {

	lastName := sql.NullString{Valid: false}
	if apiContact.LastName != "" {
		lastName = sql.NullString{
			String: apiContact.LastName,
			Valid:  true,
		}
	}

	userID := sql.NullInt64{Valid: false}
	if apiContact.UserID > 0 {
		userID = sql.NullInt64{
			Int64: int64(apiContact.UserID),
			Valid: true,
		}
	}

	// Get TGContact entry, return if exists
	tgc, err2 := models.ReadTGContactByMeta(apiContact.PhoneNumber, apiContact.FirstName, lastName, userID)
	if err2 == nil {
		tgc.UpdateLastSeen()
		tgContact = tgc
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeMessage: error reading message from db: %s", err2)
		err = err2
		return
	}

	err2 = nil
	tgc, err2 = models.CreateTGContactFromAPI(apiContact)
	if err2 != nil {
		err = err2
		return
	}
	tgContact = tgc
	return
}