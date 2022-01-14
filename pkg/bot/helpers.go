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

// Checks if user has First/Last name and return correct full name
func genUserFullName(firstName string, lastName string) string {
	var fullName string
	if lastName == "" {
		log.Println("user has no last name")
		fullName = firstName
	} else {
		fullName = firstName + " " + lastName
	}
	return fullName
}

func genInlineYesNoKeyboardWithData(yes *callbackMessage, no *callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Generate jsons for data
	jsonYes, err := json.Marshal(yes)
	if err != nil {
		log.Println(err)
	}
	jsonNo, err := json.Marshal(no)
	if err != nil {
		log.Println(err)
	}

	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonNo) > 64 {
		return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonNo))
	} else if len(jsonYes) > 64 {
		return nil, fmt.Errorf("jsonYes size is greater then 64b: %v", len(jsonNo))
	}

	// Create numeric inline keyboard
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", string(jsonYes)),
			tgbotapi.NewInlineKeyboardButtonData("No", string(jsonNo)),
		),
	)
	return &numericKeyboard, nil
}

func genInlineOffDutyKeyboardWithData(offDutyList []string, cm callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, v := range offDutyList {
		cm.Answer = strconv.Itoa(i) // Save current index to data
		jsonData, err := json.Marshal(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
		}
		// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
		if len(jsonData) > 64 {
			return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
		}
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, string(jsonData)))
		rows = append(rows, row)
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &numericKeyboard, nil
}

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
func (t *TgBot) sendMessage(message string, chatId int64, replyId *int, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ParseMode = "markdown"
	if replyId != nil {
		msg.ReplyToMessageID = *replyId
	}
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	// Reply to message
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (t *TgBot) checkIsUserRegistered(tgID string) bool {
	// Check if user is registered
	if !t.dc.IsInDutyList(tgID) {
		messageText := "Привет.\n" +
			"Это бот команды DSO.\n\n" +
			"*Вы не зарегестрированы.*\n\n" +
			"Используйте команду */register* для того, чтобы уведомить администраторов, о новом участнике.\n\n"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return false
	}
	return true
}

func genHelpCmdText(commands []botCommand) string {
	var cmdList string
	for i, cmd := range commands {
		var argList string
		if cmd.command.args != nil {
			argList = fmt.Sprintf("*Возможные значения аргумента:*\n")
			for index, arg := range *cmd.command.args {
				argList += fmt.Sprintf("*%s*: *%s* %q\n",
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
		cmdList += fmt.Sprintf("*%d*: */%s*%s - %s\n%s",
			i+1,
			cmd.command.name,
			argType,
			cmd.description,
			argList)
	}
	return cmdList
}

// Check if we have date in the command argument
func checkArgHasDate(arg string) (time.Time, error) {
	tn := time.Time{}
	s := strings.Split(arg, " ")
	if len(s) == 2 {
		var err error
		tn, err = time.Parse(botDataShort1, s[1])
		if err != nil {
			tn, err = time.Parse(botDataShort2, s[1])
			if err != nil {
				tn, err = time.Parse(botDataShort3, s[1])
				if err != nil {
					return tn, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
						"Доступны следующие форматы:\n"+
						"*%q*\n"+
						"*%q*\n"+
						"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
				}
			}
		}
	}
	return tn, nil
}

// Check if we have two dates in the command argument
func checkArgIsOffDutyRange(arg string) ([]time.Time, error) {
	var timeRange []time.Time
	dates := strings.Split(arg, "-")
	if len(dates) == 2 {
		for _, date := range dates {
			//var err error
			parsedTime, err := time.Parse(botDataShort1, date)
			if err != nil {
				parsedTime, err = time.Parse(botDataShort2, date)
				if err != nil {
					parsedTime, err = time.Parse(botDataShort3, date)
					if err != nil {
						return timeRange, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
							"Доступны следующие форматы:\n"+
							"*%q*\n"+
							"*%q*\n"+
							"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
					}
					timeRange = append(timeRange, parsedTime)
				}
				timeRange = append(timeRange, parsedTime)
			}
			timeRange = append(timeRange, parsedTime)
		}
		// If valid - return true
		return timeRange, nil
	}
	return timeRange, fmt.Errorf("формат аргумента должен быть: " +
		"*DDMMYYYY-DDMMYYYY* (период _'от-до'_ через дефис)")
}

// Generate keyboard with available args
func genArgsKeyboard(bc *botCommands, command tCmd) [][]tgbotapi.KeyboardButton {
	var rows [][]tgbotapi.KeyboardButton
	for _, cmd := range bc.commands {
		if cmd.command.name == command {
			for _, arg := range *cmd.command.args {
				row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("/%s %s",
					cmd.command.name, arg.name)))
				rows = append(rows, row)
			}
		}
	}
	return rows
}

// Generate keyboard with men on-duty indexes
func genIndexKeyboard(dm *[]data.DutyMan, cm callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, v := range *dm {
		cm.Answer = strconv.Itoa(i) // Save current index to data
		jsonData, err := json.Marshal(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
		}
		// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
		if len(jsonData) > 64 {
			return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
		}
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d. %s (%s)",
			i+1, v.FullName, v.UserName), string(jsonData)))
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := json.Marshal(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := json.Marshal(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &numericKeyboard, nil
}

// Generate keyboard with edit-duty data
func genEditDutyKeyboard(dm *[]data.DutyMan, cm callbackMessage) (*[][]tgbotapi.InlineKeyboardButton, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	// Iterate over all duty men
	for manIndex, man := range *dm {
		jsonData, err := marshalCallbackDataForEditDuty(cm, manIndex, 0)
		if err != nil {
			return nil, err
		}
		// Add leftmost button to hold man name
		var keyboardButtons []tgbotapi.InlineKeyboardButton
		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d. %s",
				manIndex+1,
				man.FullName),
				string(jsonData)))
		// Iterate over currently supported duty types
		for _, dt := range data.DutyTypes {
			for dutyIndex, d := range man.DutyType {
				if dt == d.Type {
					// Generate jsonData with current man's duty type state (false/true)
					jsonData, err := marshalCallbackDataForEditDuty(cm, manIndex, dutyIndex, d.Enabled)
					if err != nil {
						return nil, err
					}
					// Generate correct buttons based on current duty type state
					if d.Enabled {
						keyboardButtons = append(keyboardButtons,
							tgbotapi.NewInlineKeyboardButtonData("✅", string(jsonData)))
					} else {
						keyboardButtons = append(keyboardButtons,
							tgbotapi.NewInlineKeyboardButtonData("❌", string(jsonData)))
					}
				}
			}
		}
		// Check if keyboard is generated correctly
		if len(keyboardButtons) == 1 {
			return nil, fmt.Errorf("unable to generate keyboard buttons for: *@%s*", man.FullName)
		}
		row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := json.Marshal(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := json.Marshal(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	return &rows, nil
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

// Generate final message for user after he is hit "ok" button at inline keyboard and delete keyboard with message
func (t *TgBot) delInlineKeyboardWithMessage(messageText string, chatId int64, messageId int) {
	if err := t.sendMessage(messageText,
		chatId,
		&messageId,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID,
		t.update.CallbackQuery.Message.MessageID)
	_, err := t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}
}

// Return string with duty types data for specified man
func typesOfDuties(m *data.DutyMan) string {
	if len(m.DutyType) == 0 {
		return "❗️"
	}
	var list string
	for _, dt := range m.DutyType {
		list += fmt.Sprintf("%s ", dt.Name)
	}
	return strings.Trim(list, " ")
}

// Return json marshaled object for callback message data
func marshalCallbackDataForEditDuty(cm callbackMessage, manIndex int, buttonIndex int, enabled ...bool) ([]byte, error) {
	// Generate callback data
	// Format: 'manIndex-buttonIndex-Answer'
	cm.Answer = strconv.Itoa(manIndex)                         // Save current man index to data
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
