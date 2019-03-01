package telegram

import (
	"database/sql"
	"time"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
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

func seeMessage(apiMessage *tgbotapi.Message) (tgMessage *models.TGMessage, err error) {
	// Get TGMessage entry, return if exists
	tgm, err2 := models.ReadTGMessageByAPIID(apiMessage.MessageID)
	if err2 == nil {
		tgMessage = tgm
		return
	} else if err2 != nil && err2 != models.ErrDoesNotExist {
		logger.Errorf("seeMessage: error reading message from db: %s", err2)
		err = err2
		return
	}

	// See Relationships
	from, err2 := seeUser(apiMessage.From)
	if err2 != nil {
		logger.Errorf("seeMessage: error seeing user: %s", err2)
		err = err2
		return
	}

	date := time.Unix(int64(apiMessage.Date), 0)

	chat, err2 := seeChat(apiMessage.Chat)
	if err2 != nil {
		logger.Errorf("seeMessage: error seeing user: %s", err2)
		err = err2
		return
	}

	var forwardedFrom *models.TGUser
	if apiMessage.ForwardFrom != nil {
		forwardedFrom, err2 = seeUser(apiMessage.ForwardFrom)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing forwarded from user: %s", err2)
			err = err2
			return
		}
	}

	var forwardedFromChat *models.TGChat
	if apiMessage.ForwardFromChat != nil {
		forwardedFromChat, err2 = seeChat(apiMessage.ForwardFromChat)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing forward from chat: %s", err2)
			err = err2
			return
		}
	}

	forwardedFromMessageID := sql.NullInt64{Valid: false}
	if apiMessage.ForwardFromMessageID != 0 {
		forwardedFromMessageID = sql.NullInt64{
			Int64: int64(forwardedFromChat.ID),
			Valid: true,
		}
	}

	forwardDate := pq.NullTime{Valid: false}
	if apiMessage.ForwardDate != 0 {
		forwardDate = pq.NullTime{
			Time: time.Unix(int64(apiMessage.ForwardDate), 0),
			Valid: true,
		}
	}

	var replyToMessage *models.TGMessage
	if apiMessage.ReplyToMessage != nil {
		replyToMessage, err2 = seeMessage(apiMessage.ReplyToMessage)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing user: %s", err2)
			err = err2
			return
		}
	}

	editDate := pq.NullTime{Valid: false}
	if apiMessage.EditDate != 0 {
		editDate = pq.NullTime{
			Time: time.Unix(int64(apiMessage.EditDate), 0),
			Valid: true,
		}
	}

	var text sql.NullString
	if apiMessage.Text != "" {
		text = sql.NullString{
			String: apiMessage.Text,
			Valid: true,
		}
	}

	var sticker *models.TGSticker

	tgm, err2 = models.CreateTGMessage(apiMessage.MessageID, from, date, chat, forwardedFrom, forwardedFromChat,
		forwardedFromMessageID, forwardDate, replyToMessage, editDate, text, sticker)
	if err2 != nil {
		err = err2
		return
	}

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
