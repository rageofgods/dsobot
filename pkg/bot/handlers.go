package bot

import (
	"fmt"
	"log"
)

// handle '/start' command
func (t *TgBot) handleStart() {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.Chat.UserName) {
		return
	}

	// Create welcome message
	var cmdList string
	for i, cmd := range t.BotCommands().commands {
		cmdList += fmt.Sprintf("%d: */%s* - %s\n", i+1, cmd.command, cmd.description)
	}

	// Check if user is registered
	t.msg.Text = "Вы уже зарегестрированы.\n\n" +
		"Вам доступны следующие команды:\n" +
		cmdList
}

// Register new user as DSO team member
func (t *TgBot) handleRegister() {
	// Check if user is already registered
	if t.dc.IsInDutyList(t.update.Message.Chat.UserName) {
		t.msg.Text = "Вы уже зарегестрированы.\n" +
			"Используйте команду */unregister* для того, чтобы исключить себя из списка участников."
		return
	}

	// Send info to user
	t.msg.Text = "Запрос на добавление отправлен администраторам.\n" +
		"По факту согласования вам придет уведомление.\n"
	_, err := t.bot.Send(t.msg)
	if err != nil {
		log.Println(err)
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleRegister,
	}
	callbackDataNo := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleRegister,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	// Create human-readable variables
	uTgID := t.update.Message.Chat.UserName
	uFirstName := t.update.Message.Chat.FirstName
	uLastName := t.update.Message.Chat.LastName

	// Generate correct username
	uFullName := genUserFullName(uFirstName, uLastName)

	t.msg.ReplyMarkup = numericKeyboard
	t.msg.Text = fmt.Sprintf("Новый запрос на добавление от пользователя:\n\n *@%s* (%s).\n\n Добавить?",
		uTgID,
		uFullName)
	t.msg.ChatID = t.adminGroupId
}

func (t *TgBot) handleUnregister() {
	// Check if user is already register. Return if it was.
	if !t.checkIsUserRegistered(t.update.Message.Chat.UserName) {
		return
	}

	// Create returned data with Yes/No button
	callbackDataYes := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		Answer:     inlineKeyboardYes,
		FromHandle: callbackHandleUnregister,
	}
	callbackDataNo := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		Answer:     inlineKeyboardNo,
		FromHandle: callbackHandleUnregister,
	}

	numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	t.msg.ReplyMarkup = numericKeyboard
	t.msg.Text = fmt.Sprintf("Вы уверены, что хотите выйти?")
}

// handle unknown command
func (t *TgBot) handleNotFound() {
	t.msg.Text = "Команда не найдена"
}
