package telegram

import (
	"fmt"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/patrickmn/go-cache"
)

func GetUserProfilePhotos(apiid int) (up *tgbotapi.UserProfilePhotos, err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	// check cache
	apiidStr := strconv.Itoa(apiid)
	if u, found := cUserProfilePhotos.Get(apiidStr); found {
		up = u.(*tgbotapi.UserProfilePhotos)
		logger.Tracef("GetUserProfilePhotos(%s) [HIT]", apiidStr)
		return
	}

	// Get from API
	config := tgbotapi.UserProfilePhotosConfig{
		UserID: apiid,
	}
	newUPP, err := bot.GetUserProfilePhotos(config)

	// update cache
	cUserProfilePhotos.Set(apiidStr, &newUPP, cache.DefaultExpiration)

	// See Photos
	err2 := seeUserProfilePhotos(&newUPP)
	if err2 != nil {
		logger.Errorf("error seeing UserProfilePhotos: %S", err2)
	}

	// return value
	up = &newUPP
	logger.Tracef("GetUserProfilePhotos(%s) [MISS]", apiidStr)
	return
}


func GetUserProfilePhotoCurrent(apiid int, size int) (photoURI string, err error) {
	upp, err := GetUserProfilePhotos(apiid)
	if err != nil {
		logger.Errorf("GetUserProfilePhotoCurrent: could not populate photo: %s", err)
		return
	}

	// Get Smallest image larger than request
	var fileID string
	var foundW, foundH int

	if len(upp.Photos) > 0 {
		for _, photo := range upp.Photos[0] {
			if photo.Width > size && photo.Height > size {
				// Init if zero
				if foundW == 0 || foundH == 0 {
					foundW = photo.Width
					foundH = photo.Height
					fileID = photo.FileID
				}

				if photo.Width < foundW || photo.Height < foundH {
					foundW = photo.Width
					foundH = photo.Height
					fileID = photo.FileID
				}
			}
		}

		// If Empty return the largest image
		if fileID == "" {
			foundW = 0
			foundH = 0
			for _, photo := range upp.Photos[0] {
				if foundW == 0 || foundH == 0 {
					foundW = photo.Width
					foundH = photo.Height
					fileID = photo.FileID
				}

				if photo.Width > foundW || photo.Height > foundH {
					foundW = photo.Width
					foundH = photo.Height
					fileID = photo.FileID
				}
			}

		}
		photoURI = fmt.Sprintf("/web/chatbot/tg/photos/%s/file", fileID)
	} else {
		// Return generic if no photo
		photoURI = fmt.Sprint("https://o.pup.haus/public/img/user-icon.jpg")
	}
	return
}