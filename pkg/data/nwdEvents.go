package data

import (
	"fmt"
	"github.com/rageofgods/isdayoff"
	"log"
	"time"
)

// CreateNwdEvents Create non-working events
func (t *CalData) CreateNwdEvents(startFrom *time.Time) error {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCodeRussia
	pre := false
	covid := false

	_, etime, err := FirstLastMonthDay(1, startFrom.Year(), int(startFrom.Month()))
	if err != nil {
		return CtxError("data.CreateNwdEvents()", err)
	}
	dayOffStartDay := startFrom.Format(DateShortIsDayOff)
	dayOffEndDay := etime.Format(DateShortIsDayOff)

	day, err := dayOff.GetByRange(isdayoff.ParamsRange{
		StartDate: &dayOffStartDay,
		EndDate:   &dayOffEndDay,
		Params: isdayoff.Params{
			CountryCode: &countryCode,
			Pre:         &pre,
			Covid:       &covid,
		},
	})
	if err != nil {
		return CtxError("data.CreateNwdEvents()", err)
	}

	if len(day) == 0 {
		return CtxError("data.CreateNwdEvents()", fmt.Errorf("zero days is returned. check your range"))
	}

	i := 0
	for d := *startFrom; d.Before(startFrom.AddDate(0, 1, 0)); d = d.AddDate(0, 0, 1) {
		// Skip all working days
		if day[i] == isdayoff.DayTypeWorking {
			i++
			continue
		}

		event := genEvent(NonWorkingDaySum, string(NonWorkingDay), CalGray,
			d.Format(DateShort), d.Format(DateShort))

		if _, err := t.addEvent(event); err != nil {
			return CtxError("data.CreateNwdEvents()", err)
		}
		i++
	}

	return nil
}

// NwdEventsForCurMonth returns slice of days (int) with non-working period for current month
func NwdEventsForCurMonth() ([]int, error) {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCodeRussia
	pre := false
	covid := false

	stime, etime, err := FirstLastMonthDay(1)
	if err != nil {
		return nil, CtxError("data.NwdEventsForCurMonth()", err)
	}
	dayOffStartDay := stime.Format(DateShortIsDayOff)
	dayOffEndDay := etime.Format(DateShortIsDayOff)

	days, err := dayOff.GetByRange(isdayoff.ParamsRange{
		StartDate: &dayOffStartDay,
		EndDate:   &dayOffEndDay,
		Params: isdayoff.Params{
			CountryCode: &countryCode,
			Pre:         &pre,
			Covid:       &covid,
		},
	})
	if err != nil {
		return nil, CtxError("data.NwdEventsForCurMonth()", err)
	}

	var nwdDays []int
	for i := 0; i < etime.Day(); i++ {
		if days[i] == isdayoff.DayTypeNonWorking {
			nwdDays = append(nwdDays, i+1)
		}
	}
	return nwdDays, nil
}

// UpdateNwdEvents Recreate nwd events
func (t *CalData) UpdateNwdEvents() error {
	// Get first day of current month
	firstMonthDay, _, err := FirstLastMonthDay(1)
	if err != nil {
		return CtxError("data.UpdateOnDutyEvents()", err)
	}

	if err := t.DeleteDutyEvents(firstMonthDay, NonWorkingDay); err != nil {
		log.Printf("%v", err)
	}
	if err := t.CreateNwdEvents(firstMonthDay); err != nil {
		return CtxError("data.UpdateNwdEvents()", err)
	}

	return nil
}
