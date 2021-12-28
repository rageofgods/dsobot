package data

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Create events via API call
func (t *CalData) addEvent(event *calendar.Event) (*calendar.Event, error) {
	event, err := t.cal.Events.Insert(t.calID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create event: %v", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)

	return event, nil
}

// dayEvents Get events for provided day
func (t *CalData) dayEvents(day *time.Time) (*calendar.Events, error) {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return nil, err
	}
	stime := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	etime := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, loc)

	e, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(stime.Format(time.RFC3339)).
		TimeMax(etime.Format(time.RFC3339)).MaxResults(10).Do()
	if err != nil {
		return nil, err
	}

	return e, nil
}

// checkDayTag Check if provided day is working or not.
// Returns true if non-working.
// Returns on-duty man as string
// Returns false if working.
func (t *CalData) checkDayTag(day *time.Time, tag CalTag) (bool, []string, error) {
	events, err := t.dayEvents(day) // Get all events for today
	if err != nil {
		return false, nil, err
	}

	var menList []string // Slice for combining men off-duty for today
	var isDutyFound bool // Set to true if found any off-duty events for today
	if len(events.Items) != 0 {
		for _, item := range events.Items {
			if item.Description == string(tag) {
				switch tag {
				case NonWorkingDay:
					return true, nil, nil
				case OnValidationTag:
					isDutyFound = true
					menList = append(menList, item.Summary)
				case OnDutyTag:
					isDutyFound = true
					menList = append(menList, item.Summary)
				case OffDutyTag:
					isDutyFound = true
					menList = append(menList, item.Summary)
				default:
					return false, nil, nil
				}
			}
		}
	}
	if isDutyFound {
		return true, menList, nil
	}
	return false, nil, nil
}

// Returns Calendar event with provided values
func genEvent(sum string, desc string, color string, startDate string, endDate string) *calendar.Event {
	event := &calendar.Event{
		Summary:     sum,
		Description: desc,
		ColorId:     color,
		Start: &calendar.EventDateTime{
			Date:     startDate,
			TimeZone: TimeZone,
		},
		End: &calendar.EventDateTime{
			Date:     endDate,
			TimeZone: TimeZone,
		},
	}

	return event
}
