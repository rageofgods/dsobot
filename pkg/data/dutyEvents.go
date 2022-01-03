package data

import (
	"fmt"
	"log"
	"time"
)

// UpdateOnDutyEvents Recreate on-duty events
func (t *CalData) UpdateOnDutyEvents(months int, contDays int, dutyTag CalTag) error {
	if len(*t.dutyMen) == 0 {
		return fmt.Errorf("men of duty list is nil")
	}
	err := t.DeleteDutyEvents(months, dutyTag)
	if err != nil {
		return err
	}

	err = t.CreateOnDutyEvents(months, contDays, dutyTag)
	if err != nil {
		return err
	}

	return nil
}

// CreateOnDutyEvents Iterate over men on duty and create events for them
func (t *CalData) CreateOnDutyEvents(months int, contDays int, dutyTag CalTag) error {
	stime, _, err := firstLastMonthDay(months)
	if err != nil {
		return err
	}

	// Check if men on-duty is initialized
	if t.dutyMen == nil {
		return fmt.Errorf("you need to load on-duty men list first")
	}

	// Creating slice with sorted men on-duty
	menOnDuty, err := genListMenOnDuty(*t.dutyMen)
	if err != nil {
		return err
	}

	// Generate slice with valid menOnDuty count iteration (following length of duty days)
	tempMen := genContListMenOnDuty(menOnDuty, contDays)

	// Go back to previous month
	prevTime := stime.AddDate(0, 0, -1)
	// Get correct index for duty order based on previous month duties
	menCount := t.genIndexForDutyList(prevTime, dutyTag, contDays, &tempMen)

	//menCount := 0
	for d := *stime; d.Before(stime.AddDate(0, months, 0)); d = d.AddDate(0, 0, 1) {
		if menCount == len(tempMen) { // Let's start from the beginning if we reached out the end of list
			menCount = 0
		}

		nwd, _, err := t.checkDayTag(&d, NonWorkingDay) // Check if current day is non-working day
		if err != nil {
			return err
		}
		if nwd { // Don't create any on-duty events if non-working day
			continue
		}

		isOffDuty, offDutyMen, err := t.checkDayTag(&d, OffDutyTag) // Check if current day is off-duty for current man
		if err != nil {
			return err
		}

		// Check if all on-duty men is out off they duty
		// If all men is busy then go try next day
		if equalLists(offDutyMen, menOnDuty) {
			continue
		}

		if isOffDuty { // Check if current day have off-duty events
			for i := 0; i < len(tempMen); i++ { // Run until next free duty man is found
				if checkOffDutyManInList(tempMen[menCount], &offDutyMen) {
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
		_, err = t.addEvent(event)
		if err != nil {
			return err
		}

		menCount++
	}
	return nil
}

// DeleteDutyEvents Delete events by months range
func (t *CalData) DeleteDutyEvents(months int, dutyTag CalTag) error {
	stime, etime, err := firstLastMonthDay(months)
	if err != nil {
		return err
	}

	events, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(stime.Format(time.RFC3339)).
		TimeMax(etime.Format(time.RFC3339)).MaxResults(SearchMaxResults).Do()
	if err != nil {
		return err
	}

	if len(events.Items) != 0 {
		for _, item := range events.Items {
			if item.Description == string(dutyTag) {
				err := t.cal.Events.Delete(t.calID, item.Id).Do()
				if err != nil {
					return err
				}
				log.Printf("Deleted event id: %v\n", item.Id)
			}
		}
	} else {
		return fmt.Errorf("no items found for delete")
	}
	return nil
}

// CreateOffDutyEvents Create off-duty (Holiday/illness events)
func (t *CalData) CreateOffDutyEvents(manOffDuty string, fromDate time.Time, toDate time.Time) error {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return err
	}
	stime := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, loc)
	etime := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 0, loc)

	event := genEvent(manOffDuty, string(OffDutyTag), CalOrange, stime.Format(DateShort), etime.Format(DateShort))

	_, err = t.addEvent(event)
	if err != nil {
		return err
	}

	return nil
}
