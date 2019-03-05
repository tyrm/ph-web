package telegram

import (
	"database/sql"
	"time"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lib/pq"
)

func seeMessage(apiMessage *tgbotapi.Message) (tgMessage *models.TGMessage, err error) {
	start := time.Now()
	chat, err2 := seeChat(apiMessage.Chat)
	if err2 != nil {
		logger.Errorf("seeMessage: error seeing user: %s", err2)
		err = err2
		return
	}

	// Get TGMessage entry, return if exists
	tgm, err2 := models.ReadTGMessageByAPIIDChat(apiMessage.MessageID, chat, apiMessage.EditDate)
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

	var forwardedFrom *models.TGUserMeta
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
			Time:  time.Unix(int64(apiMessage.ForwardDate), 0),
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
			Time:  time.Unix(int64(apiMessage.EditDate), 0),
			Valid: true,
		}
	}

	var text sql.NullString
	if apiMessage.Text != "" {
		text = sql.NullString{
			String: apiMessage.Text,
			Valid:  true,
		}
	}

	var audio *models.TGAudio
	if apiMessage.Audio != nil {
		audio, err2 = seeAudio(apiMessage.Audio)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing audio: %s", err2)
			err = err2
			return
		}
	}

	var document *models.TGDocument
	if apiMessage.Document != nil {
		document, err2 = seeDocument(apiMessage.Document)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing audio: %s", err2)
			err = err2
			return
		}
	}

	var animation *models.TGChatAnimation
	if apiMessage.Animation != nil {
		animation, err2 = seeChatAnimation(apiMessage.Animation)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing animation: %s", err2)
			err = err2
			return
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

	var video *models.TGVideo
	if apiMessage.Video != nil {
		video, err2 = seeVideo(apiMessage.Video)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing sticker: %s", err2)
			err = err2
			return
		}
	}

	var videoNotes *models.TGVideoNote
	if apiMessage.VideoNote != nil {
		videoNotes, err2 = seeVideoNote(apiMessage.VideoNote)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing sticker: %s", err2)
			err = err2
			return
		}
	}

	var voice *models.TGVoice
	if apiMessage.Voice != nil {
		voice, err2 = seeVoice(apiMessage.Voice)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing sticker: %s", err2)
			err = err2
			return
		}
	}

	caption := sql.NullString{Valid: false}
	if apiMessage.Caption != "" {
		caption = sql.NullString{
			String: apiMessage.Caption,
			Valid:  true,
		}
	}

	var contact *models.TGContact
	if apiMessage.Contact != nil {
		contact, err2 = seeContact(apiMessage.Contact)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing animation: %s", err2)
			err = err2
			return
		}
	}

	var location *models.TGLocation
	if apiMessage.Location != nil {
		location, err2 = seeLocation(apiMessage.Location)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing location: %s", err2)
			err = err2
			return
		}
	}

	var venue *models.TGVenue
	if apiMessage.Venue != nil {
		venue, err2 = seeVenue(apiMessage.Venue)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing animation: %s", err2)
			err = err2
			return
		}
	}


	var leftChatMember *models.TGUserMeta
	if apiMessage.LeftChatMember != nil {
		leftChatMember, err2 = seeUser(apiMessage.LeftChatMember)
		if err2 != nil {
			logger.Errorf("seeMessage: error seeing animation: %s", err2)
			err = err2
			return
		}
	}

	tgm, err2 = models.CreateTGMessage(apiMessage.MessageID, from, date, chat, forwardedFrom, forwardedFromChat,
		forwardedFromMessageID, forwardDate, replyToMessage, editDate, text, audio, document, animation, sticker,
		video, videoNotes, voice, caption, contact, location, venue, leftChatMember)
	if err2 != nil {
		err = err2
		return
	}

	if apiMessage.Entities != nil {
		for _, entity := range *apiMessage.Entities {

			var user *models.TGUserMeta
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

	elapsed := time.Since(start)
	logger.Tracef("seeMessage() [%s]", elapsed)
	return
}
