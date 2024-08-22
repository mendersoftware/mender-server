// Copyright 2021 Northern.tech AS
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

package mongo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contextExpireAfterX struct {
	context.Context
	x    int64
	done chan struct{}
	err  error
}

func newContextExpireAfterX(x int) context.Context {
	return &contextExpireAfterX{
		Context: context.Background(),
		x:       int64(x),
		done:    make(chan struct{}),
	}
}

func (ctx *contextExpireAfterX) Done() <-chan struct{} {
	n := atomic.AddInt64(&ctx.x, -1)
	if n <= 0 {
		select {
		case <-ctx.done:
		default:
			close(ctx.done)
		}
		ctx.err = context.DeadlineExceeded
	}
	return ctx.done
}

func (ctx *contextExpireAfterX) Err() error {
	return ctx.err
}

func ptrNow() *time.Time {
	now := time.Now()
	return &now
}

func TestPing(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping TestPing in short mode.")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	ds := GetTestDataStore(t)
	err := ds.Ping(ctx)
	assert.NoError(t, err)
}

func TestNewMongoStore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string
		CTX  context.Context

		Config MongoStoreConfig

		Error error
	}{{
		Name: "ok",

		Config: MongoStoreConfig{
			DbName: t.Name(),
			MongoURL: func() *url.URL {
				uri, _ := url.Parse(db.URL())
				return uri
			}(),
		},
	}, {
		Name: "error, bad uri scheme",

		Config: MongoStoreConfig{
			DbName: t.Name(),
			MongoURL: func() *url.URL {
				uri, _ := url.Parse(db.URL())
				uri.Scheme = "notMongo"
				return uri
			}(),
		},
		Error: errors.New("mongo: failed to connect with server"),
	}, {
		Name: "error, wrong username/password",

		Config: MongoStoreConfig{
			DbName: t.Name(),
			MongoURL: func() *url.URL {
				uri, _ := url.Parse(db.URL())
				return uri
			}(),
			Username: "user",
			Password: "password",
		},
		Error: errors.New("mongo: error reaching mongo server"),
	}, {
		Name: "error, wrong username/password",

		Config: MongoStoreConfig{
			DbName: t.Name(),
			MongoURL: func() *url.URL {
				uri, _ := url.Parse(db.URL())
				return uri
			}(),
			Username: "user",
			Password: "password",
		},
		Error: errors.New("^mongo: error reaching mongo server: "),
	}, {
		Name: "error, missing url",

		Config: MongoStoreConfig{
			DbName: t.Name(),
		},
		Error: errors.New("^mongo: missing URL"),
	}, {
		Name: "error, context canceled",
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),

		Config: MongoStoreConfig{
			DbName: t.Name(),
			MongoURL: func() *url.URL {
				uri, _ := url.Parse(db.URL())
				return uri
			}(),
			TLSConfig: &tls.Config{},
		},
		Error: errors.New("^mongo: error reaching mongo server:.*" +
			context.Canceled.Error(),
		),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					time.Second*5,
				)
				tc.CTX = ctx
				defer cancel()
			}
			ds, err := NewMongoStore(tc.CTX, tc.Config)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ds)
			}
			if ds != nil {
				ds.Close(tc.CTX)
			}
		})
	}
}

func TestInsertDevice(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX     context.Context
		Devices []model.Device

		Error error
	}{{
		Name: "ok",

		Devices: []model.Device{{
			ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			UpdatedTS: ptrNow(),
		}},
	}, {
		Name: "error, invalid document",

		Devices: []model.Device{{
			UpdatedTS: ptrNow(),
		}},
		Error: errors.New(`^invalid device object: id: cannot be blank.$`),
	}, {
		Name: "error, context canceled",
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),

		Devices: []model.Device{{
			ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			UpdatedTS: ptrNow(),
		}},
		Error: errors.New(
			`mongo: failed to store device configuration: .*` +
				context.Canceled.Error(),
		),
	}, {
		Name: "error, duplicate key",

		Devices: []model.Device{{
			ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			UpdatedTS: ptrNow(),
		}, {
			ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			UpdatedTS: ptrNow(),
		}},
		Error: store.ErrDeviceAlreadyExists,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var err error

			ds := GetTestDataStore(t)
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				tc.CTX = ctx
				defer ds.DropDatabase(tc.CTX)
			}
			for _, dev := range tc.Devices {
				err = ds.InsertDevice(tc.CTX, dev)
				if err != nil {
					break
				}
			}
			if tc.Error != nil {
				if assert.Error(t, err) {
					t.Logf("%T", errors.Cause(err))
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	t.Parallel()
	deviceID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String()
	testCases := []struct {
		Name string

		CTX          context.Context
		Devices      []model.Device
		FoundDevices []model.Device

		Error error
	}{
		{
			Name: "ok",

			Devices: []model.Device{
				{
					ID: deviceID,
					ConfiguredAttributes: []model.Attribute{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					ReportedAttributes: []model.Attribute{
						{
							Key:   "key2",
							Value: "value2",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
			FoundDevices: []model.Device{
				{
					ID: deviceID,
					ConfiguredAttributes: []model.Attribute{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					ReportedAttributes: []model.Attribute{
						{
							Key:   "key2",
							Value: "value2",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
		},

		{
			Name: "error device not found",

			Devices: []model.Device{
				{
					ID: deviceID,
					ConfiguredAttributes: []model.Attribute{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					ReportedAttributes: []model.Attribute{
						{
							Key:   "key2",
							Value: "value2",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
			FoundDevices: []model.Device{
				{
					ConfiguredAttributes: []model.Attribute{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					ReportedAttributes: []model.Attribute{
						{
							Key:   "key2",
							Value: "value2",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
			Error: errors.New("mongo: device does not exist"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var err error

			ds := GetTestDataStore(t)
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				tc.CTX = ctx
				defer ds.DropDatabase(tc.CTX)
			}
			for _, dev := range tc.Devices {
				err = ds.InsertDevice(tc.CTX, dev)
				if err != nil {
					break
				}
			}

			for _, dev := range tc.Devices {
				if tc.Error != nil {
					_, err := ds.GetDevice(tc.CTX, uuid.New().String())
					assert.EqualError(t, tc.Error, err.Error())
					break
				}
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				timeVal := time.Unix(1, 0)
				d.UpdatedTS = &timeVal
				dev.UpdatedTS = &timeVal
				assert.Equal(t, d, dev)
			}
		})
	}
}

func TestReplaceConfiguration(t *testing.T) {
	t.Parallel()
	deviceID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String()
	testCases := []struct {
		Name string

		CTX            context.Context
		Devices        []model.Device
		UpdatedDevices []model.Device

		Error error
	}{
		{
			Name: "ok new configured set",

			Devices: []model.Device{
				{
					ID:        deviceID,
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID: deviceID,
					ConfiguredAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
		},

		{
			Name: "ok removed configured",

			Devices: []model.Device{
				{
					ID: deviceID,
					ConfiguredAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID:                   deviceID,
					ConfiguredAttributes: model.Attributes{},
					UpdatedTS:            ptrNow(),
				},
			},
		},

		{
			Name: "error not a valid device",

			Devices: []model.Device{
				{
					ConfiguredAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID:                   deviceID,
					ConfiguredAttributes: model.Attributes{},
					UpdatedTS:            ptrNow(),
				},
			},

			Error: errors.New("invalid device object: id: cannot be blank."),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var err error

			ds := GetTestDataStore(t)
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				tc.CTX = ctx
				defer ds.DropDatabase(tc.CTX)
			}

			for _, dev := range tc.Devices {
				err = ds.ReplaceConfiguration(tc.CTX, dev)
				if err != nil {
					if tc.Error != nil {
						assert.EqualError(t, tc.Error, err.Error())
						return
					}
					break
				}
			}
			for _, dev := range tc.Devices {
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				assert.Equal(t, dev.ID, d.ID)
				assert.Equal(t, dev.ConfiguredAttributes, d.ConfiguredAttributes)
				assert.Equal(t, dev.ReportedAttributes, d.ReportedAttributes)
			}

			for _, dev := range tc.UpdatedDevices {
				err = ds.ReplaceConfiguration(tc.CTX, dev)
				if err != nil {
					break
				}
			}
			for _, dev := range tc.UpdatedDevices {
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				assert.Equal(t, dev.ID, d.ID)
				assert.Equal(t, dev.ConfiguredAttributes, d.ConfiguredAttributes)
				assert.Equal(t, dev.ReportedAttributes, d.ReportedAttributes)
			}

			assert.NoError(t, err)
		})
	}
}

func TestUpdateConfiguration(t *testing.T) {
	t.Parallel()
	var testSet = []model.Device{{
		ID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte("0")).String(),
		UpdatedTS: ptrNow(),
	}, {
		ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String(),
		ConfiguredAttributes: model.Attributes{{
			Key:   "key0",
			Value: "value0",
		}},
		ReportedAttributes: model.Attributes{{
			Key:   "key0",
			Value: "value0",
		}},
		UpdatedTS: ptrNow(),
	}, {
		ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("2")).String(),
		ConfiguredAttributes: model.Attributes{{
			Key:   "foo",
			Value: "bar",
		}},
		ReportedAttributes: model.Attributes{},
		UpdatedTS:          ptrNow(),
	}}
	testCases := []struct {
		Name string

		CTX            context.Context
		Devices        []model.Device
		UpdatedDevices []model.Device

		Error       error
		ErrorUpdate error
	}{{
		Name: "ok new configured set",

		UpdatedDevices: []model.Device{{
			ID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
			ConfiguredAttributes: model.Attributes{
				{
					Key:   "key0",
					Value: "value0",
				},
			},
			UpdatedTS: ptrNow(),
		}},
	}, {
		Name: "ok update configured attribute",

		UpdatedDevices: []model.Device{{
			ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String(),
			ConfiguredAttributes: model.Attributes{{
				Key:   "key1",
				Value: "value1",
			}},
			UpdatedTS: ptrNow(),
		}},
	}, {
		Name: "error/too many attributes",

		UpdatedDevices: []model.Device{{
			ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String(),
			ConfiguredAttributes: func() model.Attributes {
				attrs := make(model.Attributes, model.AttributesMaxLength+2)
				for i := range attrs {
					attrs[i] = model.Attribute{
						Key:   fmt.Sprintf("key%d", i),
						Value: fmt.Sprintf("value%d", i),
					}
				}
				return attrs
			}(),
			UpdatedTS: ptrNow(),
		}},

		ErrorUpdate: errors.New("too many configuration attributes, maximum is 100"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var err error
			ctx := context.Background()

			ds := GetTestDataStore(t)
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				tc.CTX = ctx
			}
			defer ds.DropDatabase(ctx)

			for _, dev := range testSet {
				err = ds.UpdateConfiguration(tc.CTX, dev.ID, dev.ConfiguredAttributes)
				if tc.Error != nil {
					if assert.Error(t, err) {
						assert.Regexp(t, tc.Error.Error(), err.Error())
					}
					return
				} else {
					require.NoError(t, err)

				}
			}

			for _, dev := range tc.UpdatedDevices {
				err = ds.UpdateConfiguration(tc.CTX, dev.ID, dev.ConfiguredAttributes)
				if tc.ErrorUpdate != nil {
					require.Error(t, err)
					assert.Regexp(t, tc.ErrorUpdate.Error(), err.Error())
					return
				}
			}
			for _, dev := range tc.UpdatedDevices {
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				assert.Equal(t, dev.ID, d.ID)
				for _, attr := range dev.ConfiguredAttributes {
					assert.Contains(t, d.ConfiguredAttributes, attr)
				}
			}

			assert.NoError(t, err)
		})
	}
}

func TestReplaceReportedConfiguration(t *testing.T) {
	t.Parallel()
	deviceID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String()
	testCases := []struct {
		Name string

		CTX            context.Context
		Devices        []model.Device
		UpdatedDevices []model.Device

		Error error
	}{
		{
			Name: "ok new reported set",

			Devices: []model.Device{
				{
					ID:        deviceID,
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID: deviceID,
					ReportedAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},
		},
		{
			Name: "ok removed reported",

			Devices: []model.Device{
				{
					ID: deviceID,
					ReportedAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID:                 deviceID,
					ReportedAttributes: model.Attributes{},
					UpdatedTS:          ptrNow(),
				},
			},
		},
		{
			Name: "error not a valid device",

			Devices: []model.Device{
				{
					ReportedAttributes: model.Attributes{
						{
							Key:   "key0",
							Value: "value0",
						},
					},
					UpdatedTS: ptrNow(),
				},
			},

			UpdatedDevices: []model.Device{
				{
					ID:                 deviceID,
					ReportedAttributes: model.Attributes{},
					UpdatedTS:          ptrNow(),
				},
			},

			Error: errors.New("invalid device object: id: cannot be blank."),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var err error

			ds := GetTestDataStore(t)
			if tc.CTX == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				tc.CTX = ctx
				defer ds.DropDatabase(tc.CTX)
			}

			for _, dev := range tc.Devices {
				err = ds.ReplaceReportedConfiguration(tc.CTX, dev)
				if err != nil {
					if tc.Error != nil {
						assert.EqualError(t, tc.Error, err.Error())
						return
					}
					break
				}
			}
			for _, dev := range tc.Devices {
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				assert.Equal(t, dev.ID, d.ID)
				assert.Equal(t, dev.ConfiguredAttributes, d.ConfiguredAttributes)
				assert.Equal(t, dev.ReportedAttributes, d.ReportedAttributes)
			}

			for _, dev := range tc.UpdatedDevices {
				err = ds.ReplaceReportedConfiguration(tc.CTX, dev)
				if err != nil {
					break
				}
			}
			for _, dev := range tc.UpdatedDevices {
				d, err := ds.GetDevice(tc.CTX, dev.ID)
				assert.NoError(t, err)
				assert.Equal(t, dev.ID, d.ID)
				assert.Equal(t, dev.ConfiguredAttributes, d.ConfiguredAttributes)
				assert.Equal(t, dev.ReportedAttributes, d.ReportedAttributes)
			}

			assert.NoError(t, err)
		})
	}
}

func TestSetDeploymentID(t *testing.T) {
	t.Parallel()

	var testDevice = model.Device{
		ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
		UpdatedTS: ptrNow(),
	}

	testCases := []struct {
		Name string

		CTX          context.Context
		ID           string
		DeploymentID uuid.UUID

		Error error
	}{{
		Name: "ok",

		ID:           testDevice.ID,
		DeploymentID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")),
	}, {
		Name: "ok, tenant",

		ID:           testDevice.ID,
		DeploymentID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")),

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "123456789012345678901234",
			},
		),
	}, {

		Name:         "error, device does not exist",
		DeploymentID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")),

		Error: errors.New(`^mongo: ` + store.ErrDeviceNoExist.Error()),
	}, {
		Name:         "error, context canceled",
		ID:           testDevice.ID,
		DeploymentID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")),
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),

		Error: errors.New(
			`mongo: failed to set the deployment ID: .*` +
				context.Canceled.Error() + `$`,
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			ds := GetTestDataStore(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			defer ds.DropDatabase(ctx)
			if tc.CTX == nil {
				tc.CTX = ctx
			} else if id := identity.FromContext(tc.CTX); id != nil {
				ctx = identity.WithContext(ctx, id)
			}
			err := ds.InsertDevice(ctx, testDevice)
			require.NoError(t, err)

			err = ds.SetDeploymentID(tc.CTX, tc.ID, tc.DeploymentID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	t.Parallel()

	var testDevice = model.Device{
		ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
		UpdatedTS: ptrNow(),
	}

	testCases := []struct {
		Name string

		CTX context.Context
		ID  string

		Error error
	}{{
		Name: "ok",

		ID: testDevice.ID,
	}, {
		Name: "ok, tenant",

		ID: testDevice.ID,

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "123456789012345678901234",
			},
		),
	}, {

		Name: "error, device does not exist",

		Error: errors.New(`^mongo: ` + store.ErrDeviceNoExist.Error()),
	}, {
		Name: "error, context canceled",
		ID:   testDevice.ID,
		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),

		Error: errors.New(
			`mongo: failed to delete device configuration: .*` +
				context.Canceled.Error() + `$`,
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			ds := GetTestDataStore(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			defer ds.DropDatabase(ctx)
			if tc.CTX == nil {
				tc.CTX = ctx
			} else if id := identity.FromContext(tc.CTX); id != nil {
				ctx = identity.WithContext(ctx, id)
			}
			err := ds.InsertDevice(ctx, testDevice)
			require.NoError(t, err)

			err = ds.DeleteDevice(tc.CTX, tc.ID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteTenant in short mode.")
	}

	const tenant = "foo"
	ctx := identity.WithContext(context.Background(),
		&identity.Identity{
			Tenant: tenant,
		},
	)

	d := GetTestDataStore(t)

	var testDevice = model.Device{
		ID:        uuid.NewSHA1(uuid.NameSpaceDNS, []byte("mender.io")).String(),
		UpdatedTS: ptrNow(),
	}
	err := d.InsertDevice(ctx, testDevice)
	require.NoError(t, err)

	err = d.DeleteTenant(ctx, tenant)
	assert.NoError(t, err)

	_, err = d.GetDevice(ctx, testDevice.ID)
	assert.Error(t, err, store.ErrDeviceNoExist.Error())
}
