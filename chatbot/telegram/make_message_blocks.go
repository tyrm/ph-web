package telegram

import "time"

type MessageBlock struct {
	UserAPIID       int
	UserDisplayName string
	UserPictureURL  string
	Me              bool

	Messages []MessageBlockMessage
}

type MessageBlockMessage struct {
	ID   int
	Text string
	Date time.Time
}

func MakeMessageBlocks() (blocks *[]MessageBlock, err error) {

	return
}