package telegram

import (
	"database/sql"
	"time"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

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
	logger.Tracef("(%v", err2)

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


	logger.Tracef("For (%v|%v|%v|%v)", apiMessage.ForwardFrom, apiMessage.ForwardFromChat, apiMessage.ForwardFromMessageID, apiMessage.ForwardDate)

	var forwardedFrom *models.TGUser
	if apiMessage.ForwardFrom != nil {
		forwardedFrom, err2 = seeUser(apiMessage.ForwardFrom)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing forwarded from user: %s", err2)
			err = err2
			return
		}
	}

	var forwardedFromChat *models.TGChatMeta
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
			Int64: int64(apiMessage.ForwardFromMessageID),
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
	if apiMessage.Sticker != nil {
		sticker, err2 = seeSticker(apiMessage.Sticker)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing sticker: %s", err2)
			err = err2
			return
		}
	}

	logger.Tracef("For (%v|%v|%v|%v)", forwardedFrom, forwardedFromChat, forwardedFromMessageID, forwardDate)

	tgm, err2 = models.CreateTGMessage(apiMessage.MessageID, from, date, chat, forwardedFrom, forwardedFromChat,
		forwardedFromMessageID, forwardDate, replyToMessage, editDate, text, sticker)
	if err2 != nil {
		err = err2
		return
	}

	if apiMessage.Entities != nil {
		for _, entity := range *apiMessage.Entities {

			var user *models.TGUser
			if entity.User != nil {
				user, err2 = seeUser(entity.User)
				if err2 != nil {
					logger.Errorf("seeMessage: error seeing forwarded from user: %s", err2)
					err = err2
					return
				}
			}

			_, err2 = tgm.CreateMessageEntityFromAPI(&entity, user)
			if err2 != nil {
				err = err2
				return
			}
		}
	}

	if apiMessage.Photo != nil {
		for _, photo := range *apiMessage.Photo {

			ps, err2 := seePhotoSize(&photo)
			if err2 != nil {
				logger.Errorf("seeMessage: error seeing forwarded from user: %s", err2)
				err = err2
				return
			}

			err2 = tgm.CreatePhoto(ps)
			if err2 != nil {
				err = err2
				return
			}
		}
	}

	tgMessage = tgm
	return
}