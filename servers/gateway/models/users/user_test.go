package users

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.
func TestValidate(t *testing.T) {
	cases := []struct {
		name          string
		nu            *NewUser
		expectedError error
	}{
		{
			"Valid User",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "123456",
				FirstName:    "Tom",
				LastName:     "Lu",
			},
			nil,
		},
		{
			"Invalid Email",
			&NewUser{
				Email:        "tiancyuw.edu",
				Password:     "123456",
				PasswordConf: "123456",
				UserName:     "123456",
				FirstName:    "Tom",
				LastName:     "Lu",
			},
			fmt.Errorf("Invalid email address: %v ", checkmail.ValidateFormat("tiancyuw.edu")),
		},
		{
			"Short Password",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "123",
				PasswordConf: "123",
				UserName:     "123456",
			},
			fmt.Errorf(shortPasswordError),
		},
		{
			"Unmatched Password",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "1231234",
				PasswordConf: "1231233",
				UserName:     "123456",
			},
			fmt.Errorf(passwordNotMatchError),
		},
		{
			"Invalid Username",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "1231234",
				PasswordConf: "1231234",
				UserName:     "",
			},
			fmt.Errorf(invalidUsernameError),
		},
		{
			"Invalid Username2",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "1231234",
				PasswordConf: "1231234",
				UserName:     "sd sd",
			},
			fmt.Errorf(invalidUsernameError),
		},
	}

	for _, c := range cases {
		err := c.nu.Validate()
		if c.expectedError == nil && err == nil {
			continue
		} else if c.expectedError != nil && err == nil {
			t.Errorf("case %s: expected error %s, but actually got nothing", c.name, c.expectedError)
		} else if c.expectedError == nil && err != nil {
			t.Errorf("case %s: unexpected error %s", c.name, err.Error())
		} else if err.Error() != c.expectedError.Error() {
			t.Errorf("case %s: expected error %s, but actually got %s error", c.name, c.expectedError, err.Error())
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name           string
		nu             *NewUser
		formattedEmail string
	}{
		{
			"Upper Letter Email",
			&NewUser{
				Email:        "tiancy@uw.EDU",
				Password:     "1231234",
				PasswordConf: "1231234",
				UserName:     "sdsd",
			},
			"tiancy@uw.edu",
		},
		{
			"Validation Failed",
			&NewUser{
				Email:        "tiancy@uw.edu",
				Password:     "1231234",
				PasswordConf: "123123",
				UserName:     "sdsd",
			},
			"tiancy@uw.edu",
		},
	}

	for _, c := range cases {
		u, err := c.nu.ToUser()
		if err != nil {
			if c.nu.Validate() == nil || err.Error() != c.nu.Validate().Error() {
				t.Errorf("case %s: unexpected error %s", c.name, err.Error())
			}
			continue
		}

		h := md5.New()
		h.Write([]byte(c.formattedEmail))
		newURL := gravatarBasePhotoURL + hex.EncodeToString(h.Sum(nil))
		if newURL != u.PhotoURL {
			t.Errorf("case %s: expected URL %s, actually got %s", c.name, newURL, u.PhotoURL)
		}

		if err1 := bcrypt.CompareHashAndPassword(u.PassHash, []byte(c.nu.Password)); err1 != nil {
			t.Errorf("case %s: password hash error", c.name)
		}
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		firstName string
		lastName  string
		fullName  string
	}{
		{
			"Tom",
			"Lu",
			"Tom Lu",
		},
		{
			"Tom",
			"",
			"Tom",
		},
		{
			"",
			"Lu",
			"Lu",
		},
		{
			"",
			"",
			"",
		},
	}

	for _, c := range cases {
		u := &User{
			FirstName: c.firstName,
			LastName:  c.lastName,
		}
		if u.FullName() != c.fullName {
			t.Errorf("expected %s, actually got %s", c.fullName, u.FullName())
		}
	}
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name        string
		storedPwd   string
		inputPwd    string
		expectError bool
	}{
		{
			"Valid Password",
			"abc",
			"abc",
			false,
		},
		{
			"Empty Password",
			"",
			"",
			false,
		},
		{
			"Empty Password2",
			"",
			"2",
			true,
		},
		{
			"Empty Password2",
			"a",
			"abc",
			true,
		},
	}

	for _, c := range cases {
		u := &User{}
		u.SetPassword(c.storedPwd)

		err := u.Authenticate(c.inputPwd)
		if err != nil && c.expectError == false {
			t.Errorf("case %s: unexpected error %s", c.name, err.Error())
		}
		if err == nil && c.expectError == true {
			t.Errorf("case %s: expecting error, but actually got nothing", c.name)
		}
	}
}

func TestUpdate(t *testing.T) {
	u := &User{
		FirstName: "Ai",
		LastName:  "Ko",
	}
	ud := &Updates{
		FirstName: "Bo",
		LastName:  "Gi",
	}
	err := u.ApplyUpdates(ud)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if u.FirstName != "Bo" || u.LastName != "Gi" {
		t.Errorf("update failed")
	}
}
