package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeChat(apiChat *tgbotapi.Chat) (tgChat *models.TGChat, err error) {
	// Get TGUser entry, create if not exists
	tgc, err2 := models.ReadTGChatByAPIID(apiChat.ID)
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tgc, err3 = models.CreateTGChat(apiChat.ID)
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
	tgch, err2 := tgc.ReadLatestHistory()
	if err2 == models.ErrDoesNotExist {
		logger.Tracef("seeChat: chat has no history. creating.")

		var err3 error
		tgch, err3 = models.CreateTGChatHistoryFromAPI(tgc, apiChat)
		if err3 != nil {
			err = err3
			return
		}

		tgChat = tgc
		return
	} else if err2 != nil {
		err = err2
		return
	}

	if tgch.Matches(apiChat) {
		logger.Tracef("seeChat: chat's value match history. updating last seen.")
		err2 = nil
		err2 = tgch.UpdateLastSeen()
		if err2 != nil {
			logger.Errorf("Error updating last seen for $d.", tgc.ID)
		}
	} else {
		logger.Tracef("seeChat: chat has changed value. creating.")
		err2 = nil
		tgch, err2 = models.CreateTGChatHistoryFromAPI(tgc, apiChat)
		if err2 != nil {
			err = err2
			return
		}
	}

	tgChat = tgc
	return
}

func seeUser(apiUser *tgbotapi.User) (tgUser *models.TGUser, err error) {
	// Get TGUser entry, create if not exists
	tgu, err2 := models.ReadTGUserByAPIID(apiUser.ID)
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tgu, err3 = models.CreateTGUser(apiUser.ID, apiUser.IsBot)
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
		logger.Tracef("seeUser: user has no history. creating.")
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
		logger.Tracef("seeUser: user's value match history. updating last seen.")
		err2 = nil
		err2 = tguh.UpdateLastSeen()
		if err2 != nil {
			logger.Errorf("Error updating last seen for $d.", tgUser.ID)
		}
	} else {
		logger.Tracef("seeUser: user has changed value. creating.")
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

func seeMessage(apiUser *tgbotapi.User) (tgUser *models.TGUser, err error) {
	return
}