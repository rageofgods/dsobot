package bot

import (
	data2 "dso_bot/internal/data"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// announceBirthdayAtWorkingDay it's wrapper for birthday announce only at working day
func (t *TgBot) announceBirthdayAtWorkingDay() {
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
	t.announceBirthday()
}

// announceBirthdayAtNonWorkingDay it's wrapper for birthday announce only at non-working day
func (t *TgBot) announceBirthdayAtNonWorkingDay() {
	// Setup time now
	tn := time.Now()
	// Check if current day is non-working day
	nwd, err := t.dc.IsNWD(tn)
	if err != nil {
		log.Printf("%v", err)
	}
	// Announce if non-working day
	if nwd {
		t.announceBirthday()
	}
}

func (t *TgBot) announceBirthday() {
	// Get current duty data
	dutyMen := t.dc.DutyMenData()

	var menBirthDay []data2.DutyMan
	//var menBirthDay []string
	for _, v := range *dutyMen {
		if v.Birthday != "" {
			pbd, err := time.Parse(botDataShort3, v.Birthday)
			if err != nil {
				message := fmt.Sprintf(
					"unable to parse birthday date: %s for user: %s",
					v.Birthday,
					v.CustomName,
				)
				log.Print(message)
				if err := t.sendMessage(message, t.adminGroupId, nil, nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
			if pbd.Month() == time.Now().Month() && pbd.Day() == time.Now().Day() {
				menBirthDay = append(menBirthDay, v)
			}
		}
	}

	if len(menBirthDay) != 0 {
		// iterate over all groups and announce if any
		for i, v := range t.settings.JoinedGroups {
			if v.Announce {
				// Templating announce message
				message := "🎂 Ура! Сегодня день рожденья у коллег:\n\n"
				for _, v := range menBirthDay {
					message += fmt.Sprintf("%s *(@%s)*\n", v.CustomName, v.UserName)
				}
				message += "\nПоздравляем!!! 🎉 🎈 🎁"
				if err := t.sendMessage(message,
					t.settings.JoinedGroups[i].Id,
					nil,
					nil); err != nil {
					log.Printf("%v", err)
				}
			}
		}
	}
}

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
	var dm data2.DutyMan
	var vm data2.DutyMan

	// Generate off-duty Announce message
	var offDutyAnnMessage string
	a, err := t.getOffDutyAnnounces(4)
	if err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("%v", a)
	fmt.Printf("%v", len(a))
	if len(a) != 0 {
		_, err := t.dc.SaveMenList()
		if err != nil {
			messageText := fmt.Sprintf("Не удалось сохранить событие: %v", err)
			if err := t.sendMessage(messageText,
				t.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
		offDutyAnnMessage = "------------------------\n"
		offDutyAnnMessage += formatOffDutyAnnounces(a)
	}

	// iterate over all groups and announce if any
	for i, v := range t.settings.JoinedGroups {
		if v.Announce {
			// Get on-duty data
			dutyMan, err := t.dc.WhoIsOnDuty(&tn, data2.OnDutyTag)
			if err != nil {
				log.Printf("%v", err)
			}
			for _, v := range *dutyMen {
				if v.UserName == dutyMan {
					dm = v
				}
			}
			validationMan, err := t.dc.WhoIsOnDuty(&tn, data2.OnValidationTag)
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
				dMan = fmt.Sprintf("%s *(@%s)*", dm.CustomName, dm.UserName)
			} else {
				dMan = "*-*"
			}
			if vm.TgID != 0 {
				vMan = fmt.Sprintf("%s *(@%s)*", vm.CustomName, vm.UserName)
			} else {
				vMan = "*-*"
			}

			// Setup cheer message
			var cheer string
			if dMan == vMan {
				cheer = "May the Force be with you!"
			} else {
				cheer = "Good luck and have fun!"
			}

			// Templating announce message
			message := fmt.Sprintf("📣Доброе утро!\n\n*Дежурный* сегодня: %s\n"+
				"*Валидирующий* сегодня: %s\n\n*%s*💪\n\n"+
				"*Tip*: %s\n\n",
				dMan, vMan, cheer, genRndTip())

			// Append off-duty Announce message
			message += offDutyAnnMessage

			image, err := t.genMonthDutyImage()
			if err != nil {
				log.Printf("%v", err)
				messageText := fmt.Sprintf(
					"Не удалось сгенерировать визуализацию дежурств за месяц: %v",
					err,
				)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}

			// If we don't have month duty image, then we just send plain text instead
			if image == nil {
				if err := t.sendMessage(message, t.adminGroupId, nil, nil, true); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			} else {
				msg := tgbotapi.NewPhoto(t.settings.JoinedGroups[i].Id, image)
				msg.Caption = message
				msg.ParseMode = "markdown"

				sentMessage, err := t.bot.Send(msg)
				if err != nil {
					log.Printf("unable to send message: %v", err)
				}

				if err := pinMessage(t, t.settings.JoinedGroups[i].Id, sentMessage); err != nil {
					log.Printf("announceDuty: %v", err)
				}
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
	err := t.dc.UpdateOnDutyEvents(data2.OnDutyContDays, data2.OnDutyTag)
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
		messageText := "События дежурств успешно созданы"
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
	err := t.dc.UpdateOnDutyEvents(data2.OnValidationContDays, data2.OnValidationTag)
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
		messageText := "События валидаций успешно созданы"
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}
