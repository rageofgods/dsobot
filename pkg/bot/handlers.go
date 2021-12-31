package bot

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// handle '/start' command
func (t *TgBot) handleStart() {
	// Generate commands list with descriptions
	var cmdList string
	for i, cmd := range t.BotCommands().commands {
		cmdList += fmt.Sprintf("%d: */%s* - %s\n", i+1, cmd.command, cmd.description)
	}

	// Check if user is registered
	if !t.dc.IsInDutyList(t.update.Message.Chat.UserName) {
		t.msg.Text = "Вы не зарегестрированы.\n" +
			"Используйте команду */register* для того, чтобы уведомить администраторов, о новом участнике.\n\n" +
			"После согласования, вам будут доступны следующие команды:\n" +
			cmdList
		return
	}

	// Create welcome message
	t.msg.Text = "Привет, это Telegram бот команды DSO.\n" +
		"Используйте команду */register* для того, чтобы уведомить администраторов, о новом участнике.\n\n" +
		"После согласования, вам будут доступны следующие команды:\n" +
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

	// Generate jsons for data
	jsonYes, err := json.Marshal(callbackDataYes)
	if err != nil {
		log.Println(err)
	}
	jsonNo, err := json.Marshal(callbackDataNo)
	if err != nil {
		log.Println(err)
	}

	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonNo) > 64 {
		log.Printf("handleRegister jsonNo size is greater then 64b: %v", len(jsonNo))
		return
	} else if len(jsonYes) > 64 {
		log.Printf("handleRegister jsonNo size is greater then 64b: %v", len(jsonNo))
		return
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", string(jsonYes)),
			tgbotapi.NewInlineKeyboardButtonData("No", string(jsonNo)),
		),
	)

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

// handle unknown command
func (t *TgBot) handleDefault() {
	t.msg.Text = "Команда не найдена"
}
