package telegram

import (
	"time"

	"../../registry"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/juju/loggo"
	"github.com/patrickmn/go-cache"
)

var bot *tgbotapi.BotAPI
var botConnected = false
var logger *loggo.Logger
var messageChan chan *tgbotapi.Message

// Caches
var cUserProfilePhotos *cache.Cache

func init() {
	newLogger := loggo.GetLogger("telegram")
	logger = &newLogger

	messageChan = make(chan *tgbotapi.Message, 100)
	for w := 1; w <= 3; w++ {
		go workerMessageHandler(w)
	}

	// init cache
	cUserProfilePhotos = cache.New(5*time.Minute, 10*time.Minute)
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
	_, err = seeUser(&bot.Self)
	if err != nil {
		logger.Errorf("Problem seeing telegram bot: %s", err.Error())
	}

	go workerUpdateHandler()
}

// IsInit returns true if telegram client is initialized
func IsInit() bool {
	return botConnected
}

// privates
func workerMessageHandler(id int) {
	logger.Debugf("Starting telegram message worker %v.", id)
	for message := range messageChan {
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

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		logger.Tracef("Got update: %v", update)

		messageChan <- update.Message
	}

	botConnected = false
}