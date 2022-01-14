package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	deep "github.com/mitchellh/copystructure"
	"log"
	"strconv"
	"strings"
	"time"
)

func (t *TgBot) callbackRegister(answer string, chatId int64, userId int64, messageId int) error {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uUserName := u.User.UserName
	uFirstName := u.User.FirstName
	uLastName := u.User.LastName
	uTgID := u.User.ID

	// Generate correct username
	uFullName := genUserFullName(uFirstName, uLastName)

	// Generate answer to user who was requested access
	if answer == inlineKeyboardYes {
		commands := t.UserBotCommands().commands
		cmdList := genHelpCmdText(commands)
		messageText := "*Запрошенный доступ был согласован.*\n\n" +
			"Вам доступны следующие команды управления:\n" + cmdList
		if err := t.sendMessage(messageText,
			chatId,
			&messageId,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		// Add user to duty list
		t.dc.AddManOnDuty(uFullName, uUserName, uTgID)
		// Save new data
		_, err := t.dc.SaveMenList()
		if err != nil {
			log.Printf("can't save men list: %v", err)
		} else {
			// Send message to admins
			messageText := fmt.Sprintf("Пользователь *@%s* успешно добавлен", uUserName)
			if err := t.sendMessage(messageText,
				t.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	} else {
		messageText := "Доступ не согласован."
		if err := t.sendMessage(messageText,
			chatId,
			&messageId,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}

	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID, t.update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}

	return nil
}

func (t *TgBot) callbackUnregister(answer string, chatId int64, userId int64, messageId int) error {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uTgID := u.User.UserName

	// Generate answer to user who was requested access
	if answer == inlineKeyboardYes {
		err := t.dc.DeleteManOnDuty(uTgID)
		if err != nil {
			messageText := fmt.Sprintf("Возникла ошибка при попытке произвести выход: %s", err)
			if err := t.sendMessage(messageText,
				chatId,
				&messageId,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Save new data
			_, err := t.dc.SaveMenList()
			if err != nil {
				messageText := fmt.Sprintf("Не удалось сохранить данные: %v", err)
				if err := t.sendMessage(messageText,
					chatId,
					&messageId,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				log.Printf("can't save men list: %v", err)
			} else {
				// Generate user message
				messageText := "Выход произведен успешно"
				if err := t.sendMessage(messageText,
					chatId,
					&messageId,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				// Send message to admins
				messageText = fmt.Sprintf("Пользователь *@%s* произвел выход", uTgID)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}
	} else {
		messageText := "Вы отменили выход"
		if err := t.sendMessage(messageText,
			chatId,
			&messageId,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID, t.update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}
	return nil
}

func (t *TgBot) callbackDeleteOffDuty(answer string, chatId int64, userId int64, messageId int) error {
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

	messageText := "Событие успешно удалено"
	if err := t.sendMessage(messageText,
		chatId,
		&messageId,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	// Deleting access request message
	del := tgbotapi.NewDeleteMessage(t.update.CallbackQuery.Message.Chat.ID,
		t.update.CallbackQuery.Message.MessageID)
	if _, err = t.bot.Request(del); err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}

	// Send message to admins about deleted event
	timeRageText := fmt.Sprintf("%s - %s",
		stime.Format(botDataShort3),
		etime.Format(botDataShort3))
	messageText = fmt.Sprintf("Пользователь *@%s* удалил нерабочий период:\n%s",
		t.update.CallbackQuery.From.UserName, timeRageText)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		nil,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	return nil
}

func (t *TgBot) callbackReindex(answer string, chatId int64, userId int64, messageId int) error {
	userId = 0 // userId is ignored here
	// Get current men data
	dutyMen := t.dc.DutyMenData()

	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// If we're still editing duty index
		if strings.Contains(t.update.CallbackQuery.Message.Text, msgTextAdminHandleReindex) {
			// Append absent men to tmpData
			for _, dMan := range *dutyMen {
				var manFound bool
				for _, dTmpMan := range *t.tmpData {
					if dMan.UserName == dTmpMan.UserName {
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
				list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
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
		} else { // New duty list is reviewed, and we want to save it
			_, err = t.dc.SaveMenList(t.tmpData)
			if err != nil {
				return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
			}
			messageText := "Новый порядок дежурных успешно сохранен"
			// Send final message and remove inline keyboard
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
			// Clearing index
			t.tmpData = new([]data.DutyMan)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
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
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
			t.update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		_, err := t.bot.Request(changeMsg)
		if err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackEnable(answer string, chatId int64, userId int64, messageId int) error {
	userId = 0 // userId is ignored here
	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// If we're still editing duty index
		if strings.Contains(t.update.CallbackQuery.Message.Text, msgTextAdminHandleEnable) {
			// If all buttons with men was pressed

			// Get current men data
			dutyMen := t.dc.DutyMenData(true)
			// Generate returned string
			var list string
			list = "*Новый список активных дежурных:*\n"
			var index int // Counter for list men index
			for _, v := range *dutyMen {
				index++
				list += fmt.Sprintf("*%d*: %s (*@%s*)\n", index, v.FullName, v.UserName)
			}
			for _, v := range *t.tmpData {
				index++
				list += fmt.Sprintf("*%d*: *%s* (*@%s*)\n", index, v.FullName, v.UserName)
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
		} else { // New duty list is reviewed, and we want to save it
			// Get current men data
			dutyMen := t.dc.DutyMenData()
			for i, dMan := range *dutyMen {
				for _, dTmpMan := range *t.tmpData {
					if dMan.TgID == dTmpMan.TgID {
						(*dutyMen)[i].Enabled = true
					}
				}
			}
			// Save data
			_, err = t.dc.SaveMenList(dutyMen)
			if err != nil {
				return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
			}

			messageText := "Новый список активных дежурных успешно сохранен"
			// Send final message and remove inline keyboard
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
			// Clearing index
			t.tmpData = new([]data.DutyMan)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка активных дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
		// Clearing index
		t.tmpData = new([]data.DutyMan)
	default:
		// Get passive men data
		dutyMen := t.dc.DutyMenData(false)

		// Append selected man to a new slice of DutyMan
		for i, man := range *dutyMen {
			// Found right man for clicked button index and append him to temporary data list
			if i == answerIndex {
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
		list = msgTextAdminHandleEnable
		list += "\n\n"
		for i, v := range *t.tmpData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
			t.update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		_, err := t.bot.Request(changeMsg)
		if err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackDisable(answer string, chatId int64, userId int64, messageId int) error {
	userId = 0 // userId is ignored here
	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// If we're still editing duty index
		if strings.Contains(t.update.CallbackQuery.Message.Text, msgTextAdminHandleDisable) {
			// If all buttons with men was pressed

			// Get current men data
			dutyMen := t.dc.DutyMenData(false)
			// Generate returned string
			var list string
			list = "*Новый список неактивных дежурных:*\n"
			var index int // Counter for list men index
			for _, v := range *dutyMen {
				index++
				list += fmt.Sprintf("*%d*: %s (*@%s*)\n", index, v.FullName, v.UserName)
			}
			for _, v := range *t.tmpData {
				index++
				list += fmt.Sprintf("*%d*: *%s* (*@%s*)\n", index, v.FullName, v.UserName)
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
		} else { // New duty list is reviewed, and we want to save it
			// Get current men data
			dutyMen := t.dc.DutyMenData()
			for i, dMan := range *dutyMen {
				for _, dTmpMan := range *t.tmpData {
					if dMan.TgID == dTmpMan.TgID {
						(*dutyMen)[i].Enabled = false
					}
				}
			}
			// Save data
			if _, err := t.dc.SaveMenList(dutyMen); err != nil {
				return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
			}

			messageText := "Новый список неактивных дежурных успешно сохранен"
			// Send final message and remove inline keyboard
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
			// Clearing index
			t.tmpData = new([]data.DutyMan)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка неактивных дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
		// Clearing index
		t.tmpData = new([]data.DutyMan)
	default:
		// Get passive men data
		dutyMen := t.dc.DutyMenData(true)

		// Append selected man to a new slice of DutyMan
		for i, man := range *dutyMen {
			// Found right man for clicked button index and append him to temporary data list
			if i == answerIndex {
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
		list = msgTextAdminHandleDisable
		list += "\n\n"
		for i, v := range *t.tmpData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(t.update.CallbackQuery.Message.Chat.ID,
			t.update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		_, err := t.bot.Request(changeMsg)
		if err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackEditDuty(answer string, chatId int64, userId int64, messageId int) error {
	// If we just spawn inline keyboard let's load our temporary data
	if *t.tmpData == nil {
		// Get current men data
		origData := *t.dc.DutyMenData()
		// Deep copy original data
		d, err := deep.Copy(origData)
		if err != nil {
			return err
		}
		// Assign deep copied data to tmpData
		*t.tmpData = d.([]data.DutyMan)
	}

	switch answer {
	case inlineKeyboardYes:
		// Save data
		if _, err := t.dc.SaveMenList(t.tmpData); err != nil {
			return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
		}
		messageText := "Список типов дежурств успешно сохранен"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
		// Clearing index
		t.tmpData = new([]data.DutyMan)
	case inlineKeyboardNo:
		messageText := "Редактирование списка типов дежурств отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId)
		// Clearing index
		t.tmpData = new([]data.DutyMan)
	default:
		// Format: 'manIndex-buttonIndex-Answer'
		splitAnswer := strings.Split(answer, "-")
		// Check only duty type button (separated by '-')
		if len(splitAnswer) == 3 {
			manIndex := splitAnswer[0] // Save man index
			mi, err := strconv.Atoi(manIndex)
			if err != nil {
				return err
			}

			buttonIndex := splitAnswer[1] // Save button index
			bi, err := strconv.Atoi(buttonIndex)
			if err != nil {
				return err
			}

			buttonState := splitAnswer[2] // Save button state

			// Edit tmp dutyMan data
			if buttonState == inlineKeyboardEditDutyYes {
				(*t.tmpData)[mi].DutyType[bi].Enabled = false
			} else {
				(*t.tmpData)[mi].DutyType[bi].Enabled = true
			}

			// Create returned data (without data)
			callbackData := &callbackMessage{
				UserId:     userId,
				ChatId:     chatId,
				MessageId:  messageId,
				FromHandle: callbackHandleEditDuty,
			}
			// Generate edited keyboard
			rows, err := genEditDutyKeyboard(t.tmpData, *callbackData)
			if err != nil {
				if err := t.sendMessage("Не удалось создать клавиатуру для отображения списка дежурных",
					t.update.Message.Chat.ID,
					&t.update.Message.MessageID,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				log.Printf("unable to generate new inline keyboard: %v", err)
				return err
			}

			// Create edited message (with correct keyboard)
			changeMsg := tgbotapi.NewEditMessageReplyMarkup(t.update.CallbackQuery.Message.Chat.ID,
				t.update.CallbackQuery.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup(*rows...))

			// Change keyboard
			if _, err := t.bot.Request(changeMsg); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		}
	}
	return nil
}
