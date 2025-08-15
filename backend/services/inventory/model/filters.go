// Copyright 2024 Northern.tech AS
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
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

var validSelectors = []interface{}{"$eq", "$in", "$nin"}
var validSortOrders = []interface{}{"asc", "desc"}
var validScopes = []string{"system", "identity", "inventory", "monitor", "tags"}

type SearchParams struct {
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	Filters    []FilterPredicate `json:"filters"`
	Sort       []SortCriteria    `json:"sort"`
	Attributes []SelectAttribute `json:"attributes"`
	DeviceIDs  []string          `json:"device_ids"`
	Text       string            `json:"text"`
}

type Filter struct {
	Id    string            `json:"id" bson:"_id"`
	Name  string            `json:"name" bson:"name"`
	Terms []FilterPredicate `json:"terms" bson:"terms"`
}

type FilterPredicate struct {
	Scope     Scope       `json:"scope" bson:"scope"`
	Attribute string      `json:"attribute" bson:"attribute"`
	Type      string      `json:"type" bson:"type"`
	Value     interface{} `json:"value" bson:"value"`
}

type SortCriteria struct {
	Scope     Scope  `json:"scope"`
	Attribute string `json:"attribute"`
	Order     string `json:"order"`
}

type SelectAttribute struct {
	Scope     Scope  `json:"scope" bson:"scope"`
	Attribute string `json:"attribute" bson:"attribute"`
}

type Scope string

func (s Scope) Validate() error {
	if !slices.Contains(validScopes, string(s)) {
		return errors.Errorf("must be one of %s", strings.Join(validScopes, ", "))
	}
	return nil
}

func (sp SearchParams) Validate() error {
	for _, f := range sp.Filters {
		err := f.Validate()
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Sort {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Scope, validation.Required),
			validation.Field(&s.Attribute, validation.Required),
			validation.Field(&s.Order, validation.Required, validation.In(validSortOrders...)))
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Attributes {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Scope, validation.Required),
			validation.Field(&s.Attribute, validation.Required))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f Filter) Validate() error {
	err := validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required))
	if err != nil {
		return err
	}

	if len(f.Terms) == 0 {
		return errors.New("at least one filter term must be provided")
	}

	for _, fp := range f.Terms {
		err = fp.Validate()
		if err != nil {
			return errors.Wrap(err, "validation failed for term")
		}
	}

	return nil
}

func (f FilterPredicate) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Scope, validation.Required),
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Type, validation.Required, validation.In(validSelectors...)),
		validation.Field(&f.Value, validation.NotNil))
}
