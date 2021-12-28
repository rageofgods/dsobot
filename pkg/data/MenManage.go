package data

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"
)

// WhoIsOnDuty Returns duty engineer name
func (t *CalData) WhoIsOnDuty(day *time.Time, dutyTag CalTag) (string, error) {
	events, err := t.dayEvents(day)
	if err != nil {
		return "", err
	}

	if len(events.Items) == 0 {
		return "", fmt.Errorf("no upcoming events found for %s", day.Format(DateShort))
	}

	for _, item := range events.Items {
		e, err := t.cal.Events.Get(t.calID, item.Id).Do()
		if err != nil {
			return "", err
		}
		if e.Description == string(dutyTag) {
			return e.Summary, nil
		}

	}
	return "", fmt.Errorf("duty not found for %s", day.Format(DateShort))
}

// WhoWasOnDuty Returns man name who was the last on duty in the previous month with the number of days done.
func (t *CalData) WhoWasOnDuty(lastYear int,
	lastMonth time.Month, dutyTag CalTag) (name string, daysDone int, err error) {
	// Get first and last date of provided month
	firstMonthDay, lastMonthDay, err := firstLastMonthDay(1, lastYear, int(lastMonth))
	if err != nil {
		return "", 0, err
	}

	var foundPrevMan string // Save founded on-duty man to check next iteration
	manCounter := 0         // Counter for save count of duties for founded man
	for d := *lastMonthDay; d.After(firstMonthDay.Local()); d = d.AddDate(0, 0, -1) {
		nwd, _, err := t.checkDayTag(&d, NonWorkingDay) // Check if current past day is non-working day
		if err != nil {
			return "", 0, err
		}
		if nwd { // If day is non-working when go to the next past day
			continue
		}
		if foundPrevMan == "" { // If we starts from the beginning
			man, err := t.WhoIsOnDuty(&d, dutyTag)
			if err != nil {
				return "", 0, err
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
	return "", 0, fmt.Errorf("can't find previous month on-duty man")
}

// SaveMenList Create events via API call
func (t *CalData) SaveMenList() (*string, error) {
	if t.menOnDuty == nil {
		return nil, fmt.Errorf("no data for saving")
	}
	jsonStr, err := json.Marshal(t.menOnDuty)
	if err != nil {
		return nil, err
	}

	event := genEvent(SaveListName, string(jsonStr), CalOrange, SaveListDate, SaveListDate)

	tn, err := time.Parse(DateShort, SaveListDate)
	if err != nil {
		return nil, err
	}

	events, err := t.dayEvents(&tn)
	if err != nil {
		return nil, err
	}

	if len(events.Items) == 0 { // If it's new save
		event, err = t.cal.Events.Insert(t.calID, event).Do()
		if err != nil {
			return nil, err
		}
	} else { // if we're updating existing save
		for _, item := range events.Items {
			e, err := t.cal.Events.Get(t.calID, item.Id).Do()
			if err != nil {
				return nil, err
			}
			if e.Summary == SaveListName {
				event, err = t.cal.Events.Update(t.calID, e.Id, event).Do()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	log.Println(event.HtmlLink)
	return &event.HtmlLink, nil
}

// LoadMenList load duty order data
func (t *CalData) LoadMenList() (*map[int]map[string]string, error) {
	men := map[int]map[string]string{}
	tn, err := time.Parse(DateShort, SaveListDate)
	if err != nil {
		return nil, err
	}

	events, err := t.dayEvents(&tn)
	if err != nil {
		return nil, err
	}

	if len(events.Items) == 0 {
		return nil, fmt.Errorf("data not found")
	}
	for _, item := range events.Items {
		e, err := t.cal.Events.Get(t.calID, item.Id).Do()
		if err != nil {
			return nil, err
		}
		if e.Summary == SaveListName {
			err := json.Unmarshal([]byte(e.Description), &men)
			if err != nil {
				return nil, err
			}
			if men == nil {
				return nil, fmt.Errorf("can't load null data")
			}

			t.menOnDuty = men
			return &men, nil
		}
	}
	return nil, fmt.Errorf("data not found")
}

// AddManOnDuty Add new man to duty list
func (t *CalData) AddManOnDuty(name string, tgID string) {
	l := len(t.menOnDuty)
	l++ // Go to next available index
	t.menOnDuty[l] = map[string]string{name: tgID}

	///////////
	ll := len(t.dutyMen)
	ll++
	m := &DutyMan{Name: name, Index: ll, TgID: tgID}
	t.dutyMen = append(t.dutyMen, *m)
	//////////
}

// DeleteManOnDuty Remove man from duty list
func (t *CalData) DeleteManOnDuty(tgID string) error {
	var isDeleted bool
	for index, man := range t.menOnDuty {
		for _, v := range man {
			if v == tgID {
				delete(t.menOnDuty, index)
				isDeleted = true
			}
		}
	}
	// Reindex only if something was changed
	if isDeleted {
		t.reIndexManOnDutyList()
		return nil
	}
	return fmt.Errorf("search string not found in map. nothing was deleted")
}

// Recreate indexes for on-duty men to persist duty order
func (t *CalData) reIndexManOnDutyList() {
	reMap := map[int]map[string]string{}
	keys := sortMenOnDuty(t.menOnDuty) // Sort by index
	i := 1                             // Index starts from one
	for _, k := range keys {           // Use sorted slice for persist men order through reindex
		for key, value := range t.menOnDuty[k] {
			reMap[i] = map[string]string{key: value}
		}
		i++
	}
	t.menOnDuty = reMap // Rewrite original map
}

// ShowMenOnDutyList Show current men on duty list
func (t *CalData) ShowMenOnDutyList() []string {
	return genListMenOnDuty(sortMenOnDuty(t.menOnDuty), t.menOnDuty)
}

// Return correct index for duty flow
func (t *CalData) genIndexForDutyList(prevTime time.Time,
	dutyTag CalTag, contDays int, tempMen *[]string) int {
	var menCount int
	man, manPrevDutyCount, err := t.WhoWasOnDuty(prevTime.Year(), prevTime.Month(), dutyTag)
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

// Sorting men on-duty in the right order (following map index)
func sortMenOnDuty(m map[int]map[string]string) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

// Creating slice with sorted men on-duty
func genListMenOnDuty(keys []int, m map[int]map[string]string) []string {
	var menOnDuty []string
	for _, k := range keys {
		for man := range m[k] { // Get on-duty Names
			menOnDuty = append(menOnDuty, man)
		}
	}
	return menOnDuty
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
		return 0, fmt.Errorf("men list is empty")
	}
	for i, name := range men {
		if name == man {
			manIndex = i
		}
	}

	debtDutyDays := contDays - manPrevDutyCount
	if debtDutyDays == 0 { // Checking man is done with his duty
		manIndex++ // Go to next man
	} else { // Checking man still awaits his duty
		if debtDutyDays > 0 { // We only wanted positive digits here
			manIndex -= debtDutyDays // Subtract debt from latest index value
		} else { // If we got below zero value, when something goes wrong (Previous month contDays was bigger than current)
			manIndex++ // Go to next man
		}
	}

	if manIndex == len(men) { // Reset index to zero if we iterate over all men
		manIndex = 0
	}
	return manIndex, nil
}
