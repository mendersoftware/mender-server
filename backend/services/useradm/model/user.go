// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"
)

const (
	MinPasswordLength = 8
)

var (
	ErrPasswordTooShort = errors.Errorf(
		"password: must be minimum %d characters long",
		MinPasswordLength,
	)
	ErrEmptyUpdate = errors.New("no update information provided")
)

type Email string

func (email *Email) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*email = Email(strings.ToLower(raw))
	return nil
}

func (email Email) Validate() error {
	return validation.Validate(string(email),
		lessThan4096,
		is.ASCII, is.EmailFormat,
	)
}

type ETag [12]byte

var (
	ETagNil ETag
)

func (t *ETag) Increment() {
	c := binary.BigEndian.Uint32((*t)[8:])
	c += 1
	binary.BigEndian.PutUint32((*t)[8:], c)
	if *t == ETagNil {
		t.Increment()
	}
}

func (t *ETag) UnmarshalText(b []byte) error {
	if t == nil {
		return errors.New("nil ETag")
	}
	if len(b) == 0 {
		// Treat an empty string as a special case
		*t = [12]byte{}
		return nil
	}
	if len(b) != 24 {
		return errors.New("invalid ETag length")
	}
	_, err := hex.Decode((*t)[:], b)
	return err
}

func (t ETag) MarshalText() (b []byte, err error) {
	b = make([]byte, 24)
	hex.Encode(b, t[:])
	return b, err
}

func (t ETag) String() string {
	b, _ := t.MarshalText()
	return string(b)
}

type User struct {
	// system-generated user ID
	ID string `json:"id" bson:"_id"`

	// ETag is the entity tag that together with ID uniquely identifies
	// the User document.
	// NOTE: The v1 API does not support ETags, so this is only used
	// internally for checking pre-conditions before performing updates.
	ETag *ETag `json:"-" bson:"etag,omitempty"`

	// user email address
	Email Email `json:"email" bson:"email"`

	// user password
	Password string `json:"password,omitempty" bson:"password"`

	// timestamp of the user creation
	CreatedTs *time.Time `json:"created_ts,omitempty" bson:"created_ts,omitempty"`

	// timestamp of the last user information update
	UpdatedTs *time.Time `json:"updated_ts,omitempty" bson:"updated_ts,omitempty"`

	// LoginTs is the timestamp of the last login for this user.
	LoginTs *time.Time `json:"login_ts,omitempty" bson:"login_ts,omitempty"`
}

func (u User) NextETag() (ret ETag) {
	if u.ETag == nil {
		u.ETag = new(ETag)
	}
	if u.CreatedTs != nil {
		// Weak part of the ETag
		lsb := uint64(u.CreatedTs.Unix())
		binary.BigEndian.PutUint64(ret[:8], lsb)
	}
	c := binary.BigEndian.Uint32(u.ETag[8:])
	c += 1
	binary.BigEndian.PutUint32(ret[8:], c)
	if ret == ETagNil {
		ret.Increment()
	}
	return ret
}

func (u User) Validate() error {
	if err := validation.ValidateStruct(&u,
		validation.Field(&u.Email, validation.Required),
		validation.Field(&u.Password, validation.Required, lessThan4096),
	); err != nil {
		return err
	}
	if len(u.Password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

type UserInternal struct {
	User
	PasswordHash string `json:"password_hash,omitempty" bson:"-"`
	Propagate    *bool  `json:"propagate,omitempty" bson:"-"`
}

func (u UserInternal) Validate() error {
	if u.Password == "" && u.PasswordHash == "" ||
		u.Password != "" && u.PasswordHash != "" {
		return errors.New("password *or* password_hash must be provided")
	} else if u.PasswordHash != "" {
		if u.ShouldPropagate() {
			return errors.New(
				"password_hash is not supported with 'propagate'; use 'password' instead",
			)
		}
		u.User.Password = u.PasswordHash
		defer func() { u.User.Password = "" }()
	}

	return validation.ValidateStruct(&u,
		validation.Field(&u.User),
	)
}

func (u UserInternal) ShouldPropagate() bool {
	return u.Propagate == nil || *u.Propagate
}

type UserUpdate struct {
	// ETag selects user ETag for the user to update
	// NOTE: This is the only parameter that goes into the query condition if set.
	// NOTE: If set to ETagNil, it will match users without an ETag.
	ETag *ETag `json:"-" bson:"-"`

	// ETagUpdate sets the updated ETag value. If not set, it is incremented from the
	// ETag field if that field is set.
	ETagUpdate *ETag `json:"-" bson:"etag,omitempty"`

	// user email address
	Email Email `json:"email,omitempty" bson:",omitempty" valid:"email"`

	// user password
	Password string `json:"password,omitempty" bson:"password,omitempty"`

	// user password
	CurrentPassword string `json:"current_password,omitempty" bson:"-"`

	// timestamp of the last user information update
	UpdatedTs *time.Time `json:"-" bson:"updated_ts,omitempty"`

	// token used to update the user, optional
	Token *jwt.Token `json:"-" bson:"-"`

	LoginTs *time.Time `json:"-" bson:"login_ts,omitempty"`
}

func (u UserUpdate) Validate() error {
	if u.Email == "" && u.Password == "" {
		return ErrEmptyUpdate
	}

	if err := validation.ValidateStruct(&u,
		validation.Field(&u.Email),
		validation.Field(&u.Password,
			validation.When(len(u.Password) > 0, lessThan4096),
		),
	); err != nil {
		return err
	}

	if len(u.Password) > 0 && len(u.Password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

type UserFilter struct {
	ID    []string `json:"id,omitempty"`
	Email []Email  `json:"email,omitempty"`

	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
}

func (fltr *UserFilter) ParseForm(form url.Values) error {
	if ids, ok := form["id"]; ok {
		fltr.ID = ids
	}
	if emails, ok := form["email"]; ok {
		fltr.Email = make([]Email, len(emails))
		for i := range emails {
			fltr.Email[i] = Email(strings.ToLower(emails[i]))
		}
	}
	if ca := form.Get("created_after"); ca != "" {
		caInt, err := strconv.ParseInt(ca, 10, 64)
		if err != nil {
			return errors.Wrap(err,
				`invalid form parameter "created_after"`)
		}
		caUnix := time.Unix(caInt, 0)
		fltr.CreatedAfter = &caUnix
	}

	if cb := form.Get("created_before"); cb != "" {
		cbInt, err := strconv.ParseInt(cb, 10, 64)
		if err != nil {
			return errors.Wrap(err,
				`invalid form parameter "created_before"`)
		}
		cbUnix := time.Unix(cbInt, 0)
		fltr.CreatedBefore = &cbUnix
	}

	if ua := form.Get("updated_after"); ua != "" {
		uaInt, err := strconv.ParseInt(ua, 10, 64)
		if err != nil {
			return errors.Wrap(err,
				`invalid form parameter "updated_after"`)
		}
		uaUnix := time.Unix(uaInt, 0)
		fltr.UpdatedAfter = &uaUnix
	}

	if ub := form.Get("updated_before"); ub != "" {
		ubInt, err := strconv.ParseInt(ub, 10, 64)
		if err != nil {
			return errors.Wrap(err,
				`invalid form parameter "updated_before"`)
		}
		ubUnix := time.Unix(ubInt, 0)
		fltr.UpdatedBefore = &ubUnix
	}
	return nil
}
