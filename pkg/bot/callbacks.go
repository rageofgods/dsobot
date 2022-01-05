package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"time"
)

func (t *TgBot) callbackRegister(answer string, chatId int64, userId int64, messageId int) {
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
		msg.ReplyToMessageID = messageId
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
		msg.ReplyToMessageID = messageId
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

func (t *TgBot) callbackUnregister(answer string, chatId int64, userId int64, messageId int) {
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
			msg.ReplyToMessageID = messageId
		} else {
			msg = tgbotapi.NewMessage(chatId, "Выход произведен успешно")
			msg.ReplyToMessageID = messageId
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
		msg.ReplyToMessageID = messageId
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

func (t *TgBot) callbackDeleteOffDuty(answer string, chatId int64, userId int64, messageId int) error {
	var msg tgbotapi.MessageConfig
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

	// Get slice with off-duty data
	offduty, err := t.dc.ShowOffDutyForMan(uTgID)
	// Converting answer to integer value
	a, err := strconv.Atoi(answer)
	if err != nil {
		return fmt.Errorf("ошибка конвертации строки в число: %v", err)
	}

	// Converting date string to time.Time
	stime, err := time.Parse(data.DateShortSaveData, (*offduty)[a].OffDutyStart)
	if err != nil {
		return fmt.Errorf("ошибка конвертации даты начала нерабочего периода: %v", err)
	}
	etime, err := time.Parse(data.DateShortSaveData, (*offduty)[a].OffDutyEnd)
	if err != nil {
		return fmt.Errorf("ошибка конвертации даты конца нерабочего периода: %v", err)
	}

	// Delete calendar events
	err = t.dc.DeleteOffDutyEvents(uFullName, stime, etime)
	if err != nil {
		return fmt.Errorf("ошибка удаления события нерабочего периода: %v", err)
	}

	// Delete saved data
	t.dc.DeleteOffDutyFromMan(uTgID, a)
	_, err = t.dc.SaveMenList()
	if err != nil {
		return fmt.Errorf("ошибка сохранения данных: %v", err)
	}

	msg.Text = "Событие успешно удалено"
	msg.ReplyToMessageID = messageId
	msg.ChatID = chatId
	// Send a message to user who was request access.
	if _, err := t.bot.Send(msg); err != nil {
		log.Printf("unable to send message to user who was requested an access: %v", err)
	}

	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID, t.update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}
	return nil
}
