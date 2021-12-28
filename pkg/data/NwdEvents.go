package data

import (
	"fmt"
	"github.com/rageofgods/isdayoff"
)

// CreateNwdEvents Create non-working events
func (t *CalData) CreateNwdEvents(months int) error {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCodeRussia
	pre := false
	covid := false

	stime, etime, err := firstLastMonthDay(months)
	if err != nil {
		return err
	}
	dayOffStartDay := stime.Format(DateShortIsDayOff)
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
		return err
	}

	if len(day) == 0 {
		return fmt.Errorf("zero days is returned. check your range")
	}

	i := 0
	for d := *stime; d.Before(stime.AddDate(0, months, 0)); d = d.AddDate(0, 0, 1) {
		// Skip all working days
		if day[i] == isdayoff.DayTypeWorking {
			i++
			continue
		}

		event := genEvent(NonWorkingDaySum, string(NonWorkingDay), CalGray,
			d.Format(DateShort), d.Format(DateShort))

		_, err := t.addEvent(event)
		if err != nil {
			return err
		}
		i++
	}

	return nil
}

// UpdateNwdEvents Recreate nwd events
func (t *CalData) UpdateNwdEvents(months int) error {
	err := t.DeleteDutyEvents(months, NonWorkingDay)
	if err != nil {
		return err
	}

	err = t.CreateNwdEvents(months)
	if err != nil {
		return err
	}

	return nil
}
