// Copyright 2023 Northern.tech AS
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
		"ok, filters": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     "scope",
						Attribute: "attribute",
						Type:      "$eq",
						Value:     "value",
					},
				},
			},
		},
		"ko, filters": {
			params: &SearchParams{
				Filters: []FilterPredicate{
					{
						Scope: "scope",
						Type:  "$eq",
						Value: "value",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
		},
		"ok, sort": {
			params: &SearchParams{
				Sort: []SortCriteria{
					{
						Scope:     "scope",
						Attribute: "attribute",
						Order:     "asc",
					},
				},
			},
		},
		"ko, sort": {
			params: &SearchParams{
				Sort: []SortCriteria{
					{
						Scope: "scope",
						Order: "asc",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
		},
		"ok, attributes": {
			params: &SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope:     "scope",
						Attribute: "attribute",
					},
				},
			},
		},
		"ko, attributes": {
			params: &SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope: "scope",
					},
				},
			},
			err: errors.New("attribute: cannot be blank."),
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
						Scope:     "scope",
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
						Scope: "scope",
						Type:  "$eq",
						Value: "",
					},
				},
			},
			err: errors.New("validation failed for term: attribute: cannot be blank."),
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
