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
	// Get data for all men
	dutyMen := t.dc.DutyMenData()
	// Generate returned string
	for _, v := range *dutyMen {
		if v.TgID == man {
			man = fmt.Sprintf("%s (*@%s*)", v.Name, v.TgID)
		}
	}

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
	// Get data for all men
	dutyMen := t.dc.DutyMenData()
	// Generate returned string
	for _, v := range *dutyMen {
		if v.TgID == man {
			man = fmt.Sprintf("%s (*@%s*)", v.Name, v.TgID)
		}
	}

	if err != nil {
		log.Printf("error in event creating: %v", err)
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	} else {
		t.msg.Text = fmt.Sprintf("Валидирующий: %s", man)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
	}
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyDuty(arg string) {
	arg = "" // Ignore cmdArgs

	dates, err := t.dc.ManDutiesList(t.update.Message.From.UserName, data.OnDutyTag)
	if err != nil {
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	if len(*dates) == 0 {
		t.msg.Text = "Дежурства в текущем месяце не найдены"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	list := fmt.Sprintf("*Список дней дежурств в текущем месяце (%d):*\n", len(*dates))
	for index, date := range *dates {
		list += fmt.Sprintf("*%d.* - %s (%s)\n",
			index+1,
			date.Format(botDataShort3),
			locWeekday(date.Weekday()))
	}
	t.msg.Text = list
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// Handle 'duty' user arg for 'showmy' command
func (t *TgBot) handleShowMyValidation(arg string) {
	arg = "" // Ignore cmdArgs

	dates, err := t.dc.ManDutiesList(t.update.Message.From.UserName, data.OnValidationTag)
	if err != nil {
		t.msg.Text = fmt.Sprintf("Не удалось выполнить запрос: %s", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	if len(*dates) == 0 {
		t.msg.Text = "Валидации в текущем месяце не найдены"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	list := fmt.Sprintf("*Список дней валидаций в текущем месяце (%d):*\n", len(*dates))
	for index, date := range *dates {
		list += fmt.Sprintf("*%d.* - %s (%s)\n",
			index+1,
			date.Format(botDataShort3),
			locWeekday(date.Weekday()))
	}
	t.msg.Text = list
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}
