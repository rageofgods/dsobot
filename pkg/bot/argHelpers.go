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
