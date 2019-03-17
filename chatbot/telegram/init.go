package telegram

import (
	"errors"
	"time"

	"../../config"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/juju/loggo"
	"github.com/patrickmn/go-cache"
)

var bot *tgbotapi.BotAPI
var botConnected = false
var logger *loggo.Logger
var messageLoggingChan chan *tgbotapi.Message

// Caches
var cUserProfilePhotos *cache.Cache

var (
	ErrNotInit = errors.New("chatbot not initialized")
)

func init() {
	newLogger := loggo.GetLogger("telegram")
	logger = &newLogger

	messageLoggingChan = make(chan *tgbotapi.Message, 100)
	go workerMessageHandler(1)

	// init cache
	cUserProfilePhotos = cache.New(5*time.Minute, 10*time.Minute)
}

// InitClient for telegram
func InitClient(config config.Config, force bool) {
	if botConnected && !force {
		return
	}

	logger.Infof("Initializing telegram")

	var err error
	bot, err = tgbotapi.NewBotAPI(config.TGToken)
	if err != nil {
		logger.Errorf("Problem starting telegram bot: %s", err.Error())
		return
	}

	go workerUpdateHandler()
}

// IsInit returns true if telegram client is initialized
func IsInit() bool {
	return botConnected
}

func MyApiID() int {
	return bot.Self.ID
}

// privates
func workerMessageHandler(id int) {
	logger.Debugf("Starting telegram message worker %v.", id)
	for message := range messageLoggingChan {
		// See Message P
		_, err := seeMessage(message)
		if err != nil {
			logger.Errorf("Error seeing from: %s", err.Error())
		}
	}
	logger.Debugf("Closing telegram message worker %v.", id)
}

func workerUpdateHandler() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Errorf("Problem starting telegram bot: %s", err.Error())
		return
	}

	botConnected = true

	logger.Infof("Telegram connected as %s", bot.Self.UserName)
	_, err = seeUser(&bot.Self)
	if err != nil {
		logger.Errorf("Problem seeing telegram bot: %s", err.Error())
	}

	for update := range updates {
		logger.Tracef("Got update: %v", update)

		if update.Message != nil {
			messageLoggingChan <- update.Message
		}
		if update.EditedMessage != nil {
			messageLoggingChan <- update.EditedMessage
		}
		if update.ChannelPost != nil {
			messageLoggingChan <- update.ChannelPost
		}
		if update.EditedChannelPost != nil {
			messageLoggingChan <- update.EditedChannelPost
		}
	}

	botConnected = false
}