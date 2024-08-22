// Copyright 2023 Northern.tech AS
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
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AttrScopeInventory = "inventory"
	AttrScopeIdentity  = "identity"
	AttrScopeSystem    = "system"
	AttrScopeTags      = "tags"
	AttrScopeMonitor   = "monitor"

	AttrNameID             = "id"
	AttrNameGroup          = "group"
	AttrNameUpdated        = "updated_ts"
	AttrNameCreated        = "created_ts"
	AttrNameTagsEtag       = "tags_etag"
	AttrNameNumberOfAlerts = "alert_count"
	AttrNameAlerts         = "alerts"
)

const (
	runeDollar = '\uFF04'
	runeDot    = '\uFF0E'
)

var validGroupNameRegex = regexp.MustCompile("^[A-Za-z0-9_-]*$")

type DeviceID string

var NilDeviceID DeviceID //TODO: how to make it NilDeviceID:=DeviceID(primitive.NilObjectID)

type GroupName string

type DeviceGroups struct {
	Groups []string `json:"groups" bson:"-"`
}

type DeviceAttribute struct {
	Name        string      `json:"name" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
	Value       interface{} `json:"value" bson:",omitempty"`
	Scope       string      `json:"scope" bson:",omitempty"`
	Timestamp   *time.Time  `json:"timestamp,omitempty" bson:",omitempty"`
}

func (da DeviceAttribute) Validate() error {
	return validation.ValidateStruct(&da,
		validation.Field(&da.Name, validation.Required, validation.Length(1, 1024)),
		validation.Field(&da.Scope, validation.Required, validation.Length(1, 1024)),
		validation.Field(&da.Value, validation.By(validateDeviceAttrVal)),
		validation.Field(&da.Timestamp, validation.Date(time.RFC3339)),
	)
}

func validateDeviceAttrVal(i interface{}) error {
	if i == nil {
		return errors.New("supported types are string, float64, and arrays thereof")
	}
	rType := reflect.TypeOf(i)
	if rType.Kind() == reflect.Interface {
		rType = rType.Elem()
	}

	switch rType.Kind() {
	case reflect.Float64, reflect.String:
		return nil
	case reflect.Slice:
		elemKind := rType.Elem().Kind()
		if elemKind == reflect.Float64 || elemKind == reflect.String {
			return nil
		} else if elemKind == reflect.Interface {
			return validateDeviceAttrValArray(i)
		}
	}
	return errors.New("supported types are string, float64, and arrays thereof")
}

func validateDeviceAttrValArray(arr interface{}) error {
	rVal := reflect.ValueOf(arr)
	rLen := rVal.Len()
	if rLen == 0 {
		return nil
	}
	elem := rVal.Index(0)
	kind := elem.Kind()
	if elem.Kind() == reflect.Interface {
		elem = elem.Elem()
		kind = elem.Kind()
	}
	if kind != reflect.String && kind != reflect.Float64 {
		return errors.New(
			"array values must be either string or float64, not: " +
				kind.String())
	}
	for i := 1; i < rLen; i++ {
		elem = rVal.Index(i)
		elemKind := elem.Kind()
		if elemKind == reflect.Interface {
			elemKind = elem.Elem().Kind()
		}
		if elemKind != kind {
			return errors.New(
				"array values must be of consistent type: " +
					"string or float64",
			)
		}
	}
	return nil
}

// Device wrapper
type Device struct {
	//system-generated device ID
	ID DeviceID `json:"id" bson:"_id,omitempty"`

	//a map of attributes names and their values.
	Attributes DeviceAttributes `json:"attributes,omitempty" bson:"attributes,omitempty"`

	//device's group name
	Group GroupName `json:"-" bson:"group,omitempty"`

	CreatedTs time.Time `json:"-" bson:"created_ts,omitempty"`
	//Timestamp of the last attribute update.
	UpdatedTs *time.Time `json:"updated_ts,omitempty" bson:"updated_ts,omitempty"`

	//device object revision
	Revision uint `json:"-" bson:"revision,omitempty"`

	//tags attributes ETag
	TagsEtag string `json:"-" bson:"tags_etag,omitempty"`

	//text attribute for the full-text search
	Text string `json:"-" bson:"text,omitempty"`
}

// internalDevice is only used internally to avoid recursive type-loops for
// member functions.
type internalDevice Device

func (d *Device) UnmarshalBSON(b []byte) error {
	if err := bson.Unmarshal(b, (*internalDevice)(d)); err != nil {
		return err
	}
	for _, attr := range d.Attributes {
		if attr.Scope == AttrScopeSystem {
			switch attr.Name {
			case AttrNameGroup:
				group := attr.Value.(string)
				d.Group = GroupName(group)
			case AttrNameUpdated:
				if attr.Value != nil {
					dateTime := attr.Value.(primitive.DateTime).Time()
					d.UpdatedTs = &dateTime
				} else {
					d.UpdatedTs = nil
				}
			case AttrNameCreated:
				dateTime := attr.Value.(primitive.DateTime)
				d.CreatedTs = dateTime.Time()
			}
		}
	}
	return nil
}

func (d Device) MarshalBSON() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if d.Group != "" {
		d.Attributes = append(d.Attributes, DeviceAttribute{
			Scope: AttrScopeSystem,
			Name:  AttrNameGroup,
			Value: d.Group,
		})
	}
	return bson.Marshal(internalDevice(d))
}

func (d Device) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.ID, validation.Required, validation.Length(1, 1024)),
		validation.Field(&d.Attributes),
		validation.Field(&d.TagsEtag, validation.Length(0, 1024)),
	)
}

func (did DeviceID) String() string {
	return string(did)
}

func (gn GroupName) String() string {
	return string(gn)
}

func (gn GroupName) Validate() error {
	if len(gn) > 1024 {
		return errors.New(
			"Group name can at most have 1024 characters",
		)
	} else if len(gn) == 0 {
		return errors.New(
			"Group name cannot be blank",
		)
	} else if !validGroupNameRegex.MatchString(string(gn)) {
		return errors.New(
			"Group name can only contain: upper/lowercase " +
				"alphanum, -(dash), _(underscore)",
		)
	}
	return nil
}

// wrapper for device attributes names and values
type DeviceAttributes []DeviceAttribute

func (d *DeviceAttributes) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*[]DeviceAttribute)(d))
	if err != nil {
		return err
	}
	for i := range *d {
		if (*d)[i].Scope == "" {
			(*d)[i].Scope = AttrScopeInventory
		}
	}

	return nil
}

// MarshalJSON ensures that an empty array is returned if DeviceAttributes is
// empty.
func (d DeviceAttributes) MarshalJSON() ([]byte, error) {
	if d == nil {
		return json.Marshal([]DeviceAttribute{})
	}
	return json.Marshal([]DeviceAttribute(d))
}

func (d DeviceAttributes) Validate() error {
	for _, a := range d {
		if err := a.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func GetDeviceAttributeNameReplacer() *strings.Replacer {
	return strings.NewReplacer(".", string(runeDot), "$", string(runeDollar))
}

// UnmarshalBSONValue correctly unmarshals DeviceAttributes from Device
// documents stored in the DB.
func (d *DeviceAttributes) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	raw := bson.Raw(b)
	elems, err := raw.Elements()
	if err != nil {
		return err
	}
	*d = make(DeviceAttributes, len(elems))
	for i, elem := range elems {
		err = elem.Value().Unmarshal(&(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// MarshalBSONValue marshals the DeviceAttributes to a mongo-compatible
// document. That is, each attribute is given a unique field consisting of
// "<scope>-<name>".
func (d DeviceAttributes) MarshalBSONValue() (bsontype.Type, []byte, error) {
	attrs := make(bson.D, len(d))
	replacer := GetDeviceAttributeNameReplacer()
	for i := range d {
		attr := DeviceAttribute{
			Name:        d[i].Name,
			Description: d[i].Description,
			Value:       d[i].Value,
			Scope:       d[i].Scope,
			Timestamp:   d[i].Timestamp,
		}
		attrs[i].Key = attr.Scope + "-" + replacer.Replace(d[i].Name)
		attrs[i].Value = &attr
	}
	return bson.MarshalValue(attrs)
}

type DeviceUpdate struct {
	Id       DeviceID `json:"id"`
	Revision uint     `json:"revision"`
}
