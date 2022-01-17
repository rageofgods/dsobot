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

func (t *TgBot) checkIsUserRegistered(tgID string, update *tgbotapi.Update) bool {
	// Check if user is registered
	if !t.dc.IsInDutyList(tgID) {
		messageText := "Привет.\n" +
			"Это бот команды DSO.\n\n" +
			"*Вы не зарегестрированы.*\n\n" +
			"Используйте команду */register* для того, чтобы уведомить администраторов, о новом участнике.\n\n"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
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
		// Check if dates is in future
		for _, v := range timeRange {
			if v.Before(time.Now()) {
				return nil, fmt.Errorf("указанные даты не должны находится в прошлом: %v",
					v.Format(botDataShort3))
			}
		}
		// Check if dates is on valid order (first must be older than second)
		if timeRange[1].Before(timeRange[0]) {
			return nil, fmt.Errorf("дата %v должна быть старше, чем %v",
				timeRange[1].Format(botDataShort3),
				timeRange[0].Format(botDataShort3))
		}
		// If valid - return true
		return timeRange, nil
	}
	return nil, fmt.Errorf("формат аргумента должен быть: " +
		"*DDMMYYYY-DDMMYYYY* (период _'от-до'_ через дефис)")
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

func (t *TgBot) userHandleRegisterHelper(messageId int, update *tgbotapi.Update) {
	// Deleting register request message
	del := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, messageId)
	if _, err := t.bot.Request(del); err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}
	// Check if user is already registered
	if t.dc.IsInDutyList(update.Message.From.UserName) {
		messageText := "Вы уже зарегестрированы.\n" +
			"Используйте команду */unregister* для того, чтобы исключить себя из списка участников."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleRegisterHelper,
	}
	callbackDataNo := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleRegisterHelper,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := fmt.Sprintf("Проверьте ваши данные перед отправкой"+
		" запроса на согласование администраторам:\n\n*%s (@%s)*\n\nПродолжить?",
		update.Message.Text, update.Message.From.UserName)
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	// Save user data to process later in callback
	t.addTmpRegisterDataForUser(update.Message.From.ID, update.Message.Text, update)
}

func (t *TgBot) tmpRegisterDataForUser(userId int64) (string, error) {
	for _, v := range t.tmpData.tmpRegisterData {
		if v.userId == userId {
			return v.data, nil
		}
	}
	return "", fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpRegisterDataForUser(userId int64, name string, update *tgbotapi.Update) {
	var isUserIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpRegisterData {
		if v.userId == userId {
			t.tmpData.tmpRegisterData[i].data = name
			isUserIdFound = true
		}
	}
	if isUserIdFound {
		return
	} else {
		// If it's a fresh new data
		tmpCustomName := tmpRegisterData{userId: update.Message.From.ID, data: name}
		t.tmpData.tmpRegisterData = append(t.tmpData.tmpRegisterData, tmpCustomName)
	}
}

func (t *TgBot) tmpDutyManDataForUser(userId int64) ([]data.DutyMan, error) {
	for _, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			if v.data != nil {
				return v.data, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpDutyManDataForUser(userId int64, man data.DutyMan) {
	var isUserIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			t.tmpData.tmpDutyManData[i].data = append(t.tmpData.tmpDutyManData[i].data, man)
			isUserIdFound = true
		}
	}
	if isUserIdFound {
		return
	} else {
		// If it's a fresh new data
		var tmp []data.DutyMan
		tmp = append(tmp, man)
		tmpNewMan := tmpDutyManData{userId: userId, data: tmp}
		t.tmpData.tmpDutyManData = append(t.tmpData.tmpDutyManData, tmpNewMan)
	}
}

func (t *TgBot) clearTmpDutyManDataForUser(userId int64) {
	// CLear user temp data
	for i, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			t.tmpData.tmpDutyManData[i].data = nil
		}
	}
}

// Return true if tmpData is still in use by another call
func (t *TgBot) checkTmpDutyMenDataIsEditing(userId int64, update *tgbotapi.Update) bool {
	// If we got error here we can safely continue
	// Because tmpData is empty
	// If we get err == nil - some other function is still running
	if _, err := t.tmpDutyManDataForUser(userId); err == nil {
		messageText := "Вы уже работаете с данными дежурных. Для того, чтобы продолжить, пожалуйста " +
			"сохраните или отмените работу с текущими данными."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return true
	}
	return false
}

func (t *TgBot) botAddedToGroup(title string, id int64) error {
	group := &data.JoinedGroup{Id: id, Title: title}
	t.settings.JoinedGroups = append(t.settings.JoinedGroups, *group)
	if err := t.dc.SaveBotSettings(&t.settings); err != nil {
		return fmt.Errorf("unable to save bot settings: %v", err)
	}
	return nil
}

func (t *TgBot) botRemovedFromGroup(id int64) error {
	for i, v := range t.settings.JoinedGroups {
		if v.Id == id {
			// Remove founded group id from settings
			t.settings.JoinedGroups = append(t.settings.JoinedGroups[:i], t.settings.JoinedGroups[i+1:]...)
			if err := t.dc.SaveBotSettings(&t.settings); err != nil {
				return fmt.Errorf("unable to save bot settings: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("group id is not found in bot settings data")
}
