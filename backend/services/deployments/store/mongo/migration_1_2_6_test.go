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
	"reflect"
	"testing"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/pkg/store"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	tDeviceDeploymentStatus = reflect.TypeOf(model.DeviceDeploymentStatus(0))

	oldBSONReg = bson.NewRegistryBuilder().
			RegisterTypeEncoder(tDeviceDeploymentStatus, oldStatusCodec{}).
			RegisterTypeDecoder(tDeviceDeploymentStatus, oldStatusCodec{}).
			Build()
)

type oldStatusCodec struct{}

func (_ oldStatusCodec) EncodeValue(
	ec bsoncodec.EncodeContext,
	vw bsonrw.ValueWriter,
	val reflect.Value,
) error {
	if !val.IsValid() || val.Type() != tDeviceDeploymentStatus {
		return bsoncodec.ValueEncoderError{
			Name:     "oldStatusCodec",
			Types:    []reflect.Type{tDeviceDeploymentStatus},
			Received: val,
		}
	}
	status := val.Interface().(model.DeviceDeploymentStatus)
	return vw.WriteString(status.String())
}

func (_ oldStatusCodec) DecodeValue(
	dc bsoncodec.DecodeContext,
	vr bsonrw.ValueReader,
	val reflect.Value,
) error {
	if !val.IsValid() || val.Type() != tDeviceDeploymentStatus {
		return bsoncodec.ValueEncoderError{
			Name:     "oldStatusCodec",
			Types:    []reflect.Type{tDeviceDeploymentStatus},
			Received: val,
		}
	}
	s, err := vr.ReadString()
	if err != nil {
		return err
	}
	var status model.DeviceDeploymentStatus
	err = status.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(status))
	return nil
}

var testSet126 = []interface{}{model.DeviceDeployment{
	Id:           "000000000000000000000000",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000000",
	Status:       model.DeviceDeploymentStatusFailure,
}, model.DeviceDeployment{
	Id:           "000000000000000000000001",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000001",
	Status:       model.DeviceDeploymentStatusPauseBeforeInstall,
}, model.DeviceDeployment{
	Id:           "000000000000000000000002",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000002",
	Status:       model.DeviceDeploymentStatusPauseBeforeCommit,
}, model.DeviceDeployment{
	Id:           "000000000000000000000003",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000003",
	Status:       model.DeviceDeploymentStatusPauseBeforeReboot,
}, model.DeviceDeployment{
	Id:           "000000000000000000000004",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000004",
	Status:       model.DeviceDeploymentStatusDownloading,
}, model.DeviceDeployment{
	Id:           "000000000000000000000005",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000005",
	Status:       model.DeviceDeploymentStatusInstalling,
}, model.DeviceDeployment{
	Id:           "000000000000000000000006",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000006",
	Status:       model.DeviceDeploymentStatusRebooting,
}, model.DeviceDeployment{
	Id:           "000000000000000000000007",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000007",
	Status:       model.DeviceDeploymentStatusPending,
}, model.DeviceDeployment{
	Id:           "000000000000000000000008",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000008",
	Status:       model.DeviceDeploymentStatusSuccess,
}, model.DeviceDeployment{
	Id:           "000000000000000000000009",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-000000000009",
	Status:       model.DeviceDeploymentStatusAborted,
}, model.DeviceDeployment{
	Id:           "00000000000000000000000a",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-00000000000a",
	Status:       model.DeviceDeploymentStatusNoArtifact,
}, model.DeviceDeployment{
	Id:           "00000000000000000000000b",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-00000000000b",
	Status:       model.DeviceDeploymentStatusAlreadyInst,
}, model.DeviceDeployment{
	Id:           "00000000000000000000000c",
	DeploymentId: "00000000-0000-0000-0000-000000000000",
	DeviceId:     "00000000-0000-0000-0000-00000000000c",
	Status:       model.DeviceDeploymentStatusDecommissioned,
}}

func TestMigration126(t *testing.T) {
	testCases := []struct {
		Name string
		// TODO add test case for single and multi-tenant setup
		CTX context.Context
	}{{
		Name: "ok, single tenant mode",
		CTX:  context.Background(),
	}, {
		Name: "ok, multi-tenant mode",
		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "123456789012345678901234",
		}),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			defer db.Wipe()
			client := db.Client()
			dbName := store.DbFromContext(tc.CTX, DbName)

			db := client.Database(dbName)
			collDevs := db.Collection(CollectionDevices)
			collDevsOldCodec := db.Collection(
				CollectionDevices,
				options.Collection().SetRegistry(oldBSONReg),
			)
			// Insert with old schema
			_, err := collDevsOldCodec.InsertMany(tc.CTX, testSet126)
			if err != nil {
				panic(err)
			}
			cur, err := collDevs.Find(tc.CTX, bson.M{})
			assert.NoError(t, err)
			var res []model.DeviceDeployment
			// Trying to decode using default decoder will fail
			err = cur.All(tc.CTX, &res)
			assert.Error(t, err)

			// 2. Run migrations
			migrations := []migrate.Migration{
				&migration_1_2_6{
					client: client,
					db:     dbName,
				},
			}
			migrator := migrate.SimpleMigrator{
				Client:      client,
				Db:          dbName,
				Automigrate: true,
			}
			err = migrator.Apply(tc.CTX, migrate.MakeVersion(1, 2, 6), migrations)
			assert.NoError(t, err)

			cur, err = collDevs.Find(tc.CTX, bson.M{})
			assert.NoError(t, err)

			err = cur.All(tc.CTX, &res)
			// If the migration was unsuccessful, there will
			// be an error decoding the documents.
			assert.NoError(t, err)

			// Run explain on query from GetDevicesListForDeployment
			// and check that the newly created index is used.
			singleRes := db.RunCommand(tc.CTX, bson.D{{
				Key: "explain", Value: bson.D{{
					Key: "find", Value: CollectionDevices,
				}, {
					Key: "filter", Value: bson.D{{
						Key:   StorageKeyDeviceDeploymentDeploymentID,
						Value: "00000000-0000-0000-0000-000000000000",
					}},
				}, {
					Key: "sort", Value: bson.D{
						{Key: StorageKeyDeviceDeploymentStatus, Value: 1},
						{Key: StorageKeyDeviceDeploymentDeviceId, Value: 1},
					},
				}},
			}, {
				Key: "verbosity", Value: "queryPlanner",
			}})
			type InputStage struct {
				IndexName string `json:"indexName"`
			}
			var explain struct {
				ExplainVersion string `json:"explainVersion"`
				QueryPlanner   struct {
					WinningPlan struct {
						InputStage InputStage `json:"inputStage"`
						QueryPlan  struct {
							// For schema v2 the input stage is here
							InputStage InputStage `json:"inputStage"`
						}
					} `json:"winningPlan"`
				} `json:"queryPlanner"`
			}
			assert.NoError(t, singleRes.Decode(&explain))
			var indexName string
			switch explain.ExplainVersion {
			case "1", "":
				indexName = explain.QueryPlanner.WinningPlan.InputStage.IndexName
			case "2":
				indexName = explain.QueryPlanner.WinningPlan.
					QueryPlan.InputStage.IndexName
			default:
				t.Error("could not determine which index was used in query")
				t.FailNow()
			}
			assert.Equal(t, IndexDeviceDeploymentStatusName, indexName)
		})
	}
}
