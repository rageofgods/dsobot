package data

import (
	"fmt"
	"log"
	"time"
)

// WhoIsOnDuty Returns duty engineer tgId
func (t *CalData) WhoIsOnDuty(day *time.Time, dutyTag CalTag) (string, error) {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return "", CtxError("data.WhoIsOnDuty()",
			fmt.Errorf("unable to load timezone %s", err))
	}
	d := day.In(loc)

	events, err := t.dayEvents(&d)
	if err != nil {
		return "", CtxError("data.WhoIsOnDuty()", err)
	}

	if len(events.Items) == 0 {
		return "", CtxError("data.WhoIsOnDuty()",
			fmt.Errorf("no upcoming events found for %s", day.Format(DateShort)))
	}

	for _, item := range events.Items {
		e, err := t.cal.Events.Get(t.calID, item.Id).Do()
		if err != nil {
			return "", CtxError("data.WhoIsOnDuty()", err)
		}
		if e.Description == string(dutyTag) {
			return e.Summary, nil
		}

	}
	return "", CtxError("data.WhoIsOnDuty()",
		fmt.Errorf("дежурств не найдено для %s", day.Format(DateShort)))
}

// ShowOffDutyForMan returns slice of OffDutyData with start/end off-duty dates
func (t *CalData) ShowOffDutyForMan(tgID string) (*[]OffDutyData, error) {
	for _, man := range *t.dutyMen {
		if man.UserName == tgID {
			return &man.OffDuty, nil
		}
	}
	return nil, CtxError("data.ShowOffDutyForMan()",
		fmt.Errorf("can't find user with tgID: @%s in saved data", tgID))
}

// ManDutiesList returns slice of dates with requested duty type for specified man tgId
func (t *CalData) ManDutiesList(tgId string, dutyTag CalTag) (*[]time.Time, error) {
	// Define returned slice of dates
	var dutyDates []time.Time

	// Get events for current month
	events, err := t.monthEventsFor(tgId, dutyTag)
	if err != nil {
		return nil, CtxError("data.ManDutiesList()", err)
	}
	if len(events.Items) == 0 {
		return nil, CtxError("data.ManDutiesList()",
			fmt.Errorf("no upcoming events found for current month"))
	}
	// Fill up dutyDates slice
	for _, event := range events.Items {
		sdate, err := time.Parse(DateShort, event.Start.Date)
		if err != nil {
			return nil, CtxError("data.ManDutiesList()", err)
		}
		edate, err := time.Parse(DateShort, event.End.Date)
		if err != nil {
			return nil, CtxError("data.ManDutiesList()", err)
		}
		// Duty events must be presented as single day events.
		if sdate == edate {
			dutyDates = append(dutyDates, sdate)
		}
	}
	return &dutyDates, nil
}

// WhoWasOnDuty Returns man name who was the last on duty in the previous month with the number of days done.
func (t *CalData) WhoWasOnDuty(lastMonthDay *time.Time, dutyTag CalTag) (name string, daysDone int, err error) {
	// Get first and last date of provided month
	firstMonthDay, _, err := FirstLastMonthDay(1, lastMonthDay.Year(), int(lastMonthDay.Month()))
	if err != nil {
		return "", 0, CtxError("data.WhoWasOnDuty()", err)
	}

	var foundPrevMan string // Save founded on-duty man to check next iteration
	manCounter := 0         // Counter for save count of duties for founded man
	for d := *lastMonthDay; d.After(firstMonthDay.Local()); d = d.AddDate(0, 0, -1) {
		nwd, _, err := t.checkDayTag(&d, NonWorkingDay) // Check if current past day is non-working day
		if err != nil {
			return "", 0, CtxError("data.WhoWasOnDuty()", err)
		}
		if nwd { // If day is non-working when go to the next past day
			continue
		}
		if foundPrevMan == "" { // If we start from the beginning
			man, err := t.WhoIsOnDuty(&d, dutyTag)
			if err != nil {
				return "", 0, CtxError("data.WhoWasOnDuty()", err)
			}
			foundPrevMan = man
			manCounter++
		} else { // If we already found on-duty man in the previous iteration
			man, err := t.WhoIsOnDuty(&d, dutyTag)
			if err != nil { // Return founded previous man if we can't found any on-duty events on current day
				return foundPrevMan, manCounter, nil
			}
			if man == foundPrevMan { // Check if previous found on-duty man is the same as man of current day
				manCounter++
				continue
			}
			return foundPrevMan, manCounter, nil // If we found different man we're done now
		}
	}
	return "", 0, CtxError("data.WhoWasOnDuty()",
		fmt.Errorf("can't find previous month on-duty man"))
}

// AddManOnDuty Add new man to duty list
func (t *CalData) AddManOnDuty(fullName string, userName string, customName string, tgID int64) {
	// Check if added user is unique in the data
	for _, man := range *t.dutyMen {
		if tgID == man.TgID {
			log.Printf("user %s with tgid: %d is already exists in the data", fullName, tgID)
			return
		}
	}
	ln := len(*t.dutyMen)
	ln++
	m := &DutyMan{
		FullName:   fullName,
		Index:      ln,
		UserName:   userName,
		CustomName: customName,
		TgID:       tgID,
		Enabled:    true, // Set every new user as an active by default
		DutyType: []Duty{
			{Type: OrdinaryDutyType, Name: OrdinaryDutyName, Enabled: false},     // Set Duty type to disabled by default
			{Type: ValidationDutyType, Name: ValidationDutyName, Enabled: false}, // Set Duty type to disabled by default
		},
	}
	*t.dutyMen = append(*t.dutyMen, *m)
}

// AddOffDutyToMan Add off-duty event data to man
func (t *CalData) AddOffDutyToMan(tgID string, startDate time.Time, endDate time.Time) {
	stime := startDate.Format(DateShortSaveData)
	etime := endDate.Format(DateShortSaveData)
	for i, man := range *t.dutyMen {
		if man.UserName == tgID {
			m := &OffDutyData{OffDutyStart: stime, OffDutyEnd: etime}
			(*t.dutyMen)[i].OffDuty = append((*t.dutyMen)[i].OffDuty, *m)
		}
	}
}

// DeleteOffDutyFromMan Removes off-duty period from specified man
func (t *CalData) DeleteOffDutyFromMan(tgID string, offDutyDataIndex int) {
	for i, man := range *t.dutyMen {
		if man.UserName == tgID {
			tmp := (*t.dutyMen)[i].OffDuty
			tmp = append(tmp[:offDutyDataIndex], tmp[offDutyDataIndex+1:]...)
			(*t.dutyMen)[i].OffDuty = tmp
		}
	}
}

// Return new slice with removed element with provided index
func deleteMan(sl []DutyMan, s int) []DutyMan {
	return append(sl[:s], sl[s+1:]...)
}

// DeleteManOnDuty Remove man from duty list
func (t *CalData) DeleteManOnDuty(tgID string) error {
	var isDeleted bool
	for index, man := range *t.dutyMen {
		if tgID == man.UserName {
			*t.dutyMen = deleteMan(*t.dutyMen, index)
			isDeleted = true
		}
	}
	// Reindex only if something was changed
	if isDeleted {
		t.reIndexManOnDutyList()
		return nil
	}
	return CtxError("data.DeleteManOnDuty()",
		fmt.Errorf("search string not found in map. nothing was deleted"))
}

// Recreate indexes for on-duty men to persist duty order
func (t *CalData) reIndexManOnDutyList() {
	var reMap []DutyMan
	for index, man := range *t.dutyMen {
		man.Index = index + 1
		reMap = append(reMap, man)
	}
	t.dutyMen = &reMap
}

// DutyMenData Returns current men on duty list
// Optional argument returns only "active" men if true
// Returns only "passive" men if false
func (t *CalData) DutyMenData(enabled ...bool) *[]DutyMan {
	// Check if we have some argument
	if len(enabled) == 1 {
		switch enabled[0] {
		case true:
			var r []DutyMan
			for _, v := range *t.dutyMen {
				if v.Enabled {
					r = append(r, v)
				}
			}
			return &r
		case false:
			var r []DutyMan
			for _, v := range *t.dutyMen {
				if !v.Enabled {
					r = append(r, v)
				}
			}
			return &r
		}
	}
	// Return full by default
	return t.dutyMen
}

// Return correct index for duty flow
func (t *CalData) genIndexForDutyList(prevTime *time.Time,
	dutyTag CalTag, contDays int, tempMen *[]string) int {
	var menCount int
	man, manPrevDutyCount, err := t.WhoWasOnDuty(prevTime, dutyTag)
	index, err := indexOfCurrentOnDutyMan(contDays, *tempMen, man, manPrevDutyCount)
	if err != nil {
		menCount = 0
	} else {
		menCount = index
	}
	return menCount
}

// Check if name is in offDuty list
func checkOffDutyManInList(man string, offDutyList *[]string) bool {
	for _, m := range *offDutyList {
		if m == man {
			return true
		}
	}
	return false
}

// Creating slice with sorted men on-duty
func genListMenOnDuty(m []DutyMan, dutyTag CalTag) ([]string, error) {
	var retStr []string

	if m == nil || len(m) == 0 {
		return nil, CtxError("data.genListMenOnDuty()",
			fmt.Errorf("unable to load men list, please load it first"))
	}
	for _, man := range m {
		// Add only active men to menOnDuty list
		if man.Enabled {
			// Check requested duty type
			for _, duty := range man.DutyType {
				switch dutyTag {
				case OnDutyTag:
					// Check if user is up for duty type
					if duty.Type == OrdinaryDutyType && duty.Enabled {
						retStr = append(retStr, man.UserName)
					}
				case OnValidationTag:
					// Check if user is up for duty type
					if duty.Type == ValidationDutyType && duty.Enabled {
						retStr = append(retStr, man.UserName)
					}
				}
			}
		}
	}
	return retStr, nil
}

// Generate slice with valid menOnDuty count iteration (following length of duty days)
func genContListMenOnDuty(menOnDuty []string, contDays int) []string {
	var tempMen []string
	for _, p := range menOnDuty {
		for i := 0; i < contDays; i++ {
			tempMen = append(tempMen, p)
		}
	}
	return tempMen
}

// Return correct index for continues duty list
func indexOfCurrentOnDutyMan(contDays int, men []string, man string, manPrevDutyCount int) (int, error) {
	var manIndex int // index for founding man latest position in the men slice
	if len(men) == 0 {
		return 0, CtxError("data.indexOfCurrentOnDutyMan()",
			fmt.Errorf("men list is empty"))
	}
	var isManFound bool
	for i, name := range men {
		if name == man {
			manIndex = i
			isManFound = true
			break
		}
	}

	// If previous duty man is out of current menOnDuty list we return zero index
	if !isManFound {
		return 0, nil
	}

	debtDutyDays := contDays - manPrevDutyCount
	if debtDutyDays == 0 { // Checking man is done with his duty
		manIndex += contDays // Go to next man
	} else { // Checking man still awaits his duty
		if debtDutyDays > 0 { // We only wanted positive digits here
			manIndex += manPrevDutyCount // Append previously duty days to current index
		} else {
			// If we got below zero value, when something goes wrong
			// (Previous month contDays was bigger than current)
			manIndex += contDays // Go to next man
		}
	}

	if manIndex == len(men) { // Reset index to zero if we iterate over all men
		manIndex = 0
	}
	return manIndex, nil
}

// IsInDutyList returns true if provided Telegram ID is in duty list
func (t *CalData) IsInDutyList(tgID string) bool {
	for _, man := range *t.dutyMen {
		if man.UserName == tgID {
			return true
		}
	}
	return false
}

// ListMenTgID returns slice of Telegram IDs of all registered men
func (t *CalData) ListMenTgID() []string {
	var menIDs []string
	for _, man := range *t.dutyMen {
		menIDs = append(menIDs, man.UserName)
	}
	return menIDs
}
