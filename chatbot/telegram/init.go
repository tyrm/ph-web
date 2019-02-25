package telegram

import (
	"log"

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
			missingReg = append(missingReg, "endpoint")
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
		log.Panic(err)
	}

}

// IsInit returns true if telegram client is initialized
func IsInit() bool {
	return botConnected
}