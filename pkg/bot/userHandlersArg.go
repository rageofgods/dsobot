package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
	"strings"
	"time"
)

// Handle 'duty' user arg for 'whoison' command
func (t *TgBot) handleWhoIsOnDuty(arg string) {
	// Set current day for request by default
	tn := time.Now()
	// Check if we have two arguments
	if len(strings.Split(arg, " ")) == 2 {
		var err error
		tn, err = checkArgHasDate(arg)
		if err != nil {
			t.msg.Text = fmt.Sprintf("%v", err)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
			return
		}
	}

	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = fmt.Sprintf("Дежурный: %s", man)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}

// Handle 'duty' user arg for 'whoison' command
func (t *TgBot) handleWhoIsOnValidation(arg string) {
	// Set current day for request by default
	tn := time.Now()
	// Check if we have two arguments
	if len(strings.Split(arg, " ")) == 2 {
		var err error
		tn, err = checkArgHasDate(arg)
		if err != nil {
			t.msg.Text = fmt.Sprintf("%v", err)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
			return
		}
	}

	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = fmt.Sprintf("Валидирующий: %s", man)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}
