package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
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
		commands := t.UserBotCommands().commands
		cmdList := genHelpCmdText(commands)
		msgText := "*Запрошенный доступ был согласован.*\n\n" +
			"Вам доступны следующие команды управления:\n" + cmdList
		msg = tgbotapi.NewMessage(chatId, msgText)
		msg.ReplyToMessageID = messageId
		msg.ParseMode = "markdown"
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
			// Save new data
			_, err := t.dc.SaveMenList()
			if err != nil {
				msg = tgbotapi.NewMessage(chatId, fmt.Sprintf("Не удалось сохранить данные: %v", err))
				msg.ReplyToMessageID = messageId
				log.Printf("can't save men list: %v", err)
			} else {
				// Generate user message
				msg = tgbotapi.NewMessage(chatId, "Выход произведен успешно")
				msg.ReplyToMessageID = messageId
				// Send message to admins
				err = t.sendMessageToAdmins(fmt.Sprintf("Пользователь *@%s* произвел выход", uTgID))
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
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID,
		t.update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}
	return nil
}

func (t *TgBot) callbackReindex(answer string, chatId int64, userId int64, messageId int) error {
	userId = 0 // userId is ignored here
	// Get current men data
	dutyMen := t.dc.DutyMenData()
	var msg tgbotapi.MessageConfig

	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// If we're still editing duty index
		if strings.Contains(t.update.CallbackQuery.Message.Text, msgTextAdminHandleReindex) {
			// If all buttons with men was pressed
			if len(*dutyMen) == len(*t.tmpData) {
				// Generate returned string
				var list string
				list = "*Новый порядок дежурных:*\n"
				for i, v := range *t.tmpData {
					list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.Name, v.TgID)
				}
				list += "\nСохранить?"

				editedMessage := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
					t.update.CallbackQuery.Message.MessageID, list, *t.update.CallbackQuery.Message.ReplyMarkup)
				editedMessage.ParseMode = "markdown"
				// Change original message
				_, err := t.bot.Request(editedMessage)
				if err != nil {
					log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
				}

			} else { // If some buttons with men wasn't pressed
				// Append absent men to tmpData
				for _, dMan := range *dutyMen {
					var manFound bool
					for _, dTmpMan := range *t.tmpData {
						if dMan.TgID == dTmpMan.TgID {
							manFound = true
						}
					}
					if !manFound {
						dMan.Index = len(*t.tmpData) + 1 // Generate correct new man index value
						*t.tmpData = append(*t.tmpData, dMan)
					}
				}
				// Generate returned string
				var list string
				list = "*Новый порядок дежурных:*\n"
				for i, v := range *t.tmpData {
					list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.Name, v.TgID)
				}
				list += "\nСохранить?"

				// Get last row of current keyboard (with yes/no buttons)
				yesNoKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					t.update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard[len(
						t.update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard)-1])

				// Generate new keyboard with final message
				editedMessage := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
					t.update.CallbackQuery.Message.MessageID, list, yesNoKeyboard)
				editedMessage.ParseMode = "markdown"
				// Change original message
				_, err := t.bot.Request(editedMessage)
				if err != nil {
					log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
				}
			}
			if _, err := t.bot.Send(msg); err != nil {
				log.Printf("unable to send message to user who was requested an access: %v", err)
			}
		} else { // New duty list is reviewed, and we want to save it
			_, err = t.dc.SaveMenList(t.tmpData)
			if err != nil {
				return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
			}
			msg.Text = "Новый порядок дежурных успешно сохранен"
			msg.ReplyToMessageID = messageId
			msg.ChatID = chatId
			// Send a message to user who was request access.
			if _, err := t.bot.Send(msg); err != nil {
				log.Printf("unable to send message to user who was requested an access: %v", err)
			}
			// Deleting message with keyboard
			del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID,
				t.update.CallbackQuery.Message.MessageID)
			_, err := t.bot.Request(del)
			if err != nil {
				log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
			}
			// Clearing index
			t.tmpData = new([]data.DutyMan)
		}
	case inlineKeyboardNo:
		msg.Text = "Редактирование списка дежурных отменено"
		msg.ReplyToMessageID = messageId
		msg.ChatID = chatId
		// Send a message to user who was request access.
		if _, err := t.bot.Send(msg); err != nil {
			log.Printf("unable to send message to user who was requested an access: %v", err)
		}
		// Deleting access request message in admin group
		del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID,
			t.update.CallbackQuery.Message.MessageID)
		_, err := t.bot.Request(del)
		if err != nil {
			log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
		}
		// Clearing index
		t.tmpData = new([]data.DutyMan)
	default:
		// Append selected man to a new slice of DutyMan
		for _, man := range *dutyMen {
			// Find dutyMan for reindex
			if man.Index == answerIndex+1 {
				man.Index = len(*t.tmpData) + 1 // Generate correct new man index value
				*t.tmpData = append(*t.tmpData, man)
			}
		}

		// Get current keyboard
		curCallbackKeyboard := t.update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard
		// Create returned keyboard
		var newCallbackKeyboard [][]tgbotapi.InlineKeyboardButton
		// Go through all element of current keyboard
		for index, button := range curCallbackKeyboard {
			var message callbackMessage
			err := json.Unmarshal([]byte(*button[0].CallbackData), &message)
			if err != nil {
				log.Printf("Can't unmarshal data json: %v", err)
				continue
			}
			// If we found a button which was pressed
			if message.Answer == answer {
				// Delete current index (button) from the rows
				newCallbackKeyboard = append(curCallbackKeyboard[:index], curCallbackKeyboard[index+1:]...)
			}
		}
		// Generate returned string
		var list string
		list = msgTextAdminHandleReindex
		list += "\n\n"
		for i, v := range *t.tmpData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.Name, v.TgID)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
			t.update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"
		//changeMsg := tgbotapi.NewEditMessageReplyMarkup(t.update.CallbackQuery.Message.Chat.ID,
		//	t.update.CallbackQuery.Message.MessageID,
		//	tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))

		// Change keyboard
		_, err := t.bot.Request(changeMsg)
		if err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}
