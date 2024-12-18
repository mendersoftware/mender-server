package utils

import (
	"testing"

	"github.com/mendersoftware/mender-server/services/useradm/model"
)

func TestSimilarChecker(t *testing.T) {
	testCases := map[string]struct {
		inUser   model.User
		inUpdate model.UserUpdate

		expected bool
	}{
		"true": {
			inUser: model.User{
				Email: "foo@bar.com",
			},
			inUpdate: model.UserUpdate{
				Password: "foo@bar",
			},
			expected: true,
		},
		"true, new passowrd and email is similar": {
			inUser: model.User{
				Email: "foo@bar.com",
			},
			inUpdate: model.UserUpdate{
				Password: "correcthorsebatterystaple",
				Email:    "correcthorsebatterystaple@bar.com",
			},
			expected: true,
		},
		"true, new email similar to current password": {
			inUser: model.User{
				Email: "foo@bar.com",
			},
			inUpdate: model.UserUpdate{
				Email:           "correcthorsebatterystaple@bar.com",
				CurrentPassword: "correcthorsebatterystaple",
			},
			expected: true,
		},
		"false, passowrd diffrent from email": {
			inUser: model.User{
				Email: "foo@bar.com",
			},
			inUpdate: model.UserUpdate{
				Password: "correcthorsebatterystaple",
			},
			expected: false,
		},
		"false, new email not similar to new password": {
			inUser: model.User{
				Email: "correcthorsebatterystaple@bar.com",
			},
			inUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				CurrentPassword: "correcthorsebatterystaple",
			},
			expected: false,
		},
		"false, no new password or email": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "foobar",
			},
			inUpdate: model.UserUpdate{},
			expected: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res := CheckIfPassSimilarToEmail(&tc.inUser, &tc.inUpdate)
			if res != tc.expected {
				t.Error()
			}
		})
	}
}
