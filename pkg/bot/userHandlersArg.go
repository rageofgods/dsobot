package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

// Handle 'duty' user arg for 'whoison_duty' command
func (t *TgBot) handleWhoIsOnDutyToday(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	// Set current day for request by default
	tn := time.Now()
	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := "Дежурства не найдены."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		// Get data for all men
		dutyMen := t.dc.DutyMenData()
		// Generate returned string
		for _, v := range *dutyMen {
			if v.UserName == man {
				man = fmt.Sprintf("%s (*@%s*)", v.CustomName, v.UserName)
			}
		}
		messageText := fmt.Sprintf("Дежурный: %s", man)
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'duty' user arg for 'whoison_duty date' command
func (t *TgBot) handleWhoIsOnDutyAtDate(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	if _, err := t.tmpOffDutyDataForUser(update.Message.From.ID); err == nil {
		messageText := "Вы уже работаете с данными нерабочего периода"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleWhoIsOnDutyAtDate,
	}

	numericKeyboard, err := genInlineCalendarKeyboard(time.Now(), *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextUserHandleWhoIsOnDutyAtDate
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Handle 'duty' user arg for 'whoison_validation' command
func (t *TgBot) handleWhoIsOnValidationToday(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	// Set current day for request by default
	tn := time.Now()
	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := "Валидации не найдены."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		// Get data for all men
		dutyMen := t.dc.DutyMenData()
		// Generate returned string
		for _, v := range *dutyMen {
			if v.UserName == man {
				man = fmt.Sprintf("%s (*@%s*)", v.CustomName, v.UserName)
			}
		}
		messageText := fmt.Sprintf("Валидирующий: %s", man)
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'duty' user arg for 'whoison_validation date' command
func (t *TgBot) handleWhoIsOnValidationAtDate(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	if _, err := t.tmpOffDutyDataForUser(update.Message.From.ID); err == nil {
		messageText := "Вы уже работаете с данными нерабочего периода"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleWhoIsOnValidationAtDate,
	}

	numericKeyboard, err := genInlineCalendarKeyboard(time.Now(), *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextUserHandleWhoIsOnValidationAtDate
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyDuty(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here

	dates, err := t.dc.ManDutiesList(update.Message.From.UserName, data.OnDutyTag)
	if err != nil {
		messageText := "Дежурства в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	if len(*dates) == 0 {
		messageText := "Дежурства в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	list := fmt.Sprintf("*Список дней дежурств в текущем месяце (%d):*\n", len(*dates))
	for index, date := range *dates {
		list += fmt.Sprintf("*%d.* - %s (%s)\n",
			index+1,
			date.Format(botDataShort3),
			locWeekday(date.Weekday()))
	}
	messageText := list
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyValidation(arg string, update *tgbotapi.Update) {
	log.Println(arg) // Ignore arg here

	dates, err := t.dc.ManDutiesList(update.Message.From.UserName, data.OnValidationTag)
	if err != nil {
		messageText := "Валидации в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	if len(*dates) == 0 {
		messageText := "Валидации в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	list := fmt.Sprintf("*Список дней валидаций в текущем месяце (%d):*\n", len(*dates))
	for index, date := range *dates {
		list += fmt.Sprintf("*%d.* - %s (%s)\n",
			index+1,
			date.Format(botDataShort3),
			locWeekday(date.Weekday()))
	}
	messageText := list
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}
