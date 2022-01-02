package data

import (
	"time"
)

// Return first and last date for provided months period.
// startYearMonth is optional. Starts from current Now() month if not provided.
func firstLastMonthDay(monthsCount int, startYearMonth ...int) (firstDay *time.Time, lastDay *time.Time, err error) {
	tn := time.Now()
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return nil, nil, err
	}
	// If startYearMonth not provided let init it from time.Now()
	if len(startYearMonth) == 0 {
		startYearMonth = append(startYearMonth, tn.Year(), int(tn.Month()))
	}

	stime := time.Date(startYearMonth[0], time.Month(startYearMonth[1]), 1, 0, 0, 0, 0, loc)
	etime := stime.AddDate(0, monthsCount, 0).Add(time.Nanosecond * -1)

	return &stime, &etime, nil

}
