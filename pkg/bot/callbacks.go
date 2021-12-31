package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (t *TgBot) callbackRegister(answer string, chatId int64, userId int64) {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
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
		// Add user to duty list
		t.dc.AddManOnDuty(uFullName, uTgID)
		// Save new data
		_, err := t.dc.SaveMenList()
		if err != nil {
			log.Printf("can't save men list: %v", err)
		} else {
			// Send message to admins
			err = t.sendMessageToAdmins(fmt.Sprintf("Пользователь @%s успешно добавлен", uTgID))
			if err != nil {
				log.Printf("unable to send message admins group: %v", err)
			}
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

func (t *TgBot) callbackUnregister(answer string, chatId int64, userId int64) {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uTgID := u.User.UserName

	// Generate answer to user who was requested access
	var msg tgbotapi.MessageConfig
	if answer == inlineKeyboardYes {
		err := t.dc.DeleteManOnDuty(uTgID)
		if err != nil {
			msg = tgbotapi.NewMessage(chatId,
				fmt.Sprintf("Возникла ошибка при попытке произвести выход: %s", err))
		} else {
			msg = tgbotapi.NewMessage(chatId, "Выход произведен успешно")
			// Save new data
			_, err := t.dc.SaveMenList()
			if err != nil {
				log.Printf("can't save men list: %v", err)
			} else {
				// Send message to admins
				err = t.sendMessageToAdmins(fmt.Sprintf("Пользователь @%s произвел выход", uTgID))
				if err != nil {
					log.Printf("unable to send message admins group: %v", err)
				}
			}
		}
	} else {
		msg = tgbotapi.NewMessage(chatId, "Вы отменили выход")
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
