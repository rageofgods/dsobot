package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"log"
)

func (t *TgBot) botAddedToGroup(title string, id int64) error {
	// Check if we already have group with this id
	for _, v := range t.settings.JoinedGroups {
		if v.Id == id {
			return fmt.Errorf("this group id (%d) is already in the list", id)
		}
	}
	group := &data.JoinedGroup{Id: id, Title: title}
	t.settings.JoinedGroups = append(t.settings.JoinedGroups, *group)
	if err := t.dc.SaveBotSettings(&t.settings); err != nil {
		return fmt.Errorf("unable to save bot settings: %v", err)
	}
	return nil
}

func (t *TgBot) botRemovedFromGroup(id int64) error {
	for i, v := range t.settings.JoinedGroups {
		if v.Id == id {
			// Remove founded group id from settings
			t.settings.JoinedGroups = append(t.settings.JoinedGroups[:i], t.settings.JoinedGroups[i+1:]...)
			if err := t.dc.SaveBotSettings(&t.settings); err != nil {
				return fmt.Errorf("unable to save bot settings: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("group id is not found in bot settings data")
}

func (t *TgBot) botCheckVersion(version string, build string) {
	// Check is version was updated
	if version == t.settings.Version {
		// Send message to admin group about current running bot build version
		messageText := fmt.Sprintf("⚠️ *%s (@%s)* был перезапущен без обновления версии.\n\n"+
			"*Возможный крэш?*\n\n_версия_: %q\n_билд_: %q",
			t.bot.Self.FirstName,
			t.bot.Self.UserName,
			version,
			build)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	} else {
		// Set and save new version info
		t.settings.Version = version
		if err := t.dc.SaveBotSettings(&t.settings); err != nil {
			log.Printf("%v", err)
		}

		// Send message to admin group about current running bot build version
		messageText := fmt.Sprintf("✅ *%s (@%s)* был обновлен до новой версии.\n\n"+
			"_новая версия_: %q\n_новый билд_: %q",
			t.bot.Self.FirstName,
			t.bot.Self.UserName,
			version,
			build)
		if err := t.sendMessage(messageText,
			t.adminGroupId,
			nil,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
	}
}
