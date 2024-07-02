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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateParamsValidate(t *testing.T) {
	tooManyAggregationTerms := make([]AggregationTerm, maxAggregationTerms+1)
	for i := 0; i < maxAggregationTerms+1; i++ {
		tooManyAggregationTerms[i] = AggregationTerm{
			Name:      "mac",
			Scope:     ScopeIdentity,
			Attribute: "mac",
			Limit:     10,
		}
	}

	testCases := map[string]struct {
		params AggregateParams
		err    error
	}{
		"ok, empty": {
			params: AggregateParams{
				Aggregations: []AggregationTerm{
					{
						Name:      "mac",
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
		},
		"ok, full example": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name:      "mac",
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
		},
		"ko, filter fails validation": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Value: "",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name:      "mac",
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
			err: errors.New("attribute: cannot be blank; scope: cannot be blank; type: cannot be blank."),
		},
		"ko, aggregation fails validation": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name: "",
					},
				},
			},
			err: errors.New("aggregations: (0: (attribute: cannot be blank; name: cannot be blank; scope: cannot be blank.).)."),
		},
		"ko, nested aggregation fails validation": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name:      "mac",
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Aggregations: []AggregationTerm{
							{
								Name: "",
							},
						},
						Limit: 10,
					},
				},
			},
			err: errors.New("aggregations: (0: (aggregations: (0: (attribute: cannot be blank; name: cannot be blank; scope: cannot be blank.).).).)."),
		},
		"ko, nested aggregation fails validation (too many terms)": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name:         "mac",
						Scope:        ScopeIdentity,
						Attribute:    "mac",
						Aggregations: tooManyAggregationTerms,
						Limit:        10,
					},
				},
			},
			err: errors.New("aggregations: (0: (aggregations: the length must be no more than 100.).)."),
		},
		"ko, too many nested aggregations": {
			params: AggregateParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []AggregationTerm{
					{
						Name:      "mac",
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Limit:     10,
						Aggregations: []AggregationTerm{
							{
								Name:      "mac",
								Scope:     ScopeIdentity,
								Attribute: "mac",
								Limit:     10,
								Aggregations: []AggregationTerm{
									{
										Name:      "mac",
										Scope:     ScopeIdentity,
										Attribute: "mac",
										Limit:     10,
										Aggregations: []AggregationTerm{
											{
												Name:      "mac",
												Scope:     ScopeIdentity,
												Attribute: "mac",
												Limit:     10,
												Aggregations: []AggregationTerm{
													{
														Name:      "mac",
														Scope:     ScopeIdentity,
														Attribute: "mac",
														Limit:     10,
														Aggregations: []AggregationTerm{
															{
																Name:      "mac",
																Scope:     ScopeIdentity,
																Attribute: "mac",
																Limit:     10,
																Aggregations: []AggregationTerm{
																	{
																		Name:      "mac",
																		Scope:     ScopeIdentity,
																		Attribute: "mac",
																		Limit:     10,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			err: errors.New("aggregations: (0: (aggregations: too many nested aggregations, limit is 5.).)."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBuildAggregations(t *testing.T) {
	testCases := map[string]struct {
		terms []AggregationTerm
		res   *Aggregations
		err   error
	}{
		"ok": {
			terms: []AggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
					Scope:     "scope",
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "scope_attribute_str",
						"size":  defaultAggregationLimit,
					},
				},
			},
		},
		"ok, with limit": {
			terms: []AggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
					Scope:     "scope",
					Limit:     11,
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "scope_attribute_str",
						"size":  11,
					},
				},
			},
		},
		"ok, with subaggrgations": {
			terms: []AggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
					Scope:     "scope",
					Aggregations: []AggregationTerm{
						{
							Name:      "aggregation",
							Attribute: "attribute",
							Scope:     "scope",
						},
					},
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "scope_attribute_str",
						"size":  defaultAggregationLimit,
					},
					"aggs": &Aggregations{
						"aggregation": map[string]interface{}{
							"terms": map[string]interface{}{
								"field": "scope_attribute_str",
								"size":  defaultAggregationLimit,
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res, err := BuildAggregations(tc.terms)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Equal(t, tc.res, res)
			}
		})
	}
}
