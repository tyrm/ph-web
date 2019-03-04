package telegram

import (
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/patrickmn/go-cache"
)

func GetUserProfilePhotos(apiid int) (up *tgbotapi.UserProfilePhotos, err error) {
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
