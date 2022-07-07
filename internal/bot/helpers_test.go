package bot

import (
	"dso_bot/internal/data"
	"reflect"
	"testing"
	"time"
)

func Test_genUserFullName(t *testing.T) {
	testFirstName1 := "Eugene"
	testLastName1 := "Khokhlov"
	testFirstName2 := "Eugene"
	testLastName2 := ""
	wantString1 := "Eugene Khokhlov"
	wantString2 := "Eugene"

	type args struct {
		firstName string
		lastName  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Check full name 1", args: args{firstName: testFirstName1, lastName: testLastName1}, want: wantString1},
		{name: "Check full name 2", args: args{firstName: testFirstName2, lastName: testLastName2}, want: wantString2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genUserFullName(tt.args.firstName, tt.args.lastName); got != tt.want {
				t.Errorf("genUserFullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nextMonth(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")

	timeIn1 := time.Date(2021, time.January, 30, 0, 0, 0, 0, loc)
	timeWant1 := time.Date(2021, time.February, 1, 0, 0, 0, 0, loc)

	timeIn2 := time.Date(2021, time.February, 25, 0, 0, 0, 0, loc)
	timeWant2 := time.Date(2021, time.March, 1, 0, 0, 0, 0, loc)

	timeIn3 := time.Date(2021, time.May, 11, 0, 0, 0, 0, loc)
	timeWant3 := time.Date(2021, time.June, 1, 0, 0, 0, 0, loc)

	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "Add month to January", args: args{t: timeIn1}, want: timeWant1, wantErr: false},
		{name: "Add month to February", args: args{t: timeIn2}, want: timeWant2, wantErr: false},
		{name: "Add month to May", args: args{t: timeIn3}, want: timeWant3, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextMonth(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("nextMonth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nextMonth() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prevMonth(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")

	timeIn1 := time.Date(2021, time.March, 30, 0, 0, 0, 0, loc)
	timeWant1 := time.Date(2021, time.February, 28, 0, 0, 0, 0, loc)

	timeIn2 := time.Date(2021, time.February, 25, 0, 0, 0, 0, loc)
	timeWant2 := time.Date(2021, time.January, 31, 0, 0, 0, 0, loc)

	timeIn3 := time.Date(2021, time.November, 11, 0, 0, 0, 0, loc)
	timeWant3 := time.Date(2021, time.October, 31, 0, 0, 0, 0, loc)

	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "Substruct month from March", args: args{t: timeIn1}, want: timeWant1, wantErr: false},
		{name: "Substruct month from February", args: args{t: timeIn2}, want: timeWant2, wantErr: false},
		{name: "Substruct month from November", args: args{t: timeIn3}, want: timeWant3, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prevMonth(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("prevMonth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prevMonth() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isDayInOffDutyRange(t *testing.T) {
	od1 := &data.OffDutyData{OffDutyStart: "02/01/2022", OffDutyEnd: "10/01/2022"}
	od2 := &data.OffDutyData{OffDutyStart: "10/02/2022", OffDutyEnd: "10/02/2022"}

	type args struct {
		offDuty *data.OffDutyData
		day     int
		month   time.Month
		year    int
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "First day test", args: args{offDuty: od1, day: 1, month: 1, year: 2022}, want: false, wantErr: false},
		{name: "Second day test", args: args{offDuty: od1, day: 2, month: 1, year: 2022}, want: true, wantErr: false},
		{name: "Third day test", args: args{offDuty: od1, day: 10, month: 1, year: 2022}, want: true, wantErr: false},
		{name: "Fourth day test", args: args{offDuty: od1, day: 11, month: 1, year: 2022}, want: false, wantErr: false},
		{name: "Fifth day test", args: args{offDuty: od2, day: 9, month: 2, year: 2022}, want: false, wantErr: false},
		{name: "Sixth day test", args: args{offDuty: od2, day: 10, month: 2, year: 2022}, want: true, wantErr: false},
		{name: "Seventh day test", args: args{offDuty: od2, day: 11, month: 2, year: 2022}, want: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isDayInOffDutyRange(tt.args.offDuty, tt.args.day, tt.args.month, tt.args.year)
			if (err != nil) != tt.wantErr {
				t.Errorf("isDayInOffDutyRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isDayInOffDutyRange() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMonthInOffDutyData(t *testing.T) {
	data1 := []data.OffDutyData{
		{OffDutyStart: "02/01/2021", OffDutyEnd: "10/01/2021"},
		{OffDutyStart: "02/01/2022", OffDutyEnd: "10/01/2022"},
		{OffDutyStart: "10/02/2022", OffDutyEnd: "10/02/2022"},
	}

	type args struct {
		offDutyData []data.OffDutyData
		month       time.Month
		year        int
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "First test", args: args{offDutyData: data1, month: time.January, year: 2022}, want: true, wantErr: false},
		{name: "Second test", args: args{offDutyData: data1, month: time.February, year: 2022}, want: true, wantErr: false},
		{name: "Third test", args: args{offDutyData: data1, month: time.May, year: 2022}, want: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isMonthInOffDutyData(tt.args.offDutyData, tt.args.month, tt.args.year)
			if (err != nil) != tt.wantErr {
				t.Errorf("isMonthInOffDutyData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isMonthInOffDutyData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
