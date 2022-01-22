package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

// Get requested user info
func (t *TgBot) getChatMember(userId int64, chatId int64) (*tgbotapi.ChatMember, error) {
	u, err := t.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId}})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Reply to user message
func (t *TgBot) sendMessage(message string, chatId int64, replyId *int, keyboard interface{}, pin ...bool) error {
	msg := tgbotapi.NewMessage(chatId, message)

	msg.ParseMode = "markdown"
	if replyId != nil {
		msg.ReplyToMessageID = *replyId
	}
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	// Reply to message
	sentMessage, err := t.bot.Send(msg)
	if err != nil {
		return err
	}
	// Check if we want to sent message
	if len(pin) == 1 {
		if pin[0] {
			pin := tgbotapi.PinChatMessageConfig{MessageID: sentMessage.MessageID, ChatID: chatId}
			_, err = t.bot.Request(pin)
			if err != nil {
				message = fmt.Sprintf("Не удалось закрепить сообщение для chatID: %d\nОшибка: (%v)", chatId, err)
				if err := t.sendMessage(message, t.adminGroupId, nil, nil); err != nil {
					log.Printf("%v", err)
				}
				return err
			}
		}
	}
	return nil
}

func genHelpCmdText(commands []botCommand) string {
	var cmdList string
	for i, cmd := range commands {
		var argList string
		if cmd.command.args != nil {
			argList = fmt.Sprintf("*Возможные значения аргумента:*\n")
			for index, arg := range *cmd.command.args {
				argList += fmt.Sprintf("  *%s*: *%s* %q\n",
					string(rune('a'-1+index+1)), // Convert number 1,2,3,etc. to char accordingly a,b,c,etc.
					arg.name,
					arg.description,
				)
			}
		}
		// Append <argument> suffix to command help if any arguments was found
		var argType string
		if argList != "" {
			argType = " *<аргумент>*"
		}
		// Generate lit of commands
		cmdList += fmt.Sprintf("*%d*: */%s*%s - `%s`\n%s",
			i+1,
			cmd.command.name,
			argType,
			cmd.description,
			argList)
	}
	return cmdList
}

// Covert weekday to localized weekday
func locWeekday(weekday time.Weekday) string {
	var locWeekday string
	switch weekday {
	case time.Sunday:
		locWeekday = "Воскресенье"
	case time.Monday:
		locWeekday = "Понедельник"
	case time.Tuesday:
		locWeekday = "Вторник"
	case time.Wednesday:
		locWeekday = "Среда"
	case time.Thursday:
		locWeekday = "Четверг"
	case time.Friday:
		locWeekday = "Пятница"
	case time.Saturday:
		locWeekday = "Суббота"
	}
	return locWeekday
}

// Return string with duty types data for specified man
func typesOfDuties(m *data.DutyMan) string {
	var list string
	var isAnyDuty bool
	for _, dt := range m.DutyType {
		if dt.Enabled {
			list += fmt.Sprintf("%s, ", dt.Name)
			isAnyDuty = true
		}
	}
	if !isAnyDuty {
		return "❗️"
	}
	return strings.Trim(strings.Trim(list, " "), ",")
}

// Return json marshaled object for callback message data
func marshalCallbackData(cm callbackMessage, itemIndex int, buttonIndex int, enabled ...bool) ([]byte, error) {
	// Generate callback data
	// Format: 'itemIndex-buttonIndex-Answer'
	cm.Answer = strconv.Itoa(itemIndex)                        // Save current item index to data
	cm.Answer += fmt.Sprintf("-%s", strconv.Itoa(buttonIndex)) // Save current button index to data
	// If we have optional argument
	if len(enabled) == 1 {
		// Append current button state as suffix
		// '-1' - button is 'active' ✅
		// '-0' - button is 'passive' ❌
		if enabled[0] {
			cm.Answer += fmt.Sprintf("-%s", inlineKeyboardEditDutyYes)
		} else {
			cm.Answer += fmt.Sprintf("-%s", inlineKeyboardEditDutyNo)
		}
	}
	// Save our data
	jsonData, err := json.Marshal(cm)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonData) > 64 {
		return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
	}
	return jsonData, nil
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

// Run separated goroutines to recreate duty event started from specific date
func (t *TgBot) updateOnDutyEvents(startFrom *time.Time, update *tgbotapi.Update, timeRangeText string) {
	// Recreate calendar duty event from current date if added duty in landed at this month
	if startFrom.Month() == time.Now().Month() {
		go func() {
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnDutyContDays, data.OnDutyTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события дежурства при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					update.Message.From.UserName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
		go func() {
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnValidationContDays, data.OnValidationTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события валидации при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					update.Message.From.UserName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
	}
}
