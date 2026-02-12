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

package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
)

func castInterfaceSlice(sliceIn interface{}) (sliceOut []interface{}) {
	rSlice := reflect.ValueOf(sliceIn)
	if rSlice.Kind() != reflect.Slice {
		panic("[PROG ERR] sliceIn is not a Slice")
	}
	l := rSlice.Len()
	sliceOut = make([]interface{}, l)
	for i := 0; i < l; i++ {
		sliceOut[i] = rSlice.Index(i).Interface()
	}
	return sliceOut
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	testCases := []struct {
		Name string

		CTX         context.Context
		Integration model.Integration

		Error error
	}{{
		Name: "ok",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "1234567890",
		}),
		Integration: model.Integration{
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type: "connection_string",
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      crypto.String("secret"),
					Name:     "foobar",
				},
			},
		},
	}, {
		Name: "ok, no tenant context",

		CTX: context.Background(),
		Integration: model.Integration{
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type: "connection_string",
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      crypto.String("secret"),
					Name:     "foobar",
				},
			},
		},
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cc := context.WithCancel(context.Background())
			cc()
			return ctx
		}(),
		Integration: model.Integration{
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type: "connection_string",
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      crypto.String("secret"),
					Name:     "foobar",
				},
			},
		},
		Error: context.Canceled,
	}}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		ds := NewDataStoreWithClient(dbClient, NewConfig().
			SetDbName(dbName))
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			_, err := ds.CreateIntegration(tc.CTX, tc.Integration)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				idty := identity.FromContext(tc.CTX)
				var tenantID string
				if idty != nil {
					tenantID = idty.Tenant
				}
				fltr := bson.D{{
					Key: "tenant_id", Value: tenantID,
				}}

				var doc bson.Raw
				err := dbClient.Database(dbName).
					Collection(CollNameIntegrations).
					FindOne(tc.CTX, fltr).
					Decode(&doc)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				field := doc.Lookup(KeyTenantID)
				actualTID, ok := field.StringValueOK()
				assert.True(t, ok, "bson document does not contain tenant_id field")
				assert.Equal(t, tenantID, actualTID)

				var integration model.Integration
				bson.UnmarshalWithRegistry(newRegistry(), doc, &integration)
				assert.True(t, uuid.Validate(integration.ID.String()) == nil)
				integration.ID = uuid.Nil
				tc.Integration.ID = uuid.Nil
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}

func TestGetIntegrations(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	type testCase struct {
		Name string

		CTX context.Context

		InitDatabase func(self *testCase, coll *mongo.Collection)

		IntegrationFilter model.IntegrationFilter

		Integrations []model.Integration
		Error        error
	}
	testCases := []testCase{
		{
			Name: "ok got integration",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(self.Integrations)
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			Integrations: []model.Integration{
				{
					Provider: model.ProviderIoTHub,
					Credentials: model.Credentials{
						Type: "connection_string",
						ConnectionString: &model.ConnectionString{
							HostName: "localhost",
							Key:      crypto.String("secret"),
							Name:     "foobar",
						},
					},
				},
			},
		},
		{
			Name: "ok filter by many IDs",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(append(self.Integrations,
					model.Integration{
						ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'3'}),
						Provider: model.ProviderIoTHub,
						Credentials: model.Credentials{
							Type: "connection_string",
							ConnectionString: &model.ConnectionString{
								HostName: "localhost",
								Key:      crypto.String("secret"),
								Name:     "idk",
							},
						},
					},
					model.Integration{
						ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'4'}),
						Provider: model.ProviderIoTHub,
						Credentials: model.Credentials{
							Type: "connection_string",
							ConnectionString: &model.ConnectionString{
								HostName: "localhost",
								Key:      crypto.String("secret"),
								Name:     "srsly/idk",
							},
						},
					},
				))
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			IntegrationFilter: model.IntegrationFilter{
				IDs: []uuid.UUID{
					uuid.NewSHA1(uuid.NameSpaceOID, []byte{'0'}),
					uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
					uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}),
				},
			},

			Integrations: []model.Integration{{
				ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: "connection_string",
					ConnectionString: &model.ConnectionString{
						HostName: "localhost",
						Key:      crypto.String("supersecret"),
						Name:     "barbaz",
					},
				},
			}, {
				ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: "connection_string",
					ConnectionString: &model.ConnectionString{
						HostName: "localhost",
						Key:      crypto.String("secret"),
						Name:     "foobar",
					},
				},
			}},
		},
		{
			Name: "ok filter by ID",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(append(self.Integrations,
					model.Integration{
						ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'3'}),
						Provider: model.ProviderIoTHub,
						Credentials: model.Credentials{
							Type: "connection_string",
							ConnectionString: &model.ConnectionString{
								HostName: "localhost",
								Key:      crypto.String("secret"),
								Name:     "idk",
							},
						},
					},
					model.Integration{
						ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte{'4'}),
						Provider: model.ProviderIoTHub,
						Credentials: model.Credentials{
							Type: "connection_string",
							ConnectionString: &model.ConnectionString{
								HostName: "localhost",
								Key:      crypto.String("secret"),
								Name:     "srsly/idk",
							},
						},
					},
				))
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			IntegrationFilter: model.IntegrationFilter{
				IDs:      []uuid.UUID{uuid.Nil},
				Provider: model.ProviderIoTHub,
				Limit:    1,
			},

			Integrations: []model.Integration{{
				ID:       uuid.Nil,
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: "connection_string",
					ConnectionString: &model.ConnectionString{
						HostName: "localhost",
						Key:      crypto.String("supersecret"),
						Name:     "barbaz",
					},
				},
			}},
		},
		{
			Name: "ok/noop",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),

			IntegrationFilter: model.IntegrationFilter{
				IDs: []uuid.UUID{},
			},
			Integrations: []model.Integration{},
		},
		{
			Name: "no integrations for tenant",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
			Integrations: []model.Integration{},
		},
		{
			Name: "error/foul document",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				doc := mstore.WithTenantID(self.CTX, map[string]interface{}{
					"_id":         "1234567890",
					"credentials": "correcthorsebatterystaple",
				})
				_, err := coll.InsertOne(context.Background(), doc)
				if err != nil {
					panic(err)
				}
			},
			Error: errors.New("error retrieving integrations collection results"),
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collIntegrations := dbClient.
				Database(dbName).
				Collection(CollNameIntegrations)

			if tc.InitDatabase != nil {
				tc.InitDatabase(&tc, collIntegrations)
			}
			db := NewDataStoreWithClient(dbClient, NewConfig().
				SetDbName(dbName))
			integrations, err := db.GetIntegrations(tc.CTX, tc.IntegrationFilter)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integrations, integrations)
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const deviceID = "1"
	const tenantID = "123456789012345678901234"
	testCases := []struct {
		Name string

		CTX context.Context

		Device *model.Device
		Error  error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Device: &model.Device{
				ID: deviceID,
			},
		},
		{
			Name: "not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),

			Error: store.ErrObjectNotFound,
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collDevices := dbClient.
				Database(dbName).
				Collection(CollNameDevices)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Device != nil {
				_, err := collDevices.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Device),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(dbClient, NewConfig().
				SetDbName(dbName))
			device, err := db.GetDevice(tc.CTX, deviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Device, device)
			}
		})
	}
}

func TestGetIntegrationById(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		CTX context.Context

		Integration *model.Integration
		Error       error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Integration: &model.Integration{
				ID: integrationID,
			},
		},
		{
			Name: "not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
			Error: store.ErrObjectNotFound,
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			client := db.Client()
			collIntegrations := client.
				Database(dbName).
				Collection(CollNameIntegrations)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Integration != nil {
				_, err := collIntegrations.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Integration),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(dbClient, NewConfig().
				SetDbName(dbName))
			integration, err := db.GetIntegrationById(tc.CTX, integrationID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}

func TestSetIntegrationCredentials(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	integrationID := uuid.New()
	testCases := []struct {
		Name string

		CTX           context.Context
		Credentials   *model.Credentials
		IntegrationID uuid.UUID
		Error         error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),

			Credentials: &model.Credentials{
				Type: model.CredentialTypeSAS,
				ConnectionString: &model.ConnectionString{
					HostName: "test-update.azure-devices.net",
					Name:     "new-test-policy",
					Key:      crypto.String("new-key"),
				},
			},
			IntegrationID: integrationID,
		},
		{
			Name: "error, integration not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			IntegrationID: uuid.New(),
			Credentials: &model.Credentials{
				Type: model.CredentialTypeSAS,
				ConnectionString: &model.ConnectionString{
					HostName: "test-update.azure-devices.net",
					Name:     "new-test-policy",
					Key:      crypto.String("new-key"),
				},
			},
			Error: store.ErrObjectNotFound,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collIntegrations := dbClient.Database(dbName).Collection(CollNameIntegrations)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Credentials != nil {
				_, err := collIntegrations.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, model.Integration{
						ID: integrationID,
						Credentials: model.Credentials{
							Type: model.CredentialTypeSAS,
							ConnectionString: &model.ConnectionString{
								HostName: "test.azure-devices.net",
								Name:     "test-policy",
								Key:      crypto.String("eMB7VENgpPsIl+aVeAYjstMpuIyoQxY2eOqpzpqI/LF8="),
							},
						},
					}),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(dbClient, NewConfig().SetDbName(dbName))
			err := db.SetIntegrationCredentials(tc.CTX, tc.IntegrationID, *tc.Credentials)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveIntegration(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	integrationID := uuid.New()
	testCases := []struct {
		Name string

		CTX           context.Context
		IntegrationID uuid.UUID
		Error         error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			IntegrationID: integrationID,
		},
		{
			Name: "error: object not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			IntegrationID: uuid.New(),
			Error:         store.ErrObjectNotFound,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			client := db.Client()
			collIntegrations := client.Database(dbName).Collection(CollNameIntegrations)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			_, err := collIntegrations.InsertMany(ctx, []interface{}{
				mstore.WithTenantID(ctx, model.Integration{
					ID:          integrationID,
					Provider:    model.ProviderIoTHub,
					Credentials: model.Credentials{},
				}),
			})
			assert.NoError(t, err)

			db := NewDataStoreWithClient(dbClient, NewConfig().SetDbName(dbName))
			err = db.RemoveIntegration(tc.CTX, tc.IntegrationID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDoDevicesExistByIntegrationID(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	integrationID := uuid.New()
	testCases := []struct {
		Name string

		CTX            context.Context
		Devices        []*model.Device
		IntegrationID  uuid.UUID
		Integration    *model.Integration
		ExpectedResult bool
		Error          error
	}{
		{
			Name: "ok, devices exist",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			IntegrationID: integrationID,
			Integration: &model.Integration{
				ID: integrationID,
			},
			Devices: []*model.Device{{
				ID:             uuid.NewString(),
				IntegrationIDs: []uuid.UUID{integrationID},
			}},
			ExpectedResult: true,
		},
		{
			Name: "ok, no devices with given integration ID",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			IntegrationID: integrationID,
			Integration: &model.Integration{
				ID: integrationID,
			},
			ExpectedResult: false,
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collIntegrations := dbClient.Database(dbName).Collection(CollNameIntegrations)
			collDevices := dbClient.Database(dbName).Collection(CollNameDevices)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Integration != nil {
				_, err := collIntegrations.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Integration),
				})
				assert.NoError(t, err)
			}

			if tc.Devices != nil {
				_, err := collDevices.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, model.Device{
						ID:             uuid.NewString(),
						IntegrationIDs: []uuid.UUID{tc.IntegrationID},
					}),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(dbClient, NewConfig().SetDbName(dbName))
			result, err := db.DoDevicesExistByIntegrationID(tc.CTX, tc.IntegrationID)
			if tc.Error != nil {
				assert.Equal(t, tc.ExpectedResult, result)
				assert.Equal(t, tc.Error, err)
			} else {
				assert.Equal(t, tc.ExpectedResult, result)
				assert.NoError(t, err)
			}
		})
	}
}

func testSetDevices() []model.Device {

	newUUID := func(dgst string) string {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(dgst)).String()
	}
	return []model.Device{{
		ID: newUUID("1"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")),
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("2")),
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("3")),
		},
	}, {
		ID: newUUID("2"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")),
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("2")),
		},
	}, {
		ID: newUUID("3"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("2")),
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("3")),
		},
	}, {
		ID: newUUID("4"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")),
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("3")),
		},
	}, {
		ID: newUUID("5"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")),
		},
	}, {
		ID: newUUID("6"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("2")),
		},
	}, {
		ID: newUUID("7"),
		IntegrationIDs: []uuid.UUID{
			uuid.NewSHA1(uuid.NameSpaceOID, []byte("3")),
		},
	}, {
		ID:             newUUID("8"),
		IntegrationIDs: []uuid.UUID{},
	}}
}

func insertDevices(ctx context.Context, db *mongo.Database, devices []model.Device) {
	collDevices := db.Collection(CollNameDevices)
	for _, device := range devices {
		_, err := collDevices.InsertOne(ctx, mstore.WithTenantID(ctx, device))
		if err != nil {
			panic(err)
		}
	}
}

func TestGetAndDeleteDevice(t *testing.T) {
	t.Parallel()
	dbName := t.Name()
	ds := NewDataStoreWithClient(
		db.Client(),
		NewConfig().SetDbName(dbName),
	)

	ctxEmpty := context.Background()
	ctxTenant := identity.WithContext(ctxEmpty, &identity.Identity{
		Tenant: "123456789012345678901234",
	})
	database := db.Client().Database(dbName)
	defer database.Drop(ctxEmpty)
	devices := testSetDevices()
	insertDevices(ctxEmpty, database, devices[:5])
	insertDevices(ctxTenant, database, devices[5:])

	dev, err := ds.GetDevice(ctxEmpty, uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String())
	assert.NoError(t, err)
	assert.Equal(t, &devices[0], dev)
	dev, err = ds.GetDevice(ctxTenant, uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String())
	assert.EqualError(t, err, store.ErrObjectNotFound.Error())

	err = ds.DeleteDevice(ctxEmpty, uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String())
	assert.NoError(t, err)
	err = ds.DeleteDevice(ctxEmpty, uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String())
	assert.EqualError(t, err, store.ErrObjectNotFound.Error())
	dev, err = ds.GetDevice(ctxEmpty, uuid.NewSHA1(uuid.NameSpaceOID, []byte("1")).String())
	assert.EqualError(t, err, store.ErrObjectNotFound.Error())

	// Context with identity
	dev, err = ds.GetDevice(ctxTenant, uuid.NewSHA1(uuid.NameSpaceOID, []byte("7")).String())
	assert.NoError(t, err)
	assert.Equal(t, &devices[6], dev)

	err = ds.DeleteDevice(ctxTenant, uuid.NewSHA1(uuid.NameSpaceOID, []byte("7")).String())
	assert.NoError(t, err)
	err = ds.DeleteDevice(ctxTenant, uuid.NewSHA1(uuid.NameSpaceOID, []byte("7")).String())
	assert.EqualError(t, err, store.ErrObjectNotFound.Error())

	dev, err = ds.GetDevice(ctxTenant, uuid.NewSHA1(uuid.NameSpaceOID, []byte("7")).String())
	assert.EqualError(t, err, store.ErrObjectNotFound.Error())

	ctxCancelled, cancel := context.WithCancel(ctxTenant)
	cancel()
	err = ds.DeleteDevice(ctxCancelled,
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("6")).String(),
	)
	assert.Error(t, err)

}

func TestGetDeviceByIntegrationID(tp *testing.T) {
	tp.Parallel()
	testCases := []struct {
		Name string

		CTX           context.Context
		DeviceID      string
		IntegrationID uuid.UUID

		Device *model.Device
		Error  error
	}{{
		Name: "ok",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "123456789012345678901234",
			},
		),
		DeviceID:      uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}).String(),
		IntegrationID: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),

		Device: &testSetDevices()[0],
	}, {
		Name: "ok/no tenant",

		CTX:           context.Background(),
		DeviceID:      uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}).String(),
		IntegrationID: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),

		Device: &testSetDevices()[0],
	}, {
		Name: "error/not found",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Tenant: "123456789012345678901234",
			},
		),
		DeviceID:      uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}).String(),
		IntegrationID: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'9'}),

		Error: store.ErrObjectNotFound,
	}, {
		Name: "error/context cancelled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		DeviceID:      uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}).String(),
		IntegrationID: uuid.NewSHA1(uuid.NameSpaceOID, []byte{'9'}),

		Error: context.Canceled,
	}}
	for i := range testCases {
		tc := testCases[i]
		ti := i
		tp.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			dbName := fmt.Sprintf("%s-%d", tp.Name(), ti)
			ds := NewDataStoreWithClient(
				db.Client(),
				NewConfig().SetDbName(dbName),
			)
			database := db.Client().Database(dbName)
			defer database.Drop(context.Background())

			id := identity.FromContext(tc.CTX)
			ctx := identity.WithContext(context.Background(), id)
			insertDevices(ctx, database, testSetDevices())

			dev, err := ds.GetDeviceByIntegrationID(
				tc.CTX,
				tc.DeviceID,
				tc.IntegrationID,
			)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tc.Device, dev)
				}
			}
		})
	}
}

func TestUpsertDeviceIntegrations(t *testing.T) {
	t.Parallel()
	client := db.Client()
	type testCase struct {
		Name string

		CTX          context.Context
		InitDatabase func(self *testCase, coll *mongo.Collection)
		UpsertDevice model.Device

		Device *model.Device
		Error  error
	}
	testCases := []testCase{{
		Name: "ok/create devices",

		CTX: context.Background(),
		UpsertDevice: model.Device{
			ID: uuid.Nil.String(),
			IntegrationIDs: []uuid.UUID{
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
			},
		},

		Device: &model.Device{
			ID: uuid.Nil.String(),
			IntegrationIDs: []uuid.UUID{
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
			},
		},
	}, {
		Name: "ok/add integration",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{Tenant: "123456789012345678901234"},
		),
		InitDatabase: func(self *testCase, coll *mongo.Collection) {
			doc := mstore.WithTenantID(self.CTX, model.Device{
				ID: uuid.Nil.String(),
				IntegrationIDs: []uuid.UUID{
					uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
				},
			})
			_, err := coll.InsertOne(self.CTX, doc)
			if err != nil {
				panic(err)
			}
		},
		UpsertDevice: model.Device{
			ID: uuid.Nil.String(),
			IntegrationIDs: []uuid.UUID{
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}),
			},
		},

		Device: &model.Device{
			ID: uuid.Nil.String(),
			IntegrationIDs: []uuid.UUID{
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'2'}),
			},
		},
	}, {
		Name: "ok/noop",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{Tenant: "123456789012345678901234"},
		),
		InitDatabase: func(self *testCase, coll *mongo.Collection) {
			doc := mstore.WithTenantID(self.CTX, model.Device{
				ID: uuid.Nil.String(),
				IntegrationIDs: []uuid.UUID{
					uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
				},
			})
			_, err := coll.InsertOne(self.CTX, doc)
			if err != nil {
				panic(err)
			}
		},
		UpsertDevice: model.Device{
			ID: uuid.Nil.String(),
		},

		Device: &model.Device{
			ID: uuid.Nil.String(),
			IntegrationIDs: []uuid.UUID{
				uuid.NewSHA1(uuid.NameSpaceOID, []byte{'1'}),
			},
		},
	}, {
		Name: "error/context canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		UpsertDevice: model.Device{
			ID: uuid.Nil.String(),
		},

		Error: context.Canceled,
	}}
	for i := range testCases {
		tc := testCases[i]
		ds := NewDataStoreWithClient(client, NewConfig().
			SetDbName(fmt.Sprintf("%s-%d", t.Name(), i)),
		)
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			if tc.InitDatabase != nil {
				collDevs := client.Database(*ds.DbName).
					Collection(CollNameDevices)
				tc.InitDatabase(&tc, collDevs)
			}

			new, err := ds.UpsertDeviceIntegrations(tc.CTX,
				tc.UpsertDevice.ID,
				tc.UpsertDevice.IntegrationIDs,
			)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tc.Device, new)
				}
			}
		})
	}
}

func TestDeleteTenantData(t *testing.T) {
	t.Parallel()
	client := db.Client()
	type testCase struct {
		Name            string
		TenantToDelete  string
		SomeOtherTenant string
	}
	testCases := []testCase{
		{
			Name: "ok",

			TenantToDelete:  primitive.NewObjectID().Hex(),
			SomeOtherTenant: primitive.NewObjectID().Hex(),
		},
		{
			Name:            "no id",
			SomeOtherTenant: primitive.NewObjectID().Hex(),
		},
		{
			Name:            "no tenant id in id",
			TenantToDelete:  "-",
			SomeOtherTenant: primitive.NewObjectID().Hex(),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		ds := NewDataStoreWithClient(client, NewConfig().
			SetDbName(fmt.Sprintf("%s-%d", t.Name(), i)),
		)
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			database := client.Database(*ds.DbName)
			collections := []*mongo.Collection{
				database.Collection(CollNameLog),
				database.Collection(CollNameDevices),
				database.Collection(CollNameIntegrations),
			}
			ctx := context.Background()

			if len(tc.TenantToDelete) > 0 {
				tenantId := tc.TenantToDelete
				if tenantId == "-" {
					tenantId = ""
				}
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tenantId,
				})
			}
			var err error
			for _, c := range collections {
				_, _ = c.DeleteMany(ctx, bson.M{})
				_, err = c.InsertOne(
					ctx,
					map[string]string{
						KeyTenantID: tc.TenantToDelete,
						"data":      uuid.New().String(),
					},
				)
				assert.NoError(t, err)
				_, err = c.InsertOne(
					ctx,
					map[string]string{
						KeyTenantID: tc.SomeOtherTenant,
						"data":      uuid.New().String(),
					},
				)
				assert.NoError(t, err)
			}
			for _, c := range collections {
				count, _ := c.CountDocuments(ctx, bson.M{KeyTenantID: tc.TenantToDelete})
				assert.Equal(t, 1, int(count), "inserted documents count to delete")
				count, _ = c.CountDocuments(ctx, bson.M{KeyTenantID: tc.SomeOtherTenant})
				assert.Equal(t, 1, int(count), "inserted documents count to keep")
			}
			err = ds.DeleteTenantData(ctx)
			if len(tc.TenantToDelete) == 0 {
				assert.Error(t, err)
			} else if tc.TenantToDelete == "-" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, c := range collections {
				count, _ := c.CountDocuments(ctx, bson.M{KeyTenantID: tc.TenantToDelete})
				expected := 0
				if tc.TenantToDelete == "" || tc.TenantToDelete == "-" {
					expected = 1 // we expect all the data to be left alone if tenant idinthe ctx was empty
				}
				assert.Equal(t, expected, int(count), "inserted documents count to delete")
				count, _ = c.CountDocuments(ctx, bson.M{KeyTenantID: tc.SomeOtherTenant})
				assert.Equal(t, 1, int(count), "inserted documents count to keep")
			}
		})
	}
}
