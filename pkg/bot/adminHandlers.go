package bot

import (
	"fmt"
	"log"
)

// handle '/help' command
func (t *TgBot) adminHandleHelp(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	// Create help message
	var cmdList string
	for i, cmd := range t.AdminBotCommands().commands {
		var argList string
		if cmd.command.args != nil {
			argList = fmt.Sprintf("Возможные значения аргументов:\n")
			for index, arg := range *cmd.command.args {
				argList += fmt.Sprintf("%d: *%s* %q\n",
					index+1,
					arg.name,
					arg.description)
			}
		}
		// Append <argument> suffix to command help is any arguments was found
		var argType string
		if argList != "" {
			argType = " <аргумент>"
		}
		// Generate lit of commands
		cmdList += fmt.Sprintf("%d: */%s*%s - %s\n%s",
			i+1,
			cmd.command.name,
			argType,
			cmd.description,
			argList)
	}

	// Check if user is registered
	t.msg.Text = "Доступны следующие команды администрирования:\n\n" +
		cmdList
}

// handle '/list' command
func (t *TgBot) adminHandleList(cmdArgs string) {
	cmdArgs = "" // Ignore cmdArgs
	var list string
	// Get menOnDuty list
	menList, err := t.dc.ShowMenOnDutyList()
	if err != nil {
		log.Printf("Возникла ошибка при загрузке: %v", err)
		err := t.sendMessageToAdmins(fmt.Sprintf("Возникла ошибка при загрузке: %v", err))
		if err != nil {
			log.Printf("unable to send message to admins: %v", err)
		}
		return
	}
	// Generate returned string
	for _, i := range menList {
		list += fmt.Sprintf("%s\n", i)
	}
	t.msg.Text = fmt.Sprintf("Список дежурных: \n%s", list)
}

// Parent function for handling args commands
func (t *TgBot) adminHandleRollout(cmdArgs string) {
	abc := t.AdminBotCommands()
	var isArgValid bool
	for _, cmd := range abc.commands {
		if cmd.command.args != nil {
			for _, arg := range *cmd.command.args {
				// Check if user command arg is supported
				if cmdArgs == string(arg.name) {
					// Run dedicated child argument function
					arg.handleFunc()
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
			t.msg.Text = fmt.Sprintf("Необходимо указать аргумент")
			t.msg.ReplyToMessageID = t.update.Message.MessageID
		}
	}
}

// Handle 'duty' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutDuty() {
	t.msg.Text = "DUTY"
}

// Handle 'validation' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutValidation() {
	t.msg.Text = "Validation"
}

// Handle 'nwd' user arg for 'rollout' command
func (t *TgBot) adminHandleRolloutNonWorkingDay() {
	t.msg.Text = "NWD"
}
