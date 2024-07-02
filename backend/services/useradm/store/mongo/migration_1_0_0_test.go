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

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration_1_0_0(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestMigration_1_0_0 in short mode")
	}
	testCases := []struct {
		Name string

		Tenant string

		TokensIn  bson.A
		TokensOut []jwt.Token
	}{
		{
			Name: "Successful migration",

			TokensIn: bson.A{
				jwt.Token{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("foo"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
				jwt.Token{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("baz"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
			},
			TokensOut: []jwt.Token{
				{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("foo"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
				{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("baz"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
			},
		},
		{
			Name:   "Successful migration, MT expire token",
			Tenant: primitive.NewObjectID().Hex(),

			TokensIn: bson.A{
				jwt.Token{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("foo"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
				jwt.Token{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("baz"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(-time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
			},
			TokensOut: []jwt.Token{
				{Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("foo"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().
							Add(time.Hour).
							Round(time.Second),
					},
					Issuer: "Mender",
					Scope:  "mender.*",
				}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db.Wipe()
			ctx := context.Background()
			if testCase.Tenant != "" {
				ctx = identity.WithContext(
					ctx,
					&identity.Identity{
						Tenant: testCase.Tenant,
					})
			}
			client := db.Client()
			ds, err := NewDataStoreMongoWithClient(client)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			database := client.Database(ctxstore.
				DbFromContext(ctx, DbName))
			collTokens := database.Collection(DbTokensColl)

			collTokens.InsertMany(ctx, testCase.TokensIn)

			migration := &migration_1_0_0{
				ds:  ds,
				ctx: ctx,
			}
			err = migration.Up(migration.Version())
			assert.NoError(t, err)

			N, err := collTokens.CountDocuments(ctx, bson.M{})
			assert.NoError(t, err)
			assert.Equal(t, int64(0), N)

			res, err := collTokens.InsertMany(ctx, testCase.TokensIn)
			assert.NoError(t, err)
			assert.Equal(t,
				len(testCase.TokensIn),
				len(res.InsertedIDs),
			)

			idx := collTokens.Indexes()
			cur, _ := idx.List(ctx)
			defer cur.Close(ctx)
			var idxModel bson.M
			for cur.Next(ctx) {
				cur.Decode(&idxModel)
				if idxModel["name"] == "TokenExpiration" {
					break
				}
			}
			if expire, ok := idxModel["expireAfterSeconds"]; assert.
				True(t, ok) {
				assert.Equal(t, expire, int32(0))
			}

			cur, err = collTokens.Find(ctx, bson.M{})
			assert.NoError(t, err)
			var results []jwt.Token
			if len(testCase.TokensIn) != len(testCase.TokensOut) {
				// Poll server for up to 5 minutes until it
				// has deleted the expired token
				for i := 0; i < 60; i++ {
					err = cur.All(ctx, &results)
					if err != nil {
						break
					} else if len(results) ==
						len(testCase.TokensOut) {
						break
					}
					time.Sleep(time.Second * 5)
					cur, err = collTokens.Find(ctx, bson.M{})
				}
				cur.Close(ctx)
			} else {
				cur.All(ctx, &results)
			}
			if assert.Len(t, results, len(testCase.TokensOut)) {
				for _, tokenOut := range testCase.TokensOut {
					var tokenPtr *jwt.Token
					for _, tokenRes := range results {
						if tokenOut.ID == tokenRes.ID {
							tokenPtr = &tokenOut
							break
						}
					}
					if assert.NotNil(t,
						tokenPtr,
						"Expected token not found: %v",
						tokenOut) {
						assertEqualTokens(
							t,
							&tokenOut,
							tokenPtr,
						)
					}
				}
			}
		})
	}
}
