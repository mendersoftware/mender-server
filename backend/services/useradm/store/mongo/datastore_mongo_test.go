// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package mongo

import (
	"context"
	"errors"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mongostore "github.com/mendersoftware/mender-server/pkg/mongo"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	mstore "github.com/mendersoftware/mender-server/pkg/store/v2"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"

	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/store"
)

func assertEqualTokens(t *testing.T, expected, actual *jwt.Token) bool {
	var ret bool
	ret = assert.Equal(t, expected.ID, actual.ID)
	ret = ret && assert.Equal(t, expected.Subject, actual.Subject)
	ret = ret && assert.Equal(t, expected.Audience, actual.Audience)
	ret = ret && assert.WithinDuration(t,
		expected.ExpiresAt.Time,
		actual.ExpiresAt.Time, time.Second)
	ret = ret && assert.WithinDuration(t,
		expected.IssuedAt.Time,
		actual.IssuedAt.Time, time.Second)
	ret = ret && assert.WithinDuration(t,
		expected.NotBefore.Time,
		actual.NotBefore.Time, time.Second)
	ret = ret && assert.Equal(t, expected.Issuer, actual.Issuer)
	ret = ret && assert.Equal(t, expected.Scope, actual.Scope)
	ret = ret && assert.Equal(t, expected.Tenant, actual.Tenant)
	return ret && assert.Equal(t, expected.User, actual.User)
}

func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestPing in short mode")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	store, _ := NewDataStoreMongoWithClient(db.Client())
	err := store.Ping(ctx)
	assert.NoError(t, err)
}

func TestMongoCreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	exisitingUsers := []interface{}{
		model.User{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "pretenditsahash",
		},
		model.User{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "pretenditsahash",
		},
	}

	testCases := map[string]struct {
		inUser model.User
		tenant string
		outErr string
	}{
		"ok": {
			inUser: model.User{
				ID:       "1234",
				Email:    "baz@bar.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "",
		},
		"ok with tenant": {
			inUser: model.User{
				ID:       "1234",
				Email:    "baz@bar.com",
				Password: "correcthorsebatterystaple",
			},
			tenant: "foo",
			outErr: "",
		},
		"duplicate email error": {
			inUser: model.User{
				ID:       "1234",
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			outErr: "user with a given email already exists",
		},
		"duplicate email error with tenant": {
			inUser: model.User{
				ID:       "1234",
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			tenant: "foo",
			outErr: "user with a given email already exists",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			store = store.WithAutomigrate()
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}
			err = store.Migrate(ctx, DbVersion)
			assert.NoError(t, err)

			_, err = client.
				Database(DbName).
				Collection(DbUsersColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, exisitingUsers))
			assert.NoError(t, err)

			err = store.CreateUser(ctx, &tc.inUser)

			if tc.outErr == "" {
				//fetch user by id, verify password checks out
				var user model.User
				err := client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbUsersColl).
					FindOne(ctx, mongostore.WithTenantID(ctx, bson.M{"_id": "1234"})).
					Decode(&user)
				assert.NoError(t, err)
				assert.Equal(t, tc.inUser.Password, user.Password)

			} else {
				err = store.CreateUser(ctx, &tc.inUser)
				assert.EqualError(t, err, tc.outErr)
			}
		})
	}
}

func TestMongoUpdateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	exisitingUsers := []interface{}{
		model.User{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "pretenditsahash",
		},
		model.User{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "pretenditsahash",
		},
		model.User{
			ID:       "2bbde4d1-2a4c-47dc-9df4-f048285d2704",
			Email:    "baz+mcetagface@bar.com",
			Password: "pretenditsahash",
			ETag:     &model.ETag{1},
		},
	}

	testCases := map[string]struct {
		inUserUpdate model.UserUpdate
		inUserId     string
		tenant       string
		outErr       string
	}{
		"update email and password: ok": {
			inUserUpdate: model.UserUpdate{
				Email:    "baz@bar.com",
				Password: "correcthorsebatterystaple",
			},
			inUserId: "1",
			outErr:   "",
		},
		"update email: ok": {
			inUserUpdate: model.UserUpdate{
				Email: "baz@bar.com",
			},
			inUserId: "1",
			outErr:   "",
		},
		"update password: ok": {
			inUserUpdate: model.UserUpdate{
				Password: "correcthorsebatterystaple",
			},
			inUserId: "1",
			outErr:   "",
		},
		"ok with tenant": {
			inUserUpdate: model.UserUpdate{
				Email:    "baz@bar.com",
				Password: "correcthorsebatterystaple",
			},
			inUserId: "1",
			tenant:   "foo",
			outErr:   "",
		},
		"ok with tenant and etag": {
			inUserUpdate: model.UserUpdate{
				ETag:       &model.ETag{1},
				Email:      "baz@bar.com",
				Password:   "correcthorsebatterystaple",
				ETagUpdate: &model.ETag{2},
			},
			inUserId: "2bbde4d1-2a4c-47dc-9df4-f048285d2704",
			tenant:   "foo",
			outErr:   "",
		},
		"error, etag mismatch": {
			inUserUpdate: model.UserUpdate{
				ETag:       &model.ETag{3},
				Email:      "baz@bar.com",
				Password:   "correcthorsebatterystaple",
				ETagUpdate: &model.ETag{4},
			},
			inUserId: "2bbde4d1-2a4c-47dc-9df4-f048285d2704",
			tenant:   "foo",
			outErr:   store.ErrUserNotFound.Error(),
		},
		"duplicate email error": {
			inUserUpdate: model.UserUpdate{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			inUserId: "2",
			outErr:   "user with a given email already exists",
		},
		"error, user not found": {
			inUserUpdate: model.UserUpdate{
				Email:    "foo@acme.com",
				Password: "correcthorsebatterystaple",
			},
			inUserId: "0",
			outErr:   store.ErrUserNotFound.Error(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			store = store.WithAutomigrate()
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}
			err = store.Migrate(ctx, DbVersion)
			assert.NoError(t, err)

			_, err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbUsersColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, exisitingUsers))
			assert.NoError(t, err)

			pass := tc.inUserUpdate.Password
			_, err = store.UpdateUser(ctx, tc.inUserId, &tc.inUserUpdate)

			if tc.outErr == "" {
				var user model.User
				err := client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbUsersColl).
					FindOne(ctx, mongostore.WithTenantID(ctx, bson.M{"_id": tc.inUserId})).
					Decode(&user)

				assert.NoError(t, err)
				if tc.inUserUpdate.Password != "" {
					err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
					assert.NoError(t, err)
				}
				if tc.inUserUpdate.Email != "" {
					assert.Equal(t, user.Email, tc.inUserUpdate.Email)
				}
			} else {
				assert.EqualError(t, err, tc.outErr)
			}
		})
	}
}

func TestMongoUpdateLoginTs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}
	client := db.Client()
	store, err := NewDataStoreMongoWithClient(client)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	user := &model.User{
		ID:       oid.NewUUIDv5("userid").String(),
		Password: "123456",
		Email:    "foo@bar.bz",
	}
	ctx := context.Background()
	err = store.CreateUser(ctx, user)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	insertedUser, err := store.GetUserById(ctx, user.ID)
	assert.NoError(t, err)
	assert.Nil(t, insertedUser.LoginTs)

	err = store.UpdateLoginTs(ctx, user.ID)
	assert.NoError(t, err)
	resultUser, err := store.GetUserById(ctx, user.ID)
	assert.NoError(t, err)
	if assert.NotNil(t, resultUser.LoginTs) {
		assert.WithinDuration(
			t, time.Now(), *resultUser.LoginTs, time.Second*10,
		)
	}

}

func TestMongoGetUserByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	existingUsers := []interface{}{
		model.User{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "passwordhash12345",
		},
		model.User{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "passwordhashqwerty",
		},
	}

	testCases := map[string]struct {
		inEmail model.Email
		tenant  string
		outUser *model.User
	}{
		"ok - found 1": {
			inEmail: "foo@bar.com",
			outUser: &model.User{
				ID:       "1",
				Email:    "foo@bar.com",
				Password: "passwordhash12345",
			},
		},
		"ok - case case insensitive": {
			inEmail: "fOO@bAr.cOm",
			outUser: &model.User{
				ID:       "1",
				Email:    "foo@bar.com",
				Password: "passwordhash12345",
			},
		},
		"ok - found 2": {
			inEmail: "bar@bar.com",
			outUser: &model.User{
				ID:       "2",
				Email:    "bar@bar.com",
				Password: "passwordhashqwerty",
			},
		},
		"ok - found 2 with tenant": {
			inEmail: "bar@bar.com",
			outUser: &model.User{
				ID:       "2",
				Email:    "bar@bar.com",
				Password: "passwordhashqwerty",
			},
		},
		"not found": {
			inEmail: "baz@bar.com",
			outUser: nil,
		},
		"not found with tenant": {
			inEmail: "baz@bar.com",
			tenant:  "foo",
			outUser: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			_, err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbUsersColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, existingUsers))
			assert.NoError(t, err)

			user, err := store.GetUserByEmail(ctx, tc.inEmail)

			if tc.outUser != nil {
				assert.Equal(t, *tc.outUser, *user)
			} else {
				assert.Nil(t, user)
				assert.Nil(t, err)
			}
		})
	}
}

func TestMongoGetUserById(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	existingUsers := []interface{}{
		model.User{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "passwordhash12345",
		},
		model.User{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "passwordhashqwerty",
		},
	}

	testCases := map[string]struct {
		inId    string
		tenant  string
		outUser *model.User
	}{
		"ok - found 1": {
			inId: "1",
			outUser: &model.User{
				ID:    "1",
				Email: "foo@bar.com",
			},
		},
		"ok - found 1 with context": {
			inId:   "1",
			tenant: "foo",
			outUser: &model.User{
				ID:    "1",
				Email: "foo@bar.com",
			},
		},
		"ok - found 2": {
			inId: "2",
			outUser: &model.User{
				ID:    "2",
				Email: "bar@bar.com",
			},
		},
		"not found": {
			inId:    "3",
			outUser: nil,
		},
		"not found with tenant": {
			inId:    "3",
			tenant:  "foo",
			outUser: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			_, err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbUsersColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, existingUsers))
			assert.NoError(t, err)

			user, err := store.GetUserById(ctx, tc.inId)

			if tc.outUser != nil {
				assert.Equal(t, *tc.outUser, *user)
			} else {
				assert.Nil(t, user)
				assert.Nil(t, err)
			}
		})
	}
}

func TestMongoGetTokenById(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}
	existing := bson.A{
		&jwt.Token{
			Claims: jwt.Claims{
				ID:        oid.NewUUIDv5("id-1"),
				Subject:   oid.NewUUIDv5("sub-1"),
				Audience:  "audience",
				ExpiresAt: &jwt.Time{Time: time.Now().Add(time.Hour)},
				IssuedAt:  jwt.Time{Time: time.Now()},
				Issuer:    "iss-1",
				NotBefore: jwt.Time{Time: time.Unix(7890, 0)},
				Scope:     "scope-1",
				Tenant:    "tenantID1",
				User:      true,
			},
		},
		&jwt.Token{
			Claims: jwt.Claims{
				ID:        oid.NewUUIDv5("id-2"),
				Subject:   oid.NewUUIDv5("sub-2"),
				Audience:  "audience",
				ExpiresAt: &jwt.Time{Time: time.Now().Add(time.Hour)},
				IssuedAt:  jwt.Time{Time: time.Now()},
				Issuer:    "iss-2",
				NotBefore: jwt.Time{Time: time.Unix(7890, 0)},
				Scope:     "scope-2",
				Tenant:    "tenantID2",
				User:      true,
			},
		},
	}

	testCases := map[string]struct {
		id       string
		tenant   string
		outToken *jwt.Token
	}{
		"ok - found 1": {
			id:       "id-1",
			outToken: existing[0].(*jwt.Token),
		},
		"ok - found 1, MT": {
			id:       "id-1",
			tenant:   "tenantID1",
			outToken: existing[0].(*jwt.Token),
		},
		"ok - found 2": {
			id:       "id-2",
			outToken: existing[1].(*jwt.Token),
		},
		"not found": {
			id:       "id-3",
			outToken: nil,
		},
		"not found, MT": {
			id:       "id-3",
			tenant:   "tenantID1",
			outToken: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			t.Log(existing[0])
			_, err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbTokensColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, existing))
			assert.NoError(t, err)

			token, err := store.GetTokenById(ctx, oid.NewUUIDv5(tc.id))

			if tc.outToken != nil {
				assertEqualTokens(t, tc.outToken, token)
			} else {
				assert.Nil(t, token)
				assert.Nil(t, err)
			}
		})
	}
}

func TestMongoGetUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	ts, err := time.Parse(time.RFC3339, "2017-01-31T16:32:05Z")
	assert.NoError(t, err)

	testCases := map[string]struct {
		ctx      context.Context
		inUsers  []interface{}
		outUsers []model.User
		filter   model.UserFilter
		error    error
	}{
		"ok: list": {
			ctx: func() context.Context {
				ctx := context.Background()
				return identity.WithContext(ctx,
					&identity.Identity{
						Tenant: "foo",
					},
				)
			}(),
			inUsers: []interface{}{
				model.User{
					ID:       "1",
					Email:    "foo@bar.com",
					Password: "passwordhash12345",
				},
				model.User{
					ID:        "2",
					Email:     "bar@bar.com",
					Password:  "passwordhashqwerty",
					CreatedTs: &ts,
				},
				model.User{
					ID:        "3",
					Email:     "baz@bar.com",
					Password:  "passwordhash1sdf2345",
					UpdatedTs: &ts,
				},
			},
			outUsers: []model.User{
				{
					ID:    "1",
					Email: "foo@bar.com",
				},
				{
					ID:        "2",
					Email:     "bar@bar.com",
					CreatedTs: &ts,
				},
				{
					ID:        "3",
					Email:     "baz@bar.com",
					UpdatedTs: &ts,
				},
			},
		},
		"ok: with filter": {
			ctx: context.Background(),
			inUsers: []interface{}{
				model.User{
					ID:       "1",
					Email:    "foo@bar.com",
					Password: "passwordhash12345",
				},
				model.User{
					ID:        "2",
					Email:     "bar@bar.com",
					Password:  "passwordhashqwerty",
					CreatedTs: &ts,
					UpdatedTs: func() *time.Time {
						t := ts.Add(time.Minute)
						return &t
					}(),
				},
				model.User{
					ID:        "3",
					Email:     "baz@bar.com",
					Password:  "passwordhash1sdf2345",
					UpdatedTs: &ts,
				},
			},
			filter: model.UserFilter{
				ID: []string{"1", "2"},
				Email: []model.Email{
					"foo@bar.com",
					"bar@bar.com",
					"baz@bar.com",
					"user@acme.io",
				},
				CreatedAfter: func() *time.Time {
					t := ts.Add(-time.Hour)
					return &t
				}(),
				CreatedBefore: func() *time.Time {
					t := ts.Add(time.Hour)
					return &t
				}(),
				UpdatedAfter: func() *time.Time {
					t := ts.Add(-time.Hour)
					return &t
				}(),
				UpdatedBefore: func() *time.Time {
					t := ts.Add(time.Hour)
					return &t
				}(),
			},

			outUsers: []model.User{
				{
					ID:        "2",
					Email:     "bar@bar.com",
					CreatedTs: &ts,
					UpdatedTs: func() *time.Time {
						t := ts.Add(time.Minute)
						return &t
					}(),
				},
			},
		},
		"ok: empty": {
			ctx:      context.Background(),
			inUsers:  []interface{}{},
			outUsers: []model.User{},
		},
		"error: context canceled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(
					context.Background(),
				)
				cancel()
				return ctx
			}(),
			inUsers: []interface{}{
				model.User{
					ID:       "1",
					Email:    "foo@bar.com",
					Password: "passwordhash12345",
				},
			},
			error: errors.New("store: failed to fetch users"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			db.Wipe()

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if len(tc.inUsers) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(tc.ctx, DbName)).
					Collection(DbUsersColl).
					InsertMany(context.Background(), mongostore.ArrayWithTenantID(tc.ctx, tc.inUsers))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			}

			users, err := store.GetUsers(tc.ctx, tc.filter)
			if tc.error != nil {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tc.error.Error())
				}
			} else {
				assert.NoError(t, err)

				// transform times to utc
				// bson encoder uses time.Local before writing
				// to mongo, which can be e.g. 'CET'
				// this won't match with assert.Equal
				for i := range users {
					if users[i].CreatedTs != nil {
						t := users[i].CreatedTs.UTC()
						users[i].CreatedTs = &t
					}
					if users[i].UpdatedTs != nil {
						t := users[i].UpdatedTs.UTC()
						users[i].UpdatedTs = &t
					}
				}

				assert.Equal(t, tc.outUsers, users)
			}
		})
	}
}

func TestMongoDeleteUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	existingUsers := []interface{}{
		model.User{
			ID:       "1",
			Email:    "foo@bar.com",
			Password: "passwordhash12345",
		},
		model.User{
			ID:       "2",
			Email:    "bar@bar.com",
			Password: "passwordhashqwerty",
		},
	}

	testCases := map[string]struct {
		inId     string
		tenant   string
		outUsers []model.User
	}{
		"ok": {
			inId: "1",
			outUsers: []model.User{
				{
					ID:       "2",
					Email:    "bar@bar.com",
					Password: "passwordhashqwerty",
				},
			},
		},
		"ok - with tenant": {
			inId:   "1",
			tenant: "foo",
			outUsers: []model.User{
				{
					ID:       "2",
					Email:    "bar@bar.com",
					Password: "passwordhashqwerty",
				},
			},
		},
		"ok - not found": {
			inId: "3",
			outUsers: []model.User{
				{
					ID:       "1",
					Email:    "foo@bar.com",
					Password: "passwordhash12345",
				},
				{
					ID:       "2",
					Email:    "bar@bar.com",
					Password: "passwordhashqwerty",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			_, err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbUsersColl).
				InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, existingUsers))
			assert.NoError(t, err)

			err = store.DeleteUser(ctx, tc.inId)
			assert.NoError(t, err)

			var users []model.User
			c, err := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbUsersColl).
				Find(ctx, mongostore.WithTenantID(ctx, bson.M{}))

			assert.NoError(t, err)

			err = c.All(ctx, &users)
			assert.NoError(t, err)

			assert.Equal(t, tc.outUsers, users)
		})
	}
}

func TestMongoSaveToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}
	testCases := map[string]struct {
		token  *jwt.Token
		tenant string
	}{
		"ok 1": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:       oid.NewUUIDv5("id-3"),
					Subject:  oid.NewUUIDv5("sub-3"),
					Audience: "audience",
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					IssuedAt: jwt.Time{Time: time.Now()},
					Issuer:   "iss-3",
					NotBefore: jwt.Time{
						Time: time.Unix(7890, 0),
					},
					Scope:  "scope-3",
					Tenant: "tenantID3",
					User:   true,
				},
			},
		},
		"ok 2": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("id-4"),
					Subject: oid.NewUUIDv5("sub-4"),
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					IssuedAt: jwt.Time{Time: time.Now()},
					Tenant:   "tenantID4",
					User:     true,
				},
			},
		},
		"ok 3, MT": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:       oid.NewUUIDv5("id-3"),
					Subject:  oid.NewUUIDv5("sub-3"),
					Audience: "audience",
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					IssuedAt: jwt.Time{Time: time.Now()},
					Issuer:   "iss-3",
					NotBefore: jwt.Time{
						Time: time.Unix(7890, 0),
					},
					Scope:  "scope-3",
					Tenant: "tenantID3",
					User:   true,
				},
			},
			tenant: "tenantID1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			//setup
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			//test
			err = store.SaveToken(ctx, tc.token)
			assert.NoError(t, err)

			var token jwt.Token
			err = client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbTokensColl).
				FindOne(ctx, mongostore.WithTenantID(ctx, bson.M{"_id": tc.token.Claims.ID})).
				Decode(&token)

			assert.NoError(t, err)

			assertEqualTokens(t, tc.token, &token)
		})
	}
}

func TestMongoEnsureSessionTokensLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	userID := oid.NewUUIDv5("sub-3")

	now := time.Now()
	testCases := map[string]struct {
		tokens []*jwt.Token
		limit  int
		count  int64
		tenant string
	}{
		"ok, single token": {
			tokens: []*jwt.Token{
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  userID,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now.Add(time.Hour - 1),
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID3",
						User:   true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("sub-4"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now.Add(time.Hour - 1),
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID3",
						User:   true,
					},
				},
			},
			count: 1,
		},
		"ok, two tokens": {
			tokens: []*jwt.Token{
				{
					Claims: jwt.Claims{
						ID:      oid.NewUUIDv5("id-1"),
						Subject: userID,
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now.Add(time.Hour - 1),
						},
						Tenant: "tenantID4",
						User:   true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:      oid.NewUUIDv5("id-2"),
						Subject: userID,
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{Time: now},
						Tenant:   "tenantID4",
						User:     true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("sub-4"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now.Add(time.Hour - 1),
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID3",
						User:   true,
					},
				},
			},
			count: 1,
		},
		"ok, two tokens with MT": {
			tokens: []*jwt.Token{
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  userID,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now.Add(time.Hour - 1),
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID1",
						User:   true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  userID,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now,
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID1",
						User:   true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  userID,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: now.Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: now,
						},
						Issuer: "iss-3",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-3",
						Tenant: "tenantID2",
						User:   true,
					},
				},
			},
			tenant: "tenantID1",
			count:  1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			//setup
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			for _, token := range tc.tokens {
				err = store.SaveToken(ctx, token)
				assert.NoError(t, err)
			}

			// test
			err = store.EnsureSessionTokensLimit(ctx, userID, 1)
			assert.NoError(t, err)

			filter := mongostore.WithTenantID(ctx, bson.M{
				DbTokenSubject: userID,
			})
			c, err := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbTokensColl).
				CountDocuments(ctx, filter)
			assert.NoError(t, err)
			assert.Equal(t, tc.count, c)
		})
	}
}

func TestMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigrate in short mode.")
	}

	testCases := map[string]struct {
		automigrate bool

		version string
		err     string
	}{
		DbVersion: {
			automigrate: true,
			version:     DbVersion,
			err:         "",
		},
		DbVersion + ", no automigrate": {
			automigrate: false,
			version:     DbVersion,
			err:         "",
		},
		"0.1 error": {
			automigrate: true,
			version:     "0.1",
			err:         "failed to parse service version: failed to parse Version: unexpected EOF",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			store, err := NewDataStoreMongoWithClient(db.Client())
			assert.NoError(t, err)

			// set up automigration
			if tc.automigrate {
				store = store.WithAutomigrate()
			}

			ctx := context.Background()
			err = store.Migrate(ctx, tc.version)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				var out []migrate.MigrationEntry
				findOpts := mopts.Find().SetSort(bson.D{
					{Key: "version.major", Value: -1},
					{Key: "version.minor", Value: -1},
					{Key: "version.patch", Value: -1},
				})
				c, err := store.client.
					Database(DbName).
					Collection(migrate.DbMigrationsColl).
					Find(ctx, bson.M{}, findOpts)
				assert.NoError(t, err)

				err = c.All(ctx, &out)
				assert.NoError(t, err)

				if tc.automigrate {
					assert.Len(t, out, 6)
					assert.NoError(t, err)

					v, _ := migrate.NewVersion(tc.version)
					assert.Equal(t, *v, out[0].Version)
				} else {
					assert.Len(t, out, 0)
				}
			}
		})
	}
}

func TestWithAutomigrate(t *testing.T) {
	db.Wipe()

	client := db.Client()

	store, err := NewDataStoreMongoWithClient(client)
	assert.NoError(t, err)

	new_store := store.WithAutomigrate()

	assert.NotEqual(t, unsafe.Pointer(store), unsafe.Pointer(new_store))
}

func TestWithMultitenant(t *testing.T) {
	db.Wipe()

	client := db.Client()

	store, err := NewDataStoreMongoWithClient(client)
	assert.NoError(t, err)

	new_store := store.WithMultitenant()

	assert.NotEqual(t, unsafe.Pointer(store), unsafe.Pointer(new_store))
}

func TestMongoDeleteToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	tokenID := oid.NewUUIDv5("id-1")
	tokenSubject := oid.NewUUIDv5("sub-1")

	testCases := map[string]struct {
		tenant   string
		token    *jwt.Token
		inTokens []interface{}

		outError string
	}{
		"ok": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      tokenID,
					Subject: tokenSubject,
				},
			},
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       tokenID,
						Subject:  tokenSubject,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("sub-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-2",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-2",
						User:  true,
					},
				},
			},
		},
		"ok - tenant": {
			tenant: "tenant-1",
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      tokenID,
					Subject: tokenSubject,
				},
			},
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       tokenID,
						Subject:  tokenSubject,
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-1",
						Tenant: "tenantID1",
						User:   true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-1",
						Tenant: "tenantID1",
						User:   true,
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, tc.inTokens))
				assert.NoError(t, err)
			}

			err = store.DeleteToken(ctx, tc.token.Subject, tc.token.ID)
			assert.NoError(t, err)

			var tokens []jwt.Token
			c, err := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbTokensColl).
				Find(ctx, mongostore.WithTenantID(ctx, bson.M{"_id": tc.token.ID}))
			assert.NoError(t, err)

			err = c.All(ctx, &tokens)
			assert.NoError(t, err)
			assert.Nil(t, tokens)
		})
	}
}

func TestMongoDeleteTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	testCases := map[string]struct {
		tenant   string
		inTokens []interface{}

		outError string
	}{
		"ok": {
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("sub-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-2",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-2",
						User:  true,
					},
				},
			},
		},
		"ok - tenant": {
			tenant: "tenant-1",
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-1",
						Tenant: "tenantID1",
						User:   true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("sub-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope:  "scope-1",
						Tenant: "tenantID1",
						User:   true,
					},
				},
			},
		},
		"tenant, no tokens": {
			tenant:   "tenantID2",
			outError: store.ErrTokenNotFound.Error(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, tc.inTokens))
				assert.NoError(t, err)
			}

			err = store.DeleteTokens(ctx)
			if tc.outError != "" {
				assert.EqualError(t, err, tc.outError)
			} else {
				assert.NoError(t, err)
			}

			var tokens []jwt.Token
			c, err := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbTokensColl).
				Find(ctx, mongostore.WithTenantID(ctx, bson.M{}))
			assert.NoError(t, err)

			err = c.All(ctx, &tokens)

			assert.NoError(t, err)
			assert.Nil(t, tokens)
		})
	}
}

func TestMongoDeleteTokensByUserId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	testCases := map[string]struct {
		tenant   string
		user     string
		inTokens []interface{}

		outTokens []jwt.Token
		outError  string
	}{
		"ok": {
			user: oid.NewUUIDv5("user-1").String(),
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
			outTokens: []jwt.Token{
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
		},
		"ok - tenant": {
			user:   oid.NewUUIDv5("user-1").String(),
			tenant: "tenant-1",
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
			outTokens: []jwt.Token{
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
		},
		"ok - no tokens": {
			user: oid.NewUUIDv5("user2").String(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, tc.inTokens))
				assert.NoError(t, err)
			}

			err = store.DeleteTokensByUserId(ctx, tc.user)
			if tc.outError != "" {
				assert.EqualError(t, err, tc.outError)
			} else {
				assert.NoError(t, err)
			}

			var tokens []jwt.Token
			if tc.outTokens != nil {
				c, err := client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					Find(ctx, mongostore.WithTenantID(ctx, bson.M{}))
				assert.NoError(t, err)

				err = c.All(ctx, &tokens)
				assert.NoError(t, err)
			}

			if assert.Len(t, tokens, len(tc.outTokens)) {
				for i, token := range tokens {
					assertEqualTokens(t, &tc.outTokens[i], &token)
				}
			}
		})
	}
}

func TestMongoDeleteTokensByUserIdExceptCurrentOne(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	testCases := map[string]struct {
		tenant   string
		user     string
		inTokens []interface{}

		outTokens []jwt.Token
		outError  string
	}{
		"ok": {
			user: oid.NewUUIDv5("user-1").String(),
			inTokens: []interface{}{
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-2"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				jwt.Token{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
			outTokens: []jwt.Token{
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-1"),
						Subject:  oid.NewUUIDv5("user-1"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
				{
					Claims: jwt.Claims{
						ID:       oid.NewUUIDv5("id-3"),
						Subject:  oid.NewUUIDv5("user-2"),
						Audience: "audience",
						ExpiresAt: &jwt.Time{
							Time: time.Now().
								Add(time.Hour),
						},
						IssuedAt: jwt.Time{
							Time: time.Now(),
						},
						Issuer: "iss-1",
						NotBefore: jwt.Time{
							Time: time.Unix(7890, 0),
						},
						Scope: "scope-1",
						User:  true,
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, tc.inTokens))
				assert.NoError(t, err)
			}

			err = store.DeleteTokensByUserIdExceptCurrentOne(ctx, tc.user, tc.inTokens[0].(jwt.Token).ID)
			if tc.outError != "" {
				assert.EqualError(t, err, tc.outError)
			} else {
				assert.NoError(t, err)
			}

			var tokens []jwt.Token
			if tc.outTokens != nil {
				c, err := client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					Find(ctx, mongostore.WithTenantID(ctx, bson.M{}))
				assert.NoError(t, err)

				err = c.All(ctx, &tokens)
				assert.NoError(t, err)
			}

			if assert.Len(t, tokens, len(tc.outTokens)) {
				for i, token := range tokens {
					assertEqualTokens(t, &tc.outTokens[i], &token)
				}
			}
		})
	}
}

func TestMongoSaveSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	// we'll preset settings _id for easy 1:1 comparison on test (normally autogenerated)
	testCases := map[string]struct {
		etag             string
		settingsIn       *model.Settings
		settingsExisting *model.Settings
		settingsOut      *model.Settings
		tenant           string
		err              error
	}{
		"ok: insert": {
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok: insert, etag matches": {
			etag: "etag",
			settingsExisting: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ko: insert, etag mismatches": {
			etag: "mismatch",
			settingsExisting: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			err: store.ErrETagMismatch,
		},
		"ok: insert, tenant": {
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
			tenant: "acme",
		},
		"ok: overwrite with exact same fields": {
			settingsExisting: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val-old",
					"bar": 0,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok: overwrite with different fields": {
			settingsExisting: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val-old",
					"bar": 0,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"baz1": "baz",
					"baz2": 420,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"baz1": "baz",
					"baz2": int32(420),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			store = store.WithAutomigrate()
			err = store.Migrate(ctx, DbVersion)
			assert.NoError(t, err)

			if tc.settingsExisting != nil {
				doc := mongostore.WithTenantID(ctx, tc.settingsExisting)
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbSettingsColl).
					InsertOne(ctx, doc)
				assert.NoError(t, err)
			}

			err = store.SaveSettings(ctx, tc.settingsIn, tc.etag)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err)

				var settings *model.Settings
				err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbSettingsColl).
					FindOne(ctx, mongostore.WithTenantID(ctx, bson.M{})).
					Decode(&settings)

				// ignore the randomly generated ETag
				tc.settingsOut.ETag = settings.ETag

				assert.Equal(t, tc.settingsOut, settings)
			}
		})
	}
}

func TestMongoGetSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	testCases := map[string]struct {
		settingsExisting *model.Settings
		settingsOut      *model.Settings
		tenant           string
		err              string
	}{
		"ok": {
			settingsExisting: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok, tenant": {
			settingsExisting: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
			tenant: "tenant-foo",
		},
		"ok, empty": {
			settingsOut: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if tc.settingsExisting != nil {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbSettingsColl).
					InsertOne(ctx, mongostore.WithTenantID(ctx, tc.settingsExisting))
				assert.NoError(t, err)
			}

			out, err := store.GetSettings(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.settingsOut, out)
		})
	}
}

func TestMongoSaveUserSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	userID := uuid.NewString()

	// we'll preset settings _id for easy 1:1 comparison on test (normally autogenerated)
	testCases := map[string]struct {
		etag             string
		settingsIn       *model.Settings
		settingsExisting *model.Settings
		settingsOut      *model.Settings
		tenant           string
		err              error
	}{
		"ok: insert": {
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok: insert, etag matches": {
			etag: "etag",
			settingsExisting: &model.Settings{
				ID:     "1",
				ETag:   "etag",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ko: insert, etag mismatches": {
			etag: "mismatch",
			settingsExisting: &model.Settings{
				ID:     "1",
				ETag:   "etag",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			err: store.ErrETagMismatch,
		},
		"ok: insert, tenant": {
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
			tenant: "acme",
		},
		"ok: overwrite with exact same fields": {
			settingsExisting: &model.Settings{
				ID:     "1",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val-old",
					"bar": 0,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok: overwrite with different fields": {
			settingsExisting: &model.Settings{
				ID:     "1",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val-old",
					"bar": 0,
				},
			},
			settingsIn: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"baz1": "baz",
					"baz2": 420,
				},
			},
			settingsOut: &model.Settings{
				ID: "1",
				Values: model.SettingsValues{
					"baz1": "baz",
					"baz2": int32(420),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			store = store.WithAutomigrate()
			err = store.Migrate(ctx, DbVersion)
			assert.NoError(t, err)

			if tc.settingsExisting != nil {
				doc := mongostore.WithTenantID(ctx, tc.settingsExisting)
				doc = append(doc, primitive.E{
					Key:   DbSettingsUserID,
					Value: userID,
				})
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbUserSettingsColl).
					InsertOne(ctx, doc)
				assert.NoError(t, err)
			}

			err = store.SaveUserSettings(ctx, userID, tc.settingsIn, tc.etag)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err)

				var settings *model.Settings
				err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbUserSettingsColl).
					FindOne(ctx, mongostore.WithTenantID(ctx, bson.M{DbSettingsUserID: userID})).
					Decode(&settings)

				// ignore the randomly generated ETag
				tc.settingsOut.ETag = settings.ETag

				assert.Equal(t, tc.settingsOut, settings)
			}
		})
	}
}

func TestMongoGetUserSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	userID := uuid.NewString()

	testCases := map[string]struct {
		settingsExisting *model.Settings
		settingsOut      *model.Settings
		tenant           string
		err              string
	}{
		"ok": {
			settingsExisting: &model.Settings{
				ID:     "1",
				ETag:   "etag",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
		},
		"ok, tenant": {
			settingsExisting: &model.Settings{
				ID:     "1",
				ETag:   "etag",
				UserID: userID,
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": 42,
				},
			},
			settingsOut: &model.Settings{
				ID:   "1",
				ETag: "etag",
				Values: model.SettingsValues{
					"foo": "foo-val",
					"bar": int32(42),
				},
			},
			tenant: "tenant-foo",
		},
		"ok, empty": {
			settingsOut: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			if tc.settingsExisting != nil {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbUserSettingsColl).
					InsertOne(ctx, mongostore.WithTenantID(ctx, tc.settingsExisting))
				assert.NoError(t, err)
			}

			out, err := store.GetUserSettings(ctx, userID)

			assert.NoError(t, err)
			assert.Equal(t, tc.settingsOut, out)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestMongoGetPersonalAccessTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	tokens := []jwt.Token{
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-1"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-2"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
			TokenName: strPtr("my_personal_access_token"),
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-3"),
				Subject:  oid.NewUUIDv5("sub-2"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-2",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-2",
				User:  true,
			},
		},
	}

	testCases := map[string]struct {
		tenant   string
		inTokens []jwt.Token

		outTokens []model.PersonalAccessToken
		outError  string
	}{
		"ok": {
			inTokens: tokens,
			outTokens: []model.PersonalAccessToken{
				{
					ID:   tokens[1].ID,
					Name: tokens[1].TokenName,
				},
			},
		},
		"ok, tenant": {
			tenant:   "tenant-1",
			inTokens: tokens,
			outTokens: []model.PersonalAccessToken{
				{
					ID:   tokens[1].ID,
					Name: tokens[1].TokenName,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			inputData := make([]interface{}, len(tc.inTokens))
			for i, v := range tc.inTokens {
				inputData[i] = v
			}
			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, inputData))
				assert.NoError(t, err)
			}

			dbTokens, err := store.GetPersonalAccessTokens(ctx, tokens[1].Subject.String())
			assert.NoError(t, err)
			//clear dates
			for i, _ := range dbTokens {
				dbTokens[i].ExpirationDate = nil
				dbTokens[i].CreatedTs = jwt.Time{}
			}

			assert.Equal(t, tc.outTokens, dbTokens)
		})
	}
}

func TestMongoCountPersonalAccessTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	tokens := []jwt.Token{
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-1"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-2"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
			TokenName: strPtr("my_personal_access_token"),
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-3"),
				Subject:  oid.NewUUIDv5("sub-2"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-2",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-2",
				User:  true,
			},
		},
	}

	testCases := map[string]struct {
		tenant   string
		inTokens []jwt.Token

		expectedCount int64
		outError      string
	}{
		"ok": {
			inTokens:      tokens,
			expectedCount: 1,
		},
		"ok, tenant": {
			tenant:        "tenant-1",
			inTokens:      tokens,
			expectedCount: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			inputData := make([]interface{}, len(tc.inTokens))
			for i, v := range tc.inTokens {
				inputData[i] = v
			}
			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, inputData))
				assert.NoError(t, err)
			}

			count, err := store.CountPersonalAccessTokens(ctx, tokens[1].Subject.String())
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, count)
		})
	}
}

func TestMongoUpdateTokenLastUsed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode.")
	}

	tokens := []jwt.Token{
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-1"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-2"),
				Subject:  oid.NewUUIDv5("sub-1"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-1",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-1",
				User:  true,
			},
			TokenName: strPtr("my_personal_access_token"),
		},
		{
			Claims: jwt.Claims{
				ID:       oid.NewUUIDv5("id-3"),
				Subject:  oid.NewUUIDv5("sub-2"),
				Audience: "audience",
				ExpiresAt: &jwt.Time{
					Time: time.Now().
						Add(time.Hour),
				},
				IssuedAt: jwt.Time{
					Time: time.Now(),
				},
				Issuer: "iss-2",
				NotBefore: jwt.Time{
					Time: time.Unix(7890, 0),
				},
				Scope: "scope-2",
				User:  true,
			},
		},
	}

	testCases := map[string]struct {
		tenant   string
		inTokens []jwt.Token

		outTokens []model.PersonalAccessToken
		outError  string
	}{
		"ok": {
			inTokens: tokens,
			outTokens: []model.PersonalAccessToken{
				{
					ID:   tokens[1].ID,
					Name: tokens[1].TokenName,
				},
			},
		},
		"ok, tenant": {
			tenant:   "tenant-1",
			inTokens: tokens,
			outTokens: []model.PersonalAccessToken{
				{
					ID:   tokens[1].ID,
					Name: tokens[1].TokenName,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			client := db.Client()
			store, err := NewDataStoreMongoWithClient(client)
			assert.NoError(t, err)

			inputData := make([]interface{}, len(tc.inTokens))
			for i, v := range tc.inTokens {
				inputData[i] = v
			}
			if len(tc.inTokens) > 0 {
				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbTokensColl).
					InsertMany(ctx, mongostore.ArrayWithTenantID(ctx, inputData))
				assert.NoError(t, err)
			}

			//get the token
			dbTokens, err := store.GetPersonalAccessTokens(ctx, tokens[1].Subject.String())
			assert.NoError(t, err)
			assert.Equal(t, len(dbTokens), 1)
			assert.Nil(t, dbTokens[0].LastUsed)
			// update last used timestamp
			err = store.UpdateTokenLastUsed(ctx, dbTokens[0].ID)
			assert.NoError(t, err)
			//get the token once again
			dbTokens, err = store.GetPersonalAccessTokens(ctx, tokens[1].Subject.String())
			assert.NoError(t, err)
			assert.NotNil(t, dbTokens[0].LastUsed)
		})
	}
}
