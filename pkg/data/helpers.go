package data

import "time"

// Compare two slices and return true if they are equal (don't care about order)
func equalLists(searchList []string, searchInList []string) bool {
	var i int
	for _, sl1 := range searchList {
		for _, sl2 := range searchInList {
			if sl2 == sl1 {
				i++
			}
		}
	}
	// If count is equal len
	if i == len(searchInList) {
		return true
	}
	return false
}

// FirstLastMonthDay Return first and last date for provided months period.
// startYearMonth is optional. Starts from current Now() month if not provided.
func FirstLastMonthDay(monthsCount int, startYearMonth ...int) (firstDay *time.Time, lastDay *time.Time, err error) {
	tn := time.Now()
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return nil, nil, CtxError("data.FirstLastMonthDay()", err)
	}
	// If startYearMonth not provided let init it from time.Now()
	if len(startYearMonth) == 0 {
		startYearMonth = append(startYearMonth, tn.Year(), int(tn.Month()))
	}

	stime := time.Date(startYearMonth[0], time.Month(startYearMonth[1]), 1, 0, 0, 0, 0, loc)
	etime := stime.AddDate(0, monthsCount, 0).Add(time.Nanosecond * -1)

	return &stime, &etime, nil
}
