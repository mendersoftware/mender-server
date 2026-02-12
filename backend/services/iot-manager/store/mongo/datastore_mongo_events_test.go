// Copyright 2022 Northern.tech AS
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
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

func TestGetEvents(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	integrationId := uuid.New().String()
	success200 := 200
	now := time.Now().Local()
	type testCase struct {
		Name string

		CTX context.Context

		InitDatabase func(self *testCase, coll *mongo.Collection)

		EventFilter model.EventsFilter

		InEvents  []model.Event
		OutEvents []model.Event
		Error     error
	}
	testCases := []testCase{
		{
			Name: "ok got event",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(self.InEvents)
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			InEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
					},
				},
			},
			OutEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceStatusChanged,
						Data: bson.M{
							"id":     "foo",
							"status": "bar",
						},
					},
				},
			},
		},
		{
			Name: "ok, many",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(self.InEvents)
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			InEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceDecommissioned,
						Data: model.DeviceEvent{
							ID: "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "baz",
						},
					},
				},
			},
			OutEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceStatusChanged,
						Data: bson.M{
							"id":     "foo",
							"status": "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceDecommissioned,
						Data: bson.M{
							"id": "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceProvisioned,
						Data: bson.M{
							"id": "baz",
						},
					},
				},
			},
		},
		{
			Name: "ok, with filter",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(self.InEvents)
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			EventFilter: model.EventsFilter{
				Skip:  1,
				Limit: 1,
			},

			InEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
						EventTS: now,
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceDecommissioned,
						Data: model.DeviceEvent{
							ID: "bar",
						},
						EventTS: now.Add(time.Second),
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "baz",
						},
						EventTS: now.Add(time.Second * 2),
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "foo-bar-baz",
						},
						EventTS: now.Add(time.Second * 3),
					},
				},
			},
			OutEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceProvisioned,
						Data: bson.M{
							"id": "baz",
						},
					},
				},
			},
		},
		{
			Name: "ok, with filter with integration id",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			InitDatabase: func(
				self *testCase,
				coll *mongo.Collection,
			) {
				docFace := castInterfaceSlice(self.InEvents)
				docs := mstore.ArrayWithTenantID(self.CTX, docFace)
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					panic(err)
				}
			},

			EventFilter: model.EventsFilter{
				IntegrationID: &integrationId,
			},

			InEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
						EventTS: now,
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceDecommissioned,
						Data: model.DeviceEvent{
							ID: "bar-" + uuid.MustParse(integrationId).String(),
						},
						EventTS: now.Add(time.Second),
					},
					DeliveryStatus: []model.DeliveryStatus{
						{
							IntegrationID: uuid.MustParse(integrationId),
							Success:       true,
							StatusCode:    &success200,
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "baz",
						},
						EventTS: now.Add(time.Second * 2),
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						ID:   uuid.New(),
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "foo-bar-baz",
						},
						EventTS: now.Add(time.Second * 3),
					},
				},
			},
			OutEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceDecommissioned,
						Data: bson.M{
							"id": "bar-" + uuid.MustParse(integrationId).String(),
						},
					},
					DeliveryStatus: []model.DeliveryStatus{
						{
							IntegrationID: uuid.MustParse(integrationId),
							Success:       true,
							StatusCode:    &success200,
						},
					},
				},
			},
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collEvents := dbClient.
				Database(dbName).
				Collection(CollNameLog)

			if tc.InitDatabase != nil {
				tc.InitDatabase(&tc, collEvents)
			}
			db := NewDataStoreWithClient(dbClient, NewConfig().
				SetDbName(dbName))
			events, err := db.GetEvents(tc.CTX, tc.EventFilter)
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
				for i := range events {
					events[i].ID = uuid.Nil
					events[i].EventTS = time.Time{}
				}
				assert.Equal(t, tc.OutEvents, events)
			}
		})
	}
}

func TestSaveEvent(t *testing.T) {
	t.Parallel()
	dbClient := db.Client()
	const tenantID = "123456789012345678901234"
	type testCase struct {
		Name string

		CTX context.Context

		InEvents []model.Event

		OutEvents []model.Event
		Error     error
	}
	testCases := []testCase{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),

			InEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "baz",
						},
					},
				},
			},
			OutEvents: []model.Event{
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceStatusChanged,
						Data: model.DeviceEvent{
							ID:     "foo",
							Status: "bar",
						},
					},
				},
				{
					WebhookEvent: model.WebhookEvent{
						Type: model.EventTypeDeviceProvisioned,
						Data: model.DeviceEvent{
							ID: "baz",
						},
					},
				},
			},
		},
	}
	for i := range testCases {
		dbName := fmt.Sprintf("%s-%d", t.Name(), i)
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			defer dbClient.Database(dbName).Drop(context.Background())
			collEvents := dbClient.
				Database(dbName).
				Collection(CollNameLog)

			db := NewDataStoreWithClient(dbClient, NewConfig().
				SetDbName(dbName))
			for _, e := range tc.InEvents {
				err := db.SaveEvent(tc.CTX, e)
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
			}
			if tc.Error != nil {
				cur, err := collEvents.Find(tc.CTX,
					mstore.WithTenantID(tc.CTX, bson.D{}),
				)
				assert.NoError(t, err)
				events := []model.Event{}
				err = cur.All(tc.CTX, &events)
				assert.NoError(t, err)
				for i := range events {
					events[i].ID = uuid.Nil
					events[i].EventTS = time.Time{}
					events[i].ExpireTS = time.Time{}
				}
				assert.Equal(t, tc.OutEvents, events)
			}
		})
	}
}
