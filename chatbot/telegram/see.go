package telegram

import (
	"database/sql"

	"../../models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func seeUser(apiUser *tgbotapi.User) (tgUser *models.TGUser, err error) {
	// Get TGUser entry, create if not exists
	tgu, err2 := models.ReadTGUserByAPIID(apiUser.ID)
	if err2 == models.ErrDoesNotExist {
		var err3 error
		tgu, err3 = models.CreateTGUser(apiUser.ID, apiUser.IsBot)
		if err3 != nil {
			err = err3
			return
		}
	} else if err2 != nil {
		err = err2
		return
	}

	// Check History is up to date
	err2 = nil
	tguh, err2 := tgu.ReadLatestHistory()
	if err2 == models.ErrDoesNotExist {
		logger.Tracef("seeUser: user has no history. creating.")
		lastName := &sql.NullString{Valid: false}
		if apiUser.LastName != "" {
			lastName = &sql.NullString{
				String: apiUser.LastName,
				Valid: true,
			}
		}

		username := &sql.NullString{Valid: false}
		if apiUser.UserName != "" {
			username = &sql.NullString{
				String: apiUser.UserName,
				Valid: true,
			}
		}

		languageCode := &sql.NullString{Valid: false}
		if apiUser.LanguageCode != "" {
			languageCode = &sql.NullString{
				String: apiUser.LanguageCode,
				Valid: true,
			}
		}

		var err3 error
		tguh, err3 = models.CreateTGUserHistory(tgu, apiUser.FirstName, *lastName, *username, *languageCode)
		if err3 != nil {
			err = err3
			return
		}

		tgUser = tgu
		return
	} else if err2 != nil {
		err = err2
		return
	}

	if tguh.Matches(apiUser) {
		logger.Tracef("seeUser: user's value match history. updating last seen.")
		err2 = nil
		err2 = tguh.UpdateLastSeen()
		if err2 != nil {
			logger.Errorf("Error updating last seen for $d.", tgUser.ID)
		}
	} else {
		logger.Tracef("seeUser: user has changed value. creating.")
		lastName := &sql.NullString{Valid: false}
		if apiUser.LastName != "" {
			lastName = &sql.NullString{
				String: apiUser.LastName,
				Valid: true,
			}
		}

		username := &sql.NullString{Valid: false}
		if apiUser.UserName != "" {
			username = &sql.NullString{
				String: apiUser.UserName,
				Valid: true,
			}
		}

		languageCode := &sql.NullString{Valid: false}
		if apiUser.LanguageCode != "" {
			languageCode = &sql.NullString{
				String: apiUser.LanguageCode,
				Valid: true,
			}
		}

		err2 = nil
		tguh, err2 = models.CreateTGUserHistory(tgu, apiUser.FirstName, *lastName, *username, *languageCode)
		if err2 != nil {
			err = err2
			return
		}
	}

	tgUser = tgu
	return
}
