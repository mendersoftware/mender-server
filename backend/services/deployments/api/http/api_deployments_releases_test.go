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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	mapp "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	dmodel "github.com/mendersoftware/mender-server/services/deployments/model"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	store_mocks "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
	deployments_testing "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	mt "github.com/mendersoftware/mender-server/pkg/testing"
)

func TestGetReleases(t *testing.T) {
	testCases := map[string]struct {
		filter        *dmodel.ReleaseOrImageFilter
		storeReleases []dmodel.Release
		storeErr      error
		checker       mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ReleaseOrImageFilter{},
			storeReleases: []dmodel.Release{
				{
					Artifacts: []model.Image{
						{
							Id: "1",
							ImageMeta: &model.ImageMeta{
								Description: "description",
							},

							ArtifactMeta: &model.ArtifactMeta{
								Name:                  "App1 v1.0",
								DeviceTypesCompatible: []string{"bar", "baz"},
								Updates:               []model.Update{},
							},
						},
					},
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.ReleaseV1{
					{
						Artifacts: []model.Image{
							{
								Id: "1",
								ImageMeta: &model.ImageMeta{
									Description: "description",
								},

								ArtifactMeta: &model.ArtifactMeta{
									Name:                  "App1 v1.0",
									DeviceTypesCompatible: []string{"bar", "baz"},
									Updates:               []model.Update{},
								},
							},
						},
					},
				}),
		},
		"ok, empty": {
			filter:        &dmodel.ReleaseOrImageFilter{},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.Release{}),
		},
		"ok, filter": {
			filter:        &dmodel.ReleaseOrImageFilter{Name: "foo"},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.Release{}),
		},
		"error: generic": {
			filter:        &dmodel.ReleaseOrImageFilter{},
			storeReleases: nil,
			storeErr:      errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error")),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			store := &store_mocks.DataStore{}

			store.On("GetReleases", deployments_testing.ContextMatcher(), tc.filter).
				Return(tc.storeReleases, len(tc.storeReleases), tc.storeErr)

			fileStorage := &fs_mocks.ObjectStorage{}

			restView := new(view.RESTView)
			app := app.NewDeployments(store, fileStorage, 0, false)

			c := NewDeploymentsApiHandlers(store, restView, app)

			api := deployments_testing.SetUpTestApi("/api/management/v1/deployments/releases", rest.Get, c.GetReleases)

			reqUrl := "http://1.2.3.4/api/management/v1/deployments/releases"

			if tc.filter != nil {
				reqUrl += "?name=" + tc.filter.Name
			}

			req := test.MakeSimpleRequest("GET",
				reqUrl,
				nil)

			req.Header.Add(requestid.RequestIdHeader, "test")

			recorded := test.RunRequest(t, api, req)

			mt.CheckResponse(t, tc.checker, recorded)
		})
	}
}

func TestGetReleasesFilter(t *testing.T) {
	testCases := map[string]struct {
		queryString string
		version     listReleasesVersion
		paginated   bool
		filter      *dmodel.ReleaseOrImageFilter
	}{
		"ok, empty": {
			version: listReleasesV1,
			filter:  &dmodel.ReleaseOrImageFilter{},
		},
		"ok, name": {
			queryString: "name=foo",
			version:     listReleasesV1,
			filter:      &dmodel.ReleaseOrImageFilter{Name: "foo"},
		},
		"ok, description": {
			queryString: "description=foo",
			version:     listReleasesV1,
			filter:      &dmodel.ReleaseOrImageFilter{Description: "foo"},
		},
		"ok, device type": {
			queryString: "device_type=foo",
			version:     listReleasesV1,
			filter:      &dmodel.ReleaseOrImageFilter{DeviceType: "foo"},
		},
		"ok, paginated, empty": {
			paginated: true,
			version:   listReleasesV1,
			filter: &dmodel.ReleaseOrImageFilter{
				Page:    1,
				PerPage: DefaultPerPage,
			},
		},
		"ok, paginated, name": {
			queryString: "name=foo",
			version:     listReleasesV1,
			paginated:   true,
			filter: &dmodel.ReleaseOrImageFilter{
				Name:    "foo",
				Page:    1,
				PerPage: DefaultPerPage,
			},
		},
		"ok, paginated, full options": {
			queryString: "name=foo&page=2&per_page=200&sort=name:asc",
			version:     listReleasesV1,
			paginated:   true,
			filter: &dmodel.ReleaseOrImageFilter{
				Name:    "foo",
				Page:    2,
				PerPage: 200,
				Sort:    "name:asc",
			},
		},
		"ok, paginated, per page too high": {
			queryString: "per_page=10000000",
			version:     listReleasesV1,
			paginated:   true,
			filter: &dmodel.ReleaseOrImageFilter{
				Page:    1,
				PerPage: DefaultPerPage,
			},
		},
		"ok, v2, name": {
			queryString: "name=foo",
			version:     listReleasesV2,
			filter:      &dmodel.ReleaseOrImageFilter{Name: "foo"},
		},
		"ok, v2, tags, name": {
			queryString: "tag=foo&tag=bar",
			version:     listReleasesV2,
			filter: &dmodel.ReleaseOrImageFilter{
				Tags: []string{"foo", "bar"},
			},
		},
		"ok, v2, tags, name, case": {
			queryString: "tag=fOO&tag=bAr",
			version:     listReleasesV2,
			filter: &dmodel.ReleaseOrImageFilter{
				Tags: []string{"foo", "bar"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			reqUrl := "http://1.2.3.4/api/management/v1/deployments/releases"
			req := &rest.Request{
				Request: test.MakeSimpleRequest("GET", reqUrl+"?"+tc.queryString, nil),
			}
			req.Header.Add(requestid.RequestIdHeader, "test")
			out := getReleaseOrImageFilter(req, tc.version, tc.paginated)
			assert.Equal(t, out, tc.filter)
		})
	}
}

func TestListReleases(t *testing.T) {
	testCases := map[string]struct {
		filter        *dmodel.ReleaseOrImageFilter
		storeReleases []dmodel.Release
		storeErr      error
		checker       mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{
				{
					Artifacts: []model.Image{
						{
							Id: "1",
							ImageMeta: &model.ImageMeta{
								Description: "description",
							},

							ArtifactMeta: &model.ArtifactMeta{
								Name:                  "App1 v1.0",
								DeviceTypesCompatible: []string{"bar", "baz"},
								Updates:               []model.Update{},
							},
						},
					},
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.ReleaseV1{
					{
						Artifacts: []model.Image{
							{
								Id: "1",
								ImageMeta: &model.ImageMeta{
									Description: "description",
								},

								ArtifactMeta: &model.ArtifactMeta{
									Name:                  "App1 v1.0",
									DeviceTypesCompatible: []string{"bar", "baz"},
									Updates:               []model.Update{},
								},
							},
						},
					},
				}),
		},
		"ok, empty": {
			filter:        &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.ReleaseV1{}),
		},
		"ok, filter": {
			filter:        &dmodel.ReleaseOrImageFilter{Name: "foo", Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.ReleaseV1{}),
		},
		"error: generic": {
			filter:        &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: nil,
			storeErr:      errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error")),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			store := &store_mocks.DataStore{}

			store.On("GetReleases", deployments_testing.ContextMatcher(), tc.filter).
				Return(tc.storeReleases, len(tc.storeReleases), tc.storeErr)

			fileStorage := &fs_mocks.ObjectStorage{}

			restView := new(view.RESTView)
			app := app.NewDeployments(store, fileStorage, 0, false)

			c := NewDeploymentsApiHandlers(store, restView, app)

			api := deployments_testing.SetUpTestApi("/api/management/v1/deployments/releases/list", rest.Get, c.ListReleases)

			reqUrl := "http://1.2.3.4/api/management/v1/deployments/releases/list"

			if tc.filter != nil {
				reqUrl += "?name=" + tc.filter.Name
			}

			req := test.MakeSimpleRequest("GET",
				reqUrl,
				nil)

			req.Header.Add(requestid.RequestIdHeader, "test")

			recorded := test.RunRequest(t, api, req)

			mt.CheckResponse(t, tc.checker, recorded)
		})
	}
}

func TestListReleasesV2(t *testing.T) {
	testCases := map[string]struct {
		filter        *dmodel.ReleaseOrImageFilter
		storeReleases []dmodel.Release
		storeErr      error
		checker       mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{
				{
					Artifacts: []model.Image{
						{
							Id: "1",
							ImageMeta: &model.ImageMeta{
								Description: "description",
							},

							ArtifactMeta: &model.ArtifactMeta{
								Name:                  "App1 v1.0",
								DeviceTypesCompatible: []string{"bar", "baz"},
								Updates:               []model.Update{},
							},
						},
					},
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.Release{
					{
						Artifacts: []model.Image{
							{
								Id: "1",
								ImageMeta: &model.ImageMeta{
									Description: "description",
								},

								ArtifactMeta: &model.ArtifactMeta{
									Name:                  "App1 v1.0",
									DeviceTypesCompatible: []string{"bar", "baz"},
									Updates:               []model.Update{},
								},
							},
						},
					},
				}),
		},
		"ok, empty": {
			filter:        &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.Release{}),
		},
		"ok, filter": {
			filter:        &dmodel.ReleaseOrImageFilter{Name: "foo", Page: 1, PerPage: 20},
			storeReleases: []dmodel.Release{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]dmodel.Release{}),
		},
		"error: generic": {
			filter:        &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			storeReleases: nil,
			storeErr:      errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error")),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			store := &store_mocks.DataStore{}

			store.On("GetReleases", deployments_testing.ContextMatcher(), tc.filter).
				Return(tc.storeReleases, len(tc.storeReleases), tc.storeErr)

			fileStorage := &fs_mocks.ObjectStorage{}

			restView := new(view.RESTView)
			app := app.NewDeployments(store, fileStorage, 0, false)

			c := NewDeploymentsApiHandlers(store, restView, app)

			api := deployments_testing.SetUpTestApi("/api/management/v2/deployments/releases", rest.Get, c.ListReleasesV2)

			reqUrl := "http://1.2.3.4/api/management/v2/deployments/releases"

			if tc.filter != nil {
				reqUrl += "?name=" + tc.filter.Name
			}

			req := test.MakeSimpleRequest("GET",
				reqUrl,
				nil)

			req.Header.Add(requestid.RequestIdHeader, "test")

			recorded := test.RunRequest(t, api, req)

			mt.CheckResponse(t, tc.checker, recorded)
		})
	}
}

func TestPutReleaseTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		App func(t *testing.T, self *testCase) *mapp.App
		*http.Request

		StatusCode int
	}

	testCases := []testCase{{
		Name: "ok",

		Request: func() *http.Request {
			b, _ := json.Marshal(model.Tags{"one", "one", "two", "three"})

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			expectedTags := model.Tags{"one", "two", "three"}
			appie.On("ReplaceReleaseTags",
				contextMatcher(),
				"release-mc-release-face",
				expectedTags).
				Return(nil)
			return appie
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error/internal",

		Request: func() *http.Request {
			b, _ := json.Marshal(model.Tags{"one", "two", "three"})

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			expectedTags := model.Tags{"one", "two", "three"}
			appie.On("ReplaceReleaseTags",
				contextMatcher(),
				"release-mc-release-face",
				expectedTags).
				Return(errors.New("internal error"))
			return appie
		},

		StatusCode: http.StatusInternalServerError,
	}, {
		Name: "error/too many unique tags",

		Request: func() *http.Request {
			b, _ := json.Marshal(model.Tags{"one", "two", "three"})

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			expectedTags := model.Tags{"one", "two", "three"}
			appie.On("ReplaceReleaseTags",
				contextMatcher(),
				"release-mc-release-face",
				expectedTags).
				Return(model.ErrTooManyUniqueTags)
			return appie
		},

		StatusCode: http.StatusConflict,
	}, {
		Name: "error/release not found",

		Request: func() *http.Request {
			b, _ := json.Marshal(model.Tags{"one", "two", "three"})

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			expectedTags := model.Tags{"one", "two", "three"}
			appie.On("ReplaceReleaseTags",
				contextMatcher(),
				"release-mc-release-face",
				expectedTags).
				Return(app.ErrReleaseNotFound)
			return appie
		},

		StatusCode: http.StatusNotFound,
	}, {
		Name: "error/too many tags",

		Request: func() *http.Request {
			tags := make(model.Tags, model.TagsMaxPerRelease+1)
			for i := range tags {
				tags[i] = model.Tag("field" + strconv.Itoa(i))
			}
			b, _ := json.Marshal(tags)

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusBadRequest,
	}, {
		Name: "ok/many duplicate tags",

		Request: func() *http.Request {
			tags := make(model.Tags, model.TagsMaxPerRelease+1)
			for i := range tags {
				tags[i] = model.Tag("field")
			}
			b, _ := json.Marshal(tags)

			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader(b),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			expectedTags := model.Tags{"field"}
			appie.On("ReplaceReleaseTags",
				contextMatcher(),
				"release-mc-release-face",
				expectedTags).
				Return(nil)
			return appie
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error/malformed JSON",

		Request: func() *http.Request {
			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "release-mc-release-face")),
				bytes.NewReader([]byte("not json")),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusBadRequest,
	}, {
		Name: "error/empty release name",

		Request: func() *http.Request {
			req, _ := http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseTags, "#name", "")),
				bytes.NewReader([]byte("[]")),
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusNotFound,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			appie := tc.App(t, &tc)
			defer appie.AssertExpectations(t)

			handlers := NewDeploymentsApiHandlers(nil, &view.RESTView{}, appie)
			routes := ReleasesRoutes(handlers)
			router, _ := rest.MakeRouter(routes...)
			api := rest.NewApi()
			api.SetApp(router)
			handler := api.MakeHandler()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tc.Request)

			rsp := w.Result()
			assert.Equal(t, tc.StatusCode, rsp.StatusCode,
				"unexpected status code from request")
		})
	}
}

func TestListReleaseTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		App func(t *testing.T, self *testCase) *mapp.App
		*http.Request

		StatusCode int
		Tags       model.Tags
	}

	testCases := []testCase{{
		Name: "ok",

		Request: func() *http.Request {
			req, _ := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseAllTags,
						"#name", "release-mc-release-face"),
				),
				nil,
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			appie.On("ListReleaseTags",
				contextMatcher()).
				Return(self.Tags, nil)
			return appie
		},

		StatusCode: http.StatusOK,
		Tags:       model.Tags{"bar", "baz", "foo"},
	}, {
		Name: "error/internal",

		Request: func() *http.Request {
			req, _ := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseAllTags,
						"#name", "release-mc-release-face"),
				),
				nil,
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			appie.On("ListReleaseTags",
				contextMatcher()).
				Return(nil, errors.New("internal error"))
			return appie
		},

		StatusCode: http.StatusInternalServerError,
	}, {
		Name: "error/internal",

		Request: func() *http.Request {
			req, _ := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://localhost:1234%s",
					strings.ReplaceAll(ApiUrlManagementV2ReleaseAllTags,
						"#name", "release-mc-release-face"),
				),
				nil,
			)
			return req
		}(),

		App: func(t *testing.T, self *testCase) *mapp.App {
			appie := new(mapp.App)
			appie.On("ListReleaseTags",
				contextMatcher()).
				Return(nil, errors.New("internal error"))
			return appie
		},

		StatusCode: http.StatusInternalServerError,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			appie := tc.App(t, &tc)
			defer appie.AssertExpectations(t)

			handlers := NewDeploymentsApiHandlers(nil, &view.RESTView{}, appie)
			routes := ReleasesRoutes(handlers)
			router, _ := rest.MakeRouter(routes...)
			api := rest.NewApi()
			api.SetApp(router)
			handler := api.MakeHandler()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tc.Request)

			rsp := w.Result()
			assert.Equal(t, tc.StatusCode, rsp.StatusCode,
				"unexpected status code from request")
			if tc.Tags != nil {
				var actual model.Tags
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				if assert.NoError(t, err, "unexpected request body") {
					assert.Equal(t, tc.Tags, actual)
				}
			}
		})
	}
}

func TestGetReleasesUpdateTypes(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		App func(t *testing.T, self *testCase) *mapp.App
		*http.Request

		StatusCode int
		Types      []string
	}

	testCases := []testCase{
		{
			Name: "ok",

			Request: func() *http.Request {
				req, _ := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("http://localhost:1234%s",
						ApiUrlManagementV2ReleaseAllUpdateTypes,
					),
					nil,
				)
				return req
			}(),

			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("GetReleasesUpdateTypes",
					contextMatcher()).
					Return(self.Types, nil)
				return appie
			},

			StatusCode: http.StatusOK,
			Types:      []string{"bar", "baz", "foo"},
		},
		{
			Name: "error/internal",

			Request: func() *http.Request {
				req, _ := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("http://localhost:1234%s",
						ApiUrlManagementV2ReleaseAllUpdateTypes,
					),
					nil,
				)
				return req
			}(),

			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("GetReleasesUpdateTypes",
					contextMatcher()).
					Return([]string{}, errors.New("internal"))
				return appie
			},

			StatusCode: http.StatusInternalServerError,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			appie := tc.App(t, &tc)
			defer appie.AssertExpectations(t)

			handlers := NewDeploymentsApiHandlers(nil, &view.RESTView{}, appie)
			routes := ReleasesRoutes(handlers)
			router, _ := rest.MakeRouter(routes...)
			api := rest.NewApi()
			api.SetApp(router)
			handler := api.MakeHandler()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tc.Request)

			rsp := w.Result()
			assert.Equal(t, tc.StatusCode, rsp.StatusCode,
				"unexpected status code from request")
			if tc.Types != nil {
				var actual []string
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				if assert.NoError(t, err, "unexpected request body") {
					assert.Equal(t, tc.Types, actual)
				}
			}
		})
	}
}

func TestPatchRelease(t *testing.T) {
	t.Parallel()

	longReleaseNotes := make([]byte, model.NotesLengthMaximumCharacters+1)
	for i := range longReleaseNotes {
		longReleaseNotes[i] = '1'
	}

	type testCase struct {
		Name string

		App func(t *testing.T, self *testCase) *mapp.App
		*http.Request

		StatusCode int
	}

	testCases := []testCase{
		{
			Name: "ok",

			Request: func() *http.Request {
				data, _ := json.Marshal(model.ReleasePatch{Notes: "New Release and fixes 2023"})
				req, _ := http.NewRequest(
					http.MethodPatch,
					fmt.Sprintf("http://localhost:1234%s",
						strings.ReplaceAll(ApiUrlManagementV2ReleasesName,
							"#name", "release-mc-release-face"),
					),
					bytes.NewReader(data),
				)
				return req
			}(),

			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("UpdateRelease",
					contextMatcher(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("model.ReleasePatch"),
				).Return(nil)
				return appie
			},

			StatusCode: http.StatusNoContent,
		},
		{
			Name: "error/notes too long",

			Request: func() *http.Request {
				data, _ := json.Marshal(model.ReleasePatch{Notes: model.Notes(longReleaseNotes)})
				req, _ := http.NewRequest(
					http.MethodPatch,
					fmt.Sprintf("http://localhost:1234%s",
						strings.ReplaceAll(ApiUrlManagementV2ReleasesName,
							"#name", "release-mc-release-face"),
					),
					bytes.NewReader(data),
				)
				return req
			}(),

			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				return appie
			},

			StatusCode: http.StatusBadRequest,
		},
		{
			Name: "error/internal",

			Request: func() *http.Request {
				data, _ := json.Marshal(model.ReleasePatch{Notes: "New Release and fixes 2023"})
				req, _ := http.NewRequest(
					http.MethodPatch,
					fmt.Sprintf("http://localhost:1234%s",
						strings.ReplaceAll(ApiUrlManagementV2ReleasesName,
							"#name", "release-mc-release-face"),
					),
					bytes.NewReader(data),
				)
				return req
			}(),

			App: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("UpdateRelease",
					contextMatcher(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("model.ReleasePatch"),
				).Return(errors.New("internal error"))
				return appie
			},

			StatusCode: http.StatusInternalServerError,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			appie := tc.App(t, &tc)
			defer appie.AssertExpectations(t)

			handlers := NewDeploymentsApiHandlers(nil, &view.RESTView{}, appie)
			routes := ReleasesRoutes(handlers)
			router, _ := rest.MakeRouter(routes...)
			api := rest.NewApi()
			api.SetApp(router)
			handler := api.MakeHandler()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tc.Request)

			rsp := w.Result()
			assert.Equal(t, tc.StatusCode, rsp.StatusCode,
				"unexpected status code from request")
		})
	}
}

func TestDeleteReleases(t *testing.T) {
	type testCase struct {
		name         string
		releaseNames []string
		app          func(t *testing.T, self *testCase) *mapp.App
		checker      mt.ResponseChecker
	}
	testCases := []testCase{
		{
			name:         "ok",
			releaseNames: []string{"foo", "bar"},
			app: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("DeleteReleases",
					contextMatcher(),
					self.releaseNames,
				).Return([]string{}, nil)
				return appie
			},
			checker: mt.NewJSONResponse(
				http.StatusNoContent,
				nil,
				nil,
			),
		},
		{
			name:         "no release name",
			releaseNames: []string{},
			app: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				return appie
			},
			checker: mt.NewJSONResponse(
				http.StatusBadRequest,
				nil,
				deployments_testing.RestError(ErrReleaseNameNotProvided.Error()),
			),
		},
		{
			name:         "conflict",
			releaseNames: []string{"foo", "bar"},
			app: func(t *testing.T, self *testCase) *mapp.App {
				appie := new(mapp.App)
				appie.On("DeleteReleases",
					contextMatcher(),
					self.releaseNames,
				).Return([]string{"id1", "id2"}, nil)
				return appie
			},
			checker: mt.NewJSONResponse(
				http.StatusConflict,
				nil,
				model.ReleasesDeleteError{
					Error:             ErrReleaseUsedInActiveDeployment.Error(),
					RequestID:         "test",
					ActiveDeployments: []string{"id1", "id2"},
				},
			),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			restView := new(view.RESTView)
			appie := tc.app(t, &tc)
			defer appie.AssertExpectations(t)

			c := NewDeploymentsApiHandlers(nil, restView, appie)

			api := deployments_testing.SetUpTestApi(ApiUrlManagementV2Releases, rest.Delete, c.DeleteReleases)

			reqUrl := "http://1.2.3.4" + ApiUrlManagementV2Releases

			if len(tc.releaseNames) > 0 {
				reqUrl += "?"
				for i, n := range tc.releaseNames {
					if i > 0 {
						reqUrl += "&"
					}
					reqUrl += "name=" + n
				}
			}

			req := test.MakeSimpleRequest("DELETE",
				reqUrl,
				nil)

			req.Header.Add(requestid.RequestIdHeader, "test")

			recorded := test.RunRequest(t, api, req)

			mt.CheckResponse(t, tc.checker, recorded)
		})
	}
}
