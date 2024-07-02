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

func TestAggregateDeploymentsParamsValidate(t *testing.T) {
	tooManyAggregationTerms := make([]DeploymentsAggregationTerm, maxAggregationTerms+1)
	for i := 0; i < maxAggregationTerms+1; i++ {
		tooManyAggregationTerms[i] = DeploymentsAggregationTerm{
			Name:      "mac",
			Attribute: "mac",
			Limit:     10,
		}
	}

	testCases := map[string]struct {
		params AggregateDeploymentsParams
		err    error
	}{
		"ok, empty": {
			params: AggregateDeploymentsParams{
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:      "mac",
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
		},
		"ok, full example": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:      "mac",
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
		},
		"ko, filter fails validation": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Value: "",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:      "mac",
						Attribute: "mac",
						Limit:     10,
					},
				},
			},
			err: errors.New("attribute: cannot be blank; type: cannot be blank."),
		},
		"ko, aggregation fails validation": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name: "",
					},
				},
			},
			err: errors.New("aggregations: (0: (attribute: cannot be blank; name: cannot be blank.).)."),
		},
		"ko, nested aggregation fails validation": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:      "mac",
						Attribute: "mac",
						Aggregations: []DeploymentsAggregationTerm{
							{
								Name: "",
							},
						},
						Limit: 10,
					},
				},
			},
			err: errors.New("aggregations: (0: (aggregations: (0: (attribute: cannot be blank; name: cannot be blank.).).).)."),
		},
		"ko, nested aggregation fails validation (too many terms)": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:         "mac",
						Attribute:    "mac",
						Aggregations: tooManyAggregationTerms,
						Limit:        10,
					},
				},
			},
			err: errors.New("aggregations: (0: (aggregations: the length must be no more than 100.).)."),
		},
		"ko, too many nested aggregations": {
			params: AggregateDeploymentsParams{
				Filters: []DeploymentsFilterPredicate{
					{
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Aggregations: []DeploymentsAggregationTerm{
					{
						Name:      "mac",
						Attribute: "mac",
						Limit:     10,
						Aggregations: []DeploymentsAggregationTerm{
							{
								Name:      "mac",
								Attribute: "mac",
								Limit:     10,
								Aggregations: []DeploymentsAggregationTerm{
									{
										Name:      "mac",
										Attribute: "mac",
										Limit:     10,
										Aggregations: []DeploymentsAggregationTerm{
											{
												Name:      "mac",
												Attribute: "mac",
												Limit:     10,
												Aggregations: []DeploymentsAggregationTerm{
													{
														Name:      "mac",
														Attribute: "mac",
														Limit:     10,
														Aggregations: []DeploymentsAggregationTerm{
															{
																Name:      "mac",
																Attribute: "mac",
																Limit:     10,
																Aggregations: []DeploymentsAggregationTerm{
																	{
																		Name:      "mac",
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

func TestBuildDeploymentsAggregations(t *testing.T) {
	testCases := map[string]struct {
		terms []DeploymentsAggregationTerm
		res   *Aggregations
		err   error
	}{
		"ok": {
			terms: []DeploymentsAggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "attribute",
						"size":  defaultAggregationLimit,
					},
				},
			},
		},
		"ok, with limit": {
			terms: []DeploymentsAggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
					Limit:     11,
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "attribute",
						"size":  11,
					},
				},
			},
		},
		"ok, with subaggrgations": {
			terms: []DeploymentsAggregationTerm{
				{
					Name:      "aggregation",
					Attribute: "attribute",
					Aggregations: []DeploymentsAggregationTerm{
						{
							Name:      "aggregation",
							Attribute: "attribute",
						},
					},
				},
			},
			res: &Aggregations{
				"aggregation": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "attribute",
						"size":  defaultAggregationLimit,
					},
					"aggs": &Aggregations{
						"aggregation": map[string]interface{}{
							"terms": map[string]interface{}{
								"field": "attribute",
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
			res, err := BuildDeploymentsAggregations(tc.terms)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Equal(t, tc.res, res)
			}
		})
	}
}
