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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rbac"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	mapp "github.com/mendersoftware/mender-server/services/reporting/app/reporting/mocks"
	"github.com/mendersoftware/mender-server/services/reporting/model"
)

func TestManagementAggregateDeployments(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		App    func(*testing.T, testCase) *mapp.App
		CTX    context.Context
		Params interface{} // *model.AggregateDeploymentsParams

		Code     int
		Response interface{}
	}
	testCases := []testCase{{
		Name: "ok",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("AggregateDeployments",
				contextMatcher,
				mock.MatchedBy(func(*model.AggregateDeploymentsParams) bool {
					return true
				})).
				Return(self.Response, nil)
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.AggregateDeploymentsParams{
			Aggregations: []model.DeploymentsAggregationTerm{
				{
					Name:      "mac",
					Attribute: "mac",
					Limit:     10,
				},
			},
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			TenantID: "123456789012345678901234",
		},

		Code:     http.StatusOK,
		Response: []model.DeviceAggregation{},
	}, {
		Name: "ok, with scope",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("AggregateDeployments",
				contextMatcher,
				mock.MatchedBy(func(*model.AggregateDeploymentsParams) bool {
					return true
				})).
				Return(self.Response, nil)
			return app
		},
		CTX: rbac.WithContext(identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		), &rbac.Scope{
			DeviceGroups: []string{"group1", "group2"},
		}),
		Params: &model.AggregateDeploymentsParams{
			Aggregations: []model.DeploymentsAggregationTerm{
				{
					Name:      "mac",
					Attribute: "mac",
					Limit:     10,
				},
			},
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			TenantID: "123456789012345678901234",
		},

		Code:     http.StatusOK,
		Response: []model.DeviceAggregation{},
	}, {
		Name: "error, malformed request body",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params:   &model.AggregateDeploymentsParams{},
		Code:     http.StatusBadRequest,
		Response: rest.Error{Err: "malformed request body: aggregations: cannot be blank."},
	}, {
		Name: "error, internal app error",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("AggregateDeployments",
				contextMatcher,
				mock.MatchedBy(func(*model.AggregateDeploymentsParams) bool {
					return true
				})).
				Return(nil, errors.New("internal error"))

			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.AggregateDeploymentsParams{
			Aggregations: []model.DeploymentsAggregationTerm{
				{
					Name:      "mac",
					Attribute: "mac",
					Limit:     10,
				},
			},
			Filters: []model.DeploymentsFilterPredicate{{
				Type:      "$eq",
				Attribute: "rootpwd",
				Value:     true,
			}},
		},

		Code:     http.StatusInternalServerError,
		Response: rest.Error{Err: "internal error"},
	}, {
		Name: "error, request identity not present",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX:    identity.WithContext(context.Background(), nil),
		Params: &model.AggregateDeploymentsParams{},

		Code:     http.StatusUnauthorized,
		Response: rest.Error{Err: "Authorization not present in header"},
	}, {
		Name: "error, malformed request body",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: map[string]string{
			"filters": "foo",
		},

		Code: http.StatusBadRequest,
		Response: rest.Error{
			Err: "malformed request body: json: " +
				"cannot unmarshal string into Go struct field " +
				"AggregateDeploymentsParams.filters of type []model.DeploymentsFilterPredicate",
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var app *mapp.App
			if tc.App == nil {
				app = new(mapp.App)
			} else {
				app = tc.App(t, tc)
			}
			defer app.AssertExpectations(t)
			router := NewRouter(app)

			b, _ := json.Marshal(tc.Params)
			req, _ := http.NewRequest(
				http.MethodPost,
				URIManagement+URIDeploymentsAggregate,
				bytes.NewReader(b),
			)
			if id := identity.FromContext(tc.CTX); id != nil {
				req.Header.Set("Authorization", "Bearer "+GenerateJWT(*id))
			}
			if scope := rbac.FromContext(tc.CTX); scope != nil {
				req.Header.Set(rbac.ScopeHeader, strings.Join(scope.DeviceGroups, ","))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code)

			switch res := tc.Response.(type) {
			case []model.DeviceAggregation:
				b, _ := json.Marshal(res)
				assert.JSONEq(t, string(b), w.Body.String())

			case rest.Error:
				var actual rest.Error
				dec := json.NewDecoder(w.Body)
				dec.DisallowUnknownFields()
				err := dec.Decode(&actual)
				if assert.NoError(t, err, "response schema did not match expected rest.Error") {
					assert.EqualError(t, res, actual.Error())
				}

			case nil:
				assert.Empty(t, w.Body.String())

			default:
				panic("[TEST ERR] Dunno what to compare!")
			}

		})
	}
}

func time2ptr(t time.Time) *time.Time {
	return &t
}

func TestManagementSearchDeployments(t *testing.T) {
	t.Parallel()
	var newSearchParamMatcher = func(expected *model.DeploymentsSearchParams) interface{} {
		return mock.MatchedBy(func(actual *model.DeploymentsSearchParams) bool {
			if expected.Page <= 0 {
				expected.Page = ParamPageDefault
			}
			if expected.PerPage <= 0 {
				expected.PerPage = ParamPerPageDefault
			}
			if assert.NotNil(t, actual) {
				return assert.Equal(t, *expected, *actual)
			}
			return false
		})
	}
	type testCase struct {
		Name string

		App    func(*testing.T, testCase) *mapp.App
		CTX    context.Context
		Params interface{} // *model.SearchParams

		Code     int
		Response interface{}
	}
	testCases := []testCase{{
		Name: "ok",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("SearchDeployments",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.DeploymentsSearchParams))).
				Return(self.Response, 0, nil)
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.DeploymentsSearchParams{
			PerPage: 10,
			Page:    2,
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "ip4",
				Order:     model.SortOrderAsc,
			}},
			TenantID: "123456789012345678901234",
		},

		Code: http.StatusOK,
		Response: []model.Deployment{{
			ID:             "5975e1e6-49a6-4218-a46d-f181154a98cc",
			DeviceCreated:  time2ptr(time.Now().Add(-time.Hour)),
			DeviceFinished: time2ptr(time.Now().Add(-time.Minute)),
		}, {
			ID:             "83bce0e4-c4c0-4995-b8b7-f056da7fc8f6",
			DeviceCreated:  time2ptr(time.Now().Add(-2 * time.Hour)),
			DeviceFinished: time2ptr(time.Now().Add(-5 * time.Minute)),
		}},
	}, {
		Name: "ok, empty result",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("SearchDeployments",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.DeploymentsSearchParams))).
				Return([]model.Deployment{}, 0, nil)
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.DeploymentsSearchParams{
			TenantID: "123456789012345678901234",
		},

		Code:     http.StatusOK,
		Response: []model.Deployment{},
	}, {
		Name: "ok, with scope, empty results",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("SearchDeployments",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.DeploymentsSearchParams))).
				Return([]model.Deployment{}, 0, nil)
			return app
		},
		CTX: rbac.WithContext(identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		), &rbac.Scope{
			DeviceGroups: []string{"group1", "group2"},
		}),
		Params: &model.DeploymentsSearchParams{
			DeploymentGroups: []string{"group1", "group2"},
			TenantID:         "123456789012345678901234",
		},

		Code:     http.StatusOK,
		Response: []model.Deployment{},
	}, {
		Name: "error, malformed request body",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.DeploymentsSearchParams{
			Filters: []model.DeploymentsFilterPredicate{{
				Type:      "$maybethiswillfindsomethinginterresting",
				Attribute: "rootpwd",
				Value:     true,
			}},
			TenantID: "123456789012345678901234",
		},
		Code:     http.StatusBadRequest,
		Response: rest.Error{Err: "malformed request body: type: must be a valid value."},
	}, {
		Name: "error, internal app error",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("SearchDeployments",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.DeploymentsSearchParams))).
				Return(nil, 0, errors.New("internal error"))
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.DeploymentsSearchParams{
			PerPage: 10,
			Page:    2,
			Filters: []model.DeploymentsFilterPredicate{{
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			Sort: []model.DeploymentsSortCriteria{{
				Attribute: "ip4",
				Order:     model.SortOrderAsc,
			}},
			TenantID: "123456789012345678901234",
		},

		Code:     http.StatusInternalServerError,
		Response: rest.Error{Err: "internal error"},
	}, {
		Name: "error, request identity not present",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX:    identity.WithContext(context.Background(), nil),
		Params: &model.DeploymentsSearchParams{},

		Code:     http.StatusUnauthorized,
		Response: rest.Error{Err: "Authorization not present in header"},
	}, {
		Name: "error, malformed request body",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: map[string]string{
			"filters": "foo",
		},

		Code: http.StatusBadRequest,
		Response: rest.Error{
			Err: "malformed request body: json: " +
				"cannot unmarshal string into Go struct field " +
				"DeploymentsSearchParams.filters of type []model.DeploymentsFilterPredicate",
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var app *mapp.App
			if tc.App == nil {
				app = new(mapp.App)
			} else {
				app = tc.App(t, tc)
			}
			defer app.AssertExpectations(t)
			router := NewRouter(app)

			b, _ := json.Marshal(tc.Params)
			req, _ := http.NewRequest(
				http.MethodPost,
				URIManagement+URIDeploymentsSearch,
				bytes.NewReader(b),
			)
			if id := identity.FromContext(tc.CTX); id != nil {
				req.Header.Set("Authorization", "Bearer "+GenerateJWT(*id))
			}
			if scope := rbac.FromContext(tc.CTX); scope != nil {
				req.Header.Set(rbac.ScopeHeader, strings.Join(scope.DeviceGroups, ","))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code)

			switch res := tc.Response.(type) {
			case []model.Deployment:
				b, _ := json.Marshal(res)
				assert.JSONEq(t, string(b), w.Body.String())

			case rest.Error:
				var actual rest.Error
				dec := json.NewDecoder(w.Body)
				dec.DisallowUnknownFields()
				err := dec.Decode(&actual)
				if assert.NoError(t, err, "response schema did not match expected rest.Error") {
					assert.EqualError(t, res, actual.Error())
				}

			case nil:
				assert.Empty(t, w.Body.String())

			default:
				panic("[TEST ERR] Dunno what to compare!")
			}

		})
	}
}
