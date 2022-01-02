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

	// Create welcome message
	var cmdList string
	for i, cmd := range t.BotCommands().commands {
		cmdList += fmt.Sprintf("%d: */%s* - %s\n", i+1, cmd.command.name, cmd.description)
	}

	// Check if user is registered
	t.msg.Text = "Вы уже зарегестрированы.\n\n" +
		"Вам доступны следующие команды:\n" +
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

// handle unknown command
func (t *TgBot) handleNotFound() {
	t.msg.Text = "Команда не найдена"
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}
