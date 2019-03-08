package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeUser(apiUser *tgbotapi.User) (tgUser *models.TGUserMeta, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// Get TGUserMeta entry, create if not exists
	tgu, err2 := models.ReadTGUserMetaByAPIID(apiUser.ID)
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tgu, err3 = models.CreateTGUserMeta(apiUser.ID, apiUser.IsBot)
		if err3 != nil {
			err = err3
			return
		}
	} else if err2 != nil {
		err = err2
		return
	}

	// Check History is up to date
	err2 = nil
	tguh, err2 := tgu.ReadLatestHistory()
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tguh, err3 = models.CreateTGUserHistoryFromAPI(tgu, apiUser)
		if err3 != nil {
			err = err3
			return
		}

		tgUser = tgu
		return
	} else if err2 != nil {
		err = err2
		return
	}

	if tguh.Matches(apiUser) {
		err2 = nil
		err2 = tguh.UpdateLastSeen()
		if err2 != nil {
			logger.Errorf("Error updating last seen for $d.", tgUser.ID)
		}
	} else {
		err2 = nil
		tguh, err2 = models.CreateTGUserHistoryFromAPI(tgu, apiUser)
		if err2 != nil {
			err = err2
			return
		}
	}

	tgUser = tgu
	return
}
