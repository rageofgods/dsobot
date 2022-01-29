package bot

import (
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
