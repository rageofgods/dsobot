package data

import (
	"testing"
	"time"
)

func Test_firstLastMonthDay(t *testing.T) {
	tn := time.Now()
	loc, _ := time.LoadLocation(TimeZone)
	monthsCount := 2
	wantFirst := time.Date(tn.Year(), tn.Month(), 1, 0, 0, 0, 0, loc)
	wantLast := wantFirst.AddDate(0, monthsCount, -1)
	type args struct {
		months int
	}
	tests := []struct {
		name         string
		args         args
		wantFirstDay *time.Time
		wantLastDay  *time.Time
		wantErr      bool
	}{
		{name: "First and Last day for two months", args: args{months: monthsCount},
			wantFirstDay: &wantFirst, wantLastDay: &wantLast, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFirstDay, gotLastDay, err := firstLastMonthDay(tt.args.months)
			if (err != nil) != tt.wantErr {
				t.Errorf("firstLastMonthDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFirstDay.Format(DateShort) != tt.wantFirstDay.Format(DateShort) {
				t.Errorf("firstLastMonthDay() gotFirstDay = %v, want %v", gotFirstDay, tt.wantFirstDay)
			}
			if gotLastDay.Format(DateShort) != tt.wantLastDay.Format(DateShort) {
				t.Errorf("firstLastMonthDay() gotLastDay = %v, want %v", gotLastDay, tt.wantLastDay)
			}
		})
	}
}
