package user

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/mendersoftware/mender-server/pkg/rules"
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
	return rules.Email(string(email))
}

type User struct {
	// system-generated user ID
	ID string `json:"id" bson:"_id"`

	// user email address
	Email Email `json:"email" bson:"email"`

	// timestamp of the user creation
	CreatedTs *time.Time `json:"created_ts,omitempty" bson:"created_ts,omitempty"`

	// timestamp of the last user information update
	UpdatedTs *time.Time `json:"updated_ts,omitempty" bson:"updated_ts,omitempty"`

	// LoginTs is the timestamp of the last login for this user.
	LoginTs *time.Time `json:"login_ts,omitempty" bson:"login_ts,omitempty"`
}
