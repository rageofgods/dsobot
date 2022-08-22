package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

// handle '/start' command
func (t *TgBot) handleStart(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	commands := t.UserBotCommands().commands
	cmdList := genHelpCmdText(commands)

	// If user was already registered send a message
	messageText := "*Вы уже зарегестрированы.*\n\n" +
		"Вам доступны следующие команды:\n" +
		cmdList
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/help' command
func (t *TgBot) handleHelp(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	commands := t.UserBotCommands().commands
	cmdList := genHelpCmdText(commands)

	// Check if user is registered
	messageText := "Вам доступны следующие команды управления:\n" +
		cmdList
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Register new user as DSO team member
func (t *TgBot) handleRegister(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
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

	// Check if user have telegram id
	if update.Message.From.UserName == "" {
		messageText := "У вас отсутствует Telegram Username (@username)\n" +
			"Пожалуйста, укажите его в настройках вашего профиля Telegram"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Request and process user custom name
	// Send info to user
	// Generate correct username
	userFullName := genUserFullName(update.Message.From.FirstName, update.Message.From.LastName)
	msgText := msgTextUserHandleRegister + fmt.Sprintf("Эта информация будет использоваться"+
		" для корректного отображения имен участников т.к. ваши текущие данные из Telegram (%s) могут не "+
		"соответствовать реальным.", userFullName)
	if err := t.sendMessage(msgText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

func (t *TgBot) handleUnregister(_ string, update *tgbotapi.Update) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleUnregister,
	}
	callbackDataNo := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleUnregister,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := "Вы уверены, что хотите выйти?"
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Parent function for handling args commands
func (t *TgBot) handleWhoIsOnDuty(cmdArgs string, update *tgbotapi.Update) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	bc := t.UserBotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil && cmd.command.name == botCmdWhoIsOnDuty {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc(cmdArgs, update)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			messageText := fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(bc, botCmdWhoIsOnDuty)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			numericKeyboard.Selective = true
			messageText := "Необходимо указать аргумент"
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				numericKeyboard); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	}
}

// Parent function for handling args commands
func (t *TgBot) handleWhoIsOnValidation(cmdArgs string, update *tgbotapi.Update) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	bc := t.UserBotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil && cmd.command.name == botCmdWhoIsOnValidation {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc(cmdArgs, update)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			messageText := fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(bc, botCmdWhoIsOnValidation)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			numericKeyboard.Selective = true
			messageText := "Необходимо указать аргумент"
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				numericKeyboard); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	}
}

// handle '/addoffduty' command
func (t *TgBot) handleAddOffDuty(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	if _, err := t.tmpOffDutyDataForUser(update.Message.From.ID); err == nil {
		messageText := "Вы уже работаете с данными нерабочего периода"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleAddOffDuty,
	}

	numericKeyboard, err := genInlineCalendarKeyboard(time.Now(), *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextUserHandleAddOffDuty1
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/showofduty' command
func (t *TgBot) handleShowOffDuty(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	offduty, err := t.dc.ShowOffDutyForMan(update.Message.From.UserName)
	if err != nil {
		messageText := fmt.Sprintf("%v", err)
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	if len(*offduty) == 0 {
		messageText := "У вас нет нерабочих периодов"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	msgText := "Список нерабочих периодов:\n"
	for i, od := range *offduty {
		msgText += fmt.Sprintf("*%d.* Начало: %q - Конец: %q\n", i+1, od.OffDutyStart, od.OffDutyEnd)
	}
	messageText := msgText
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/deleteoffduty' command
func (t *TgBot) handleDeleteOffDuty(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}

	offduty, err := t.dc.ShowOffDutyForMan(update.Message.From.UserName)
	if err != nil {
		messageText := fmt.Sprintf("%v", err)
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	if len(*offduty) == 0 {
		messageText := "У вас нет нерабочих периодов"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
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
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleDeleteOffDuty,
	}

	numericKeyboard, err := genInlineOffDutyKeyboardWithData(msgText, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := "Выберите нерабочий период для удаления:"
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/showmyduties' command
func (t *TgBot) handleShowMy(cmdArgs string, update *tgbotapi.Update) {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(update.Message.From.UserName, update) {
		return
	}
	// Check args for valid values
	bc := t.UserBotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil && cmd.command.name == botCmdShowMy {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc(cmdArgs, update)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			messageText := fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(bc, botCmdShowMy)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			numericKeyboard.Selective = true
			messageText := "Необходимо указать аргумент"
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				numericKeyboard); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	}
}

// handle unknown command
func (t *TgBot) handleNotFound(update *tgbotapi.Update) {
	messageText := "Команда не найдена"
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/duties_csv' command
func (t *TgBot) handleShowMonthDuty(cmdArgs string, update *tgbotapi.Update) {
	t.adminHandleShowMonthDuty(cmdArgs, update)
}

// handle '/validation_csv' command
func (t *TgBot) handleShowMonthValidation(cmdArgs string, update *tgbotapi.Update) {
	t.adminHandleShowMonthValidation(cmdArgs, update)
}

// handle '/birthday' command
func (t *TgBot) handleBirthday(_ string, update *tgbotapi.Update) {
	messageText := msgTextUserHandleBirthday
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}
