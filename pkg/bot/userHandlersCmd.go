package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// handle '/start' command
func (t *TgBot) handleStart(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.From.UserName) {
		return
	}

	cmdList := t.genHelpCmdText()

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

	cmdList := t.genHelpCmdText()

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
	bc := t.BotCommands()
	var isArgValid bool
	for _, cmd := range bc.commands {
		if cmd.command.args != nil {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc()
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
			t.msg.Text = fmt.Sprintf("Необходимо указать аргумент")
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		}
	}
}

// handle unknown command
func (t *TgBot) handleNotFound() {
	t.msg.Text = "Команда не найдена"
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}
