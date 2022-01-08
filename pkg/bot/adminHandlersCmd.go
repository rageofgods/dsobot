package bot

import (
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
	t.msg.Text = "Доступны следующие команды администрирования:\n\n" +
		cmdList
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}

// handle '/list' command
func (t *TgBot) adminHandleList(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	var list string
	// Get menOnDuty list
	menData := t.dc.DutyMenData()

	// Generate returned string
	for i, v := range *menData {
		list += fmt.Sprintf("*%d*: %s (*@%s*)\n", i+1, v.Name, v.TgID)
	}
	t.msg.Text = fmt.Sprintf("*Список дежурных:*\n%s", list)
	t.msg.ReplyToMessageID = t.update.Message.MessageID
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
					arg.handleFunc(cmdArgs)
					isArgValid = true
				}
			}
		}
	}
	// If provided argument is missing or invalid show error to user
	if !isArgValid {
		if cmdArgs != "" {
			t.msg.Text = fmt.Sprintf("Неверный аргумент - %q", cmdArgs)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		} else {
			// Show keyboard with available args
			rows := genArgsKeyboard(abc, botCmdRollout)
			var numericKeyboard = tgbotapi.NewOneTimeReplyKeyboard(rows...)
			t.msg.Text = "Необходимо указать аргумент"
			t.msg.ReplyMarkup = numericKeyboard
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		}
	}
}

// handle '/showoffduty' command
func (t *TgBot) adminHandleShowOffDuty(arg string) {
	arg = "" // Ignore cmdArgs

	men := t.dc.DutyMenData()
	var msgText string
	for _, man := range *men {
		offduty, err := t.dc.ShowOffDutyForMan(man.TgID)
		if err != nil {
			t.msg.Text = fmt.Sprintf("%v", err)
			t.msg.ReplyToMessageID = t.update.Message.MessageID
			return
		}

		if len(*offduty) == 0 {
			continue
		}

		msgText += fmt.Sprintf("Нерабочие периоды для *%s* (*@%s*):\n", man.Name, man.TgID)
		for i, od := range *offduty {
			msgText += fmt.Sprintf("*%d.* Начало: %q - Конец: %q\n", i+1, od.OffDutyStart, od.OffDutyEnd)
		}
		msgText += "\n"
	}

	t.msg.Text = msgText
	t.msg.ReplyToMessageID = t.update.Message.MessageID
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
	numericKeyboard, err := genIndexKeyboard(men, *callbackData)
	if err != nil {
		log.Printf("unable to generate new inline keyboard: %v", err)
	}

	t.msg.ReplyMarkup = numericKeyboard
	t.msg.Text = msgTextAdminHandleReindex
	t.msg.ReplyToMessageID = t.update.Message.MessageID
}
