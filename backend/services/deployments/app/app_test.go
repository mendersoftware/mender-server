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
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"

	inventory_mocks "github.com/mendersoftware/mender-server/services/deployments/client/inventory/mocks"
	reporting_mocks "github.com/mendersoftware/mender-server/services/deployments/client/reporting/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/client/workflows"
	workflows_mocks "github.com/mendersoftware/mender-server/services/deployments/client/workflows/mocks"
	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
	h "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

const (
	validUUIDv4  = "d50eda0d-2cea-4de1-8d42-9cd3e7e8670d"
	artifactSize = 10000
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		DataStoreError error
		FileStoreError error
		WorkflowsError error
		InventoryError error
		ReportingError error
	}{{
		Name: "ok",
	}, {
		Name:           "error: datastore",
		DataStoreError: errors.New("connection error"),
	}, {
		Name:           "error: filestore",
		FileStoreError: errors.New("connection error"),
	}, {
		Name:           "error: workflows",
		WorkflowsError: errors.New("connection error"),
	}, {
		Name:           "error: inventory",
		InventoryError: errors.New("connection error"),
	}, {
		Name:           "error: reporting",
		ReportingError: errors.New("connection error"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.TODO()
			mDStore := &mocks.DataStore{}
			mFStore := &fs_mocks.ObjectStorage{}
			mWorkflows := &workflows_mocks.Client{}
			mInventory := &inventory_mocks.Client{}
			mReporting := &reporting_mocks.Client{}
			dep := &Deployments{
				db:              mDStore,
				objectStorage:   mFStore,
				workflowsClient: mWorkflows,
				inventoryClient: mInventory,
			}
			dep = dep.WithReporting(mReporting)
			switch {
			default:
				mReporting.On("CheckHealth", ctx).
					Return(tc.ReportingError)
				fallthrough
			case tc.InventoryError != nil:
				mInventory.On("CheckHealth", ctx).
					Return(tc.InventoryError)
				fallthrough
			case tc.WorkflowsError != nil:
				mWorkflows.On("CheckHealth", ctx).
					Return(tc.WorkflowsError)
				fallthrough
			case tc.FileStoreError != nil:
				mFStore.On("HealthCheck", ctx).
					Return(tc.FileStoreError)
				fallthrough
			case tc.DataStoreError != nil:
				mDStore.On("Ping", ctx).
					Return(tc.DataStoreError)
				mDStore.On("GetStorageSettings", ctx).
					Return(&model.StorageSettings{
						Region:      config.Config.GetString(dconfig.SettingAwsS3Region),
						ExternalUri: config.Config.GetString(dconfig.SettingAwsExternalURI),
						Uri:         config.Config.GetString(dconfig.SettingAwsURI),
						Bucket:      config.Config.GetString(dconfig.SettingStorageBucket),
						Key:         config.Config.GetString(dconfig.SettingAwsAuthKeyId),
						Secret:      config.Config.GetString(dconfig.SettingAwsAuthSecret),
						Token:       config.Config.GetString(dconfig.SettingAwsAuthToken)}, nil)
			}
			err := dep.HealthCheck(ctx)
			switch {
			case tc.DataStoreError != nil:
				assert.EqualError(t, err,
					"error reaching MongoDB: "+
						tc.DataStoreError.Error(),
				)

			case tc.FileStoreError != nil:
				assert.EqualError(t, err,
					"error reaching artifact storage service: "+
						tc.FileStoreError.Error(),
				)

			case tc.WorkflowsError != nil:
				assert.EqualError(t, err,
					"Workflows service unhealthy: "+
						tc.WorkflowsError.Error(),
				)

			case tc.InventoryError != nil:
				assert.EqualError(t, err,
					"Inventory service unhealthy: "+
						tc.InventoryError.Error(),
				)

			case tc.ReportingError != nil:
				assert.EqualError(t, err,
					"Reporting service unhealthy: "+
						tc.ReportingError.Error(),
				)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeploymentModelCreateDeployment(t *testing.T) {
	longStr := strings.Repeat("a", 257)
	t.Parallel()

	testCases := map[string]struct {
		InputConstructor *model.DeploymentConstructor

		InputDeploymentStorageInsertError error
		InputImagesByNameError            error

		InvDevices        []model.InvDevice
		InvDevicesPageTwo []model.InvDevice
		TotalCount        int
		SearchError       error
		GetFilterError    error

		CallGetDeviceGroups  bool
		InventoryGroups      []string
		GetDeviceGroupsError error

		ReportingService bool

		OutputError error
		OutputBody  bool
	}{
		"model missing": {
			OutputError: ErrModelMissingInput,
		},
		"insert error": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			InputDeploymentStorageInsertError: errors.New("insert error"),
			CallGetDeviceGroups:               true,

			OutputError: errors.New("Storing deployment data: insert error"),
		},
		"ok": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			CallGetDeviceGroups: true,
			InventoryGroups:     []string{"foo", "bar"},

			OutputBody: true,
		},
		"ok with group": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "group",
				ArtifactName: "App 123",
				Group:        "group",
			},

			InvDevices: []model.InvDevice{
				{
					ID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
				},
			},
			TotalCount: 1,

			OutputBody: true,
		},
		"ok with group, two pages": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "group",
				ArtifactName: "App 123",
				Group:        "group",
			},

			InvDevices: []model.InvDevice{
				{
					ID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
				},
			},
			InvDevicesPageTwo: []model.InvDevice{
				{
					ID: "b532b01a-9313-404f-8d19-e7fcbe5cc348",
				},
			},
			TotalCount: 2,

			OutputBody: true,
		},
		"ok with group, reeporting": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "group",
				ArtifactName: "App 123",
				Group:        "group",
			},

			InvDevices: []model.InvDevice{
				{
					ID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
				},
			},
			TotalCount: 1,

			ReportingService: true,

			OutputBody: true,
		},
		"ko, with group, no device found": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "group",
				ArtifactName: "App 123",
				Group:        "group",
			},

			OutputError: ErrNoDevices,
		},
		"ko, name too long": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         longStr,
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},

			OutputError: errors.New("Validating deployment: name: the length must be no more than 256."),
		},
		"ko, with group, error while searching": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "group",
				ArtifactName: "App 123",
				Group:        "group",
			},

			SearchError: errors.New("error searching inventory"),
			OutputError: ErrModelInternal,
		},
		"ko, conflict": {
			InputConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			CallGetDeviceGroups:               true,
			InputDeploymentStorageInsertError: mongo.ErrConflictingDeployment,

			OutputError: ErrConflictingDeployment,
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {
			ctx := context.Background()

			identityObject := &identity.Identity{Tenant: "tenant_id"}
			ctx = identity.WithContext(ctx, identityObject)

			db := mocks.DataStore{}
			db.On("InsertDeployment",
				ctx,
				mock.AnythingOfType("*model.Deployment")).
				Return(testCase.InputDeploymentStorageInsertError)

			db.On("ImagesByName",
				ctx,
				mock.AnythingOfType("string")).
				Return(
					[]*model.Image{model.NewImage(
						validUUIDv4,
						&model.ImageMeta{},
						&model.ArtifactMeta{
							Name: "App 123",
							DeviceTypesCompatible: []string{
								"hammer",
							},
							Depends: map[string]interface{}{},
						}, artifactSize)},
					testCase.InputImagesByNameError)

			fs := &fs_mocks.ObjectStorage{}
			ds := NewDeployments(&db, fs, 0, false)

			mockInventoryClient := &inventory_mocks.Client{}
			if testCase.CallGetDeviceGroups {
				mockInventoryClient.On("GetDeviceGroups",
					ctx,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string")).
					Return(testCase.InventoryGroups, testCase.GetDeviceGroupsError)
			}

			mockReportingClient := &reporting_mocks.Client{}
			if testCase.InputConstructor != nil && testCase.InputConstructor.Group != "" && len(testCase.InputConstructor.Devices) == 0 {
				if testCase.ReportingService {
					mockReportingClient.On("Search", ctx,
						"tenant_id",
						model.SearchParams{
							Page:    1,
							PerPage: PerPageInventoryDevices,
							Filters: []model.FilterPredicate{
								{
									Scope:     InventoryIdentityScope,
									Attribute: InventoryStatusAttributeName,
									Type:      "$eq",
									Value:     InventoryStatusAccepted,
								},
								{
									Scope:     InventoryGroupScope,
									Attribute: InventoryGroupAttributeName,
									Type:      "$eq",
									Value:     testCase.InputConstructor.Group,
								},
							},
						},
					).Return(testCase.InvDevices, testCase.TotalCount, testCase.SearchError)
				} else {
					mockInventoryClient.On("Search", ctx,
						"tenant_id",
						model.SearchParams{
							Page:    1,
							PerPage: PerPageInventoryDevices,
							Filters: []model.FilterPredicate{
								{
									Scope:     InventoryIdentityScope,
									Attribute: InventoryStatusAttributeName,
									Type:      "$eq",
									Value:     InventoryStatusAccepted,
								},
								{
									Scope:     InventoryGroupScope,
									Attribute: InventoryGroupAttributeName,
									Type:      "$eq",
									Value:     testCase.InputConstructor.Group,
								},
							},
						},
					).Return(testCase.InvDevices, testCase.TotalCount, testCase.SearchError)

					if testCase.TotalCount > len(testCase.InvDevices) {
						mockInventoryClient.On("Search", ctx,
							"tenant_id",
							model.SearchParams{
								Page:    2,
								PerPage: PerPageInventoryDevices,
								Filters: []model.FilterPredicate{
									{
										Scope:     InventoryIdentityScope,
										Attribute: InventoryStatusAttributeName,
										Type:      "$eq",
										Value:     InventoryStatusAccepted,
									},
									{
										Scope:     InventoryGroupScope,
										Attribute: InventoryGroupAttributeName,
										Type:      "$eq",
										Value:     testCase.InputConstructor.Group,
									},
								},
							},
						).Return(testCase.InvDevicesPageTwo, testCase.TotalCount, testCase.SearchError)
					}
				}
			}

			ds.SetInventoryClient(mockInventoryClient)
			if testCase.ReportingService {
				ds.WithReporting(mockReportingClient)
			}

			out, err := ds.CreateDeployment(ctx, testCase.InputConstructor)
			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}
			if testCase.OutputBody {
				assert.NotNil(t, out)
			}

			mockInventoryClient.AssertExpectations(t)
		})
	}

}

func TestUploadLink(t *testing.T) {
	t.Parallel()

	regexMatcher := func(pattern string) interface{} {
		return mock.MatchedBy(func(value string) bool {
			return assert.Regexp(t, pattern, value)
		})
	}

	link := &model.Link{
		Uri:    "http://localhost:8080",
		Method: "PUT",
		Expire: time.Now().Add(time.Hour),
	}
	matchUpLink := mock.MatchedBy(func(value *model.UploadLink) bool {
		return assert.Equal(t, *link, value.Link)
	})

	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		objStore := new(fs_mocks.ObjectStorage)
		ds := new(mocks.DataStore)
		deploy := NewDeployments(ds, objStore, 0, false)
		objStore.On("PutRequest",
			h.ContextMatcher(),
			regexMatcher(`^[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}\`+
				fileSuffixTmp),
			time.Minute,
			true,
		).Return(link, nil)

		ds.On("GetStorageSettings", ctx).
			Return(nil, nil).
			Once().
			On("InsertUploadIntent", h.ContextMatcher(), matchUpLink).
			Return(nil).
			Once()
		upLink, err := deploy.UploadLink(ctx, time.Minute, false)
		assert.NoError(t, err)
		assert.NotNil(t, upLink)
		objStore.AssertExpectations(t)
		ds.AssertExpectations(t)
	})

	t.Run("ok/multi-tenancy", func(t *testing.T) {
		ctx := identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		})
		objStore := new(fs_mocks.ObjectStorage)
		ds := new(mocks.DataStore)
		deploy := NewDeployments(ds, objStore, 0, false)
		objStore.On("PutRequest",
			h.ContextMatcher(),
			regexMatcher(`^123456789012345678901234/`+
				`[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}\`+
				fileSuffixTmp),
			time.Minute,
			true,
		).Return(link, nil)

		ds.On("GetStorageSettings", h.ContextMatcher()).
			Return(nil, nil).
			Once().
			On("InsertUploadIntent", h.ContextMatcher(), matchUpLink).
			Return(nil).
			Once()
		upLink, err := deploy.UploadLink(ctx, time.Minute, false)
		assert.NoError(t, err)
		assert.NotNil(t, upLink)
		objStore.AssertExpectations(t)
		ds.AssertExpectations(t)
	})

	t.Run("error/signing request", func(t *testing.T) {
		ctx := identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		})
		objStore := new(fs_mocks.ObjectStorage)
		ds := new(mocks.DataStore)
		deploy := NewDeployments(ds, objStore, 0, false)
		errInternal := errors.New("internal error")
		ds.On("GetStorageSettings", ctx).
			Return(nil, nil).
			Once()
		objStore.On("PutRequest",
			h.ContextMatcher(),
			regexMatcher(`^123456789012345678901234/`+
				`[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}\`+
				fileSuffixTmp),
			time.Minute,
			true,
		).Return(nil, errInternal)

		upLink, err := deploy.UploadLink(ctx, time.Minute, false)
		assert.ErrorIs(t, err, errInternal)
		assert.Nil(t, upLink)
		objStore.AssertExpectations(t)
		ds.AssertExpectations(t)
	})

	t.Run("error/recording upload intent", func(t *testing.T) {
		ctx := identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		})
		objStore := new(fs_mocks.ObjectStorage)
		ds := new(mocks.DataStore)
		deploy := NewDeployments(ds, objStore, 0, false)
		errInternal := errors.New("internal error")
		objStore.On("PutRequest",
			h.ContextMatcher(),
			regexMatcher(`^123456789012345678901234/`+
				`[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}\`+
				fileSuffixTmp),
			time.Minute,
			true,
		).Return(link, nil)

		ds.On("GetStorageSettings", ctx).
			Return(nil, nil).
			Once().
			On("InsertUploadIntent", h.ContextMatcher(), matchUpLink).
			Return(errInternal).
			Once()
		upLink, err := deploy.UploadLink(ctx, time.Minute, false)
		assert.ErrorIs(t, err, errInternal)
		assert.Nil(t, upLink)
		objStore.AssertExpectations(t)
		ds.AssertExpectations(t)
	})
	t.Run("error/getting storage settings", func(t *testing.T) {
		ctx := identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		})
		objStore := new(fs_mocks.ObjectStorage)
		ds := new(mocks.DataStore)
		deploy := NewDeployments(ds, objStore, 0, false)
		errInternal := errors.New("internal error")
		ds.On("GetStorageSettings", ctx).
			Return(nil, errInternal).
			Once()
		upLink, err := deploy.UploadLink(ctx, time.Minute, false)
		assert.ErrorIs(t, err, errInternal)
		assert.Nil(t, upLink)
		objStore.AssertExpectations(t)
		ds.AssertExpectations(t)
	})
}

type eofReadCloser struct {
	ch   chan struct{}
	once *sync.Once
	err  error
}

func newEOFReadCloser(closeErr error) *eofReadCloser {
	return &eofReadCloser{
		ch:   make(chan struct{}),
		once: new(sync.Once),
		err:  closeErr,
	}
}

func (r *eofReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (r *eofReadCloser) Close() error {
	r.once.Do(func() { close(r.ch) })
	return r.err
}

func TestCompleteUpload(t *testing.T) {
	t.Parallel()

	const intentID = "9bf1bfff-eeb4-49d4-b55d-d717d407888a"
	var testErr = errors.New("test error")

	type testCase struct {
		Name string

		Identity      *identity.Identity
		Database      func(t *testing.T, self *testCase) *mocks.DataStore
		ObjectStorage func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage
		SkipVerify    bool

		syncChan chan struct{}

		ErrorAssertionFunc func(t *testing.T, self *testCase, err error)
	}
	contextHasIdentity := func(t *testing.T, expected *identity.Identity) interface{} {
		return mock.MatchedBy(func(ctx context.Context) bool {
			actual := identity.FromContext(ctx)
			return assert.Equal(t, expected, actual)
		})
	}
	testCases := []*testCase{{
		Name: "ok",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusProcessing,
					model.LinkStatusAborted).
				Return(nil)

			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(nil)
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID+fileSuffixTmp).
				Return(r, nil).
				Once().
				On("PutObject",
					contextHasIdentity(t, self.Identity),
					intentID,
					mock.AnythingOfType("*io.PipeReader")).
				Return(nil)
			self.syncChan = r.ch
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.NoError(t, err)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "ok/skip verify",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusProcessing,
					model.LinkStatusAborted).
				Return(nil)

			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(nil)
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID).
				Return(r, nil)
			self.syncChan = r.ch
			return os
		},
		SkipVerify: true,

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.NoError(t, err)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "ok/multi-tenancy",

		Identity: &identity.Identity{
			Tenant: "123456789012345678901234",
		},
		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusProcessing,
					model.LinkStatusAborted).
				Return(errors.New("internal error"))

			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(nil)
			objectPath := "123456789012345678901234/" + intentID
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				objectPath+fileSuffixTmp).
				Return(r, nil).
				Once().
				On("PutObject",
					contextHasIdentity(t, self.Identity),
					objectPath,
					mock.AnythingOfType("*io.PipeReader")).
				Return(nil)
			self.syncChan = r.ch
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.NoError(t, err)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "ok/multi-tenancy/skip verify",

		Identity: &identity.Identity{
			Tenant: "123456789012345678901234",
		},
		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusProcessing,
					model.LinkStatusAborted).
				Return(errors.New("internal error"))

			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(nil)
			objectPath := "123456789012345678901234/" + intentID
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				objectPath).
				Return(r, nil)
			self.syncChan = r.ch
			return os
		},
		SkipVerify: true,

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.NoError(t, err)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "error/set status to processing",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(errors.New("internal error")).
				Once()
			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(errors.New("close error"))
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID+fileSuffixTmp).
				Return(r, nil).
				Once()
			self.syncChan = r.ch
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.Error(t, err)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "error/already in progress",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once().
				On("UpdateUploadIntentStatus",
					contextHasIdentity(t, self.Identity),
					intentID,
					model.LinkStatusPending,
					model.LinkStatusProcessing).
				Return(store.ErrNotFound).
				Once()
			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			r := newEOFReadCloser(nil)
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID+fileSuffixTmp).
				Return(r, nil).
				Once()
			self.syncChan = r.ch
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			deadline, ok := t.Deadline()
			if !ok || time.Until(deadline) > time.Minute {
				deadline = time.Now().Add(time.Minute)
			}
			select {
			case <-self.syncChan:
				assert.ErrorIs(t, err, ErrUploadNotFound)
			case <-time.After(time.Until(deadline)):
				assert.FailNow(t,
					"timed out waiting for processUploadedArtifact"+
						"to be called")
			}
		},
	}, {
		Name: "error/object not found",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once()
			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID+fileSuffixTmp).
				Return(nil, storage.ErrObjectNotFound).
				Once()
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			assert.ErrorIs(t, err, ErrUploadNotFound)
		},
	}, {
		Name: "error/internal storage error",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, nil).
				Once()
			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			os.On("GetObject",
				contextHasIdentity(t, self.Identity),
				intentID+fileSuffixTmp).
				Return(nil, testErr).
				Once()
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			assert.ErrorIs(t, err, testErr)
		},
	}, {
		Name: "error/retrieve storage settings",

		Database: func(t *testing.T, self *testCase) *mocks.DataStore {
			ds := new(mocks.DataStore)
			ds.On("GetStorageSettings", contextHasIdentity(t, self.Identity)).
				Return(nil, testErr).
				Once()
			return ds
		},
		ObjectStorage: func(t *testing.T, self *testCase) *fs_mocks.ObjectStorage {
			os := new(fs_mocks.ObjectStorage)
			return os
		},

		ErrorAssertionFunc: func(t *testing.T, self *testCase, err error) {
			assert.ErrorIs(t, err, testErr)
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			if tc.Identity != nil {
				ctx = identity.WithContext(ctx, tc.Identity)
			}
			ds := tc.Database(t, tc)
			defer ds.AssertExpectations(t)
			objStore := tc.ObjectStorage(t, tc)
			defer objStore.AssertExpectations(t)
			deploy := NewDeployments(ds, objStore, 0, false)

			err := deploy.CompleteUpload(ctx, intentID, tc.SkipVerify, nil)
			tc.ErrorAssertionFunc(t, tc, err)
		})
	}
}

func TestCreateDeviceConfigurationDeployment(t *testing.T) {

	t.Parallel()

	testCases := map[string]struct {
		inputConstructor  *model.ConfigurationDeploymentConstructor
		inputDeviceID     string
		inputDeploymentID string

		inputDeploymentStorageInsertError error
		inventoryError                    error

		callInventory bool
		callDb        bool

		outputError error
		outputID    string
	}{
		"ok": {
			inputConstructor: &model.ConfigurationDeploymentConstructor{
				Name:          "foo",
				Configuration: []byte("bar"),
			},
			inputDeviceID:     "foo-device",
			inputDeploymentID: "foo-deployment",
			callInventory:     true,
			callDb:            true,

			outputID: "foo-deployment",
		},
		"constructor missing": {
			outputError: ErrModelMissingInput,
		},
		"insert error": {
			inputConstructor: &model.ConfigurationDeploymentConstructor{
				Name:          "foo",
				Configuration: []byte("bar"),
			},
			inputDeploymentStorageInsertError: errors.New("insert error"),
			callInventory:                     true,
			callDb:                            true,

			outputError: errors.New("Storing deployment data: insert error"),
		},
		"inventory error": {
			inputConstructor: &model.ConfigurationDeploymentConstructor{
				Name:          "foo",
				Configuration: []byte("bar"),
			},
			inventoryError: errors.New("inventory error"),
			callInventory:  true,

			outputError: errors.New("inventory error"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {
			ctx := context.Background()

			identityObject := &identity.Identity{Tenant: "tenant_id"}
			ctx = identity.WithContext(ctx, identityObject)

			db := mocks.DataStore{}
			if tc.callDb {
				db.On("InsertDeployment",
					ctx,
					mock.AnythingOfType("*model.Deployment")).
					Return(tc.inputDeploymentStorageInsertError)
			}
			defer db.AssertExpectations(t)

			inv := &inventory_mocks.Client{}
			if tc.callInventory {
				inv.On("GetDeviceGroups", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return([]string{}, tc.inventoryError)
			}
			defer inv.AssertExpectations(t)

			ds := &Deployments{
				db:              &db,
				inventoryClient: inv,
			}

			out, err := ds.CreateDeviceConfigurationDeployment(ctx, tc.inputConstructor, tc.inputDeviceID, tc.inputDeploymentID)
			if tc.outputError != nil {
				assert.EqualError(t, err, tc.outputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, out, tc.outputID)
			}
		})
	}
}

func TestAbortDeployment(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		InputDeploymentID string

		AbortDeviceDeploymentsError error

		AggregateDeviceDeploymentByStatusStats model.Stats
		AggregateDeviceDeploymentByStatusError error
		CallAggregateDeviceDeploymentByStatus  bool

		UpdateStatsError error
		CallUpdateStats  bool

		SetDeploymentStatusError error
		CallSetDeploymentStatus  bool

		OutputError error
	}{
		"AbortDeviceDeployments error": {
			InputDeploymentID:           "f826484e-1157-4109-af21-304e6d711561",
			AbortDeviceDeploymentsError: errors.New("AbortDeviceDeploymentsError"),
			OutputError:                 errors.New("AbortDeviceDeploymentsError"),
		},
		"AggregateDeviceDeploymentByStatus error": {
			InputDeploymentID:                      "f826484e-1157-4109-af21-304e6d711561",
			CallAggregateDeviceDeploymentByStatus:  true,
			AggregateDeviceDeploymentByStatusError: errors.New("AggregateDeviceDeploymentByStatusError"),
			AggregateDeviceDeploymentByStatusStats: model.Stats{},
			OutputError:                            errors.New("AggregateDeviceDeploymentByStatusError"),
		},
		"UpdateStats error": {
			InputDeploymentID:                      "f826484e-1157-4109-af21-304e6d711561",
			CallAggregateDeviceDeploymentByStatus:  true,
			AggregateDeviceDeploymentByStatusStats: model.Stats{"aaa": 1},
			CallUpdateStats:                        true,
			UpdateStatsError:                       errors.New("UpdateStatsError"),
			OutputError:                            errors.New("failed to update deployment stats: UpdateStatsError"),
		},
		"all correct": {
			InputDeploymentID:                      "f826484e-1157-4109-af21-304e6d711561",
			CallAggregateDeviceDeploymentByStatus:  true,
			AggregateDeviceDeploymentByStatusStats: model.Stats{"aaa": 1},
			CallUpdateStats:                        true,
			CallSetDeploymentStatus:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {
			db := mocks.DataStore{}
			defer db.AssertExpectations(t)
			db.On("AbortDeviceDeployments",
				h.ContextMatcher(), tc.InputDeploymentID).
				Return(tc.AbortDeviceDeploymentsError)
			if tc.CallAggregateDeviceDeploymentByStatus {
				db.On("AggregateDeviceDeploymentByStatus",
					h.ContextMatcher(), tc.InputDeploymentID).
					Return(tc.AggregateDeviceDeploymentByStatusStats,
						tc.AggregateDeviceDeploymentByStatusError)
			}
			if tc.CallUpdateStats {
				db.On("UpdateStats",
					h.ContextMatcher(), tc.InputDeploymentID,
					mock.AnythingOfType("model.Stats")).
					Return(tc.UpdateStatsError)
			}
			if tc.CallSetDeploymentStatus {
				db.On("SetDeploymentStatus",
					h.ContextMatcher(), tc.InputDeploymentID,
					model.DeploymentStatusFinished, mock.AnythingOfType("time.Time")).
					Return(tc.SetDeploymentStatusError)
			}

			ds := &Deployments{
				db: &db,
			}
			ctx := context.Background()

			err := ds.AbortDeployment(ctx, tc.InputDeploymentID)
			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteDeviceDeploymentsHistory(t *testing.T) {
	t.Parallel()
	f := false
	ctx := context.Background()
	deviceID := "f826484e-1157-4109-af21-304e6d711561"

	testCases := map[string]struct {
		workflowsMock *workflows_mocks.Client
		storeMock     *mocks.DataStore

		OutputError error
	}{
		"ok": {
			storeMock: func() *mocks.DataStore {
				ds := new(mocks.DataStore)

				ds.On("DeleteDeviceDeploymentsHistory",
					h.ContextMatcher(), deviceID).
					Return(nil)
				ds.On("GetDeviceDeployments",
					h.ContextMatcher(),
					0,
					0,
					deviceID,
					&f,
					false,
				).Return(
					[]model.DeviceDeployment{
						{
							Id:           "foo",
							DeviceId:     "bar",
							DeploymentId: "baz",
						},
					},
					nil,
				)

				return ds
			}(),
			workflowsMock: func() *workflows_mocks.Client {
				wf := new(workflows_mocks.Client)
				wf.On(
					"StartReindexReportingDeploymentBatch",
					ctx,
					[]workflows.DeviceDeploymentShortInfo{
						{
							ID:           "foo",
							DeviceID:     "bar",
							DeploymentID: "baz",
						},
					},
				).Return(nil)
				return wf
			}(),
		},
		"error": {
			workflowsMock: func() *workflows_mocks.Client {
				wf := new(workflows_mocks.Client)
				return wf
			}(),
			storeMock: func() *mocks.DataStore {
				ds := new(mocks.DataStore)

				ds.On("DeleteDeviceDeploymentsHistory",
					h.ContextMatcher(), deviceID).
					Return(errors.New("error"))
				ds.On("GetDeviceDeployments",
					h.ContextMatcher(),
					0,
					0,
					deviceID,
					&f,
					false,
				).Return(
					[]model.DeviceDeployment{
						{
							Id:           "foo",
							DeviceId:     "bar",
							DeploymentId: "baz",
						},
					},
					nil,
				)

				return ds
			}(),

			OutputError: errors.New("error"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {

			rc := new(reporting_mocks.Client)
			defer rc.AssertExpectations(t)
			defer tc.workflowsMock.AssertExpectations(t)
			defer tc.storeMock.AssertExpectations(t)
			ds := &Deployments{
				db:              tc.storeMock,
				workflowsClient: tc.workflowsMock,
				reportingClient: rc,
			}

			err := ds.DeleteDeviceDeploymentsHistory(ctx, deviceID)
			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImageUsedInActiveDeployment(t *testing.T) {

	t.Parallel()

	testCases := map[string]struct {
		InputID string

		ExistUnfinishedByArtifactIdResponse bool
		ExistUnfinishedByArtifactIdError    error

		OutputError error
		OutputBool  bool
	}{
		"ok": {
			InputID:                             "ID:1234",
			ExistUnfinishedByArtifactIdResponse: true,

			OutputBool: true,
		},
		"ExistUnfinishedByArtifactId error": {
			InputID:                             "ID:1234",
			ExistUnfinishedByArtifactIdError:    errors.New("Some error"),
			ExistUnfinishedByArtifactIdResponse: false,

			OutputError: errors.New("Checking if image is used by active deployment: Some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {
			db := mocks.DataStore{}
			defer db.AssertExpectations(t)

			db.On("ExistUnfinishedByArtifactId",
				h.ContextMatcher(),
				mock.AnythingOfType("string")).
				Return(tc.ExistUnfinishedByArtifactIdResponse,
					tc.ExistUnfinishedByArtifactIdError)

			ds := &Deployments{
				db: &db,
			}

			found, err := ds.ImageUsedInActiveDeployment(context.Background(),
				tc.InputID)
			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.OutputBool, found)
		})
	}

}

func TestGetDeviceDeploymentListForDevice(t *testing.T) {
	const deviceID = "device_id"
	testCases := map[string]struct {
		query                  store.ListQueryDeviceDeployments
		deviceDeployments      []model.DeviceDeployment
		deviceDeploymentsCount int
		deviceDeploymentsErr   error

		deploymentsQuery model.Query
		deployments      []*model.Deployment
		deploymentsCount int64
		deploymentsErr   error

		res      []model.DeviceDeploymentListItem
		resCount int
		resErr   error
	}{
		"ok": {
			query: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Limit:    10,
			},
			deviceDeployments: []model.DeviceDeployment{
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
				},
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
			},
			deviceDeploymentsCount: 2,

			deploymentsQuery: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Limit:        2,
				DisableCount: true,
			},
			deployments: []*model.Deployment{
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
				},
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
			},
			deploymentsCount: 2,

			res: []model.DeviceDeploymentListItem{
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					Deployment: &model.Deployment{
						Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					},
					Device: &model.DeviceDeployment{
						Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
						DeviceId:     deviceID,
						DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					},
				},
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					Deployment: &model.Deployment{
						Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					},
					Device: &model.DeviceDeployment{
						Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
						DeviceId:     deviceID,
						DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					},
				},
			},
			resCount: 2,
		},
		"ok, partial match of deployments": {
			query: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Limit:    10,
			},
			deviceDeployments: []model.DeviceDeployment{
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
				},
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
			},
			deviceDeploymentsCount: 2,

			deploymentsQuery: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Limit:        2,
				DisableCount: true,
			},
			deployments: []*model.Deployment{
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
			},
			deploymentsCount: 2,

			res: []model.DeviceDeploymentListItem{
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					Device: &model.DeviceDeployment{
						Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
						DeviceId:     deviceID,
						DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					},
				},
				{
					Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					Deployment: &model.Deployment{
						Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					},
					Device: &model.DeviceDeployment{
						Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
						DeviceId:     deviceID,
						DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					},
				},
			},
			resCount: 2,
		},
		"ko, error retrieving deviceDeployments": {
			query: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Limit:    10,
			},
			deviceDeploymentsCount: -1,
			deviceDeploymentsErr:   errors.New("error"),

			resCount: -1,
			resErr:   errors.New("retrieving the list of deployment statuses: error"),
		},
		"ko, error retrieving deployments": {
			query: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Limit:    10,
			},
			deviceDeployments: []model.DeviceDeployment{
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
				},
				{
					Id:           "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					DeviceId:     deviceID,
					DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
			},
			deviceDeploymentsCount: 2,

			deploymentsQuery: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Limit:        2,
				DisableCount: true,
			},
			deploymentsCount: -1,
			deploymentsErr:   errors.New("error"),

			resCount: -1,
			resErr:   errors.New("retrieving the list of deployments: error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := mocks.DataStore{}
			defer db.AssertExpectations(t)

			db.On("GetDeviceDeploymentsForDevice",
				ctx,
				tc.query,
			).Return(
				tc.deviceDeployments,
				tc.deviceDeploymentsCount,
				tc.deviceDeploymentsErr,
			)

			if tc.deviceDeploymentsErr == nil {
				db.On("FindDeployments",
					ctx,
					tc.deploymentsQuery,
				).Return(
					tc.deployments,
					tc.deploymentsCount,
					tc.deploymentsErr,
				)
			}

			ds := &Deployments{
				db: &db,
			}

			deviceDeployments, count, err := ds.GetDeviceDeploymentListForDevice(ctx, tc.query)
			if tc.resErr != nil {
				assert.EqualError(t, err, tc.resErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.res, deviceDeployments)
				assert.Equal(t, tc.resCount, count)
			}
		})
	}
}

func TestUpdateDeploymentsWithArtifactName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		// Input
		artifactName string

		storeMock *mocks.DataStore

		// Output
		err error
	}{
		{
			name:         "ok",
			artifactName: "foo",

			storeMock: func() *mocks.DataStore {
				ds := new(mocks.DataStore)

				ds.On("ExistUnfinishedByArtifactName",
					h.ContextMatcher(),
					"foo",
				).Return(true, nil)
				ds.On("ImagesByName",
					h.ContextMatcher(),
					"foo",
				).Return([]*model.Image{{Id: "foo-id"}}, nil)
				ds.On("UpdateDeploymentsWithArtifactName",
					h.ContextMatcher(), "foo", []string{"foo-id"},
				).Return(nil)

				return ds
			}(),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			defer tc.storeMock.AssertExpectations(t)

			app := &Deployments{
				db: tc.storeMock,
			}

			err := app.UpdateDeploymentsWithArtifactName(context.Background(), tc.artifactName)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReindexDevice(t *testing.T) {
	t.Parallel()

	const deviceID = "deviceID"
	ctx := context.Background()

	testCases := []struct {
		name          string
		workflowsMock func() workflows.Client
		err           error
	}{
		{
			name: "ok",
			workflowsMock: func() workflows.Client {
				wf := &workflows_mocks.Client{}
				wf.On("StartReindexReporting", ctx, deviceID).Return(nil)
				return wf
			},
		},
		{
			name: "ko",
			workflowsMock: func() workflows.Client {
				wf := &workflows_mocks.Client{}
				wf.On("StartReindexReporting", ctx, deviceID).Return(errors.New("error"))
				return wf
			},
			err: errors.New("error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &Deployments{
				workflowsClient: tc.workflowsMock(),
				reportingClient: &reporting_mocks.Client{},
			}

			err := app.reindexDevice(ctx, deviceID)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReindexDeployment(t *testing.T) {
	t.Parallel()

	const deviceID = "deviceID"
	const deploymentID = "deploymentID"
	const ID = "ID"
	ctx := context.Background()

	testCases := []struct {
		name          string
		workflowsMock func() workflows.Client
		err           error
	}{
		{
			name: "ok",
			workflowsMock: func() workflows.Client {
				wf := &workflows_mocks.Client{}
				wf.On("StartReindexReportingDeployment", ctx, deviceID, deploymentID, ID).Return(nil)
				return wf
			},
		},
		{
			name: "ko",
			workflowsMock: func() workflows.Client {
				wf := &workflows_mocks.Client{}
				wf.On("StartReindexReportingDeployment", ctx, deviceID, deploymentID, ID).Return(errors.New("error"))
				return wf
			},
			err: errors.New("error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := &Deployments{
				workflowsClient: tc.workflowsMock(),
				reportingClient: &reporting_mocks.Client{},
			}

			err := app.reindexDeployment(ctx, deviceID, deploymentID, ID)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLookupDeployment(t *testing.T) {
	const deviceID = "device_id"
	testCases := map[string]struct {
		query              model.Query
		dbDeployments      []*model.Deployment
		dbDeploymentsCount int64
		dbErr              error

		res      []*model.Deployment
		resCount int64
		resErr   error
	}{
		"ok": {
			query: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Names: []string{"bar", "baz"},
				Limit: 10,
			},
			dbDeployments: []*model.Deployment{
				{
					Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					DeviceCount: intPtr(3),
				},
				{
					Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					DeviceCount: intPtr(3),
				},
			},
			dbDeploymentsCount: 2,
			res: []*model.Deployment{
				{
					Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					DeviceCount: intPtr(3),
				},
				{
					Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
					DeviceCount: intPtr(3),
				},
			},
			resCount: 2,
		},
		"no deployments": {
			query: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Names: []string{"bar", "baz"},
				Limit: 10,
			},
			dbDeployments:      nil,
			dbDeploymentsCount: 0,
			res:                []*model.Deployment{},
			resCount:           0,
		},
		"database error": {
			query: model.Query{
				IDs: []string{
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
					"d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				},
				Names: []string{"bar", "baz"},
				Limit: 10,
			},
			dbDeployments:      nil,
			dbDeploymentsCount: 0,
			dbErr:              errors.New("db error"),
			res:                nil,
			resCount:           0,
			resErr:             errors.New("searching for deployments: db error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := mocks.DataStore{}
			defer db.AssertExpectations(t)

			db.On("FindDeployments",
				ctx,
				tc.query,
			).Return(
				tc.dbDeployments,
				tc.dbDeploymentsCount,
				tc.dbErr,
			)

			ds := &Deployments{
				db: &db,
			}

			deployments, count, err := ds.LookupDeployment(ctx, tc.query)
			if tc.resErr != nil {
				assert.EqualError(t, err, tc.resErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.resCount, count)
				assert.Equal(t, tc.res, deployments)
			}
		})
	}
}

func TestGetDeploymentFilter(t *testing.T) {
	testCases := map[string]struct {
		inConstructor *model.DeploymentConstructor
		outFilter     *model.Filter
	}{
		"group": {
			inConstructor: &model.DeploymentConstructor{
				Group: "foo",
			},
			outFilter: &model.Filter{
				Terms: []model.FilterPredicate{
					{
						Scope:     InventoryGroupScope,
						Attribute: InventoryGroupAttributeName,
						Type:      "$eq",
						Value:     "foo",
					},
				},
			},
		},
		"all devices": {
			inConstructor: &model.DeploymentConstructor{
				AllDevices: true,
			},
			outFilter: &model.Filter{
				Terms: []model.FilterPredicate{
					{
						Scope:     InventoryIdentityScope,
						Attribute: InventoryStatusAttributeName,
						Type:      "$eq",
						Value:     InventoryStatusAccepted,
					},
				},
			},
		},
		"list of devices": {
			inConstructor: &model.DeploymentConstructor{
				Devices: []string{
					"b532b01a-9313-404f-8d19-e7fcbe5cc347",
					"b532b01a-9313-404f-8d19-e7fcbe5cc348",
				},
			},
			outFilter: &model.Filter{
				Terms: []model.FilterPredicate{
					{
						Scope:     InventoryIdentityScope,
						Attribute: InventoryIdAttributeName,
						Type:      "$in",
						Value: []string{
							"b532b01a-9313-404f-8d19-e7fcbe5cc347",
							"b532b01a-9313-404f-8d19-e7fcbe5cc348",
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			filter := getDeploymentFilter(tc.inConstructor)
			assert.Equal(t, tc.outFilter, filter)
		})
	}
}
