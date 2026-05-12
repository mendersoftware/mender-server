// Copyright 2024 Northern.tech AS
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

package app

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
)

func TestReplaceReleaseTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		context.Context
		ReleaseName string
		Tags        model.Tags

		GetDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		Error error
	}
	testCases := []testCase{{
		Name: "ok",

		Context:     context.Background(),
		ReleaseName: "foobar",
		Tags:        model.Tags{"foo", "baz"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ReplaceReleaseTags", self.Context, self.ReleaseName, self.Tags).
				Return(nil)
			return ds
		},
	}, {
		Name: "error/not found",

		Context:     context.Background(),
		ReleaseName: "foobar",
		Tags:        model.Tags{"foo", "baz"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ReplaceReleaseTags", self.Context, self.ReleaseName, self.Tags).
				Return(store.ErrNotFound)
			return ds
		},
		Error: ErrReleaseNotFound,
	}, {
		Name: "error/too many unique keys",

		Context:     context.Background(),
		ReleaseName: "foobar",
		Tags:        model.Tags{"foo", "baz"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ReplaceReleaseTags", self.Context, self.ReleaseName, self.Tags).
				Return(model.ErrTooManyUniqueTags)
			return ds
		},
		Error: model.ErrTooManyUniqueTags,
	}, {
		Name: "error/internal error",

		Context:     context.Background(),
		ReleaseName: "foobar",
		Tags:        model.Tags{"foo", "baz"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ReplaceReleaseTags", self.Context, self.ReleaseName, self.Tags).
				Return(errors.New("internal error with sensitive info"))
			return ds
		},
		Error: ErrModelInternal,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.GetDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			err := app.ReplaceReleaseTags(tc.Context, tc.ReleaseName, tc.Tags)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListReleaseTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		context.Context

		GetDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		Tags  model.Tags
		Error error
	}
	testCases := []testCase{{
		Name: "ok",

		Context: context.Background(),
		Tags:    model.Tags{"field1", "field2"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ListReleaseTags", self.Context).
				Return(self.Tags, nil)
			return ds
		},
	}, {
		Name: "error/internal error",

		Context: context.Background(),

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("ListReleaseTags", self.Context).
				Return(nil, errors.New("internal error with sensitive info"))
			return ds
		},
		Error: ErrModelInternal,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.GetDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			tags, err := app.ListReleaseTags(tc.Context)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Tags, tags)
			}
		})
	}
}

func TestGetReleasesUpdateTypes(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		context.Context

		GetDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		Types []string
		Error error
	}
	testCases := []testCase{{
		Name: "ok",

		Context: context.Background(),
		Types:   []string{"field1", "field2"},

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetUpdateTypes", self.Context).
				Return(self.Types, nil)
			return ds
		},
	}, {
		Name: "error/internal error",

		Context: context.Background(),

		GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetUpdateTypes", self.Context).
				Return([]string{}, errors.New("internal error with sensitive info"))
			return ds
		},
		Error: ErrModelInternal,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.GetDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			tags, err := app.GetReleasesUpdateTypes(tc.Context)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Types, tags)
			}
		})
	}
}

func TestUpdateRelease(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		context.Context
		ReleaseName string
		Release     model.ReleasePatch

		GetDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			Context:     context.Background(),
			ReleaseName: "foobar",
			Release:     model.ReleasePatch{Notes: "New Release fixes 2023"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("UpdateRelease", self.Context, self.ReleaseName, self.Release).
					Return(nil)
				return ds
			},
		},
		{
			Name: "error/not found",

			Context:     context.Background(),
			ReleaseName: "foobar",
			Release:     model.ReleasePatch{Notes: "New Release fixes 2023"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("UpdateRelease", self.Context, self.ReleaseName, self.Release).
					Return(store.ErrNotFound)
				return ds
			},
			Error: ErrReleaseNotFound,
		},
		{
			Name: "error/internal error",

			Context:     context.Background(),
			ReleaseName: "foobar",
			Release:     model.ReleasePatch{Notes: "New Release fixes 2023"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("UpdateRelease", self.Context, self.ReleaseName, self.Release).
					Return(errors.New("internal error with sensitive info"))
				return ds
			},
			Error: ErrModelInternal,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.GetDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			err := app.UpdateRelease(tc.Context, tc.ReleaseName, tc.Release)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteReleases(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		context.Context
		ReleaseNames []string

		GetDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		ids   []string
		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			Context:      context.Background(),
			ReleaseNames: []string{"foo", "bar"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetActiveDeploymentIDsByArtifactNames", self.Context, self.ReleaseNames).
					Return([]string{}, nil)
				ds.On("DeleteImagesByNames", self.Context, self.ReleaseNames).
					Return(nil)
				ds.On("DeleteReleasesByNames", self.Context, self.ReleaseNames).
					Return(nil)
				return ds
			},
		},
		{
			Name: "release in active deployment",

			Context:      context.Background(),
			ReleaseNames: []string{"foo", "bar"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetActiveDeploymentIDsByArtifactNames", self.Context, self.ReleaseNames).
					Return([]string{"foo"}, nil)
				return ds
			},
			ids: []string{"foo"},
		},
		{
			Name: "error: get deployment IDs error",

			Context:      context.Background(),
			ReleaseNames: []string{"foo", "bar"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetActiveDeploymentIDsByArtifactNames", self.Context, self.ReleaseNames).
					Return([]string{}, errors.New("some error"))
				return ds
			},
			Error: errors.New("some error"),
		},
		{
			Name: "error: delete images error",

			Context:      context.Background(),
			ReleaseNames: []string{"foo", "bar"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetActiveDeploymentIDsByArtifactNames", self.Context, self.ReleaseNames).
					Return([]string{}, nil)
				ds.On("DeleteImagesByNames", self.Context, self.ReleaseNames).
					Return(errors.New("some error"))
				return ds
			},
			Error: errors.New("some error"),
		},
		{
			Name: "error: delete releases error",

			Context:      context.Background(),
			ReleaseNames: []string{"foo", "bar"},

			GetDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetActiveDeploymentIDsByArtifactNames", self.Context, self.ReleaseNames).
					Return([]string{}, nil)
				ds.On("DeleteImagesByNames", self.Context, self.ReleaseNames).
					Return(nil)
				ds.On("DeleteReleasesByNames", self.Context, self.ReleaseNames).
					Return(errors.New("some error"))
				return ds
			},
			Error: errors.New("some error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ds := tc.GetDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			ids, err := app.DeleteReleases(tc.Context, tc.ReleaseNames)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
			if len(tc.ids) > 0 {
				assert.Equal(t, tc.ids, ids)
			} else {
				assert.Len(t, tc.ids, 0)
			}
		})
	}
}

func TestGetRelease(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string

		context.Context
		releaseName string

		getDatabase func(t *testing.T, self *testCase) *mocks.DataStore

		release *model.Release
		err     error
	}
	testCases := []testCase{
		{
			name: "ok",

			Context:     context.Background(),
			releaseName: "foo",

			getDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetRelease", self.Context, self.releaseName).
					Return(self.release, nil)
				return ds
			},
			release: &model.Release{},
		},
		{
			name: "not found",

			Context:     context.Background(),
			releaseName: "foo",

			getDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetRelease", self.Context, self.releaseName).
					Return(nil, store.ErrNotFound)
				return ds
			},
			release: nil,
			err:     ErrReleaseNotFound,
		},
		{
			name: "database error",

			Context:     context.Background(),
			releaseName: "foo",

			getDatabase: func(t *testing.T, self *testCase) *mocks.DataStore {
				ds := new(mocks.DataStore)
				ds.On("GetRelease", self.Context, self.releaseName).
					Return(nil, errors.New("db error"))
				return ds
			},
			release: nil,
			err:     ErrModelInternal,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ds := tc.getDatabase(t, &tc)
			defer ds.AssertExpectations(t)

			app := NewDeployments(ds, nil, 0, false)

			release, err := app.GetRelease(tc.Context, tc.releaseName)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.release, release)
			}
		})
	}
}
