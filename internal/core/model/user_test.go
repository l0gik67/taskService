package model

import "testing"

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{name: "valid", user: &User{Name: "Vadim"}},
		{name: "empty name", user: &User{Name: "   "}, wantErr: true},
		{name: "nil user", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.user.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
