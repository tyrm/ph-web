package models

import "time"

type TGMessageBlock struct {
	UserAPIID       int
	UserDisplayName string
	UserPictureURL  string
	Me              bool

	Messages []TGMessageBlockMessage
}

type TGMessageBlockMessage struct {
	ID   int
	Text string
	Date time.Time
}
