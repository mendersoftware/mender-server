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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	"github.com/mendersoftware/mender-server/services/reporting/model"
	mstore "github.com/mendersoftware/mender-server/services/reporting/store/mocks"
)

var contextMatcher = mock.MatchedBy(func(_ context.Context) bool { return true })

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		StoreErr     error
		DatastoreErr error

		Error error
	}{
		{
			Name: "ok",
		},
		{
			Name:     "ko, store",
			StoreErr: errors.New("error"),
			Error:    errors.New("error"),
		},
		{
			Name:         "ko, datastore",
			DatastoreErr: errors.New("error"),
			Error:        errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mstore.DataStore{}
			ds.On("Ping", ctx).Return(tc.DatastoreErr)

			store := &mstore.Store{}
			if tc.DatastoreErr == nil {
				store.On("Ping", ctx).Return(tc.StoreErr)
			}
			app := NewApp(store, ds)

			err := app.HealthCheck(ctx)
			if tc.Error != nil {
				assert.Equal(t, tc.Error, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetMapping(t *testing.T) {
	const tenantID = "tenant_id"
	ctx := context.Background()

	mapping := &model.Mapping{
		TenantID: tenantID,
	}

	ds := &mstore.DataStore{}
	ds.On("GetMapping", ctx, tenantID).Return(mapping, nil)

	app := NewApp(nil, ds)
	res, err := app.GetMapping(ctx, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, mapping, res)
}

func TestAggregateDevices(t *testing.T) {
	const tenantID = "tenant_id"
	t.Parallel()
	type testCase struct {
		Name string

		Params                 *model.AggregateParams
		MappedParams           *model.SearchParams
		MappedAggregatedParams []model.AggregationTerm
		Store                  func(*testing.T, testCase) *mstore.Store
		Mapping                model.Mapping

		Result []model.DeviceAggregation
		Error  error
	}
	testCases := []testCase{{
		Name: "ok",

		Params: &model.AggregateParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Aggregations: []model.AggregationTerm{
				{
					Name:      "aggr",
					Attribute: "attr",
					Scope:     "inventory",
				},
			},
			TenantID: tenantID,
		},
		MappedParams: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "attribute1",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
		},
		MappedAggregatedParams: []model.AggregationTerm{
			{
				Name:      "aggr",
				Attribute: "attribute2",
				Scope:     "inventory",
			},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.MappedParams)
			q.Must(model.M{
				"term": model.M{
					model.FieldNameTenantID: tenantID,
				},
			})
			aggrs, _ := model.BuildAggregations(self.MappedAggregatedParams)
			q = q.WithSize(0).With(map[string]interface{}{
				"aggs": aggrs,
			})
			store.On("AggregateDevices", contextMatcher, q).
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

		Params: &model.AggregateParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Aggregations: []model.AggregationTerm{
				{
					Name:      "aggr",
					Attribute: "attr",
					Scope:     "inventory",
					Aggregations: []model.AggregationTerm{
						{
							Name:      "subaggr",
							Attribute: "foo",
							Scope:     "inventory",
						},
					},
				},
			},
			TenantID: tenantID,
		},
		MappedParams: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "attribute1",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
		},
		MappedAggregatedParams: []model.AggregationTerm{
			{
				Name:      "aggr",
				Attribute: "attribute2",
				Scope:     "inventory",
				Aggregations: []model.AggregationTerm{
					{
						Name:      "subaggr",
						Attribute: "attribute1",
						Scope:     "inventory",
					},
				},
			},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.MappedParams)
			q.Must(model.M{
				"term": model.M{
					model.FieldNameTenantID: tenantID,
				},
			})
			aggrs, _ := model.BuildAggregations(self.MappedAggregatedParams)
			q = q.WithSize(0).With(map[string]interface{}{
				"aggs": aggrs,
			})
			store.On("AggregateDevices", contextMatcher, q).
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
			res, err := app.AggregateDevices(context.Background(), tc.Params)
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

func TestSearchDevices(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		Params       *model.SearchParams
		MappedParams *model.SearchParams
		Store        func(*testing.T, testCase) *mstore.Store
		Mapping      model.Mapping

		Result     []inventory.Device
		TotalCount int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok",

		Params: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Sort: []model.SortCriteria{{
				Attribute: "foo",
				Scope:     "inventory",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		MappedParams: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "attribute1",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Sort: []model.SortCriteria{{
				Attribute: "attribute1",
				Scope:     "inventory",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.MappedParams)
			q = q.Must(model.M{"terms": model.M{"id": self.Params.DeviceIDs}})
			store.On("SearchDevices", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"_source": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
						model.ToAttr("inventory", "attribute1", model.TypeStr): []string{"bar"},
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
		Result: []inventory.Device{{
			ID: "194d1060-1717-44dc-a783-00038f4a8013",
			Attributes: inventory.DeviceAttributes{{
				Name:  "foo",
				Value: []string{"bar"},
				Scope: "inventory",
			}},
		}},
	}, {
		Name: "ok with attributes",

		Params: &model.SearchParams{
			Attributes: []model.SelectAttribute{{
				Attribute: "foo",
				Scope:     "inventory",
			}},
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Sort: []model.SortCriteria{{
				Attribute: "foo",
				Scope:     "inventory",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		MappedParams: &model.SearchParams{
			Attributes: []model.SelectAttribute{{
				Attribute: "attribute1",
				Scope:     "inventory",
			}},
			Filters: []model.FilterPredicate{{
				Attribute: "attribute1",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Sort: []model.SortCriteria{{
				Attribute: "attribute1",
				Scope:     "inventory",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.MappedParams)
			q = q.Must(model.M{"terms": model.M{"id": self.Params.DeviceIDs}})
			store.On("SearchDevices", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"fields": map[string]interface{}{
						"id":        "194d1060-1717-44dc-a783-00038f4a8013",
						"tenant_id": "123456789012345678901234",
						model.ToAttr("inventory", "attribute1", model.TypeStr): []string{"bar"},
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
		Result: []inventory.Device{{
			ID: "194d1060-1717-44dc-a783-00038f4a8013",
			Attributes: inventory.DeviceAttributes{{
				Name:  "foo",
				Value: []string{"bar"},
				Scope: "inventory",
			}},
		}},
	}, {
		Name: "ok, empty result",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("SearchDevices", contextMatcher, q).
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
		Result: []inventory.Device{},
	}, {
		Name: "error, internal storage-layer error",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("SearchDevices", contextMatcher, q).
				Return(nil, errors.New("internal error"))
			return store
		},
		Result: []inventory.Device{},
		Error:  errors.New("internal error"),
	}, {
		Name: "error, parsing elastic result",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("SearchDevices", contextMatcher, q).
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
		Result: []inventory.Device{},
		Error:  errors.New("can't process total hits value"),
	}, {
		Name: "error, invalid search parameters",

		Params: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     true,
				Scope:     "inventory",
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
			res, cnt, err := app.SearchDevices(context.Background(), tc.Params)
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

func TestGetSearchableInvAttrs(t *testing.T) {
	const tenantID = "tenant_id"

	t.Parallel()
	type testCase struct {
		Name    string
		Store   func(*testing.T, testCase) *mstore.Store
		Mapping model.Mapping

		Result []model.FilterAttribute
		Error  error
	}
	testCases := []testCase{{
		Name: "ok",

		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			store.On("GetDevicesIndexMapping", contextMatcher, tenantID).
				Return(map[string]interface{}{
					"mappings": map[string]interface{}{
						"properties": map[string]interface{}{
							"inventory_attribute1_str": 1,
							"inventory_attribute2_str": 1,
							"system_attribute3_str":    1,
						},
					},
				}, nil)
			return store
		},
		Mapping: model.Mapping{
			TenantID: "",
			Inventory: []string{
				"inventory/foo",
				"inventory/bar",
			},
		},
		Result: []model.FilterAttribute{
			{
				Name:  "bar",
				Scope: "inventory",
				Count: 1,
			},
			{
				Name:  "foo",
				Scope: "inventory",
				Count: 1,
			},
			{
				Name:  "attribute3",
				Scope: "system",
				Count: 1,
			},
		},
	}, {
		Name: "ko, error in GetDevicesIndexMapping",

		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			store.On("GetDevicesIndexMapping", contextMatcher, tenantID).
				Return(nil, errors.New("error"))
			return store
		},
		Error: errors.New("error"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			store := tc.Store(t, tc)
			defer store.AssertExpectations(t)

			ds := &mstore.DataStore{}
			ds.On("GetMapping",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tenantID,
			).Return(&tc.Mapping, nil).Once()

			app := NewApp(store, ds)
			res, err := app.GetSearchableInvAttrs(context.Background(), tenantID)
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

func TestGetTime(t *testing.T) {
	const tenantID = "tenant_id"

	t.Parallel()
	now := time.Now().UTC()
	now = now.Truncate(time.Second).UTC()
	notNow := now.AddDate(-1, 0, 0)
	notNow = notNow.Truncate(time.Second)
	type testCase struct {
		Name string

		InputData      map[string]interface{}
		ExpectedResult *time.Time
		FieldName      string
	}
	testCases := []testCase{
		{
			Name: "field found",

			InputData: map[string]interface{}{
				"now": []interface{}{
					now.Format(time.RFC3339),
				},
				"not-now": []interface{}{
					&notNow,
				},
			},
			FieldName:      "now",
			ExpectedResult: &now,
		},
		{
			Name: "field not present",

			InputData: map[string]interface{}{
				"now2": []interface{}{
					now.Format(time.RFC3339),
				},
				"not-now2": []interface{}{
					&notNow,
				},
			},
			FieldName: "now",
		},
		{
			Name:      "unexpected element types",
			FieldName: "some-other",
			InputData: map[string]interface{}{
				"now":     &now,
				"not-now": &notNow,
			},
		},
		{
			Name:      "empty map",
			FieldName: "some-other",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			v := getTime(tc.InputData, tc.FieldName)
			assert.Equal(t, tc.ExpectedResult, v)
			if tc.ExpectedResult != nil {
				assert.NotNil(t, v)
				assert.Equal(t, *tc.ExpectedResult, *v)
			}
		})
	}
}
