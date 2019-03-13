package telegram

import (
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	)

func SendMessage(chat *models.TGChat, text string) error {
	letter := tgbotapi.NewMessage(chat.APIID, text)
	msg, err := bot.Send(letter)

	if err != nil {
		return err
	}

	_, err = seeMessage(&msg)
	return nil
}