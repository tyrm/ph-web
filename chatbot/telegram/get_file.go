package telegram

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"../../files"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type hasFiles interface {
	IsFileLocationValid() bool
	GetFileID() string
	GetFileLocation() string
	UpdateFileRetrieved(string, string) error
}

func GetFile(tgOjb hasFiles) (body []byte, err error) {
	start := time.Now()

	// Check if we've retrieved file already
	if tgOjb.IsFileLocationValid() {
		b, err2 := files.GetBytes(fmt.Sprintf("chatbot/telegram/%s", tgOjb.GetFileLocation()))
		if err2 == nil {
			body = *b

			elapsed := time.Since(start)
			logger.Tracef("GetFile() [%s][HIT]", elapsed)
			return
		}
		logger.Errorf("GetFile: error getting file bytes: %v", err)

	}

	// Get File Location
	fileConfig := tgbotapi.FileConfig{
		FileID: tgOjb.GetFileID(),
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
		logger.Errorf("GetPhotoFile: error getting file: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	// Put in files
	r := regexp.MustCompile(`([[:word:]]+/)[[:word:]]+\.([[:alnum:]]+)`)
	urlPieces := r.FindStringSubmatch(file.FilePath)
	filesLocation := fmt.Sprintf("%s%s.%s", urlPieces[1], tgOjb.GetFileID(), urlPieces[2])

	_, err = files.PutBytes(fmt.Sprintf("chatbot/telegram/%s", filesLocation), &body)
	if err != nil {
		logger.Errorf("Error putting file: %s", err)
	}

	err = tgOjb.UpdateFileRetrieved(filesLocation, urlPieces[2])
	if err != nil {
		logger.Errorf("Error updated record: %s", err)
	}

	elapsed := time.Since(start)
	logger.Tracef("GetFile() [%s][MISS]", elapsed)
	return
}