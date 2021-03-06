package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeChat(apiChat *tgbotapi.Chat) (tgChat *models.TGChatMeta, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// Get TGUserMeta entry, create if not exists
	tgc, err2 := models.ReadTGChatMetaByAPIID(apiChat.ID)
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tgc, err3 = models.CreateTGChatMeta(apiChat.ID)
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
		err2 = nil
		err2 = tgch.UpdateLastSeen()
		if err2 != nil {
			logger.Errorf("Error updating last seen for $d.", tgc.ID)
		}
	} else {
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

