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

type AggregateDeploymentsParams struct {
	Aggregations     []DeploymentsAggregationTerm `json:"aggregations"`
	Filters          []DeploymentsFilterPredicate `json:"filters"`
	DeploymentGroups []string                     `json:"-"`
	TenantID         string                       `json:"-"`
}

type DeploymentsAggregationTerm struct {
	Name         string                       `json:"name"`
	Attribute    string                       `json:"attribute"`
	Limit        int                          `json:"limit"`
	Aggregations []DeploymentsAggregationTerm `json:"aggregations"`
}

func checkMaxNestedDeploymentsAggregationsWithLimit(value interface{}, limit uint) error {
	if limit <= 0 {
		return errors.Errorf("too many nested aggregations, limit is %d", maxNestedAggregations)
	}
	if aggs, ok := value.([]DeploymentsAggregationTerm); ok {
		for _, agg := range aggs {
			if len(agg.Aggregations) > 0 {
				return checkMaxNestedDeploymentsAggregationsWithLimit(agg.Aggregations, limit-1)
			}
		}
	}
	return nil
}

func checkMaxNestedDeploymentsAggregations(value interface{}) error {
	return checkMaxNestedDeploymentsAggregationsWithLimit(value, maxNestedAggregations)
}

func (sp AggregateDeploymentsParams) Validate() error {
	err := validation.ValidateStruct(&sp,
		validation.Field(&sp.Aggregations, validation.Required,
			validation.Length(1, maxAggregationTerms)))
	if err != nil {
		return err
	}

	for _, f := range sp.Filters {
		err := f.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f DeploymentsAggregationTerm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Name, validation.Required),
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Limit, validation.Min(0)),
		validation.Field(&f.Aggregations, validation.When(
			len(f.Aggregations) > 0,
			validation.Length(0, maxAggregationTerms),
			validation.By(checkMaxNestedDeploymentsAggregations),
		)),
	)
}

func BuildDeploymentsAggregations(terms []DeploymentsAggregationTerm) (*Aggregations, error) {
	aggs := Aggregations{}
	for _, term := range terms {
		terms := map[string]interface{}{
			"field": term.Attribute,
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
			subaggs, err := BuildDeploymentsAggregations(term.Aggregations)
			if err != nil {
				return nil, err
			}
			agg["aggs"] = subaggs
		}
		aggs[term.Name] = agg
	}
	return &aggs, nil
}
