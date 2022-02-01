package data

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	BackupSaveDate = "2021-01-02"
)

func (t *CalData) BackupData(saveName string, rotates int) error {
	switch saveName {
	case SaveNameForBotSettings:
		// Get current men data
		botSettings, err := t.LoadBotSettings()
		if err != nil {
			return CtxError("data.BackupData()", err)
		}
		// Generate json with men save data
		jsonStr, err := json.Marshal(botSettings)
		if err != nil {
			return CtxError("data.BackupData()", err)
		}
		// Write backup data
		if err := t.rotateBackupData(SaveNameForBotSettings, jsonStr, rotates); err != nil {
			return CtxError("data.BackupData()", err)
		}
	case SaveNameForDutyMenData:
		// Get current men data
		menData, err := t.LoadMenList()
		if err != nil {
			return CtxError("data.BackupData()", err)
		}
		// Generate json with men save data
		jsonStr, err := json.Marshal(menData)
		if err != nil {
			return CtxError("data.BackupData()", err)
		}
		// Write backup data
		if err := t.rotateBackupData(SaveNameForDutyMenData, jsonStr, rotates); err != nil {
			return CtxError("data.BackupData()", err)
		}
	default:
		return CtxError("data.BackupData()", fmt.Errorf("data save with name %s not found", saveName))
	}
	return nil
}

func (t *CalData) rotateBackupData(saveName string, jsonStr []byte, rotates int) error {
	// Check if we have valid save type
	if saveName == SaveNameForBotSettings || saveName == SaveNameForDutyMenData {
		// Get current events count
		bkpDate, err := time.Parse(DateShort, BackupSaveDate)
		if err != nil {
			return CtxError("data.rotateBackupData()", err)
		}
		// Get current backup evens (cut '.json' from save name to be able to search it through Google Calendar API)
		events, err := t.dayEvents(&bkpDate, saveName[:len(saveName)-5])
		if err != nil {
			return CtxError("data.rotateBackupData()", err)
		}
		// Iterate over founded events
		if len(events.Items) == 0 { // If no backup data found
			err := t.insertBackupEvent(jsonStr, saveName)
			if err != nil {
				if err != nil {
					return CtxError("data.rotateBackupData()", err)
				}
			}
		} else { // if some backup data is already exists
			// If current backup count is less than target rotates
			if len(events.Items) < rotates {
				err := t.insertBackupEvent(jsonStr, saveName)
				if err != nil {
					if err != nil {
						return CtxError("data.rotateBackupData()", err)
					}
				}
			} else { // Need to rotate some data
				currentEvents := len(events.Items)
				// While current events count is higher than target rotates number
				for currentEvents >= rotates {
					// Get current backup evens (cut '.json' from save name to be able to search it through Google Calendar API)
					events, err := t.dayEvents(&bkpDate, saveName[:len(saveName)-5])
					if err != nil {
						return CtxError("data.rotateBackupData()", err)
					}
					oldestEvent := events.Items[0]
					for _, e := range events.Items {
						oe, err := time.Parse(time.RFC3339, e.Created)
						if err != nil {
							return CtxError("data.rotateBackupData()", err)
						}
						// If current event is older than the oldest then rewrite it
						coe, err := time.Parse(time.RFC3339, oldestEvent.Created)
						if err != nil {
							return CtxError("data.rotateBackupData()", err)
						}
						if oe.Before(coe) {
							oldestEvent = e
						}
					}
					// Delete oldest backup event
					if err := t.cal.Events.Delete(t.calID, oldestEvent.Id).Do(); err != nil {
						return CtxError("data.rotateBackupData()", err)
					}
					currentEvents--
				}
				// Add new backup
				err := t.insertBackupEvent(jsonStr, saveName)
				if err != nil {
					if err != nil {
						return CtxError("data.rotateBackupData()", err)
					}
				}
			}
		}
	} else {
		return CtxError("data.rotateBackupData()", fmt.Errorf("data save with name %s not found", saveName))
	}
	return nil
}

// Insert backup data
func (t *CalData) insertBackupEvent(jsonStr []byte, saveName string) error {
	// Generate event data
	event := genEvent(fmt.Sprintf("%s_%s.bkp", saveName, time.Now().Format(DateShort)),
		string(jsonStr),
		CalGreen,
		BackupSaveDate,
		BackupSaveDate)
	event, err := t.cal.Events.Insert(t.calID, event).Do()
	if err != nil {
		return CtxError("data.insertBackupEvent()", err)
	}
	log.Println(event.HtmlLink)
	return nil
}
