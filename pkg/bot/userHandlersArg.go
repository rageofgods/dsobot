package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
	"time"
)

// Handle 'duty' user arg for 'whoison' command
func (t *TgBot) handleWhoIsOnDuty() {
	tn := time.Now()
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = fmt.Sprintf("Дежурный: %s", man)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}

// Handle 'duty' user arg for 'whoison' command
func (t *TgBot) handleWhoIsOnValidation() {
	tn := time.Now()
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = fmt.Sprintf("Валидирующий: %s", man)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}
