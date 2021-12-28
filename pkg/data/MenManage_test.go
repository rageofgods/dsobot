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

func Test_sortMenOnDuty(t *testing.T) {
	testMap := map[int]map[string]string{
		1: {"One": "1"},
		2: {"Two": "2"},
		3: {"Three": "3"},
	}
	wantInt := []int{1, 2, 3}

	type args struct {
		m map[int]map[string]string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		// Do three iteration for correct order check
		{name: "Check correct order iter one", args: args{m: testMap}, want: wantInt},
		{name: "Check correct order iter two", args: args{m: testMap}, want: wantInt},
		{name: "Check correct order iter three", args: args{m: testMap}, want: wantInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortMenOnDuty(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortMenOnDuty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genListMenOnDuty(t *testing.T) {
	testInt := []int{1, 2, 3}
	testMap := map[int]map[string]string{
		1: {"One": "1"},
		2: {"Two": "2"},
		3: {"Three": "3"},
	}
	wantString := []string{"One", "Two", "Three"}

	type args struct {
		keys []int
		m    map[int]map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Check returned strings", args: args{keys: testInt, m: testMap}, want: wantString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genListMenOnDuty(tt.args.keys, tt.args.m); !reflect.DeepEqual(got, tt.want) {
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
