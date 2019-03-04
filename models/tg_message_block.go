package models

import (
	"database/sql"
	"time"
)

type TGMessageBlocksData struct {
	ID         int
	APIID      int64
	FirstName  sql.NullString
	LastName   sql.NullString
	Text       sql.NullString
	StickerID  sql.NullInt64
	Date       time.Time
	PhotoCount int
}

const sqlReadTGMessageBlocksData = `
SELECT DISTINCT ON (m.id) m.id, fu.api_id, fh.first_name, fh.last_name, m.text, m.sticker_id, m.date, 
sum(case when p.tgm_id = m.id then 1 else 0 end) as photos
FROM tg_messages as m
LEFT JOIN tg_message_photos p ON m.id = p.tgm_id
LEFT JOIN tg_users as fu ON m.from_id = fu."id"
LEFT JOIN tg_users_history as fh ON m.from_id = fh.tgu_id
WHERE m.chat_id = $1
GROUP BY m.id, fu.api_id, fh.first_name, fh.last_name
ORDER BY m.id DESC LIMIT $2 OFFSET $3;`

func ReadTGMessageBlocksData(chatID int64, limit uint, page uint) (messageList *[]TGMessageBlocksData, err error) {
	offset := limit * page
	var newMessageBlocks []TGMessageBlocksData

	rows, err := db.Query(sqlReadTGMessageBlocksData, chatID, limit, offset)
	if err != nil {
		logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
		return
	}

	for rows.Next() {
		var newID int
		var newAPIID int64
		var newFirstName sql.NullString
		var newLastName sql.NullString
		var newText sql.NullString
		var newStickerID sql.NullInt64
		var newDate time.Time
		var newPhotoCount int

		err = rows.Scan(&newID, &newAPIID, &newFirstName, &newLastName, &newText, &newStickerID, &newDate,
			&newPhotoCount)
		if err != nil {
			logger.Tracef("ReadUsersPage(%d, %d) (%v, %v)", limit, page, nil, err)
			return
		}
		messageBlocks := TGMessageBlocksData{
			ID         :newID,
			APIID      :newAPIID,
			FirstName  :newFirstName,
			LastName   :newLastName,
			Text       :newText,
			StickerID  :newStickerID,
			Date       :newDate,
			PhotoCount :newPhotoCount,
		}

		newMessageBlocks = append(newMessageBlocks, messageBlocks)
	}

	messageList = &newMessageBlocks
	return
}
