package bot

import "testing"

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
