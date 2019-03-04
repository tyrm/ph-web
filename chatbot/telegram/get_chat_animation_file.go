package telegram

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"../../files"
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetChatAnimationFile(tgSticker *models.TGChatAnimation) (body []byte, err error) {
	start := time.Now()

	// Check if we've retrieved file already
	if tgSticker.FileLocation.Valid {
		b, err2 := files.GetBytes(fmt.Sprintf("chatbot/telegram/%s", tgSticker.FileLocation.String))
		if err2 == nil {
			body = *b

			elapsed := time.Since(start)
			logger.Tracef("GetPhotoFile() [%s][HIT]", elapsed)
			return
		}
		logger.Errorf("GetPhotoFile: error getting file bytes: %v", err)

	}

	// Get File Location
	fileConfig := tgbotapi.FileConfig{
		FileID: tgSticker.FileID,
	}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		logger.Errorf("GetPhotoFile: error getting file config: %v", err)
		return
	}
	logger.Tracef("%v", file)

	// Generate File URL
	getURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, file.FilePath)

	// Get File from Telegram API
	resp, err := http.Get(getURL)
	if err != nil {
		logger.Errorf("GetPhotoFile: error getting file: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	// Put in files
	r := regexp.MustCompile(`([[:word:]]+/)[[:word:]]+\.([[:alnum:]]+)`)
	urlPieces := r.FindStringSubmatch(file.FilePath)
	filesLocation := fmt.Sprintf("%s%s.%s", urlPieces[1], tgSticker.FileID, urlPieces[2])

	_, err = files.PutBytes(fmt.Sprintf("chatbot/telegram/%s", filesLocation), &body)
	if err != nil {
		logger.Errorf("Error putting file: %s", err)
	}

	err = tgSticker.UpdateFileRetrieved(filesLocation, urlPieces[2])
	if err != nil {
		logger.Errorf("Error updated record: %s", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("GetPhotoFile() [%s][MISS]", elapsed)
	return
}