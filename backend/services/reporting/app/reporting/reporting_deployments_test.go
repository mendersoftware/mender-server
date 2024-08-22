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

package reporting

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/reporting/model"
	mstore "github.com/mendersoftware/mender-server/services/reporting/store/mocks"
)

func TestAggregateDeployments(t *testing.T) {
	const tenantID = "tenant_id"
	t.Parallel()
	type testCase struct {
		Name string

		Params       *model.AggregateDeploymentsParams
		SearchParams *model.DeploymentsSearchParams
		Store        func(*testing.T, testCase) *mstore.Store
		Mapping      model.Mapping

		Result []model.DeviceAggregation
		Error  error
	}
	testCases := []testCase{{
		Name: "ok",

		Params: &model.AggregateDeploymentsParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Aggregations: []model.DeploymentsAggregationTerm{
				{
					Name:      "aggr",
					Attribute: "attr",
				},
			},
			DeploymentGroups: []string{"foo", "bar"},
			TenantID:         tenantID,
		},
		SearchParams: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			DeploymentGroups: []string{"foo", "bar"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.SearchParams)
			q.Must(model.M{
				"term": model.M{
					model.FieldNameTenantID: tenantID,
				},
			})
			aggrs, _ := model.BuildDeploymentsAggregations(self.Params.Aggregations)
			q = q.WithSize(0).With(map[string]interface{}{
				"aggs": aggrs,
			})
			store.On("AggregateDeployments", contextMatcher, q).
				Return(model.M{
					"aggregations": map[string]interface{}{
						"aggr": map[string]interface{}{
							"sum_other_doc_count": float64(0),
							"buckets": []interface{}{
								map[string]interface{}{
									"key":       "group1",
									"doc_count": float64(5),
								},
								map[string]interface{}{
									"key":       "group2",
									"doc_count": float64(4),
								},
							},
						},
					},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo", "inventory/attr"},
		},
		Result: []model.DeviceAggregation{
			{
				Name: "aggr",
				Items: []model.DeviceAggregationItem{
					{
						Key:   "group1",
						Count: 5,
					},
					{
						Key:   "group2",
						Count: 4,
					},
				},
			},
		},
	}, {
		Name: "ok, subaggregations",

		Params: &model.AggregateDeploymentsParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Aggregations: []model.DeploymentsAggregationTerm{
				{
					Name:      "aggr",
					Attribute: "attr",
					Aggregations: []model.DeploymentsAggregationTerm{
						{
							Name:      "subaggr",
							Attribute: "foo",
						},
					},
				},
			},
			TenantID: tenantID,
		},
		SearchParams: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.SearchParams)
			q.Must(model.M{
				"term": model.M{
					model.FieldNameTenantID: tenantID,
				},
			})
			aggrs, _ := model.BuildDeploymentsAggregations(self.Params.Aggregations)
			q = q.WithSize(0).With(map[string]interface{}{
				"aggs": aggrs,
			})
			store.On("AggregateDeployments", contextMatcher, q).
				Return(model.M{
					"aggregations": map[string]interface{}{
						"aggr": map[string]interface{}{
							"sum_other_doc_count": float64(0),
							"buckets": []interface{}{
								map[string]interface{}{
									"key":       "group1",
									"doc_count": float64(5),
									"subaggr": map[string]interface{}{
										"sum_other_doc_count": float64(0),
										"buckets": []interface{}{
											map[string]interface{}{
												"key":       "v1",
												"doc_count": float64(5),
											},
										},
									},
								},
								map[string]interface{}{
									"key":       "group2",
									"doc_count": float64(4),
									"subaggr": map[string]interface{}{
										"sum_other_doc_count": float64(0),
										"buckets": []interface{}{
											map[string]interface{}{
												"key":       "v1",
												"doc_count": float64(4),
											},
										},
									},
								},
							},
						},
					},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo", "inventory/attr"},
		},
		Result: []model.DeviceAggregation{
			{
				Name: "aggr",
				Items: []model.DeviceAggregationItem{
					{
						Key:   "group1",
						Count: 5,
						Aggregations: []model.DeviceAggregation{{
							Name: "subaggr",
							Items: []model.DeviceAggregationItem{
								{
									Key:   "v1",
									Count: 5,
								},
							},
						}},
					},
					{
						Key:   "group2",
						Count: 4,
						Aggregations: []model.DeviceAggregation{{
							Name: "subaggr",
							Items: []model.DeviceAggregationItem{
								{
									Key:   "v1",
									Count: 4,
								},
							},
						}},
					},
				},
			},
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			var store *mstore.Store
			if tc.Store == nil {
				store = new(mstore.Store)
			} else {
				store = tc.Store(t, tc)
			}
			defer store.AssertExpectations(t)

			ds := &mstore.DataStore{}

			ds.On("UpdateAndGetMapping",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tenantID,
				[]string{"foo", "attr"},
			).Return(&tc.Mapping, nil).Once()

			ds.On("GetMapping",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tenantID,
			).Return(&tc.Mapping, nil).Once()

			app := NewApp(store, ds)
			res, err := app.AggregateDeployments(context.Background(), tc.Params)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Result, res)
			}
		})
	}
}

func TestSearchDeployments(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		Params  *model.DeploymentsSearchParams
		Store   func(*testing.T, testCase) *mstore.Store
		Mapping model.Mapping

		Result     []model.Deployment
		TotalCount int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok",

		Params: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "foo",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			q = q.Must(model.M{"terms": model.M{model.FieldNameDeviceID: self.Params.DeviceIDs}})
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"_source": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
					}}},
					"total": map[string]interface{}{
						"value": float64(1),
					}},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo"},
		},
		TotalCount: 1,
		Result: []model.Deployment{{
			ID:       "194d1060-1717-44dc-a783-00038f4a8013",
			TenantID: "123456789012345678901234",
		}},
	}, {
		Name: "ok with deployment_ids",

		Params: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "foo",
				Order:     "desc",
			}},
			DeploymentIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			q = q.Must(model.M{"terms": model.M{model.FieldNameDeploymentID: self.Params.DeploymentIDs}})
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"_source": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
					}}},
					"total": map[string]interface{}{
						"value": float64(1),
					}},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo"},
		},
		TotalCount: 1,
		Result: []model.Deployment{{
			ID:       "194d1060-1717-44dc-a783-00038f4a8013",
			TenantID: "123456789012345678901234",
		}},
	}, {
		Name: "ok with tenant_id",

		Params: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "foo",
				Order:     "desc",
			}},
			TenantID: "194d1060-1717-44dc-a783-00038f4a8013",
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			q = q.Must(model.M{"term": model.M{model.FieldNameTenantID: self.Params.TenantID}})
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"_source": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
					}}},
					"total": map[string]interface{}{
						"value": float64(1),
					}},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo"},
		},
		TotalCount: 1,
		Result: []model.Deployment{{
			ID:       "194d1060-1717-44dc-a783-00038f4a8013",
			TenantID: "123456789012345678901234",
		}},
	}, {
		Name: "ok with attributes",

		Params: &model.DeploymentsSearchParams{
			Attributes: []model.DeploymentsSelectAttribute{{
				Attribute: "foo",
			}},
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Type:      "$eq",
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "foo",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			q = q.Must(model.M{"terms": model.M{model.FieldNameDeviceID: self.Params.DeviceIDs}})
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"fields": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
					}}},
					"total": map[string]interface{}{
						"value": float64(1),
					}},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo"},
		},
		TotalCount: 1,
		Result: []model.Deployment{{
			ID:       "194d1060-1717-44dc-a783-00038f4a8013",
			TenantID: "123456789012345678901234",
		}},
	}, {
		Name: "ok, empty result",

		Params: &model.DeploymentsSearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{
					"hits": map[string]interface{}{
						"hits": []interface{}{},
						"total": map[string]interface{}{
							"value": float64(0),
						},
					},
				}, nil)
			return store
		},
		Result: []model.Deployment{},
	}, {
		Name: "error, internal storage-layer error",

		Params: &model.DeploymentsSearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			store.On("SearchDeployments", contextMatcher, q).
				Return(nil, errors.New("internal error"))
			return store
		},
		Result: []model.Deployment{},
		Error:  errors.New("internal error"),
	}, {
		Name: "error, parsing elastic result",

		Params: &model.DeploymentsSearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildDeploymentsQuery(*self.Params)
			store.On("SearchDeployments", contextMatcher, q).
				Return(model.M{
					"hits": map[string]interface{}{
						"hits": []interface{}{},
						"total": map[string]interface{}{
							"value": "doh!",
						},
					},
				}, nil)
			return store
		},
		Result: []model.Deployment{},
		Error:  errors.New("can't process total hits value"),
	}, {
		Name: "error, invalid search parameters",

		Params: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "foo",
				Value:     true,
				Type:      "$useyourimagination",
			}},
		},
		Mapping: model.Mapping{
			TenantID:  "",
			Inventory: []string{"inventory/foo"},
		},
		Error: errors.New("filter type not supported"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			var store *mstore.Store
			if tc.Store == nil {
				store = new(mstore.Store)
			} else {
				store = tc.Store(t, tc)
			}
			defer store.AssertExpectations(t)

			ds := &mstore.DataStore{}

			ds.On("UpdateAndGetMapping",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				"",
				[]string{"foo"},
			).Return(&tc.Mapping, nil).Once()

			ds.On("GetMapping",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				"",
			).Return(&tc.Mapping, nil).Once()

			app := NewApp(store, ds)
			res, cnt, err := app.SearchDeployments(context.Background(), tc.Params)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.TotalCount, cnt)
				assert.Equal(t, tc.Result, res)
			}
		})
	}
}
