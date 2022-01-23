package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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
