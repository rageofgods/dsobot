package data

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// SaveMenList Create events via API call
func (t *CalData) SaveMenList(d ...*[]DutyMan) (*string, error) {
	if len(d) != 0 {
		if d[0] == nil {
			return nil, CtxError("data.SaveMenList()", fmt.Errorf("no data for saving"))
		}
		t.dutyMen = d[0]
	}
	if t.dutyMen == nil {
		return nil, CtxError("data.SaveMenList()", fmt.Errorf("no data for saving"))
	}
	jsonStr, err := json.Marshal(t.dutyMen)
	if err != nil {
		return nil, CtxError("data.SaveMenList()", err)
	}

	event := genEvent(SaveNameForDutyMenData, string(jsonStr), CalOrange, SaveListDate, SaveListDate)

	tn, err := time.Parse(DateShort, SaveListDate)
	if err != nil {
		return nil, CtxError("data.SaveMenList()", err)
	}

	events, err := t.dayEvents(&tn, SaveNameForDutyMenData)
	if err != nil {
		return nil, CtxError("data.SaveMenList()", err)
	}

	if len(events.Items) == 0 { // If it's new save
		event, err = t.cal.Events.Insert(t.calID, event).Do()
		if err != nil {
			return nil, CtxError("data.SaveMenList()", err)
		}
	} else { // if we're updating existing save
		for _, item := range events.Items {
			e, err := t.cal.Events.Get(t.calID, item.Id).Do()
			if err != nil {
				return nil, CtxError("data.SaveMenList()", err)
			}
			if e.Summary == SaveNameForDutyMenData {
				event, err = t.cal.Events.Update(t.calID, e.Id, event).Do()
				if err != nil {
					return nil, CtxError("data.SaveMenList()", err)
				}
			}
		}
	}

	log.Println(event.HtmlLink)
	return &event.HtmlLink, nil
}

// LoadMenList is loading previously saved (at Google Calendar) men duty data into DutyMan struct
func (t *CalData) LoadMenList() (*[]DutyMan, error) {
	var men []DutyMan
	tn, err := time.Parse(DateShort, SaveListDate)
	if err != nil {
		return nil, CtxError("data.LoadMenList()", err)
	}

	events, err := t.dayEvents(&tn)
	if err != nil {
		return nil, CtxError("data.LoadMenList()", err)
	}

	if len(events.Items) == 0 {
		return nil, CtxError("data.LoadMenList()", fmt.Errorf("data not found"))
	}
	for _, item := range events.Items {
		e, err := t.cal.Events.Get(t.calID, item.Id).Do()
		if err != nil {
			return nil, CtxError("data.LoadMenList()", err)
		}
		if e.Summary == SaveNameForDutyMenData {
			err := json.Unmarshal([]byte(e.Description), &men)
			if err != nil {
				return nil, CtxError("data.LoadMenList()", err)
			}
			if men == nil {
				return nil, CtxError("data.LoadMenList()", fmt.Errorf("can't load null data"))
			}

			t.dutyMen = &men
			return &men, nil
		}
	}
	return nil, CtxError("data.LoadMenList()", fmt.Errorf("data not found"))
}

// SaveBotSettings is saving bot data in google calendar for persistence
func (t *CalData) SaveBotSettings(botSettings *BotSettings) error {
	// Generate json with data to save
	jsonStr, err := json.Marshal(botSettings)
	if err != nil {
		return CtxError("data.SaveBotSettings()", err)
	}

	// Create calendar event with data for save
	event := genEvent(SaveNameForBotSettings,
		string(jsonStr),
		CalTurquoise,
		SaveBotSettingsDate,
		SaveBotSettingsDate)

	tn, err := time.Parse(DateShort, SaveBotSettingsDate)
	if err != nil {
		return CtxError("data.SaveBotSettings()", err)
	}

	events, err := t.dayEvents(&tn, SaveNameForBotSettings)
	if err != nil {
		return CtxError("data.SaveBotSettings()", err)
	}

	if len(events.Items) == 0 { // If it's new save
		event, err = t.cal.Events.Insert(t.calID, event).Do()
		if err != nil {
			return CtxError("data.SaveBotSettings()", err)
		}
	} else { // if we're updating existing save
		for _, item := range events.Items {
			e, err := t.cal.Events.Get(t.calID, item.Id).Do()
			if err != nil {
				return CtxError("data.SaveBotSettings()", err)
			}
			if e.Summary == SaveNameForBotSettings {
				event, err = t.cal.Events.Update(t.calID, e.Id, event).Do()
				if err != nil {
					return CtxError("data.SaveBotSettings()", err)
				}
			}
		}
	}
	log.Println(event.HtmlLink)
	return nil
}

// LoadBotSettings is loading previously saved (at Google Calendar) bot settings data into BotSettings struct
func (t *CalData) LoadBotSettings() (BotSettings, error) {
	var botSettings BotSettings
	tn, err := time.Parse(DateShort, SaveBotSettingsDate)
	if err != nil {
		return BotSettings{}, CtxError("data.LoadBotSettings()", err)
	}

	events, err := t.dayEvents(&tn)
	if err != nil {
		return BotSettings{}, CtxError("data.LoadBotSettings()", err)
	}

	if len(events.Items) == 0 {
		return BotSettings{}, CtxError("data.LoadBotSettings()", fmt.Errorf("data not found"))
	}
	for _, item := range events.Items {
		e, err := t.cal.Events.Get(t.calID, item.Id).Do()
		if err != nil {
			return BotSettings{}, CtxError("data.LoadBotSettings()", err)
		}
		if e.Summary == SaveNameForBotSettings {
			err := json.Unmarshal([]byte(e.Description), &botSettings)
			if err != nil {
				return BotSettings{}, CtxError("data.LoadBotSettings()", err)
			}
			return botSettings, nil
		}
	}
	return BotSettings{}, CtxError("data.LoadBotSettings()", fmt.Errorf("BotSettings data not found"))
}
