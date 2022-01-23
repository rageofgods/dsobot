package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (t *TgBot) tmpRegisterDataForUser(userId int64) (string, error) {
	for _, v := range t.tmpData.tmpRegisterData {
		if v.userId == userId {
			return v.data, nil
		}
	}
	return "", fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpRegisterDataForUser(userId int64, name string, update *tgbotapi.Update) {
	var isUserIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpRegisterData {
		if v.userId == userId {
			t.tmpData.tmpRegisterData[i].data = name
			isUserIdFound = true
		}
	}
	if isUserIdFound {
		return
	} else {
		// If it's a fresh new data
		tmpCustomName := tmpRegisterData{userId: update.Message.From.ID, data: name}
		t.tmpData.tmpRegisterData = append(t.tmpData.tmpRegisterData, tmpCustomName)
	}
}

func (t *TgBot) tmpAnnounceDataForUser(userId int64) ([]data.JoinedGroup, error) {
	for _, v := range t.tmpData.tmpJoinedGroupData {
		if v.userId == userId {
			if v.data != nil {
				return v.data, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpAnnounceDataForUser(userId int64, group data.JoinedGroup) {
	var isGroupIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpJoinedGroupData {
		if v.userId == userId {
			t.tmpData.tmpJoinedGroupData[i].data = append(t.tmpData.tmpJoinedGroupData[i].data, group)
			isGroupIdFound = true
		}
	}
	if isGroupIdFound {
		return
	} else {
		// If it's a fresh new data
		var tmp []data.JoinedGroup
		tmp = append(tmp, group)
		tmpNewGroup := tmpJoinedGroupData{userId: userId, data: tmp}
		t.tmpData.tmpJoinedGroupData = append(t.tmpData.tmpJoinedGroupData, tmpNewGroup)
	}
}

func (t *TgBot) clearTmpAnnounceDataForUser(userId int64) {
	// CLear user temp data
	for i, v := range t.tmpData.tmpJoinedGroupData {
		if v.userId == userId {
			t.tmpData.tmpJoinedGroupData[i].data = nil
		}
	}
}

func (t *TgBot) tmpDutyManDataForUser(userId int64) ([]data.DutyMan, error) {
	for _, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			if v.data != nil {
				return v.data, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpDutyManDataForUser(userId int64, man data.DutyMan) {
	var isUserIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			t.tmpData.tmpDutyManData[i].data = append(t.tmpData.tmpDutyManData[i].data, man)
			isUserIdFound = true
		}
	}
	if isUserIdFound {
		return
	} else {
		// If it's a fresh new data
		var tmp []data.DutyMan
		tmp = append(tmp, man)
		tmpNewMan := tmpDutyManData{userId: userId, data: tmp}
		t.tmpData.tmpDutyManData = append(t.tmpData.tmpDutyManData, tmpNewMan)
	}
}

func (t *TgBot) clearTmpDutyManDataForUser(userId int64) {
	// CLear user temp data
	for i, v := range t.tmpData.tmpDutyManData {
		if v.userId == userId {
			t.tmpData.tmpDutyManData[i].data = nil
		}
	}
}

// Return true if tmpData is still in use by another call
func (t *TgBot) checkTmpDutyMenDataIsEditing(userId int64, update *tgbotapi.Update) bool {
	// If we got error here we can safely continue
	// Because tmpData is empty
	// If we get err == nil - some other function is still running
	if _, err := t.tmpDutyManDataForUser(userId); err == nil {
		messageText := "Вы уже работаете с данными дежурных. Для того, чтобы продолжить, пожалуйста " +
			"сохраните или отмените работу с текущими данными."
		if err := t.sendMessage(messageText,
			update.Message.Chat.ID,
			&update.Message.MessageID,
			nil); err != nil {
			log.Printf("unable to send message: %v", err)
		}
		return true
	}
	return false
}
