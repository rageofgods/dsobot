package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"strings"
	"time"
)

// Check if we have date in the command argument
func checkArgHasDate(arg string) (time.Time, error) {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return time.Time{}, err
	}

	tn := time.Time{}
	s := strings.Split(arg, " ")
	if len(s) == 2 {
		var err error
		tn, err = time.ParseInLocation(botDataShort1, s[1], loc)
		if err != nil {
			tn, err = time.ParseInLocation(botDataShort2, s[1], loc)
			if err != nil {
				tn, err = time.ParseInLocation(botDataShort3, s[1], loc)
				if err != nil {
					return tn, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
						"Доступны следующие форматы:\n"+
						"*%q*\n"+
						"*%q*\n"+
						"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
				}
			}
		}
	}
	return tn, nil
}

// Check if we have two dates in the command argument
func checkArgIsOffDutyRange(arg string) ([]time.Time, error) {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return nil, err
	}
	var timeRange []time.Time
	dates := strings.Split(arg, "-")
	if len(dates) == 2 {
		for _, date := range dates {
			//var err error
			parsedTime, err := time.ParseInLocation(botDataShort1, date, loc)
			if err != nil {
				parsedTime, err = time.ParseInLocation(botDataShort2, date, loc)
				if err != nil {
					parsedTime, err = time.ParseInLocation(botDataShort3, date, loc)
					if err != nil {
						return timeRange, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
							"Доступны следующие форматы:\n"+
							"*%q*\n"+
							"*%q*\n"+
							"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
					}
					timeRange = append(timeRange, parsedTime)
				}
				timeRange = append(timeRange, parsedTime)
			}
			timeRange = append(timeRange, parsedTime)
		}
		// Check if dates is in future
		for _, v := range timeRange {
			if v.Before(time.Now()) {
				return nil, fmt.Errorf("указанные даты не должны находится в прошлом: %v",
					v.Format(botDataShort3))
			}
		}
		// Check if dates is on valid order (first must be older than second)
		if timeRange[1].Before(timeRange[0]) {
			return nil, fmt.Errorf("дата %v должна быть старше, чем %v",
				timeRange[1].Format(botDataShort3),
				timeRange[0].Format(botDataShort3))
		}
		// If valid - return true
		return timeRange, nil
	}
	return nil, fmt.Errorf("формат аргумента должен быть: " +
		"*DDMMYYYY-DDMMYYYY* (период _'от-до'_ через дефис)")
}
