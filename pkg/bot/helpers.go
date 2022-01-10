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
