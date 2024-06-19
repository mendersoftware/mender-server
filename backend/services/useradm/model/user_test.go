// Copyright 2022 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateNew(t *testing.T) {

	testCases := map[string]struct {
		inUser User

		outErr string
	}{
		"email ok, pass ok": {
			inUser: User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "",
		},
		"email invalid, pass ok": {
			inUser: User{
				Email:    "foobar",
				Password: "correcthorsebatterystaple",
			},
			outErr: "email: must be a valid email address.",
		},
		"email ok (+), pass ok": {
			inUser: User{
				Email:    "foobar+org@org.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "",
		},
		"email invalid(non-ascii), pass ok": {
			inUser: User{
				Email:    "ąę@org.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "email: must contain ASCII characters only.",
		},
		"email ok, pass invalid (empty)": {
			inUser: User{
				Email:    "foo@bar.com",
				Password: "",
			},
			outErr: "password: cannot be blank.",
		},
		"email ok, pass invalid (too short)": {
			inUser: User{
				Email:    "foo@bar.com",
				Password: "asdf",
			},
			outErr: "password: must be minimum 8 characters long",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.inUser.Validate()

			if tc.outErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.outErr)
			}
		})
	}
}

func TestUserFilterParseForm(t *testing.T) {
	testCases := []struct {
		Name string

		Form url.Values

		Result UserFilter
		Error  error
	}{{
		Name: "ok",

		Form: url.Values{
			"id": []string{"1", "2", "3"},
			"email": []string{
				"user1@acme.io",
				"user2@acme.io",
				"user3@acme.io",
			},
			"created_before": []string{"1234567890"},
			"created_after":  []string{"123456789"},
			"updated_before": []string{"9876543210"},
			"updated_after":  []string{"2345678901"},
		},
		Result: UserFilter{
			ID: []string{"1", "2", "3"},
			Email: []Email{
				"user1@acme.io",
				"user2@acme.io",
				"user3@acme.io",
			},
			CreatedBefore: func() *time.Time {
				ret := time.Unix(1234567890, 0)
				return &ret
			}(),
			CreatedAfter: func() *time.Time {
				ret := time.Unix(123456789, 0)
				return &ret
			}(),
			UpdatedBefore: func() *time.Time {
				ret := time.Unix(9876543210, 0)
				return &ret
			}(),
			UpdatedAfter: func() *time.Time {
				ret := time.Unix(2345678901, 0)
				return &ret
			}(),
		},
	}, {
		Name: "error, created_after not an int",

		Form: url.Values{
			"created_after": []string{"foo"},
		},
		Error: errors.New(`invalid form parameter "created_after": ` +
			`strconv.ParseInt: parsing "foo": invalid syntax`),
	}, {
		Name: "error, created_before not an int",

		Form: url.Values{
			"created_before": []string{"foo"},
		},
		Error: errors.New(`invalid form parameter "created_before": ` +
			`strconv.ParseInt: parsing "foo": invalid syntax`),
	}, {
		Name: "error, updated_after not an int",

		Form: url.Values{
			"updated_after": []string{"foo"},
		},
		Error: errors.New(`invalid form parameter "updated_after": ` +
			`strconv.ParseInt: parsing "foo": invalid syntax`),
	}, {
		Name: "error, updated_before not an int",

		Form: url.Values{
			"updated_before": []string{"foo"},
		},
		Error: errors.New(`invalid form parameter "updated_before": ` +
			`strconv.ParseInt: parsing "foo": invalid syntax`),
	}}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var fltr UserFilter

			err := fltr.ParseForm(tc.Form)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Result, fltr)
			}
		})
	}
}

func TestEmailType(t *testing.T) {
	const (
		testEmail        = "test@mender.io"
		validEmailJSON   = `"` + testEmail + `"`
		invalidEmailJSON = `"` + testEmail
	)

	var email Email

	assert.NoError(t, json.Unmarshal([]byte(validEmailJSON), &email))
	assert.NoError(t, email.Validate())
	assert.Error(t, email.UnmarshalJSON([]byte(invalidEmailJSON)))
}

func TestETag(t *testing.T) {
	t.Parallel()

	const (
		validETag        = `00000000635bc43600000001`
		validETagPlusOne = `00000000635bc43600000002`
	)
	var tag *ETag

	err := tag.UnmarshalText([]byte{})
	assert.EqualError(t, err, "nil ETag")

	tag = new(ETag)

	err = tag.UnmarshalText([]byte(validETag))
	assert.NoError(t, err)

	assert.Equal(t, validETag, tag.String())

	tag.Increment()
	assert.Equal(t, validETagPlusOne, tag.String())

	err = tag.UnmarshalText([]byte("deadbeef"))
	assert.EqualError(t, err, "invalid ETag length")

	err = tag.UnmarshalText([]byte{})
	if assert.NoError(t, err) {
		assert.Equal(t, ETag{}, *tag)
	}

	now := time.Now()
	usr := User{
		ID:        "5a7c5462-7cb0-4b6f-9b57-43814a62b7cc",
		CreatedTs: &now,
	}
	eTag := usr.NextETag()

	assert.Equal(t, fmt.Sprintf("%016x%08x", now.Unix(), 1), eTag.String())
}
