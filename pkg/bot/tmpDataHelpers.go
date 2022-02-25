package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
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

func (t *TgBot) tmpOffDutyDataForUser(userId int64) ([]time.Time, error) {
	for _, v := range t.tmpData.tmpOffDutyData {
		if v.userId == userId {
			if v.data != nil {
				return v.data, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) addTmpOffDutyDataForUser(userId int64, date time.Time) {
	var isDateIdFound bool
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpOffDutyData {
		if v.userId == userId {
			// Check if we have correct length of current offDuty data (max = 2)
			if len(t.tmpData.tmpOffDutyData[i].data) == 1 {
				t.tmpData.tmpOffDutyData[i].data = append(t.tmpData.tmpOffDutyData[i].data, date)
				isDateIdFound = true
			}
		}
	}
	if isDateIdFound {
		return
	} else {
		// If it's a fresh new data
		var tmp []time.Time
		tmp = append(tmp, date)
		tmpNewGroup := tmpOffDutyData{userId: userId, data: tmp}
		t.tmpData.tmpOffDutyData = append(t.tmpData.tmpOffDutyData, tmpNewGroup)
	}
}

func (t *TgBot) clearTmpOffDutyDataForUser(userId int64) {
	// CLear user temp data
	for i, v := range t.tmpData.tmpOffDutyData {
		if v.userId == userId {
			t.tmpData.tmpOffDutyData[i].data = nil
		}
	}
}

// Reuse 'tmpOffDutyDataForUser' to access original user data
func (t *TgBot) tmpAdminOffDutyDataForUser(adminUserId int64) ([]time.Time, error) {
	for _, v := range t.tmpData.tmpAdminOffDutyData {
		if v.userId == adminUserId {
			if v.data != nil {
				userData, err := t.tmpAdminOffDutyDataForUserId(adminUserId)
				if err != nil {
					return nil, err
				}
				return userData, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for adminUserId: %d\n", adminUserId)
}

func (t *TgBot) addTmpAdminOffDutyDataForUserId(adminUserId int64, origUserId int64) {
	// Add original user id as a tmp data source
	tmpNewGroup := tmpOffDutyData{userId: origUserId}
	tmpNewAdminGroup := tmpAdminOffDutyData{userId: adminUserId, data: &tmpNewGroup}
	t.tmpData.tmpAdminOffDutyData = append(t.tmpData.tmpAdminOffDutyData, tmpNewAdminGroup)
}

func (t *TgBot) addTmpAdminOffDutyDataForUser(adminUserId int64, date time.Time) {
	// If we already have some previously saved data for current userId
	for i, v := range t.tmpData.tmpAdminOffDutyData {
		if v.userId == adminUserId {
			// Check if we have correct length of current offDuty data (max = 2)
			if len(t.tmpData.tmpAdminOffDutyData[i].data.data) < 2 {
				t.tmpData.tmpAdminOffDutyData[i].data.data = append(t.tmpData.tmpAdminOffDutyData[i].data.data, date)
			}
		}
	}
}

func (t *TgBot) clearTmpAdminOffDutyDataForUser(adminUserId int64) {
	// CLear user temp data
	for i, v := range t.tmpData.tmpAdminOffDutyData {
		if v.userId == adminUserId {
			t.tmpData.tmpAdminOffDutyData[i].data = &tmpOffDutyData{}
		}
	}
}

func (t *TgBot) tmpAdminOffDutyDataForUserId(userId int64) ([]time.Time, error) {
	for _, v := range t.tmpData.tmpAdminOffDutyData {
		if v.userId == userId {
			if v.data != nil {
				return v.data.data, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to find saved data for userId: %d\n", userId)
}

func (t *TgBot) tmpAdminGetOffDutyDataForUserId(adminUserId int64) (*data.DutyMan, error) {
	for i, v := range t.tmpData.tmpAdminOffDutyData {
		if v.userId == adminUserId {
			for _, man := range *t.dc.DutyMenData() {
				if t.tmpData.tmpAdminOffDutyData[i].data.userId == man.TgID {
					return &man, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no user found for id: %d", adminUserId)
}
