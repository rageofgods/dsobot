package bot

import (
	"fmt"
	"log"
)

// handle '/help' command
func (t *TgBot) adminHandleHelp() {
	// Create help message
	var cmdList string
	for i, cmd := range t.AdminBotCommands().commands {
		cmdList += fmt.Sprintf("%d: */%s* - %s\n", i+1, cmd.command, cmd.description)
	}

	// Check if user is registered
	t.msg.Text = "Доступны следующие команды администрирования:\n\n" +
		cmdList
}

// handle '/list' command
func (t *TgBot) adminHandleList() {
	var list string
	menList, err := t.dc.ShowMenOnDutyList()
	if err != nil {
		log.Printf("Возникла ошибка при загрузке: %v", err)
		err := t.sendMessageToAdmins(fmt.Sprintf("Возникла ошибка при загрузке: %v", err))
		if err != nil {
			log.Printf("unable to send message to admins: %v", err)
		}
		return
	}

	for _, i := range menList {
		list += fmt.Sprintf("%s\n", i)
	}
	t.msg.Text = fmt.Sprintf("Список дежурных: \n%s", list)
}
