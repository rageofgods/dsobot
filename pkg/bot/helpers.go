package bot

import (
	"encoding/json"
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
		//log.Printf("handleRegister jsonNo size is greater then 64b: %v", len(jsonNo))
		//return
	} else if len(jsonYes) > 64 {
		return nil, fmt.Errorf("jsonYes size is greater then 64b: %v", len(jsonNo))
		//log.Printf("jsonYes size is greater then 64b: %v", len(jsonNo))
		//return
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

// Send provided message to admins Telegram group
func (t *TgBot) sendMessageToAdmins(message string) error {
	msg := tgbotapi.NewMessage(t.adminGroupId, message)
	// Send a message to user who was request access.
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (t *TgBot) checkIsUserRegistered(tgID string) bool {
	// Generate commands list with descriptions
	var cmdList string
	for i, cmd := range t.BotCommands().commands {
		cmdList += fmt.Sprintf("%d: */%s* - %s\n", i+1, cmd.command, cmd.description)
	}

	// Check if user is registered
	if !t.dc.IsInDutyList(tgID) {
		t.msg.Text = "Вы не зарегестрированы.\n" +
			"Используйте команду */register* для того, чтобы уведомить администраторов, о новом участнике.\n\n" +
			"После согласования, вам будут доступны следующие команды:\n" +
			cmdList
		t.msg.ReplyToMessageID = t.update.Message.MessageID
		return false
	}
	return true
}
