package users

import (
	"testing"

	"gopkg.in/mgo.v2/bson"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		nu          *NewUser
		expectError bool
	}{
		{
			"Valid User",
			&NewUser{
				Email:        "fredhw@uw.edu",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "fredhw",
			},
			false,
		},
		{
			"Invalid Email Address",
			&NewUser{
				Email:        "fredhw.uw.edu",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "fredhw",
			},
			true,
		},
		{
			"Invalid Password",
			&NewUser{
				Email:        "fredhw@uw.edu",
				Password:     "1234",
				PasswordConf: "1234",
				UserName:     "fredhw",
			},
			true,
		},
		{
			"Invalid Password Confirmation",
			&NewUser{
				Email:        "fredhw@uw.edu",
				Password:     "123456",
				PasswordConf: "123466",
				UserName:     "fredhw",
			},
			true,
		},
		{
			"Invalid UserName",
			&NewUser{
				Email:        "fredhw@uw.edu",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "",
			},
			true,
		},
	}

	for _, c := range cases {
		err := (c.nu).Validate()
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error validating user: %v", c.name, err)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one", c.name)
		}
	}
}

// Test the (nu *NewUser) ToUser() function to ensure it calculates the PhotoURL field correctly,
// even when the email address has upper case letters or spaces, and sets the PassHash field to the password hash.
// Since bcrypt hashes are salted with a random value, you can't anticipate what the hash should be,
// but you can verify the generated hash by comparing it to the original password using the bcrypt package functions.
func TestToUser(t *testing.T) {
	cases := []struct {
		name        string
		nu          *NewUser
		expectError bool
	}{
		{
			"Valid User",
			&NewUser{
				Email:    "fredhw@uw.edu",
				Password: "123456",
			},
			false,
		},
		{
			"Valid User with uppercase && space Email",
			&NewUser{
				Email:    " fredHW@uw.edu ",
				Password: "123456",
			},
			false,
		},
	}

	for _, c := range cases {
		user, err := (c.nu).ToUser()
		if err != nil {
			t.Errorf("case %s: unexpected error validating user: %v", c.name, err)
		}
		if err := user.Authenticate((c.nu).Password); err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error validating user: %v", c.name, err)
		}
	}
}

// Test the (u *User) FullName() function to verify that it returns the correct results
// given the various possible inputs (no FirstName, no LastName, neither field set, both fields set).
func TestFullName(t *testing.T) {
	cases := []struct {
		name     string
		user     *User
		expected string
	}{
		{
			"No FirstName field set",
			&User{
				FirstName: "",
				LastName:  "Wijaya",
			},
			"Wijaya",
		},
		{
			"No LastName field set",
			&User{
				FirstName: "Frederick",
				LastName:  "",
			},
			"Frederick",
		},
		{
			"Neither fields set",
			&User{
				FirstName: "",
				LastName:  "",
			},
			"",
		},
		{
			"Valid User with uppercase && space Email",
			&User{
				ID:        bson.NewObjectId(),
				FirstName: "Frederick",
				LastName:  "Wijaya",
			},
			"Frederick Wijaya",
		},
	}

	for _, c := range cases {
		if fn := (c.user).FullName(); fn != c.expected {
			t.Errorf("case %s: incorrect full name: expected %v but got %v", c.name, c.expected, fn)
		}
	}
}

// Test the (u *User) Authenticate() function to verify that authentication happens correctly
// for the various possible inputs (incorrect password, correct password).
func TestAuthentication(t *testing.T) {
	cases := []struct {
		name        string
		nu          *NewUser
		input       string
		expectError bool
	}{
		{
			"Incorrect Password",
			&NewUser{
				Email:    "fredhw@uw.edu",
				Password: "3334",
			},
			"1234",
			true,
		},
		{
			"Correct Password",
			&NewUser{
				Email:    "fredhw@uw.edu",
				Password: "1234",
			},
			"1234",
			false,
		},
	}

	for _, c := range cases {
		user, err := (c.nu).ToUser()
		if err != nil {
			t.Fatalf("error getting user, %v", err)
		}

		err = user.Authenticate(c.input)
		if err != nil && !c.expectError {
			t.Errorf("case %s: unexpected error authenticating user: %v", c.name, err)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one", c.name)
		}
	}
}

// Test the (u *User) ApplyUpdates() function to ensure the user's fields are updated properly given an Updates struct.
func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name        string
		upd         *Updates
		expectError bool
	}{
		{
			"FirstName field not set",
			&Updates{
				FirstName: "",
				LastName:  "Wijaya",
			},
			true,
		},
		{
			"LastName field not set",
			&Updates{
				FirstName: "Frederick",
				LastName:  "",
			},
			true,
		},
		{
			"Both fields set",
			&Updates{
				FirstName: "Frederick",
				LastName:  "Wijaya",
			},
			false,
		},
	}

	for _, c := range cases {
		user := &User{}

		err := user.ApplyUpdates(c.upd)
		if err != nil && !c.expectError && user.FirstName == (c.upd).FirstName && user.LastName == (c.upd).LastName {
			t.Errorf("case %s: unexpected error authenticating user: %v", c.name, err)
		}
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error but didn't get one", c.name)
		}
	}
}
