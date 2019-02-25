package telegram

import (
	"../../registry"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/juju/loggo"
)

var bot *tgbotapi.BotAPI
var botConnected = false
var logger *loggo.Logger

func init() {
	newLogger := loggo.GetLogger("telegram")
	logger = &newLogger
}

// InitClient for telegram
func InitClient(force bool) {
	if botConnected && !force {
		return
	}

	logger.Infof("Initializing telegram")
	var missingReg []string
	regToken, err := registry.Get("/system/chatbot/telegram/token")
	if err != nil {
		if err == registry.ErrDoesNotExist {
			missingReg = append(missingReg, "token")
		} else {
			logger.Errorf("Problem getting [token]: %s", err.Error())
			return
		}
	}

	if len(missingReg) > 0 {
		logger.Warningf("Could not init telegram, missing registry items: %v", missingReg)
		return
	}

	token, err := regToken.GetValue()
	if err != nil {
		logger.Errorf("Problem getting [token] value: %s", err.Error())
		return
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Errorf("Problem starting telegram bot: %s", err.Error())
		return
	}

	logger.Infof("Telegram connected as %s", bot.Self.UserName)
	go workerMessageHander()
}

// IsInit returns true if telegram client is initialized
func IsInit() bool {
	return botConnected
}

// privates
func workerMessageHander() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Errorf("Problem starting telegram bot: %s", err.Error())
		return
	}

	botConnected = true

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		logger.Tracef("%v", update)
		logger.Tracef("%v", update.Message)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}

	botConnected = false
}