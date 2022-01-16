package data

import (
	"reflect"
	"testing"
)

func Test_checkOffDutyManInList(t *testing.T) {
	offDutyList := []string{"Vasia", "Petia"}

	type args struct {
		man         string
		offDutyList *[]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Man found", args: args{man: "Vasia", offDutyList: &offDutyList}, want: true},
		{name: "Man not found", args: args{man: "Kolya", offDutyList: &offDutyList}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkOffDutyManInList(tt.args.man, tt.args.offDutyList); got != tt.want {
				t.Errorf("checkOffDutyManInList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genContListMenOnDuty(t *testing.T) {
	testString := []string{"Vasia", "Petia", "Slava"}
	testInt := 3
	wantString := []string{"Vasia",
		"Vasia",
		"Vasia",
		"Petia",
		"Petia",
		"Petia",
		"Slava",
		"Slava",
		"Slava"}

	type args struct {
		menOnDuty []string
		contDays  int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Check continues string", args: args{menOnDuty: testString, contDays: testInt}, want: wantString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genContListMenOnDuty(tt.args.menOnDuty, tt.args.contDays); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genContListMenOnDuty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_indexOfCurrentOnDutyMan(t *testing.T) {
	menString := []string{"One", "One", "Two", "Two", "Three", "Three"}

	type args struct {
		countDays        int
		men              []string
		man              string
		manPrevDutyCount int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "Get index of current duty with Two",
			args: args{countDays: 2, men: menString, man: "Two", manPrevDutyCount: 1}, want: 3, wantErr: false},
		{name: "Get index of current duty with One",
			args: args{countDays: 2, men: menString, man: "One", manPrevDutyCount: 2}, want: 2, wantErr: false},
		{name: "Get index of current duty with Three",
			args: args{countDays: 2, men: menString, man: "Three", manPrevDutyCount: 2}, want: 0, wantErr: false},
		{name: "Get index of current duty with Zero man",
			args: args{countDays: 2, men: menString, man: "Zero", manPrevDutyCount: 2}, want: 0, wantErr: false},
		{name: "Get index of current duty with One man",
			args: args{countDays: 2, men: menString, man: "One", manPrevDutyCount: 1}, want: 1, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := indexOfCurrentOnDutyMan(tt.args.countDays, tt.args.men, tt.args.man, tt.args.manPrevDutyCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("indexOfCurrentOnDutyMan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("indexOfCurrentOnDutyMan() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeMan(t *testing.T) {
	sDutyMan1 := []DutyMan{
		{Index: 1, FullName: "Test1", UserName: "test_one"},
		{Index: 2, FullName: "Test2", UserName: "test_two"},
		{Index: 3, FullName: "Test3", UserName: "test_three"},
	}
	sInt1 := 1 // Deleting slice with index "2"
	w1 := []DutyMan{
		{Index: 1, FullName: "Test1", UserName: "test_one"},
		{Index: 3, FullName: "Test3", UserName: "test_three"},
	}
	sDutyMan2 := []DutyMan{
		{Index: 1, FullName: "Test1", UserName: "test_one"},
		{Index: 2, FullName: "Test2", UserName: "test_two"},
		{Index: 3, FullName: "Test3", UserName: "test_three"},
	}
	sInt2 := 2 // Deleting slice with index "3"
	w2 := []DutyMan{
		{Index: 1, FullName: "Test1", UserName: "test_one"},
		{Index: 2, FullName: "Test2", UserName: "test_two"},
	}

	type args struct {
		sl []DutyMan
		s  int
	}
	tests := []struct {
		name string
		args args
		want []DutyMan
	}{
		{name: "Remove slice element two", args: args{sl: sDutyMan1, s: sInt1}, want: w1},
		{name: "Remove slice element three", args: args{sl: sDutyMan2, s: sInt2}, want: w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteMan(tt.args.sl, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deleteMan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genListMenOnDuty(t *testing.T) {
	testMap := &[]DutyMan{
		{Index: 10,
			FullName: "One",
			UserName: "TG_ONE",
			Enabled:  true,
			DutyType: []Duty{{Type: OrdinaryDutyType, Enabled: true}}},
		{Index: 20,
			FullName: "Two",
			UserName: "TG_TWO",
			Enabled:  true,
			DutyType: []Duty{{Type: OrdinaryDutyType, Enabled: true}}},
		{Index: 30,
			FullName: "Three",
			UserName: "TG_THREE",
			Enabled:  true,
			DutyType: []Duty{{Type: OrdinaryDutyType, Enabled: true}}},
	}
	wantString := []string{"TG_ONE", "TG_TWO", "TG_THREE"}
	type args struct {
		m []DutyMan
		d CalTag
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "Check returned strings", args: args{m: *testMap, d: OnDutyTag}, want: wantString, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := genListMenOnDuty(tt.args.m, tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("genListMenOnDuty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genListMenOnDuty() got = %v, want %v", got, tt.want)
			}
		})
	}
}
