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
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	settingsID       = "_id"
	settingsETag     = "etag"
	settingsUserID   = "user_id"
	settingsTenantID = "tenant_id"

	maxSettings = 1024
)

type SettingsValues map[string]interface{}

type Settings struct {
	ID     string         `json:"id"`
	ETag   string         `json:"etag"`
	UserID string         `json:"-"`
	Values SettingsValues `json:"-"`
}

func (s Settings) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Values)
}

func (s *Settings) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.Values)
}

func (s Settings) MarshalBSON() ([]byte, error) {
	value := map[string]interface{}{}
	for k, v := range s.Values {
		value[k] = v
	}
	if s.ID != "" {
		value[settingsID] = s.ID
	}
	value[settingsETag] = s.ETag
	if s.UserID != "" {
		value[settingsUserID] = s.UserID
	}
	return bson.Marshal(value)
}

func (s *Settings) UnmarshalBSON(b []byte) error {
	value := map[string]interface{}{}
	err := bson.Unmarshal(b, &value)
	if val, ok := value[settingsID]; ok {
		if valString, ok := val.(string); ok {
			s.ID = valString
		}
	}
	if val, ok := value[settingsETag]; ok {
		if valString, ok := val.(string); ok {
			s.ETag = valString
		}
	}
	if err == nil {
		delete(value, settingsID)
		delete(value, settingsETag)
		delete(value, settingsUserID)
		delete(value, settingsTenantID)
		s.Values = value
	}
	return err
}

func ValidateKeys(value interface{}) error {
	s, _ := value.(SettingsValues)
	for k := range s {
		err := validation.Validate(k, lessThan128)
		if err != nil {
			return err
		}
	}
	return nil
}

func lessThan4096Strings(value interface{}) error {
	if _, ok := value.(string); ok {
		return validation.Validate(value, lessThan4096)
	}
	return nil
}

func (s Settings) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Values,
			validation.Length(0, maxSettings),
			validation.By(ValidateKeys),
			validation.Each(
				validation.By(lessThan4096Strings),
			),
		),
	)
}
