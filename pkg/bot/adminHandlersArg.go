package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
)

// Handle 'duty' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutDuty() {
	t.msg.Text = "Создаю записи, ждите..."
	t.msg.ReplyToMessageID = t.update.Message.MessageID
	if _, err := t.bot.Send(t.msg); err != nil {
		log.Printf("unable to send message to admins: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(1, 2, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = "События дежурства успешно созданы"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}

// Handle 'validation' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutValidation() {
	t.msg.Text = "Создаю записи, ждите..."
	t.msg.ReplyToMessageID = t.update.Message.MessageID
	if _, err := t.bot.Send(t.msg); err != nil {
		log.Printf("unable to send message to admins: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(1, 1, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = "События валидации успешно созданы"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}

// Handle 'nwd' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutNonWorkingDay() {
	t.msg.Text = "Создаю записи, ждите..."
	t.msg.ReplyToMessageID = t.update.Message.MessageID
	if _, err := t.bot.Send(t.msg); err != nil {
		log.Printf("unable to send message to admins: %v", err)
	}

	err := t.dc.UpdateNwdEvents(1)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = "События нерабочих дней успешно созданы"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}
