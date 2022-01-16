package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
)

// Handle 'all' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutAll(arg string) {
	arg = "" // Ignore cmdArgs
	go t.adminHandleRolloutDuty("")
	go t.adminHandleRolloutValidation("")
}

// Handle 'duty' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutDuty(arg string) {
	arg = "" // Ignore cmdArgs
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.OrdinaryDutyName)
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(1, onDutyContDays, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "События дежурства успешно сгенерированы"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'validation' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutValidation(arg string) {
	arg = "" // Ignore cmdArgs
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.ValidationDutyName)
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(1, onValidationContDays, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText = "События валидации успешно сгенерированы"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'nwd' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutNonWorkingDay(arg string) {
	arg = "" // Ignore cmdArgs
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.NonWorkingDaySum)
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateNwdEvents(1)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText = "События нерабочих дней успешно сгенерированы"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}
