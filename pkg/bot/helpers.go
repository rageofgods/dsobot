package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"sync"
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

// Covert weekday to localized weekday
func locMonth(month time.Month) string {
	var locMonth string
	switch month {
	case time.January:
		locMonth = "Январь"
	case time.February:
		locMonth = "Февраль"
	case time.March:
		locMonth = "Март"
	case time.April:
		locMonth = "Апрель"
	case time.May:
		locMonth = "Май"
	case time.June:
		locMonth = "Июнь"
	case time.July:
		locMonth = "Июль"
	case time.August:
		locMonth = "Август"
	case time.September:
		locMonth = "Сентябрь"
	case time.October:
		locMonth = "Октябрь"
	case time.November:
		locMonth = "Ноябрь"
	case time.December:
		locMonth = "Декабрь"
	}
	return locMonth
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

// Run separated goroutines to recreate duty event started from specific date
func (t *TgBot) updateOnDutyEvents(startFrom *time.Time, userName string, timeRangeText string) {
	// Recreate calendar duty event from current date if added duty in landed at this month
	if startFrom.Month() == time.Now().Month() {
		var wg = sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnDutyContDays, data.OnDutyTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события дежурства при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					userName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnValidationContDays, data.OnValidationTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события валидации при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					userName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
		wg.Wait()
	}
}

// Check if we have date in the command argument
func parseDateString(str string) (time.Time, error) {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return time.Time{}, err
	}

	tn := time.Time{}
	tn, err = time.ParseInLocation(botDataShort1, str, loc)
	if err != nil {
		tn, err = time.ParseInLocation(botDataShort2, str, loc)
		if err != nil {
			tn, err = time.ParseInLocation(botDataShort3, str, loc)
			if err != nil {
				return tn, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
					"Доступны следующие форматы:\n"+
					"*%q*\n"+
					"*%q*\n"+
					"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
			}
		}
	}
	return tn, nil
}

// Check if provided off-duty period is don't overlap with existing periods
func (t *TgBot) isOffDutyDatesOverlapWithCurrent(startDate time.Time, endDate time.Time,
	chatId int64, userId int64, messageId int) (bool, error) {
	// Get current MenData
	dutyMen := t.dc.DutyMenData()
	//Iterate over all men
	for _, man := range *dutyMen {
		// Find current man
		if man.TgID == userId {
			// Iterate over all available duties
			for _, offDuty := range man.OffDuty {
				// Convert string data to time object and check for parsing errors
				startOffDuty, err := parseDateString(offDuty.OffDutyStart)
				if err != nil {
					return true, err
				}
				// Convert string data to time object and check for parsing errors
				endOffDuty, err := parseDateString(offDuty.OffDutyEnd)
				if err != nil {
					messageText := fmt.Sprintf("%v", err)
					if err := t.sendMessage(messageText,
						chatId,
						&messageId,
						nil); err != nil {
						log.Printf("unable to send message: %v", err)
					}
					return true, err
				}

				// If new added start duty is in scope for old ones
				if (startDate.After(startOffDuty) || startDate.Equal(startOffDuty)) &&
					(startDate.Before(endOffDuty) || startDate.Equal(endOffDuty)) {
					messageText := fmt.Sprintf("Начало (%s) добавляемого нерабочего периода не должно "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						startDate, offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				} else if (endDate.After(startOffDuty) || endDate.Equal(startOffDuty)) &&
					(endDate.Before(endOffDuty) || endDate.Equal(endOffDuty)) {
					messageText := fmt.Sprintf("Конец (%s) добавляемого нерабочего периода не должен "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						endDate, offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				} else if startDate.Before(startOffDuty) && endDate.After(endOffDuty) {
					messageText := fmt.Sprintf("Период времени нового нерабочего периода не должен "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				}
			}
		}
	}
	return false, nil
}
