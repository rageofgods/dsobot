package bot

import (
	"bytes"
	"dso_bot/pkg/data"
	"encoding/csv"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

// handle '/help' command
func (t *TgBot) adminHandleHelp(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	// Create help message
	commands := t.AdminBotCommands().commands
	cmdList := genHelpCmdText(commands)
	messageText := "Доступны следующие команды администрирования:\n\n" +
		cmdList
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/list' command
func (t *TgBot) adminHandleList(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here
	var listActive string
	var listPassive string
	// Get menOnDuty list
	menData := t.dc.DutyMenData()

	// Generate returned string
	var indexActive int  // Index for Active list men
	var indexPassive int // Index for passive list men
	for _, v := range *menData {
		if v.Enabled {
			indexActive++
			listActive += fmt.Sprintf("*%d*: %s *(@%s)*\n`[%s]`\n",
				indexActive,
				v.CustomName,
				v.UserName,
				typesOfDuties(&v))
		} else {
			indexPassive++
			listPassive += fmt.Sprintf("*%d*: %s *(@%s)*\n`[%s]`\n",
				indexPassive,
				v.CustomName,
				v.UserName,
				typesOfDuties(&v))
		}
	}

	if indexActive == 0 {
		listActive = "*-*"
	} else if indexPassive == 0 {
		listPassive = "*-*"
	}
	messageText := fmt.Sprintf("*Список дежурных:*\n_Активные:_\n%s\n_Неактивные:_\n%s",
		listActive,
		listPassive)
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Parent function for handling args commands
func (t *TgBot) adminHandleRollout(cmdArgs string, update *tgbotapi.Update) {
	abc := t.AdminBotCommands()
	var isArgValid bool
	for _, cmd := range abc.commands {
		if cmd.command.args != nil && cmd.command.name == botCmdRollout {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					go arg.handleFunc(cmdArgs, update)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			messageText := fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(abc, botCmdRollout)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			numericKeyboard.Selective = true
			messageText := "Необходимо указать аргумент"
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				numericKeyboard); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	}
}

// handle '/showoffduty' command
func (t *TgBot) adminHandleShowOffDuty(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	men := t.dc.DutyMenData()
	var msgText string
	var isOffDutyFound bool
	for _, man := range *men {
		offduty, err := t.dc.ShowOffDutyForMan(man.UserName)
		if err != nil {
			messageText := fmt.Sprintf("%v", err)
			if err := t.sendMessage(messageText,
				update.Message.Chat.ID,
				&update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return
		}

		if len(*offduty) == 0 {
			continue
		}
		isOffDutyFound = true

		msgText += fmt.Sprintf("Нерабочие периоды для *%s* (*@%s*):\n", man.CustomName, man.UserName)
		for i, od := range *offduty {
			msgText += fmt.Sprintf("*%d.* Начало: %q - Конец: %q\n", i+1, od.OffDutyStart, od.OffDutyEnd)
		}
		msgText += "\n"
	}

	if !isOffDutyFound {
		messageText := "Нерабочие периоды не найдены"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	messageText := msgText
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/reindex' command
func (t *TgBot) adminHandleReindex(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	// Check if we are still editing tmpData at another function call
	if t.checkTmpDutyMenDataIsEditing(update.Message.From.ID, update) {
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleReindex,
	}

	men := t.dc.DutyMenData()
	if len(*men) == 0 {
		messageText := "Дежурных не найдено"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	numericKeyboard, err := genIndexKeyboard(men, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextAdminHandleReindex
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/enable' command
func (t *TgBot) adminHandleEnable(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	// Check if we are still editing tmpData at another function call
	if t.checkTmpDutyMenDataIsEditing(update.Message.From.ID, update) {
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleEnable,
	}

	men := t.dc.DutyMenData(false) // Get only passive men list
	if len(*men) == 0 {
		messageText := "Неактивных дежурных не найдено"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	numericKeyboard, err := genIndexKeyboard(men, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextAdminHandleEnable
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/disable' command
func (t *TgBot) adminHandleDisable(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	// Check if we are still editing tmpData at another function call
	if t.checkTmpDutyMenDataIsEditing(update.Message.From.ID, update) {
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleDisable,
	}

	men := t.dc.DutyMenData(true) // Get only active men list
	if len(*men) == 0 {
		messageText := "Активных дежурных не найдено"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	numericKeyboard, err := genIndexKeyboard(men, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	messageText := msgTextAdminHandleDisable
	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/editduty' command
func (t *TgBot) adminHandleEditDutyType(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	// Check if we are still editing tmpData at another function call
	if t.checkTmpDutyMenDataIsEditing(update.Message.From.ID, update) {
		return
	}

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleEditDuty,
	}

	men := t.dc.DutyMenData()
	if len(*men) == 0 {
		messageText := "Дежурных не найдено"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	rows, err := genEditDutyKeyboard(men, *callbackData)
	if err != nil {
		if err := t.sendMessage("Не удалось создать клавиатуру для отображения списка дежурных",
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		log.Printf("unable to generate new inline keyboard: %v", err)
		return
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(*rows...)

	messageText := msgTextAdminHandleEditDuty

	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/announce' command
func (t *TgBot) adminHandleAnnounce(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     update.Message.From.ID,
		ChatId:     update.Message.Chat.ID,
		MessageId:  update.Message.MessageID,
		FromHandle: callbackHandleAnnounce,
	}

	if len(t.settings.JoinedGroups) == 0 {
		messageText := "Бот не добавлен ни в одну пользовательскую группу"
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	rows, err := genAnnounceKeyboard(t.settings.JoinedGroups, *callbackData)
	if err != nil {
		if err := t.sendMessage("Не удалось создать клавиатуру для отображения списка групп добавленных чатов",
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		log.Printf("unable to generate new inline keyboard: %v", err)
		return
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)

	messageText := msgTextAdminHandleAnnounce

	if err := t.sendMessage(messageText,
		update.Message.Chat.ID,
		&update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/show_month' command
func (t *TgBot) adminHandleShowMonthDuty(cmdArgs string, update *tgbotapi.Update) {
	log.Println(cmdArgs) // Ignore arg here

	_, lastMonthDay, err := data.FirstLastMonthDay(1)
	if err != nil {
		messageText := err.Error()
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}

	// Generate header based on days count for current month
	header := []string{"Имя"}
	for i := 1; i <= lastMonthDay.Day(); i++ {
		header = append(header, strconv.Itoa(i))
	}

	var menData [][]string
	menData = append(menData, header)
	nwdDays, err := data.NwdEventsForCurMonth()
	if err != nil {
		log.Printf("unable to get non-working day data: %v", err)
		return
	}
	for _, man := range *t.dc.DutyMenData(true) {
		// Get man month duty dates
		dutyDates, err := t.dc.ManDutiesList(man.UserName, data.OnDutyTag)
		if err != nil {
			continue
		}

		manData := []string{man.CustomName}
		for i := 1; i <= lastMonthDay.Day(); i++ {
			var dayString string
			// Iterate over all man duty dates for current month
			for _, dd := range *dutyDates {
				// If man duty date is equal with current month day - add it to data slice
				if dd.Day() == i {
					// Mark duty day
					dayString = "\U0001F7E9"
					break
				} else {
					for _, n := range nwdDays {
						if n == i {
							// Mark nwd day
							dayString = "\U0001F7EB"
							break
						} else {
							// Mark free of duty day
							dayString = "⬜"
						}
					}
				}
			}
			manData = append(manData, dayString)
		}
		menData = append(menData, manData)
	}

	// Create buffer, generate csv
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	if err := w.WriteAll(menData); err != nil {
		messageText := fmt.Sprintf("Не удалось сгенерировать файл с отчетом дежурств: %v", err)
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}

	// Upload file to Telegram and sand message to user
	photoFileBytes := tgbotapi.FileBytes{
		Name:  fmt.Sprintf("Duties-%s-%d.csv", lastMonthDay.Month(), lastMonthDay.Year()),
		Bytes: buf.Bytes(),
	}
	if _, err := t.bot.Send(tgbotapi.NewDocument(t.adminGroupId, photoFileBytes)); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}
