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
			messageText := fmt.Sprintf("%v", err)
			if err := t.sendMessage(messageText,
				t.update.Message.Chat.ID,
				&t.update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return
		}
	}

	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnDutyTag)
	// Get data for all men
	dutyMen := t.dc.DutyMenData()
	// Generate returned string
	for _, v := range *dutyMen {
		if v.UserName == man {
			man = fmt.Sprintf("%s (*@%s*)", v.FullName, v.UserName)
		}
	}

	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %v", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := fmt.Sprintf("Дежурный: %s", man)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
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
			messageText := fmt.Sprintf("%v", err)
			if err := t.sendMessage(messageText,
				t.update.Message.Chat.ID,
				&t.update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return
		}
	}

	// Get on-duty data
	man, err := t.dc.WhoIsOnDuty(&tn, data.OnValidationTag)
	// Get data for all men
	dutyMen := t.dc.DutyMenData()
	// Generate returned string
	for _, v := range *dutyMen {
		if v.UserName == man {
			man = fmt.Sprintf("%s (*@%s*)", v.FullName, v.UserName)
		}
	}

	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %v", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := fmt.Sprintf("Валидирующий: %s", man)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyDuty(arg string) {
	arg = "" // Ignore cmdArgs

	dates, err := t.dc.ManDutiesList(t.update.Message.From.UserName, data.OnDutyTag)
	if err != nil {
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	if len(*dates) == 0 {
		messageText := "Дежурства в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyValidation(arg string) {
	arg = "" // Ignore cmdArgs

	dates, err := t.dc.ManDutiesList(t.update.Message.From.UserName, data.OnValidationTag)
	if err != nil {
		messageText := fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}
	if len(*dates) == 0 {
		messageText := "Валидации в текущем месяце не найдены"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}
