package chatbot

import (
	"../config"
	"./telegram"
	"github.com/juju/loggo"
)

var logger *loggo.Logger

func init() {
	newLogger := loggo.GetLogger("chatbot")
	logger = &newLogger
}

// InitClients tries to init bot clients
func InitClients(config config.Config) {
	telegram.InitClient(config, false)
}
