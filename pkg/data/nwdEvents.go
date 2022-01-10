package data

import (
	"fmt"
	"github.com/rageofgods/isdayoff"
	"log"
)

// CreateNwdEvents Create non-working events
func (t *CalData) CreateNwdEvents(months int) error {
	dayOff := isdayoff.New()
	countryCode := isdayoff.CountryCodeRussia
	pre := false
	covid := false

	stime, etime, err := firstLastMonthDay(months)
	if err != nil {
		return CtxError("data.CreateNwdEvents()", err)
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
		return CtxError("data.CreateNwdEvents()", err)
	}

	if len(day) == 0 {
		return CtxError("data.CreateNwdEvents()", fmt.Errorf("zero days is returned. check your range"))
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

		if _, err := t.addEvent(event); err != nil {
			return CtxError("data.CreateNwdEvents()", err)
		}
		i++
	}

	return nil
}

// UpdateNwdEvents Recreate nwd events
func (t *CalData) UpdateNwdEvents(months int) error {
	if err := t.DeleteDutyEvents(months, NonWorkingDay); err != nil {
		log.Printf("%v", err)
	}
	if err := t.CreateNwdEvents(months); err != nil {
		return CtxError("data.UpdateNwdEvents()", err)
	}

	return nil
}
