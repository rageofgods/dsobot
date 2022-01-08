package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

// handle '/start' command
func (t *TgBot) handleStart(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	commands := t.UserBotCommands().commands
	cmdList := genHelpCmdText(commands)

	// Check if user is registered
	t.msg.Text = "*Вы уже зарегестрированы.*\n\n" +
		"Вам доступны следующие команды:\n" +
		cmdList
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// handle '/help' command
func (t *TgBot) handleHelp(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	commands := t.UserBotCommands().commands
	cmdList := genHelpCmdText(commands)

	// Check if user is registered
	t.msg.Text = "Вам доступны следующие команды управления:\n" +
		cmdList
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// Register new user as DSO team member
func (t *TgBot) handleRegister(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already registered
	if t.dc.IsInDutyList(t.update.Message.From.UserName) {
		t.msg.Text = "Вы уже зарегестрированы.\n" +
			"Используйте команду */unregister* для того, чтобы исключить себя из списка участников."
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}

	// Send info to user
	t.msg.Text = "Запрос на добавление отправлен администраторам.\n" +
		"По факту согласования вам придет уведомление.\n"
	t.msg.ReplyToMessageID = t.update.Message.MessageID
	_, err := t.bot.Send(t.msg)
	if err != nil {
		log.Println(err)
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleRegister,
	}
	callbackDataNo := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleRegister,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	// Create human-readable variables
	uTgID := t.update.Message.From.UserName
	uFirstName := t.update.Message.From.FirstName
	uLastName := t.update.Message.From.LastName

	// Generate correct username
	uFullName := genUserFullName(uFirstName, uLastName)

	// Send message to admins with inlineKeyboard question
	*t.msg = tgbotapi.NewMessage(t.adminGroupId,
		fmt.Sprintf("Новый запрос на добавление от пользователя:\n\n *@%s* (%s).\n\n Добавить?",
			uTgID,
			uFullName))
	t.msg.ReplyMarkup = numericKeyboard
	t.msg.ParseMode = "markdown"
}

func (t *TgBot) handleUnregister(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleUnregister,
	}
	callbackDataNo := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleUnregister,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	t.msg.ReplyMarkup = numericKeyboard
	t.msg.Text = fmt.Sprintf("Вы уверены, что хотите выйти?")
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// Parent function for handling args commands
func (t *TgBot) handleWhoIsOn(cmdArgs string) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	bc := t.UserBotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil {
			for _, arg := range *cmd.command.args {
				s := strings.Split(cmdArgs, " ")
				// Check if we have two arguments (type first, date second)
				if len(s) == 2 {
					if s[0] == string(arg.name) {
						// Check args for correct date format
						_, err := checkArgHasDate(cmdArgs)
						if err != nil {
							t.msg.Text = fmt.Sprintf("%v", err)
							t.msg.ReplyToMessageID = t.update.Message.MessageID
							return
						}
						// Run dedicated child argument function
						arg.handleFunc(cmdArgs)
						isArgValid = true
					}
				} else if len(s) == 1 {
					if s[0] == string(arg.name) {
						// Run dedicated child argument function
						arg.handleFunc(cmdArgs)
						isArgValid = true
					}
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			t.msg.Text = fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(bc, botCmdWhoIsOn)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			t.msg.Text = "Необходимо указать аргумент"
			t.msg.ReplyMarkup = numericKeyboard
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		}
	}
}

// handle '/addoffduty' command
func (t *TgBot) handleAddOffDuty(cmdArgs string) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	timeRange, err := checkArgIsOffDutyRange(cmdArgs)
	if err != nil {
		t.msg.Text = fmt.Sprintf("%v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	firstName := t.update.Message.From.FirstName
	lastName := t.update.Message.From.LastName
	fullName := genUserFullName(firstName, lastName)

	err = t.dc.CreateOffDutyEvents(fullName, timeRange[0], timeRange[1])
	if err != nil {
		t.msg.Text = fmt.Sprintf("Не удалось добавить событие: %v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	// Save off-duty data
	t.dc.AddOffDutyToMan(t.update.Message.From.UserName, timeRange[0], timeRange[1])
	_, err = t.dc.SaveMenList()
	if err != nil {
		t.msg.Text = fmt.Sprintf("Не удалось сохранить событие: %v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}
	t.msg.Text = "Событие добавлено успешно"
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// handle '/showofduty' command
func (t *TgBot) handleShowOffDuty(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	offduty, err := t.dc.ShowOffDutyForMan(t.update.Message.From.UserName)
	if err != nil {
		t.msg.Text = fmt.Sprintf("%v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}

	if len(*offduty) == 0 {
		t.msg.Text = "У вас нет нерабочих периодов"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}

	msgText := "Список нерабочих периодов:\n"
	for i, od := range *offduty {
		msgText += fmt.Sprintf("*%d.* Начало: %q - Конец: %q\n", i+1, od.OffDutyStart, od.OffDutyEnd)
	}
	t.msg.Text = msgText
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// handle '/deleteoffduty' command
func (t *TgBot) handleDeleteOffDuty(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	offduty, err := t.dc.ShowOffDutyForMan(t.update.Message.From.UserName)
	if err != nil {
		t.msg.Text = fmt.Sprintf("%v", err)
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}

	if len(*offduty) == 0 {
		t.msg.Text = "У вас нет нерабочих периодов"
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return
	}

	// Generate slice with off-duty periods
	var msgText []string
	for _, od := range *offduty {
		msgText = append(msgText, fmt.Sprintf("Начало: %q - Конец: %q",
			od.OffDutyStart,
			od.OffDutyEnd))
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		FromHandle: callbackHandleDeleteOffDuty,
	}

	numericKeyboard, err := genInlineOffDutyKeyboardWithData(msgText, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	t.msg.ReplyMarkup = numericKeyboard
	t.msg.Text = fmt.Sprintf("Выберите нерабочий период для удаления:")
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// handle '/showmyduties' command
func (t *TgBot) handleShowMy(cmdArgs string) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}
	// Check args for valid values
	bc := t.UserBotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc(cmdArgs)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			t.msg.Text = fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(bc, botCmdShowMy)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			t.msg.Text = "Необходимо указать аргумент"
			t.msg.ReplyMarkup = numericKeyboard
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		}
	}
}

// handle unknown command
func (t *TgBot) handleNotFound() {
	t.msg.Text = "Команда не найдена"
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}
