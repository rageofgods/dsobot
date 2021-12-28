package data

import (
	"google.golang.org/api/calendar/v3"
	"reflect"
	"testing"
)

func Test_genEvent(t *testing.T) {
	testSum := "Sum"
	testDesc := "Desc"
	testColor := CalBlue
	testStartDate := SaveListDate
	testEndDate := SaveListDate
	wantEvent := &calendar.Event{
		Summary:     "Sum",
		Description: "Desc",
		ColorId:     "1",
		Start: &calendar.EventDateTime{
			Date:     "2021-01-01",
			TimeZone: TimeZone,
		},
		End: &calendar.EventDateTime{
			Date:     "2021-01-01",
			TimeZone: TimeZone,
		},
	}

	type args struct {
		sum       string
		desc      string
		color     string
		startDate string
		endDate   string
	}
	tests := []struct {
		name string
		args args
		want *calendar.Event
	}{
		{name: "Check returned event", args: args{sum: testSum,
			desc: testDesc, color: testColor, startDate: testStartDate, endDate: testEndDate}, want: wantEvent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genEvent(tt.args.sum, tt.args.desc, tt.args.color,
				tt.args.startDate, tt.args.endDate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
