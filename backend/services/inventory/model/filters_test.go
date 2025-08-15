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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSearchParams(t *testing.T) {
	testCases := map[string]struct {
		params *SearchParams
		err    error
	}{
		"ok, empty": {
			params: &SearchParams{},
		},
		"ok, filters - $eq": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     "system",
						Attribute: "attribute",
						Type:      "$eq",
						Value:     "value",
					},
				},
			},
		},
		"ok, filters - $in": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     "system",
						Attribute: "attribute",
						Type:      "$in",
						Value:     []string{"value1", "value2"},
					},
				},
			},
		},
		"ok, filters - $nin": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     "system",
						Attribute: "attribute",
						Type:      "$nin",
						Value:     []string{"value1", "value2"},
					},
				},
			},
		},
		"ko, filters": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope: "system",
						Type:  "$eq",
						Value: "value",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
		},
		"ko, filters - unsupported operator": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     "system",
						Attribute: "attribute",
						Type:      "$regex",
						Value:     "value",
					},
				},
			},
			err: errors.New("type: must be a valid value."),
		},
		"ok, sort": {
			params: &SearchParams{
				Sort: []SortCriteria{
					{
						Scope:     "system",
						Attribute: "attribute",
						Order:     "asc",
					},
				},
			},
		},
		"ko, sort, missing attribute": {
			params: &SearchParams{
				Sort: []SortCriteria{
					{
						Scope: "system",
						Order: "asc",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
		},
		"ko, sort, invalid scope": {
			params: &SearchParams{
				Sort: []SortCriteria{
					{
						Scope:     "scope",
						Order:     "asc",
						Attribute: "attribute",
					},
				},
			},
			err: errors.New("scope: must be one of system, identity, inventory, monitor, tags."),
		},
		"ok, attributes": {
			params: &SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope:     "system",
						Attribute: "attribute",
					},
				},
			},
		},
		"ko, attributes, missing attribute": {
			params: &SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope: "system",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
		},
		"ko, attributes, invalid scope": {
			params: &SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope:     "scope",
						Attribute: "attribute",
					},
				},
			},
			err: errors.New("scope: must be one of system, identity, inventory, monitor, tags."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		filter *Filter
		err    error
	}{
		"ok": {
			filter: &Filter{
				Name: "name",
				Terms: []FilterPredicate{
					{
						Scope:     "system",
						Attribute: "attribute",
						Type:      "$eq",
						Value:     "",
					},
				},
			},
		},
		"ko, empty": {
			filter: &Filter{},
			err:    errors.New("name: cannot be blank."),
		},
		"ko, no filter terms": {
			filter: &Filter{
				Name: "name",
			},
			err: errors.New("at least one filter term must be provided"),
		},
		"ko, term validation": {
			filter: &Filter{
				Name: "name",
				Terms: []FilterPredicate{
					{
						Scope: "system",
						Type:  "$eq",
						Value: "",
					},
				},
			},
			err: errors.New("validation failed for term: attribute: cannot be blank."),
		},
		"ko, scope": {
			filter: &Filter{
				Name: "name",
				Terms: []FilterPredicate{
					{
						Scope:     "scope",
						Type:      "$eq",
						Attribute: "attribute",
						Value:     "value",
					},
				},
			},
			err: errors.New("validation failed for term: scope: must be one of system, identity, inventory, monitor, tags."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.filter.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
