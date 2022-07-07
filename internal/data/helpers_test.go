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
			gotFirstDay, gotLastDay, err := FirstLastMonthDay(tt.args.months)
			if (err != nil) != tt.wantErr {
				t.Errorf("FirstLastMonthDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFirstDay.Format(DateShort) != tt.wantFirstDay.Format(DateShort) {
				t.Errorf("FirstLastMonthDay() gotFirstDay = %v, want %v", gotFirstDay, tt.wantFirstDay)
			}
			if gotLastDay.Format(DateShort) != tt.wantLastDay.Format(DateShort) {
				t.Errorf("FirstLastMonthDay() gotLastDay = %v, want %v", gotLastDay, tt.wantLastDay)
			}
		})
	}
}

func Test_equalLists(t *testing.T) {
	list1 := []string{"a", "b", "c", "d", "1", "2"}
	list2 := []string{"d", "c", "b", "a", "2", "1"}
	list3 := []string{"a", "1", "2"}
	type args struct {
		searchList   []string
		searchInList []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Equal compare two lists", args: args{searchList: list2, searchInList: list1}, want: true},
		{name: "Unequal compare two lists", args: args{searchList: list3, searchInList: list1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalLists(tt.args.searchList, tt.args.searchInList); got != tt.want {
				t.Errorf("equalLists() = %v, want %v", got, tt.want)
			}
		})
	}
}
