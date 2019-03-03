package telegram

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"../../files"
	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetFile(tgPhotoSize *models.TGPhotoSize) (body []byte, err error) {
	// Get File Location
	fileConfig := tgbotapi.FileConfig{
		FileID: tgPhotoSize.FileID,
	}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		logger.Errorf("GetFile: error getting file config: %v", err)
		return
	}
	logger.Tracef("%v", file)

	// Generate File URL
	getURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, file.FilePath)

	// Get File from Telegram API
	resp, err := http.Get(getURL)
	if err != nil {
		logger.Errorf("GetFile: error getting file: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	// Put in files
	r := regexp.MustCompile(`([[:word:]]+/)[[:word:]]+\.([[:alnum:]]+)`)
	urlPieces := r.FindStringSubmatch(file.FilePath)
	filesLocation := fmt.Sprintf("%s%s.%s", urlPieces[1], tgPhotoSize.FileID, urlPieces[2])

	_, err = files.PutBytes(filesLocation, &body)
	if err != nil {
		logger.Errorf("Error putting file: %s", err)
	}

	err = tgPhotoSize.UpdateFileRetrieved(filesLocation, urlPieces[2])
	if err != nil {
		logger.Errorf("Error updated record: %s", err)
	}

	return
}