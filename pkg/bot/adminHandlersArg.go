package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// Handle 'all' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutAll(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here

	go t.adminHandleRolloutDuty("", update)
	go t.adminHandleRolloutValidation("", update)
}

// Handle 'duty' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutDuty(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.OrdinaryDutyName)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(data.OnDutyContDays, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "События дежурства успешно сгенерированы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'validation' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutValidation(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.ValidationDutyName)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateOnDutyEvents(data.OnValidationContDays, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText = "События валидации успешно сгенерированы"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'nwd' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutNonWorkingDay(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	messageText := fmt.Sprintf("Создаю записи для типа событий: %q, ждите...", data.NonWorkingDaySum)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	err := t.dc.UpdateNwdEvents()
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText = "События нерабочих дней успешно сгенерированы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}
