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

func Test_genListMenOnDuty(t *testing.T) {
	testMap := &[]DutyMan{
		{Index: 10, Name: "One", TgID: "TG_ONE"},
		{Index: 20, Name: "Two", TgID: "TG_TWO"},
		{Index: 30, Name: "Three", TgID: "TG_THREE"},
	}
	wantString := []string{"One", "Two", "Three"}

	type args struct {
		m []DutyMan
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Check returned strings", args: args{m: *testMap}, want: wantString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genListMenOnDuty(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genListMenOnDuty() = %v, want %v", got, tt.want)
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
	wantInt1 := 2
	man1 := "Two"
	wantInt2 := 2
	man2 := "One"
	wantInt3 := 0
	man3 := "Three"
	wantInt4 := 0
	man4 := "Zero"

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
			args: args{countDays: 2, men: menString, man: man1, manPrevDutyCount: 1}, want: wantInt1, wantErr: false},
		{name: "Get index of current duty with One",
			args: args{countDays: 2, men: menString, man: man2, manPrevDutyCount: 2}, want: wantInt2, wantErr: false},
		{name: "Get index of current duty with Three",
			args: args{countDays: 2, men: menString, man: man3, manPrevDutyCount: 2}, want: wantInt3, wantErr: false},
		{name: "Get index of current duty with Zero man",
			args: args{countDays: 2, men: menString, man: man4, manPrevDutyCount: 2}, want: wantInt4, wantErr: false},
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
		{Index: 1, Name: "Test1", TgID: "test_one"},
		{Index: 2, Name: "Test2", TgID: "test_two"},
		{Index: 3, Name: "Test3", TgID: "test_three"},
	}
	sInt1 := 1 // Deleting slice with index "2"
	w1 := []DutyMan{
		{Index: 1, Name: "Test1", TgID: "test_one"},
		{Index: 3, Name: "Test3", TgID: "test_three"},
	}
	sDutyMan2 := []DutyMan{
		{Index: 1, Name: "Test1", TgID: "test_one"},
		{Index: 2, Name: "Test2", TgID: "test_two"},
		{Index: 3, Name: "Test3", TgID: "test_three"},
	}
	sInt2 := 2 // Deleting slice with index "3"
	w2 := []DutyMan{
		{Index: 1, Name: "Test1", TgID: "test_one"},
		{Index: 2, Name: "Test2", TgID: "test_two"},
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
			if got := removeMan(tt.args.sl, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeMan() = %v, want %v", got, tt.want)
			}
		})
	}
}
