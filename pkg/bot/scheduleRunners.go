package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
	"time"
)

// Send announce message to user group chat
func (t *TgBot) announceDuty() {
	// Setup time now
	tn := time.Now()
	// Check if current day is non-working day
	nwd, err := t.dc.IsNWD(tn)
	if err != nil {
		log.Printf("%v", err)
	}
	// Don't announce if non-working day
	if nwd {
		return
	}

	// Get current duty data
	dutyMen := t.dc.DutyMenData()
	// Define duty and validation man variables
	var dm data.DutyMan
	var vm data.DutyMan
	// iterate over all groups and announce if any
	for i, v := range t.settings.JoinedGroups {
		if v.Announce {
			// Get on-duty data
			dutyMan, err := t.dc.WhoIsOnDuty(&tn, data.OnDutyTag)
			if err != nil {
				log.Printf("%v", err)
			}
			for _, v := range *dutyMen {
				if v.UserName == dutyMan {
					dm = v
				}
			}
			validationMan, err := t.dc.WhoIsOnDuty(&tn, data.OnValidationTag)
			if err != nil {
				log.Printf("%v", err)
			}
			for _, v := range *dutyMen {
				if v.UserName == validationMan {
					vm = v
				}
			}
			// Setup men names
			var dMan string
			var vMan string
			if dm.TgID != 0 {
				dMan = fmt.Sprintf("%s *@%s*", dm.CustomName, dm.UserName)
			} else {
				dMan = "*-*"
			}
			if vm.TgID != 0 {
				vMan = fmt.Sprintf("%s *@%s*", vm.CustomName, vm.UserName)
			} else {
				vMan = "*-*"
			}
			message := fmt.Sprintf("Доброе утро!\n\n*Дежурный* на сегодня: %s\n"+
				"*Валидирующий* на сегодня: %s\n\nОтличного дня!👍",
				dMan,
				vMan)
			if err := t.sendMessage(message,
				t.settings.JoinedGroups[i].Id,
				nil,
				nil,
				true); err != nil {
				log.Printf("%v", err)
			}
		}
	}
}

// update non-working events
func (t *TgBot) updateNwd() {
	err := t.dc.UpdateNwdEvents()
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось создать события нерабочих дней: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "События нерабочих дней успешно созданы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// update on-duty events
func (t *TgBot) updateOnDuty() {
	err := t.dc.UpdateOnDutyEvents(data.OnDutyContDays, data.OnDutyTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось создать события дежурств: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "События нерабочих дней успешно созданы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}

// update on-validation events
func (t *TgBot) updateOnValidation() {
	err := t.dc.UpdateOnDutyEvents(data.OnValidationContDays, data.OnValidationTag)
	if err != nil {
		log.Printf("error in event creating: %v", err)
		messageText := fmt.Sprintf("Не удалось создать события валидаций: %s", err)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "События нерабочих дней успешно созданы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}
