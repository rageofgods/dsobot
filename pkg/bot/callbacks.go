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

func (t *TgBot) callbackRegister(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
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

		// Get saved user data
		userNameSurname, err := t.tmpRegisterDataForUser(userId)
		if err != nil {
			return err
		}

		// Add user to duty list
		t.dc.AddManOnDuty(uFullName, uUserName, userNameSurname, uTgID)
		// Save new data
		if _, err := t.dc.SaveMenList(); err != nil {
			log.Printf("can't save men list: %v", err)
		} else {
			// Send message to admins
			messageText := fmt.Sprintf("Пользователь %s *(@%s)* успешно добавлен", userNameSurname, uUserName)
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
	del := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}

	return nil
}

func (t *TgBot) callbackUnregister(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uTgID := u.User.UserName

	// Get current men data
	dutyMen := t.dc.DutyMenData()
	// Get Custom Name for deleted user
	var uCustomName string
	for _, v := range *dutyMen {
		if v.TgID == userId {
			uCustomName = v.CustomName
		}
	}
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
				messageText = fmt.Sprintf("Пользователь %s *(@%s)* произвел выход", uCustomName, uTgID)
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
	del := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}
	return nil
}

func (t *TgBot) callbackDeleteOffDuty(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Create human-readable variables
	uTgID := u.User.UserName

	// Get slice with off-duty data
	offduty, err := t.dc.ShowOffDutyForMan(uTgID)
	if err != nil {
		return err
	}
	// Converting answer to integer value
	a, err := strconv.Atoi(answer)
	if err != nil {
		return fmt.Errorf("ошибка конвертации строки в число: %v", err)
	}

	// Converting date string to time.Time
	stime, err := time.ParseInLocation(data.DateShortSaveData, (*offduty)[a].OffDutyStart, loc)
	if err != nil {
		return fmt.Errorf("ошибка конвертации даты начала нерабочего периода: %v", err)
	}
	etime, err := time.ParseInLocation(data.DateShortSaveData, (*offduty)[a].OffDutyEnd, loc)
	if err != nil {
		return fmt.Errorf("ошибка конвертации даты конца нерабочего периода: %v", err)
	}

	// Delete calendar events
	err = t.dc.DeleteOffDutyEvents(uTgID, stime, etime)
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
	del := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID)
	if _, err = t.bot.Request(del); err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}

	// Send message to admins about deleted event
	timeRangeText := fmt.Sprintf("%s - %s",
		stime.Format(botDataShort3),
		etime.Format(botDataShort3))
	messageText = fmt.Sprintf("Пользователь *@%s* удалил нерабочий период:\n%s",
		update.CallbackQuery.From.UserName, timeRangeText)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		nil,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	// Recreate calendar duty events from current date if deleted duty was landed at this month
	t.updateOnDutyEvents(&stime, update.CallbackQuery.From.UserName, timeRangeText)

	return nil
}

func (t *TgBot) callbackReindex(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// Get current men data
	dutyMen := t.dc.DutyMenData()

	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// Get saved user data
		_, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			messageText := "Нет данных для сохранения"
			// Send final message and remove inline keyboard
			if err := t.sendMessage(messageText, chatId, &messageId, nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return nil
		}
		// If we're still editing duty index
		if strings.Contains(update.CallbackQuery.Message.Text, msgTextAdminHandleReindex) {
			// Append absent men to tmpDutyData
			for _, dMan := range *dutyMen {
				var manFound bool
				tmpDutyData, err := t.tmpDutyManDataForUser(userId)
				if err != nil {
					return err
				}
				for _, dTmpMan := range tmpDutyData {
					if dMan.UserName == dTmpMan.UserName {
						manFound = true
					}
				}
				if !manFound {
					dMan.Index = len(tmpDutyData) + 1        // Generate correct new man index value
					t.addTmpDutyManDataForUser(userId, dMan) // Append edited man to tmpDutyData
				}
			}
			// Generate returned string
			var list string
			tmpDutyData, err := t.tmpDutyManDataForUser(userId)
			if err != nil {
				return err
			}
			list = "*Новый порядок дежурных:*\n"
			for i, v := range tmpDutyData {
				list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.CustomName, v.UserName)
			}
			list += "\nСохранить?"

			// Get last row of current keyboard (with yes/no buttons)
			yesNoKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard[len(
					update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard)-1])

			// Generate new keyboard with final message
			editedMessage := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, list, yesNoKeyboard)
			editedMessage.ParseMode = "markdown"
			// Change original message
			if _, err := t.bot.Request(editedMessage); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		} else { // New duty list is reviewed, and we want to save it
			tmpDutyData, err := t.tmpDutyManDataForUser(userId)
			if err != nil {
				return err
			}
			if _, err = t.dc.SaveMenList(&tmpDutyData); err != nil {
				return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
			}
			messageText := "Новый порядок дежурных успешно сохранен"
			// Send final message and remove inline keyboard
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			// Clear tmp data
			t.clearTmpDutyManDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpDutyManDataForUser(userId)
	default:
		// Get current tmpDutyData (we can safety ignore error here)
		tmpDutyData, _ := t.tmpDutyManDataForUser(userId)

		// Append selected man to a new slice of DutyMan
		for i, man := range *dutyMen {
			// Find dutyMan for reindex
			if i == answerIndex {
				man.Index = len(tmpDutyData) + 1 // Generate correct new man index value
				t.addTmpDutyManDataForUser(userId, man)
				break
			}
		}

		// Get current keyboard
		curCallbackKeyboard := update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard
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
		// Get current tmpDutyData
		tmpDutyData, err = t.tmpDutyManDataForUser(userId)
		if err != nil {
			return err
		}
		for i, v := range tmpDutyData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.CustomName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackEnable(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	//userId = 0 // userId is ignored here
	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// Get saved user data
		tmpDutyData, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			messageText := "Нет данных для сохранения"
			// Send final message and remove inline keyboard
			if err := t.sendMessage(messageText, chatId, &messageId, nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return nil
		}
		// If we're still editing duty index
		if strings.Contains(update.CallbackQuery.Message.Text, msgTextAdminHandleEnable) {
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
			for _, v := range tmpDutyData {
				index++
				list += fmt.Sprintf("*%d*: *%s* (*@%s*)\n", index, v.FullName, v.UserName)
			}
			list += "\nСохранить?"

			// Get last row of current keyboard (with yes/no buttons)
			yesNoKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard[len(
					update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard)-1])

			// Generate new keyboard with final message
			editedMessage := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, list, yesNoKeyboard)
			editedMessage.ParseMode = "markdown"
			// Change original message
			if _, err := t.bot.Request(editedMessage); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		} else { // New duty list is reviewed, and we want to save it
			// Get current men data
			dutyMen := t.dc.DutyMenData()
			for i, dMan := range *dutyMen {
				for _, dTmpMan := range tmpDutyData {
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
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			// Clear tmp data
			t.clearTmpDutyManDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка активных дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpDutyManDataForUser(userId)
	default:
		// Get passive men data
		dutyMen := t.dc.DutyMenData(false)

		// Append selected man to a new slice of DutyMan
		for i, man := range *dutyMen {
			// Found right man for clicked button index and append him to temporary data list
			if i == answerIndex {
				t.addTmpDutyManDataForUser(userId, man)
				break
			}
		}

		// Get current keyboard
		curCallbackKeyboard := update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard
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
		// Get current tmpDutyData
		tmpDutyData, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			return err
		}
		for i, v := range tmpDutyData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackDisable(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	answerIndex, err := strconv.Atoi(answer)
	if err != nil {
		log.Printf("unable to convert string answer to integer: %v", err)
		return fmt.Errorf("unable to convert string answer to integer: %v", err)
	}

	switch answer {
	case inlineKeyboardYes:
		// Get saved user data
		tmpDutyData, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			messageText := "Нет данных для сохранения"
			// Send final message and remove inline keyboard
			if err := t.sendMessage(messageText, chatId, &messageId, nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return nil
		}
		// If we're still editing duty index
		if strings.Contains(update.CallbackQuery.Message.Text, msgTextAdminHandleDisable) {
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
			for _, v := range tmpDutyData {
				index++
				list += fmt.Sprintf("*%d*: *%s* (*@%s*)\n", index, v.FullName, v.UserName)
			}
			list += "\nСохранить?"

			// Get last row of current keyboard (with yes/no buttons)
			yesNoKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard[len(
					update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard)-1])

			// Generate new keyboard with final message
			editedMessage := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, list, yesNoKeyboard)
			editedMessage.ParseMode = "markdown"
			// Change original message
			if _, err := t.bot.Request(editedMessage); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		} else { // New duty list is reviewed, and we want to save it
			// Get current men data
			dutyMen := t.dc.DutyMenData()
			for i, dMan := range *dutyMen {
				for _, dTmpMan := range tmpDutyData {
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
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			// Clear tmp data
			t.clearTmpDutyManDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Редактирование списка неактивных дежурных отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpDutyManDataForUser(userId)
	default:
		// Get passive men data
		dutyMen := t.dc.DutyMenData(true)

		// Append selected man to a new slice of DutyMan
		for i, man := range *dutyMen {
			// Found right man for clicked button index and append him to temporary data list
			if i == answerIndex {
				t.addTmpDutyManDataForUser(userId, man)
				break
			}
		}

		// Get current keyboard
		curCallbackKeyboard := update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard
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
		// Get current tmpDutyData
		tmpDutyData, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			return err
		}
		for i, v := range tmpDutyData {
			list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.FullName, v.UserName)
		}

		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, list, tgbotapi.NewInlineKeyboardMarkup(newCallbackKeyboard...))
		changeMsg.ParseMode = "markdown"

		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackEditDuty(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// If we just spawn inline keyboard let's load our temporary data
	_, err := t.tmpDutyManDataForUser(userId)
	if err != nil {
		// Get current men data
		origData := *t.dc.DutyMenData()
		// Deep copy original data
		d, err := deep.Copy(origData)
		if err != nil {
			return err
		}
		// Assign deep copied data to tmpDutyData
		for _, man := range d.([]data.DutyMan) {
			t.addTmpDutyManDataForUser(userId, man)
		}
	}
	switch answer {
	case inlineKeyboardYes:
		tmpDutyData, err := t.tmpDutyManDataForUser(userId)
		if err != nil {
			return err
		}
		// Save data
		if _, err := t.dc.SaveMenList(&tmpDutyData); err != nil {
			return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
		}
		messageText := "Список типов дежурств успешно сохранен"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpDutyManDataForUser(userId)
	case inlineKeyboardNo:
		messageText := "Редактирование списка типов дежурств отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpDutyManDataForUser(userId)
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
			for i, v := range t.tmpData.tmpDutyManData {
				if v.userId == userId {
					if buttonState == inlineKeyboardEditDutyYes {
						t.tmpData.tmpDutyManData[i].data[mi].DutyType[bi].Enabled = false
					} else {
						t.tmpData.tmpDutyManData[i].data[mi].DutyType[bi].Enabled = true
					}
				}
			}

			// Create returned data (without data)
			callbackData := &callbackMessage{
				UserId:     userId,
				ChatId:     chatId,
				MessageId:  messageId,
				FromHandle: callbackHandleEditDuty,
			}
			// Generate edited keyboard
			tmpDutyData, err := t.tmpDutyManDataForUser(userId)
			if err != nil {
				return err
			}
			rows, err := genEditDutyKeyboard(&tmpDutyData, *callbackData)
			if err != nil {
				if err := t.sendMessage("Не удалось создать клавиатуру для отображения списка дежурных",
					update.Message.Chat.ID,
					&update.Message.MessageID,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				log.Printf("unable to generate new inline keyboard: %v", err)
				return err
			}

			// Create edited message (with correct keyboard)
			changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup(*rows...))

			// Change keyboard
			if _, err := t.bot.Request(changeMsg); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		}
	}
	return nil
}

func (t *TgBot) callbackRegisterHelper(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// Get requested user info
	u, err := t.getChatMember(userId, chatId)
	if err != nil {
		log.Printf("unable to get user info: %v", err)
	}

	// Generate answer to user who was requested access
	if answer == inlineKeyboardYes {
		// Send info to user
		messageText := "Запрос на добавление отправлен администраторам.\n" +
			"По факту согласования вам придет уведомление.\n"
		if err := t.sendMessage(messageText,
			chatId,
			&messageId,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}

		// Create returned data with Yes/No button
		callbackDataYes := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			Answer:     inlineKeyboardYes,
			FromHandle: callbackHandleRegister,
		}
		callbackDataNo := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			Answer:     inlineKeyboardNo,
			FromHandle: callbackHandleRegister,
		}

		numericKeyboard, err := genInlineYesNoKeyboardWithData(callbackDataYes, callbackDataNo)
		if err != nil {
			log.Printf("unable to generate new inline keyboard: %v", err)
		}

		// Create human-readable variables
		uUserName := u.User.UserName
		uFirstName := u.User.FirstName
		uLastName := u.User.LastName

		// Generate correct username
		uFullName := genUserFullName(uFirstName, uLastName)

		// Get saved user data
		userNameSurname, err := t.tmpRegisterDataForUser(userId)
		if err != nil {
			return err
		}

		// Send message to admins with inlineKeyboard question
		messageText = fmt.Sprintf("Новый запрос на добавление от пользователя:\n\n "+
			"*@%s* - %s (%s).\n\n Добавить?",
			uUserName,
			userNameSurname,
			uFullName)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			numericKeyboard); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		messageText := "Вы отменили регистрацию"
		if err := t.sendMessage(messageText,
			chatId,
			&messageId,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}

	// Deleting register request message
	del := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete admin group message with requested access: %v", err)
	}

	return nil
}

func (t *TgBot) callbackAnnounce(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// If we just spawn inline keyboard let's load our temporary data
	_, err := t.tmpAnnounceDataForUser(userId)
	if err != nil {
		// Get current men data
		origData := t.settings.JoinedGroups
		// Deep copy original data
		d, err := deep.Copy(origData)
		if err != nil {
			return err
		}
		// Assign deep copied data to tmpAnnounceData
		for _, group := range d.([]data.JoinedGroup) {
			t.addTmpAnnounceDataForUser(userId, group)
		}
	}
	switch answer {
	case inlineKeyboardYes:
		tmpAnnounceData, err := t.tmpAnnounceDataForUser(userId)
		if err != nil {
			return err
		}
		// Save data
		t.settings.JoinedGroups = tmpAnnounceData
		if err := t.dc.SaveBotSettings(&t.settings); err != nil {
			return fmt.Errorf("не удалось сохранить список дежурных: %v", err)
		}
		messageText := "Список анонсов для групп успешно сохранен"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpAnnounceDataForUser(userId)
	case inlineKeyboardNo:
		messageText := "Редактирование списка анонсов для групп отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpAnnounceDataForUser(userId)
	default:
		// Format: 'groupIndex-buttonIndex-Answer'
		splitAnswer := strings.Split(answer, "-")
		// Check only duty type button (separated by '-')
		if len(splitAnswer) == 3 {
			buttonIndex := splitAnswer[1] // Save button index
			bi, err := strconv.Atoi(buttonIndex)
			if err != nil {
				return err
			}

			buttonState := splitAnswer[2] // Save button state

			// Edit tmp Announce data
			for i, v := range t.tmpData.tmpJoinedGroupData {
				if v.userId == userId {
					if buttonState == inlineKeyboardEditDutyYes {
						t.tmpData.tmpJoinedGroupData[i].data[bi].Announce = false
					} else {
						t.tmpData.tmpJoinedGroupData[i].data[bi].Announce = true
					}
				}
			}

			// Create returned data (without data)
			callbackData := &callbackMessage{
				UserId:     userId,
				ChatId:     chatId,
				MessageId:  messageId,
				FromHandle: callbackHandleAnnounce,
			}
			// Generate edited keyboard
			tmpAnnounceData, err := t.tmpAnnounceDataForUser(userId)
			if err != nil {
				return err
			}
			rows, err := genAnnounceKeyboard(tmpAnnounceData, *callbackData)
			if err != nil {
				if err := t.sendMessage("Не удалось создать клавиатуру для отображения групп для анонса",
					update.Message.Chat.ID,
					&update.Message.MessageID,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				log.Printf("unable to generate new inline keyboard: %v", err)
				return err
			}

			// Create edited message (with correct keyboard)
			changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup(rows...))

			// Change keyboard
			if _, err := t.bot.Request(changeMsg); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		}
	}
	return nil
}

func (t *TgBot) callbackAddOffDuty(answer string, chatId int64, userId int64, messageId int, update *tgbotapi.Update) error {
	// Split answer for two string following - 'buttonType-currentDate'
	parsedAnswer := strings.Split(answer, "-")
	var answerButtonType, answerCurrentDate string
	if len(parsedAnswer) == 2 {
		answerButtonType = parsedAnswer[0]
		answerCurrentDate = parsedAnswer[1]
	} else if len(parsedAnswer) == 1 {
		answerButtonType = parsedAnswer[0]
	} else {
		return fmt.Errorf("returned callback data has wrong format")
	}

	switch answerButtonType {
	case inlineKeyboardYes:
		// Get saved user data
		dates, err := t.tmpOffDutyDataForUser(userId)
		// Check if we have enough data to add new offDuty period
		if err != nil || len(dates) != 2 {
			messageText := "⚠️ Недостаточно данных для сохранения.\nПожалуйста выберите начало и конец нерабочего периода"
			// Send final message and remove inline keyboard
			if err := t.sendMessage(messageText, chatId, &messageId, nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return nil
		} else {
			// Check if dates is in future
			for _, v := range dates {
				if v.Before(time.Now().Add(time.Hour * -24)) {
					messageText := fmt.Sprintf("⚠️ Указанные даты не должны находится в прошлом: %v",
						v.Format(botDataShort3))
					// Send final message and remove inline keyboard
					t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
					// Clear tmp data
					t.clearTmpOffDutyDataForUser(userId)
					return nil
				}
				// Check if dates is on valid order (first must be older than second)
				if dates[1].Before(dates[0]) {
					messageText := fmt.Sprintf("⚠️ Дата %v должна быть старше, чем %v",
						dates[1].Format(botDataShort3),
						dates[0].Format(botDataShort3))
					// Send final message and remove inline keyboard
					t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
					// Clear tmp data
					t.clearTmpOffDutyDataForUser(userId)
					return nil
				}
			}

			// Check if provided off-duty period is overlap with existing off-duty periods
			if _, err := t.isOffDutyDatesOverlapWithCurrent(dates[0], dates[1], chatId, userId, messageId); err != nil {
				// Send final message and remove inline keyboard
				t.delInlineKeyboardWithMessage(err.Error(), chatId, messageId, update)
				// Clear tmp data
				t.clearTmpOffDutyDataForUser(userId)
				return nil
			}

			err = t.dc.CreateOffDutyEvents(update.CallbackQuery.From.UserName, dates[0], dates[1])
			if err != nil {
				messageText := fmt.Sprintf("Не удалось добавить событие: %v", err)
				if err := t.sendMessage(messageText,
					chatId,
					&messageId,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				return fmt.Errorf("%v", err)
			}
			// Save off-duty data
			t.dc.AddOffDutyToMan(update.CallbackQuery.From.UserName, dates[0], dates[1])
			_, err = t.dc.SaveMenList()
			if err != nil {
				messageText := fmt.Sprintf("Не удалось сохранить событие: %v", err)
				if err := t.sendMessage(messageText,
					chatId,
					&messageId,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
				return fmt.Errorf("%v", err)
			}
			messageText := "Событие добавлено успешно"
			// Send final message and remove inline keyboard
			t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)

			// Send message to admins about added event
			timeRangeText := fmt.Sprintf("%s - %s",
				dates[0].Format(botDataShort3),
				dates[1].Format(botDataShort3))
			messageText = fmt.Sprintf("Пользователь *@%s* добавил новый нерабочий период:\n%s",
				update.CallbackQuery.From.UserName, timeRangeText)
			if err := t.sendMessage(messageText,
				t.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}

			// Recreate calendar duty event from current date if added duty in landed at this month
			t.updateOnDutyEvents(&dates[0], update.CallbackQuery.From.UserName, timeRangeText)

			// Clear tmp data
			t.clearTmpOffDutyDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Действие отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardNext:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleAddOffDuty,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, 1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	case inlineKeyboardPrev:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleAddOffDuty,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, -1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
		}
	case inlineKeyboardDate:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		parsedAnswerCurrentDate, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Check if user adding off-duty start period date (first stage)
		if strings.Contains(update.CallbackQuery.Message.Text, "чтобы добавить") {
			// Save current date at tmpData
			t.addTmpOffDutyDataForUser(userId, parsedAnswerCurrentDate)

			textMessage := fmt.Sprintf(msgTextUserHandleAddOffDuty2+"\n*%s* %s",
				msgTextUserHandleAddOffDutyStart,
				parsedAnswerCurrentDate.Format(botDataShort3))
			// Create edited message (with correct keyboard)
			changeMsg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, textMessage, *update.CallbackQuery.Message.ReplyMarkup)
			changeMsg.ParseMode = "markdown"

			// Change keyboard
			if _, err := t.bot.Request(changeMsg); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		} else {
			parsedAnswerCurrentDate, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
			if err != nil {
				return err
			}
			// Save current date at tmpData
			t.addTmpOffDutyDataForUser(userId, parsedAnswerCurrentDate)
			// Get current tmpData
			dates, err := t.tmpOffDutyDataForUser(userId)
			if err != nil {
				return err
			}
			textMessage := fmt.Sprintf("📅 Проверьте ваши данные перед сохранением нового нерабочего периода:"+
				"\n\n*%s* %s\n*%s* %s\n\nСохранить?",
				msgTextUserHandleAddOffDutyStart,
				dates[0].Format(botDataShort3),
				msgTextUserHandleAddOffDutyEnd,
				dates[1].Format(botDataShort3))
			// Get last row of current keyboard (with yes/no buttons)
			yesNoKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard[len(
					update.CallbackQuery.Message.ReplyMarkup.InlineKeyboard)-1])

			// Generate new keyboard with final message
			editedMessage := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID, textMessage, yesNoKeyboard)
			editedMessage.ParseMode = "markdown"
			// Change original message
			if _, err := t.bot.Request(editedMessage); err != nil {
				log.Printf("unable to change message with on-duty index inline keyboard: %v", err)
			}
		}
	}
	return nil
}

func (t *TgBot) callbackWhoIsOnDutyAtDate(answer string,
	chatId int64,
	userId int64,
	messageId int,
	update *tgbotapi.Update) error {
	// Split answer for two string following - 'buttonType-currentDate'
	parsedAnswer := strings.Split(answer, "-")
	var answerButtonType, answerCurrentDate string
	if len(parsedAnswer) == 2 {
		answerButtonType = parsedAnswer[0]
		answerCurrentDate = parsedAnswer[1]
	} else if len(parsedAnswer) == 1 {
		answerButtonType = parsedAnswer[0]
	} else {
		return fmt.Errorf("returned callback data has wrong format")
	}

	switch answerButtonType {
	case inlineKeyboardYes:
		// Get current tmpData
		date, err := t.tmpOffDutyDataForUser(userId)
		if err != nil {
			messageText := "⚠️ Пожалуйста выберите дату дежурства, чтобы продолжить"
			if err := t.sendMessage(messageText,
				chatId,
				&messageId,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
		if len(date) == 1 {
			// Get on-duty data
			man, err := t.dc.WhoIsOnDuty(&date[0], data.OnDutyTag)
			if err != nil {
				log.Printf("error in event creating: %v", err)
				messageText := "Дежурства не найдены."
				// Send final message and remove inline keyboard
				t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			} else {
				// Get data for all men
				dutyMen := t.dc.DutyMenData()
				// Generate returned string
				for _, v := range *dutyMen {
					if v.UserName == man {
						man = fmt.Sprintf("%s (*@%s*)", v.CustomName, v.UserName)
					}
				}
				messageText := fmt.Sprintf("Дежурный: %s", man)
				// Send final message and remove inline keyboard
				t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			}
			// Clear tmp data
			t.clearTmpOffDutyDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Действие отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardNext:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnDutyAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, 1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-duty index inline keyboard: %v", err)
		}
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardPrev:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnDutyAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, -1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-duty index inline keyboard: %v", err)
		}
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardDate:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Save current date at tmpData
		t.addTmpOffDutyDataForUser(userId, pt)
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnDutyAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt, *callbackData, pt.Day())
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-duty index inline keyboard: %v", err)
		}
	}
	return nil
}

func (t *TgBot) callbackWhoIsOnValidationAtDate(answer string,
	chatId int64,
	userId int64,
	messageId int,
	update *tgbotapi.Update) error {
	// Split answer for two string following - 'buttonType-currentDate'
	parsedAnswer := strings.Split(answer, "-")
	var answerButtonType, answerCurrentDate string
	if len(parsedAnswer) == 2 {
		answerButtonType = parsedAnswer[0]
		answerCurrentDate = parsedAnswer[1]
	} else if len(parsedAnswer) == 1 {
		answerButtonType = parsedAnswer[0]
	} else {
		return fmt.Errorf("returned callback data has wrong format")
	}

	switch answerButtonType {
	case inlineKeyboardYes:
		// Get current tmpData
		date, err := t.tmpOffDutyDataForUser(userId)
		if err != nil {
			messageText := "⚠️ Пожалуйста выберите дату валидации, чтобы продолжить"
			if err := t.sendMessage(messageText,
				chatId,
				&messageId,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
		if len(date) == 1 {
			// Get on-duty data
			man, err := t.dc.WhoIsOnDuty(&date[0], data.OnDutyTag)
			if err != nil {
				log.Printf("error in event creating: %v", err)
				messageText := "Валидации не найдены."
				// Send final message and remove inline keyboard
				t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			} else {
				// Get data for all men
				dutyMen := t.dc.DutyMenData()
				// Generate returned string
				for _, v := range *dutyMen {
					if v.UserName == man {
						man = fmt.Sprintf("%s (*@%s*)", v.CustomName, v.UserName)
					}
				}
				messageText := fmt.Sprintf("Валидирующий: %s", man)
				// Send final message and remove inline keyboard
				t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
			}
			// Clear tmp data
			t.clearTmpOffDutyDataForUser(userId)
		}
	case inlineKeyboardNo:
		messageText := "Действие отменено"
		// Send final message and remove inline keyboard
		t.delInlineKeyboardWithMessage(messageText, chatId, messageId, update)
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardNext:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnValidationAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, 1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-validation index inline keyboard: %v", err)
		}
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardPrev:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnValidationAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt.AddDate(0, -1, 0), *callbackData)
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-validation index inline keyboard: %v", err)
		}
		// Clear tmp data
		t.clearTmpOffDutyDataForUser(userId)
	case inlineKeyboardDate:
		loc, err := time.LoadLocation(data.TimeZone)
		if err != nil {
			return err
		}
		pt, err := time.ParseInLocation(botDataShort4, answerCurrentDate, loc)
		if err != nil {
			return err
		}
		// Save current date at tmpData
		t.addTmpOffDutyDataForUser(userId, pt)
		// Create returned data (without data)
		callbackData := &callbackMessage{
			UserId:     userId,
			ChatId:     chatId,
			MessageId:  messageId,
			FromHandle: callbackHandleWhoIsOnValidationAtDate,
		}
		inlineKeyboard, err := genInlineCalendarKeyboard(pt, *callbackData, pt.Day())
		if err != nil {
			return err
		}
		// Create edited message (with correct keyboard)
		changeMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID, *inlineKeyboard)
		// Change keyboard
		if _, err := t.bot.Request(changeMsg); err != nil {
			log.Printf("unable to change message with whoison-validation index inline keyboard: %v", err)
		}
	}
	return nil
}
