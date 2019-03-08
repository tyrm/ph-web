package telegram

import (
	"fmt"

	"../../models"
)

func PopulateProfilePhotos(userList []models.TGUser, size int) []models.TGUser {
	if !botConnected {
		return userList
	}

	for i := range userList {
		upp, err := GetUserProfilePhotos(userList[i].APIID)
		if err != nil {
			logger.Errorf("PopulateProfilePhotos: could not populate photo: %s", err)
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

		userList[i].ProfilePhotoURL = fmt.Sprintf("/web/chatbot/tg/photos/%s/file", fileID)
	}

	return userList
}
