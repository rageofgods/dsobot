package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (t *TgBot) callbackRegister(answer string, chatId int64, userId int64) {
	// Get requested user info
	u, err := t.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId}})
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uTgID := u.User.UserName
	uFirstName := u.User.FirstName
	uLastName := u.User.LastName

	// Generate correct username
	uFullName := genUserFullName(uFirstName, uLastName)

	// Generate answer to user who was requested access
	var msg tgbotapi.MessageConfig
	if answer == inlineKeyboardYes {
		msg = tgbotapi.NewMessage(chatId, "Запрошенный доступ был согласован.")
		t.dc.AddManOnDuty(uFullName, uTgID) // Add user to duty list
		_, err := t.dc.SaveMenList()        // Save new data
		if err != nil {
			log.Printf("cant' save men list: %v", err)
		}
	} else {
		msg = tgbotapi.NewMessage(chatId, "Доступ не согласован.")
	}

	// Send a message to user who was request access.
	if _, err := t.bot.Send(msg); err != nil {
		log.Printf("unable to send message to user who was requested an access: %v", err)
	}

	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID, t.update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}
}
