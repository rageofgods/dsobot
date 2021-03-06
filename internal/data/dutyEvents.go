package data

import (
	"fmt"
	"log"
	"time"
)

// UpdateOnDutyEvents Recreate on-duty events from start of the current month
func (t *CalData) UpdateOnDutyEvents(contDays int, dutyTag CalTag) error {
	if len(*t.dutyMen) == 0 {
		return CtxError("data.UpdateOnDutyEvents()",
			fmt.Errorf("men of duty list is nil"))
	}
	// Get first day of current month
	firstMonthDay, _, err := FirstLastMonthDay(1)
	if err != nil {
		return CtxError("data.UpdateOnDutyEvents()", err)
	}
	// Delete events
	if err := t.DeleteDutyEvents(firstMonthDay, dutyTag); err != nil {
		log.Printf("%v", err)
	}
	// Create events starting from first month day
	if err := t.CreateOnDutyEvents(firstMonthDay, contDays, dutyTag); err != nil {
		return CtxError("data.UpdateOnDutyEvents()", err)
	}
	return nil
}

// UpdateOnDutyEventsFrom Recreate on-duty events from specific date
func (t *CalData) UpdateOnDutyEventsFrom(startFrom *time.Time, contDays int, dutyTag CalTag) error {
	if len(*t.dutyMen) == 0 {
		return CtxError("data.UpdateOnDutyEventsFrom()",
			fmt.Errorf("men of duty list is nil"))
	}
	// Delete events
	if err := t.DeleteDutyEvents(startFrom, dutyTag); err != nil {
		log.Printf("%v", err)
	}
	// Create events starting from first month day
	if err := t.CreateOnDutyEvents(startFrom, contDays, dutyTag); err != nil {
		return CtxError("data.UpdateOnDutyEvents()", err)
	}
	return nil
}

// CreateOnDutyEvents Iterate over men on duty and create events for them
func (t *CalData) CreateOnDutyEvents(startFrom *time.Time, contDays int, dutyTag CalTag) error {
	// Generate error is provided month is in feature\past
	if startFrom.Month() != time.Now().Month() {
		return CtxError("data.CreateOnDutyEvents()", fmt.Errorf("provided month is not present time: %v",
			startFrom.Month()))
	}

	// Check if men on-duty is initialized
	if t.dutyMen == nil {
		return CtxError("data.CreateOnDutyEvents()",
			fmt.Errorf("need to load on-duty men list first"))
	}

	// Creating slice with sorted men on-duty
	menOnDuty, err := genListMenOnDuty(*t.dutyMen, dutyTag)
	if err != nil {
		return CtxError("data.CreateOnDutyEvents()", err)
	}

	// Generate slice with valid menOnDuty count iteration (following length of duty days)
	tempMen := genContListMenOnDuty(menOnDuty, contDays)

	// Go back to previous day
	prevTime := startFrom.AddDate(0, 0, -1)
	// Get correct index for duty order based on previous month duties
	menCount := t.genIndexForDutyList(&prevTime, dutyTag, contDays, &tempMen)

	_, lastMonthDay, err := FirstLastMonthDay(1)
	if err != nil {
		return CtxError("data.CreateOnDutyEvents()", err)
	}

	// Start to generate men events
	for d := *startFrom; d.Before(*lastMonthDay); d = d.AddDate(0, 0, 1) {
		if menCount == len(tempMen) { // Let's start from the beginning if we reached out the end of list
			menCount = 0
		}

		nwd, _, err := t.checkDayTag(&d, NonWorkingDay) // Check if current day is non-working day
		if err != nil {
			return CtxError("data.CreateOnDutyEvents()", err)
		}
		if nwd { // Don't create any on-duty events if non-working day
			continue
		}

		isOffDuty, menOffDuty, err := t.checkDayTag(&d, OffDutyTag) // Check if current day is off-duty for current man
		if err != nil {
			return CtxError("data.CreateOnDutyEvents()", err)
		}

		// Check if all on-duty men is out off they duty
		// If all men is busy then go try next day
		if equalLists(menOffDuty, menOnDuty) {
			continue
		}

		if isOffDuty { // Check if current day have off-duty events
			for i := 0; i < len(tempMen); i++ { // Run until next free duty man is found
				if checkOffDutyManInList(tempMen[menCount], &menOffDuty) {
					menCount++ // Proceed to next man if current man found in off-duty for today
					if menCount == len(tempMen) {
						menCount = 0 // Go to the first man if we reach end of out men list
					}
				} else {
					break // Current man is able to do his duty
				}
			}
			if menCount == len(tempMen) {
				continue // If we reach end of men list (oll men is busy) let's go to the next day
			}
		}

		// Set calendar event color based on duty type
		var clrID string
		switch dutyTag {
		case OnValidationTag:
			clrID = CalPurple
		case OnDutyTag:
			clrID = CalBlue
		}

		// Create calendar event
		event := genEvent(tempMen[menCount], string(dutyTag), clrID, d.Format(DateShort), d.Format(DateShort))

		// Add calendar event
		if _, err = t.addEvent(event); err != nil {
			return CtxError("data.CreateOnDutyEvents()", err)
		}

		menCount++
	}
	return nil
}

// DeleteDutyEvents Delete events by months range
func (t *CalData) DeleteDutyEvents(startFrom *time.Time, dutyTag CalTag) error {
	_, etime, err := FirstLastMonthDay(1)
	if err != nil {
		return CtxError("data.DeleteDutyEvents()", err)
	}

	events, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(startFrom.Format(time.RFC3339)).
		TimeMax(etime.Format(time.RFC3339)).MaxResults(SearchMaxResults).Do()
	if err != nil {
		return CtxError("data.DeleteDutyEvents()", err)
	}

	if len(events.Items) != 0 {
		for _, item := range events.Items {
			if item.Description == string(dutyTag) {
				if err := t.cal.Events.Delete(t.calID, item.Id).Do(); err != nil {
					return CtxError("data.DeleteDutyEvents()", err)
				}
				log.Printf("Deleted event id: %v\n", item.Id)
			}
		}
	} else {
		return CtxError("data.DeleteDutyEvents()", fmt.Errorf("no items found for delete"))
	}
	return nil
}

// CreateOffDutyEvents Create off-duty (Holiday/illness events)
func (t *CalData) CreateOffDutyEvents(manOffDuty string, fromDate time.Time, toDate time.Time) error {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return CtxError("data.CreateOffDutyEvents()", err)
	}
	stime := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, loc)
	// Append 24 hours to end dated because Google API is considering END date as exclusive
	etime := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, loc).Add(time.Hour * 24)

	event := genEvent(manOffDuty, string(OffDutyTag), CalOrange, stime.Format(DateShort), etime.Format(DateShort))

	if _, err = t.addEvent(event); err != nil {
		return CtxError("data.CreateOffDutyEvents()", err)
	}
	return nil
}

// CleanUpOffDutyEvents will clean up expired off-duty events from saved running and saved data
func (t *CalData) CleanUpOffDutyEvents() error {
	tn := time.Now()
	loc, _ := time.LoadLocation(TimeZone)
	for i, v := range *t.dutyMen {
		buffer := make([]OffDutyData, 0, len(v.OffDuty))
		for ii, vv := range v.OffDuty {
			offDutyEndDate, err := time.ParseInLocation(DateShortSaveData, vv.OffDutyEnd, loc)
			if err != nil {
				return fmt.Errorf("CleanUpOffDutyEvents: %w", err)
			}
			if offDutyEndDate.After(tn) {
				buffer = append(buffer, (*t.dutyMen)[i].OffDuty[ii])
			}
		}
		(*t.dutyMen)[i].OffDuty = buffer
	}
	if _, err := t.SaveMenList(); err != nil {
		return fmt.Errorf("CleanUpOffDutyEvents: %w", err)
	}
	return nil
}

// DeleteOffDutyEvents Create off-duty (Holiday/illness events)
func (t *CalData) DeleteOffDutyEvents(manOffDuty string, fromDate time.Time, toDate time.Time) error {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return CtxError("data.DeleteOffDutyEvents()", err)
	}
	stime := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, loc)
	etime := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 0, loc)

	events, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(stime.Format(time.RFC3339)).
		TimeMax(etime.Format(time.RFC3339)).MaxResults(SearchMaxResults).Do()
	if err != nil {
		return CtxError("data.DeleteOffDutyEvents()", err)
	}

	if len(events.Items) != 0 {
		for _, item := range events.Items {
			if item.Description == string(OffDutyTag) && item.Summary == manOffDuty {
				err := t.cal.Events.Delete(t.calID, item.Id).Do()
				if err != nil {
					return CtxError("data.DeleteOffDutyEvents()", err)
				}
				log.Printf("Deleted event id: %v\n", item.Id)
			}
		}
	} else {
		return CtxError("data.DeleteOffDutyEvents()", fmt.Errorf("no items found for delete"))
	}
	return nil
}

// IsNWD checks if provided time is in non-working period
func (t *CalData) IsNWD(tn time.Time) (bool, error) {
	nwd, _, err := t.checkDayTag(&tn, NonWorkingDay) // Check if current day is non-working day
	if err != nil {
		return false, CtxError("data.IsNWD()", err)
	}
	if nwd {
		return true, nil
	}
	return false, nil
}
