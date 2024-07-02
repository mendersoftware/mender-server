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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildQuery(t *testing.T) {
	testCases := map[string]struct {
		inParams SearchParams
		outQuery Query
		outErr   error
	}{
		"empty": {
			inParams: SearchParams{
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery(),
		},
		"groups": {
			inParams: SearchParams{
				Groups:  []string{"group1", "group2"},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"terms": M{
					"system_group_str": []string{"group1", "group2"},
				},
			}),
		},
		"filter $eq": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"match": M{
					"identity_mac_str": "00:11:22:33:44",
				},
			}),
		},
		"filter $neq": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$ne",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().MustNot(M{
				"match": M{
					"identity_mac_str": "00:11:22:33:44",
				},
			}),
		},
		"filter $gt": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$gt",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"range": M{
					"identity_mac_str": M{
						"gt": "00:11:22:33:44",
					},
				},
			}),
		},
		"filter $gte": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$gte",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"range": M{
					"identity_mac_str": M{
						"gte": "00:11:22:33:44",
					},
				},
			}),
		},
		"filter $lt": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$lt",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"range": M{
					"identity_mac_str": M{
						"lt": "00:11:22:33:44",
					},
				},
			}),
		},
		"filter $lte": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$lte",
						Value:     "00:11:22:33:44",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"range": M{
					"identity_mac_str": M{
						"lte": "00:11:22:33:44",
					},
				},
			}),
		},
		"filter $in": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$in",
						Value:     []string{"00:11:22:33:44"},
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"terms": M{
					"identity_mac_str": []string{"00:11:22:33:44"},
				},
			}),
		},
		"filter $nin": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$nin",
						Value:     []string{"00:11:22:33:44"},
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().MustNot(M{
				"terms": M{
					"identity_mac_str": []string{"00:11:22:33:44"},
				},
			}),
		},
		"filter $exists": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$exists",
						Value:     true,
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"bool": M{
					"minimum_should_match": 1,
					"should": S{
						M{
							"exists": M{
								"field": "identity_mac_str",
							},
						},
						M{
							"exists": M{
								"field": "identity_mac_num",
							},
						},
						M{
							"exists": M{
								"field": "identity_mac_bool",
							},
						},
					},
				},
			}),
		},
		"filter $regex": {
			inParams: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$regex",
						Value:     "00:.*",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"regexp": M{
					"identity_mac_str": "00:.*",
				},
			}),
		},
		"sort": {
			inParams: SearchParams{
				Sort: []SortCriteria{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().WithSort(M{
				"identity_mac_str": M{
					"order":         "asc",
					"unmapped_type": "keyword",
				},
			}).WithSort(M{
				"identity_mac_num": M{
					"order":         "asc",
					"unmapped_type": "double",
				},
			}),
		},
		"attributes": {
			inParams: SearchParams{
				Attributes: []SelectAttribute{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
					},
				},
				Page:    defaultPage,
				PerPage: defaultPerPage,
			},
			outQuery: NewQuery().With(map[string]interface{}{
				"_source": false,
				"fields": []string{
					"identity_mac_str",
					"identity_mac_num",
					"identity_mac_bool",
					"id",
					FieldNameCheckIn,
				},
			}),
		},
		"device IDs": {
			inParams: SearchParams{
				DeviceIDs: []string{"1", "2"},
				Page:      defaultPage,
				PerPage:   defaultPerPage,
			},
			outQuery: NewQuery().Must(M{
				"terms": M{
					"id": []string{"1", "2"},
				},
			}),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			query, err := BuildQuery(tc.inParams)
			if tc.outErr != nil {
				assert.Equal(t, tc.outErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outQuery, query)
			}

			_, err = json.Marshal(query)
			assert.Nil(t, nil)
		})
	}
}
