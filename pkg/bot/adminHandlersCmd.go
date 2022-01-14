package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// handle '/help' command
func (t *TgBot) adminHandleHelp(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Create help message
	commands := t.AdminBotCommands().commands
	cmdList := genHelpCmdText(commands)
	messageText := "Доступны следующие команды администрирования:\n\n" +
		cmdList
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/list' command
func (t *TgBot) adminHandleList(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
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
			listActive += fmt.Sprintf("*%d*: %s (*@%s*) [[%s]]\n",
				indexActive,
				v.FullName,
				v.UserName,
				typesOfDuties(&v))
		} else {
			indexPassive++
			listPassive += fmt.Sprintf("*%d*: %s (*@%s*) [[%s]]\n",
				indexPassive,
				v.FullName,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// Parent function for handling args commands
func (t *TgBot) adminHandleRollout(cmdArgs string) {
	abc := t.AdminBotCommands()
	var isArgValid bool
	for _, cmd := range abc.commands {
		if cmd.command.args != nil && cmd.command.name == botCmdRollout {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					go arg.handleFunc(cmdArgs)
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
				t.update.Message.Chat.ID,
				&t.update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(abc, botCmdRollout)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			messageText := "Необходимо указать аргумент"
			if err := t.sendMessage(messageText,
				t.update.Message.Chat.ID,
				&t.update.Message.MessageID,
				numericKeyboard); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	}
}

// handle '/showoffduty' command
func (t *TgBot) adminHandleShowOffDuty(arg string) {
	arg = "" // Ignore cmdArgs

	men := t.dc.DutyMenData()
	var msgText string
	var isOffDutyFound bool
	for _, man := range *men {
		offduty, err := t.dc.ShowOffDutyForMan(man.UserName)
		if err != nil {
			messageText := fmt.Sprintf("%v", err)
			if err := t.sendMessage(messageText,
				t.update.Message.Chat.ID,
				&t.update.Message.MessageID,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
			return
		}

		if len(*offduty) == 0 {
			continue
		}
		isOffDutyFound = true

		msgText += fmt.Sprintf("Нерабочие периоды для *%s* (*@%s*):\n", man.FullName, man.UserName)
		for i, od := range *offduty {
			msgText += fmt.Sprintf("*%d.* Начало: %q - Конец: %q\n", i+1, od.OffDutyStart, od.OffDutyEnd)
		}
		msgText += "\n"
	}

	if !isOffDutyFound {
		messageText := "Нерабочие периоды не найдены"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	messageText := msgText
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/reindex' command
func (t *TgBot) adminHandleReindex(arg string) {
	arg = "" // Ignore cmdArgs

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		FromHandle: callbackHandleReindex,
	}

	men := t.dc.DutyMenData()
	if len(*men) == 0 {
		messageText := "Дежурных не найдено"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/enable' command
func (t *TgBot) adminHandleEnable(arg string) {
	arg = "" // Ignore cmdArgs

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		FromHandle: callbackHandleEnable,
	}

	men := t.dc.DutyMenData(false) // Get only passive men list
	if len(*men) == 0 {
		messageText := "Неактивных дежурных не найдено"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/disable' command
func (t *TgBot) adminHandleDisable(arg string) {
	arg = "" // Ignore cmdArgs

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		FromHandle: callbackHandleDisable,
	}

	men := t.dc.DutyMenData(true) // Get only active men list
	if len(*men) == 0 {
		messageText := "Активных дежурных не найдено"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
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
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}

// handle '/editduty' command
func (t *TgBot) adminHandleEditDutyType(arg string) {
	arg = "" // Ignore cmdArgs

	// Create returned data (without data)
	callbackData := &callbackMessage{
		UserId:     t.update.Message.From.ID,
		ChatId:     t.update.Message.Chat.ID,
		MessageId:  t.update.Message.MessageID,
		FromHandle: callbackHandleEditDuty,
	}

	men := t.dc.DutyMenData()
	if len(*men) == 0 {
		messageText := "Дежурных не найдено"
		if err := t.sendMessage(messageText,
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return
	}

	rows, err := genEditDutyKeyboard(men, *callbackData)
	if err != nil {
		if err := t.sendMessage("Не удалось создать клавиатуру для отображения списка дежурных",
			t.update.Message.Chat.ID,
			&t.update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		log.Printf("unable to generate new inline keyboard: %v", err)
		return
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(*rows...)

	messageText := msgTextAdminHandleEditDuty
	appendedText := "`| Имя дежурного |`"
	for _, dt := range data.DutyNames {
		appendedText += fmt.Sprintf("` %s |`", dt)
	}
	messageText += appendedText
	if err := t.sendMessage(messageText,
		t.update.Message.Chat.ID,
		&t.update.Message.MessageID,
		numericKeyboard); err != nil {
		log.Printf("unable to send message: %v", err)
	}
}
