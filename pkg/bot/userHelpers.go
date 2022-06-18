package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
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

func (t *TgBot) userHandleBirthdayHelper(messageId int, update *tgbotapi.Update) {
	loc, _ := time.LoadLocation(data.TimeZone)
	// Setup supported date layouts
	dateLayouts := []string{botDataShort3, botDataShort2}
	parsedData := time.Time{}
	// Check user provided date
	for _, v := range dateLayouts {
		var err error
		parsedData, err = time.ParseInLocation(v, update.Message.Text, loc)
		if err != nil {
			continue
		}
		break
	}
	// Report error if we are unable to parse user provided date
	if parsedData.IsZero() {
		log.Printf("userHandleBirthdayHelper: unable to parse user provided date: %s", update.Message.Text)
		messageText := "Неверный формат даты. Пожалуйста укажите дату в формате 'DD.MM.YYYY'"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Check user age (must be at least 18 years old)
	if parsedData.After(time.Now().AddDate(-18, 0, 0)) {
		messageText := "Чтобы продолжить, вам должно быть хотя бы 18 лет. Проверьте дату вашего рождения."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	// Add birthday to man
	t.dc.AddBirthdayToMan(update.Message.From.UserName, parsedData)
	// Save data
	_, err := t.dc.SaveMenList()
	if err != nil {
		log.Printf("userHandleBirthdayHelper: unable to save data %v", err)
		messageText := "Не удалось сохранить данные из-за внутренней ошибки."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	messageText := "Информация о вашем дне рожденья была успешно добавлена."
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	adminMessage := fmt.Sprintf("Пользователь *@%s* добавил дату рождения: *%s*",
		update.Message.From.UserName,
		update.Message.Text)
	if err := t.sendMessage(adminMessage, t.adminGroupId, nil, nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	// Deleting original bot message
	del := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, messageId)
	_, err = t.bot.Request(del)
	if err != nil {
		log.Printf("userHandleBirthdayHelper: unable to delete original bot message: %v", err)
	}
}
