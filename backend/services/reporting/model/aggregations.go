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
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	defaultAggregationLimit = 10
	maxAggregationTerms     = 100
	maxNestedAggregations   = 5
)

type AggregateParams struct {
	Aggregations         []AggregationTerm     `json:"aggregations"`
	Filters              []FilterPredicate     `json:"filters"`
	GeoDistanceFilter    *GeoDistanceFilter    `json:"geo_distance_filter"`
	GeoBoundingBoxFilter *GeoBoundingBoxFilter `json:"geo_bounding_box_filter"`
	Groups               []string              `json:"-"`
	TenantID             string                `json:"-"`
}

type AggregationTerm struct {
	Name         string            `json:"name"`
	Attribute    string            `json:"attribute"`
	Scope        string            `json:"scope"`
	Limit        int               `json:"limit"`
	Aggregations []AggregationTerm `json:"aggregations"`
}

func checkMaxNestedAggregationsWithLimit(value interface{}, limit uint) error {
	if limit <= 0 {
		return errors.Errorf("too many nested aggregations, limit is %d", maxNestedAggregations)
	}
	if aggs, ok := value.([]AggregationTerm); ok {
		for _, agg := range aggs {
			if len(agg.Aggregations) > 0 {
				return checkMaxNestedAggregationsWithLimit(agg.Aggregations, limit-1)
			}
		}
	}
	return nil
}

func checkMaxNestedAggregations(value interface{}) error {
	return checkMaxNestedAggregationsWithLimit(value, maxNestedAggregations)
}

func (ap AggregateParams) Validate() error {
	err := validation.ValidateStruct(&ap,
		validation.Field(&ap.Aggregations,
			validation.Required,
			validation.Length(1, maxAggregationTerms),
		),
		validation.Field(&ap.GeoDistanceFilter),
		validation.Field(&ap.GeoBoundingBoxFilter),
	)
	if err != nil {
		return err
	}

	for _, f := range ap.Filters {
		err := f.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f AggregationTerm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Scope, validation.Required),
		validation.Field(&f.Limit, validation.Min(0)),
		validation.Field(&f.Aggregations, validation.When(
			len(f.Aggregations) > 0,
			validation.Length(0, maxAggregationTerms),
			validation.By(checkMaxNestedAggregations),
		)),
	)
}

type Aggregations map[string]interface{}

func BuildAggregations(terms []AggregationTerm) (*Aggregations, error) {
	aggs := Aggregations{}
	for _, term := range terms {
		terms := map[string]interface{}{
			"field": ToAttr(term.Scope, term.Attribute, TypeStr),
		}
		limit := term.Limit
		if limit <= 0 {
			limit = defaultAggregationLimit
		}
		terms["size"] = limit
		agg := map[string]interface{}{
			"terms": terms,
		}
		if len(term.Aggregations) > 0 {
			subaggs, err := BuildAggregations(term.Aggregations)
			if err != nil {
				return nil, err
			}
			agg["aggs"] = subaggs
		}
		aggs[term.Name] = agg
	}
	return &aggs, nil
}

type DeviceAggregation struct {
	Name       string                  `json:"name"`
	Items      []DeviceAggregationItem `json:"items"`
	OtherCount int                     `json:"other_count"`
}

type DeviceAggregationItem struct {
	Key          string              `json:"key"`
	Count        int                 `json:"count"`
	Aggregations []DeviceAggregation `json:"aggregations,omitempty"`
}
