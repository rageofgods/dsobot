package data

import (
	"log"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Create events via API call
func (t *CalData) addEvent(event *calendar.Event) (*calendar.Event, error) {
	event, err := t.cal.Events.Insert(t.calID, event).Do()
	if err != nil {
		return nil, CtxError("data.addEvent()", err)
	}
	log.Printf("Event created: %s\n", event.HtmlLink)

	return event, nil
}

// dayEvents Get events for provided day
func (t *CalData) dayEvents(day *time.Time, searchStrings ...string) (*calendar.Events, error) {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return nil, CtxError("data.dayEvents()", err)
	}
	stime := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	etime := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, loc)

	// Search day events with provided patterns
	if len(searchStrings) != 0 {
		var ss string
		for _, v := range searchStrings {
			ss += v + " "
		}
		ss = strings.Trim(ss, " ")
		e, err := t.cal.Events.List(t.calID).ShowDeleted(false).
			SingleEvents(true).TimeMin(stime.Format(time.RFC3339)).
			TimeMax(etime.Format(time.RFC3339)).MaxResults(50).Q(ss).Do()
		if err != nil {
			return nil, CtxError("data.dayEvents()", err)
		}
		return e, nil
	}

	e, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(stime.Format(time.RFC3339)).
		TimeMax(etime.Format(time.RFC3339)).MaxResults(50).Do()
	if err != nil {
		return nil, CtxError("data.dayEvents()", err)
	}
	return e, nil
}

// monthEventsFor Get specified events for current month
func (t *CalData) monthEventsFor(tgId string, dutyTag CalTag) (*calendar.Events, error) {
	firstMonthDay, lastMonthDay, err := FirstLastMonthDay(1)
	if err != nil {
		return nil, CtxError("data.monthEventsFor()", err)
	}

	// Get events with tgId && dutyTag (Add 1 day to TimeMax because it's exclusive in google calendar
	e, err := t.cal.Events.List(t.calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(firstMonthDay.Format(time.RFC3339)).
		TimeMax(lastMonthDay.AddDate(0, 0, 1).Format(time.RFC3339)).
		MaxResults(100).Q(string(dutyTag) + " " + tgId).Do()
	if err != nil {
		return nil, CtxError("data.monthEventsFor()", err)
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
		return false, nil, CtxError("data.checkDayTag()", err)
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
