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
	"fmt"
	"regexp"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

var validSelectors = []interface{}{
	"$eq",
	"$gt",
	"$gte",
	"$in",
	"$lt",
	"$lte",
	"$ltne",
	"$ne",
	"$nin",
	"$exists",
	"$regex",
}

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

func (sc SortCriteria) Validate() error {
	return validation.ValidateStruct(&sc,
		validation.Field(&sc.Scope, validation.Required),
		validation.Field(&sc.Attribute, validation.Required),
		validation.Field(&sc.Order, validation.Required, validation.In(validSortOrders...)))
}

type SelectAttribute struct {
	Scope     Scope  `json:"scope" bson:"scope"`
	Attribute string `json:"attribute" bson:"attribute"`
}

func (selectAttr SelectAttribute) Validate() error {
	return validation.ValidateStruct(&selectAttr,
		validation.Field(&selectAttr.Scope, validation.Required),
		validation.Field(&selectAttr.Attribute, validation.Required),
	)
}

type Scope string

func (s Scope) Validate() error {
	if !slices.Contains(validScopes, string(s)) {
		return errors.Errorf("must be one of %s", strings.Join(validScopes, ", "))
	}
	return nil
}

func (sp SearchParams) Validate() error {
	return validation.ValidateStruct(&sp,
		validation.Field(&sp.Attributes, validation.Each()),
		validation.Field(&sp.Filters, validation.Each()),
		validation.Field(&sp.PerPage, validation.Min(1), validation.Max(rest.PerPageMax)),
		validation.Field(&sp.Sort, validation.Each()),
	)
}

func (f Filter) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.Terms,
			validation.Required.Error("at least one filter term must be provided"),
			validation.Each()),
	)
}

func (f FilterPredicate) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Scope, validation.Required),
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Type, validation.Required, validation.In(validSelectors...)),
		validation.Field(&f.Value, validation.NotNil, validation.By(func(interface{}) error {
			_, err := f.TimeRanges()
			if err != nil {
				return err
			}
			if f.Type == "$regex" {
				if value, ok := f.Value.(string); ok {
					_, err := regexp.Compile(value)
					if err != nil {
						return fmt.Errorf(
							"invalid regular expression: %w",
							err,
						)
					}
				} else {
					return fmt.Errorf(
						"invalid type %T for type $regex: must be string",
						f.Value,
					)
				}

			} else if f.Type == "$exists" {
				if _, ok := f.Value.(bool); !ok {
					return fmt.Errorf(
						"invalid type %T for type $exists: "+
							"must be boolean",
						f.Value)
				}
			}
			return err
		})))
}

type SearchIdentityParams struct {
	ValuePrefix string                   `json:"value_prefix"`
	Attribute   *SearchIdentityAttribute `json:"attribute,omitempty"`
	Attributes  []SelectAttribute        `json:"attributes"`
	Page        int                      `json:"page"`
	PerPage     int                      `json:"per_page"`
}

func (s SearchIdentityParams) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.ValuePrefix, validation.Required),
		validation.Field(&s.Attribute),
		validation.Field(&s.Attributes, validation.Each()),
		validation.Field(&s.PerPage, validation.Max(rest.PerPageMax)),
	)
}

type SearchIdentityAttribute struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
}

func (s SearchIdentityAttribute) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Scope,
			validation.Required,
			validation.In(AttrScopeIdentity, AttrScopeTags),
			validation.When(s.Name == AttrNameID, validation.In(AttrScopeIdentity).
				Error(fmt.Sprintf("must be '%s' when scope is '%s'",
					AttrScopeIdentity, AttrNameID)))),
		validation.Field(&s.Name, validation.Required,
			validation.When(s.Scope == AttrScopeTags, validation.In("name").
				Error(fmt.Sprintf("must be 'name' when scope is '%s'",
					AttrScopeTags)))),
	)
}
