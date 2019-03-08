package telegram

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeUserProfilePhotos(apiUserProfilePhotos *tgbotapi.UserProfilePhotos) (err error) {
	if !botConnected {
		err = ErrNotInit
		return
	}

	if apiUserProfilePhotos.Photos != nil {
		for _, photoSlice := range apiUserProfilePhotos.Photos {
			for _, photo := range photoSlice {
				_, err2 := seePhotoSize(&photo)
				if err2 != nil {
					logger.Errorf("seeMessage: error seeing forwarded from user: %s", err2)
					err = err2
				}
			}
		}
	}
	return
}
