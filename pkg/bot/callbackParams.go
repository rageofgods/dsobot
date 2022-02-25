package bot

import (
	"math/rand"
	"time"
)

// Aggregate struct to support multiple callback types
type callbackButton struct {
	callbackMessage callbackMessage
	data            interface{}
}

// Button struct for man index button
type callbackButtonIndex struct {
	targetUserTgId int64
}

// Button struct for cancel button
type callbackButtonCancel struct{}

// Button struct for ok button
type callbackButtonOk struct{}

// Button struct for date button
type callbackButtonDate struct {
	date time.Time
}

// Button struct for previous button
type callbackButtonPrev struct {
	date time.Time
}

// Button struct for next button
type callbackButtonNext struct {
	date time.Time
}

// Generate new callback button with interface data
func newCallbackButton(t *TgBot, cm *callbackMessage, i interface{}) string {
	id := genRandomString(64)
	c := callbackButton{
		callbackMessage: *cm,
		data:            i,
	}
	t.callbackButton[id] = c
	return id
}

// Generate random string with variable length
func genRandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// return data struct
func (c *callbackButton) getCallbackData() interface{} {
	return c.data
}

func (c *callbackButton) getCallbackMessage() callbackMessage {
	return c.callbackMessage
}
