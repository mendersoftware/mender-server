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
package useradm

import (
	"context"
	"github.com/mendersoftware/mender-server/services/useradm/common"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mendersoftware/mender-server/pkg/apiclient"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/mendersoftware/mender-server/services/useradm/client/tenant"
	ct "github.com/mendersoftware/mender-server/services/useradm/client/tenant"
	mct "github.com/mendersoftware/mender-server/services/useradm/client/tenant/mocks"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	mjwt "github.com/mendersoftware/mender-server/services/useradm/jwt/mocks"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/scope"
	"github.com/mendersoftware/mender-server/services/useradm/store"
	mstore "github.com/mendersoftware/mender-server/services/useradm/store/mocks"
)

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		MultiTenant    bool
		DataStoreError error
		TenantAdmError error
	}{{
		Name: "ok",
	}, {
		Name:        "ok, multitenant",
		MultiTenant: true,
	}, {
		Name:           "error, datastore unhealthy",
		DataStoreError: errors.New("connection refused"),
	}, {
		Name:           "error, tenantadm unhealthy",
		MultiTenant:    true,
		TenantAdmError: errors.New("connection refused"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), time.Second*5,
			)
			defer cancel()
			db := &mstore.DataStore{}
			db.On("Ping", ctx).Return(tc.DataStoreError)
			useradm := NewUserAdm(nil, db, Config{})
			if tc.MultiTenant {
				cTenant := &mct.ClientRunner{}
				cTenant.On("CheckHealth", ctx).
					Return(tc.TenantAdmError)
				useradm = useradm.WithTenantVerification(cTenant)
			}
			err := useradm.HealthCheck(ctx)
			switch {
			case tc.DataStoreError != nil:
				assert.EqualError(t, err,
					"error reaching MongoDB: "+
						tc.DataStoreError.Error(),
				)
			case tc.TenantAdmError != nil && tc.MultiTenant:
				assert.EqualError(t, err,
					"Tenantadm service unhealthy: "+
						tc.TenantAdmError.Error(),
				)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAdmSignToken(t *testing.T) {
	//cases: handler err, no handler err
	testCases := map[string]struct {
		signed  string
		signErr error

		config Config
	}{
		"ok": {
			signed:  "foo",
			signErr: nil,
		},
		"token sign error": {
			signed:  "",
			signErr: errors.New("token generation error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			mockJWTHandler := mjwt.Handler{}
			mockJWTHandler.On("ToJWT",
				mock.AnythingOfType("*jwt.Token"),
			).Return(tc.signed, tc.signErr)

			useradm := NewUserAdm(map[int]jwt.Handler{0: &mockJWTHandler}, nil, tc.config)
			signed, err := useradm.SignToken(ctx, &jwt.Token{})

			if tc.signErr != nil {
				assert.EqualError(t, err, tc.signErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.signed, signed)
			}
		})
	}

}

func TestUserAdmLogin(t *testing.T) {
	const sessionTokensLimit = 10

	testCases := map[string]struct {
		inEmail    model.Email
		inPassword string
		noExpiry   bool

		verifyTenant bool
		tenant       *ct.Tenant
		tenantErr    error

		dbUser    *model.User
		dbUserErr error

		dbTokenErr        error
		dbEnsureLimitsErr error
		dbUpdateErr       error

		outErr   error
		outToken *jwt.Token

		config Config
	}{
		"ok": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,

			outErr: nil,
			outToken: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("1234"),
					Scope:   scope.All,
				},
			},

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"ok, multitenant": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			verifyTenant: true,
			tenant: &ct.Tenant{
				ID:   "TenantID1",
				Name: "tenant1",
			},
			tenantErr: nil,

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,
			// error for updating login_ts is suppressed
			dbUpdateErr: errors.New("internal error"),

			outErr: nil,
			outToken: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("1234"),
					Scope:   scope.All,
					Tenant:  "TenantID1",
				},
			},

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error, multitenant: tenant not found": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			verifyTenant: true,
			tenant:       nil,
			tenantErr:    nil,

			outErr:   ErrUnauthorized,
			outToken: nil,
		},
		"error, multitenant: tenant verification error": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			verifyTenant: true,
			tenant:       nil,
			tenantErr:    errors.New("some error"),

			outErr:   errors.New("failed to check user's tenant: some error"),
			outToken: nil,
		},
		"error, multitenant: tenant account suspended": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			verifyTenant: true,
			tenant: &ct.Tenant{
				ID:     "TenantID1",
				Name:   "tenant1",
				Status: "suspended",
			},
			tenantErr: nil,

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,

			outErr: ErrTenantAccountSuspended,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error: no user": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser:    nil,
			dbUserErr: nil,

			outErr:   ErrUnauthorized,
			outToken: nil,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error: wrong password": {
			inEmail:    "foo@bar.com",
			inPassword: "notcorrecthorsebatterystaple",

			dbUser: &model.User{
				ID:       "1234",
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,

			outErr:   ErrUnauthorized,
			outToken: nil,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error: regular login, db.GetUserByEmail() error": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser:    nil,
			dbUserErr: errors.New("db: internal error"),

			outErr:   errors.New("useradm: failed to get user: db: internal error"),
			outToken: nil,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error: db.SaveToken() error": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,

			dbTokenErr: errors.New("db failed"),

			outErr:   errors.New("useradm: failed to save token: db failed"),
			outToken: nil,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
		"error: db.EnsureSessionTokensLimit() error": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			dbUserErr: nil,

			dbEnsureLimitsErr: errors.New("db failed"),

			outErr:   errors.New("useradm: failed to ensure session tokens limit: db failed"),
			outToken: nil,

			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitSessionsPerUser:  sessionTokensLimit,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("GetUserByEmail", ContextMatcher(), tc.inEmail).Return(tc.dbUser, tc.dbUserErr)

			db.On("SaveToken", ContextMatcher(), mock.AnythingOfType("*jwt.Token")).Return(tc.dbTokenErr)
			if tc.dbTokenErr == nil {
				db.On("EnsureSessionTokensLimit", ContextMatcher(), mock.AnythingOfType("oid.ObjectID"),
					sessionTokensLimit).Return(tc.dbEnsureLimitsErr)
			}
			if tc.dbUser != nil {
				db.On("UpdateLoginTs", ContextMatcher(), tc.dbUser.ID).
					Return(tc.dbUpdateErr)
			}

			useradm := NewUserAdm(nil, db, tc.config)
			if tc.verifyTenant {
				cTenant := &mct.ClientRunner{}
				cTenant.On("GetTenant", ContextMatcher(), string(tc.inEmail), &apiclient.HttpApi{}).
					Return(tc.tenant, tc.tenantErr)
				useradm = useradm.WithTenantVerification(cTenant)
			}

			token, err := useradm.Login(ctx, tc.inEmail, tc.inPassword, &LoginOptions{
				NoExpiry: tc.noExpiry,
			})

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
				assert.Nil(t, token)
			} else {
				if tc.outToken != nil && assert.NotNil(t, token) {
					assert.NoError(t, err)
					assert.NotEmpty(t, token.ID)
					assert.NotEmpty(t, token.Claims.ID)
					assert.Equal(t, tc.config.Issuer, token.Claims.Issuer)
					assert.Equal(t, tc.outToken.Claims.Scope, token.Claims.Scope)
					assert.WithinDuration(t,
						time.Now().Add(time.Duration(tc.config.ExpirationTimeSeconds)*time.Second),
						token.Claims.ExpiresAt.Time,
						time.Second)

				}
			}
		})
	}

}

func TestUserAdmLogout(t *testing.T) {
	testCases := map[string]struct {
		token            *jwt.Token
		deleteTokenError error

		err error
	}{
		"ok": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("1234"),
					Scope:   scope.All,
				},
			},
		},
		"ko": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("1234"),
					Scope:   scope.All,
				},
			},
			deleteTokenError: errors.New("error"),
			err:              errors.New("error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On(
				"DeleteToken",
				ContextMatcher(),
				tc.token.Subject,
				tc.token.ID,
			).Return(tc.deleteTokenError)
			defer db.AssertExpectations(t)

			useradm := NewUserAdm(nil, db, Config{})
			err := useradm.Logout(ctx, tc.token)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUserAdmDoCreateUser(t *testing.T) {
	testCases := map[string]struct {
		inUser model.User

		withTenantVerification     bool
		propagate                  bool
		tenantCreateUserErr        error
		tenantDeleteUserErr        error
		shouldVerifyTenant         bool
		shouldCompensateTenantUser bool

		dbUser       *model.User
		dbGetUserErr error

		dbErr error

		outErr error
	}{
		"ok": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			dbErr:              nil,
			outErr:             nil,
			propagate:          true,
			shouldVerifyTenant: false,
		},
		"ok, multitenant": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification: true,
			propagate:              true,
			tenantCreateUserErr:    nil,
			shouldVerifyTenant:     true,

			dbErr:  nil,
			outErr: nil,
		},
		"ok, multitenant, progate: false": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification: true,
			propagate:              false,
			tenantCreateUserErr:    nil,
			shouldVerifyTenant:     false,

			dbErr:  nil,
			outErr: nil,
		},
		"error, multitenant: duplicate user": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification: true,
			propagate:              true,
			tenantCreateUserErr:    ct.ErrDuplicateUser,
			shouldVerifyTenant:     true,
			dbUser: &model.User{
				ID:       "1234",
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},

			dbErr:  nil,
			outErr: errors.New("user with a given email already exists"),
		},
		"error, multitenant: duplicate user, no user in useradm": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification:     true,
			propagate:                  true,
			tenantCreateUserErr:        ct.ErrDuplicateUser,
			shouldVerifyTenant:         true,
			shouldCompensateTenantUser: true,

			dbErr:  nil,
			outErr: errors.New("tenant data out of sync: user with the same name already exists"),
		},
		"error, multitenant: duplicate user, no user in useradm, compensate error": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification:     true,
			propagate:                  true,
			tenantCreateUserErr:        ct.ErrDuplicateUser,
			tenantDeleteUserErr:        errors.New("delate user error"),
			shouldVerifyTenant:         true,
			shouldCompensateTenantUser: true,

			dbErr:  nil,
			outErr: errors.New("tenant data out of sync: faield to delete tenant user: delate user error"),
		},
		"error, multitenant: duplicate user, get user error": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification: true,
			propagate:              true,
			shouldVerifyTenant:     true,
			tenantCreateUserErr:    ct.ErrDuplicateUser,

			dbGetUserErr: errors.New("db error"),
			dbErr:        nil,
			outErr:       errors.New("tenant data out of sync: failed to get user from db: db error"),
		},
		"error, multitenant: generic": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			withTenantVerification: true,
			propagate:              true,
			tenantCreateUserErr:    errors.New("http 500"),
			shouldVerifyTenant:     true,

			dbErr:  nil,
			outErr: errors.New("useradm: failed to create user in tenantadm: http 500"),
		},
		"db error: duplicate email": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			dbErr:  store.ErrDuplicateEmail,
			outErr: store.ErrDuplicateEmail,
		},
		"db error, multitenant: duplicate email": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			withTenantVerification: true,
			propagate:              true,
			shouldVerifyTenant:     true,
			dbErr:                  store.ErrDuplicateEmail,
			outErr:                 store.ErrDuplicateEmail,
		},
		"db error: general": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			dbErr: errors.New("no reachable servers"),

			outErr: errors.New("useradm: failed to create user in the db: no reachable servers"),
		},
		"db error, multitenant: general": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			withTenantVerification:     true,
			propagate:                  true,
			shouldVerifyTenant:         true,
			shouldCompensateTenantUser: true,

			dbErr:  errors.New("no reachable servers"),
			outErr: errors.New("useradm: failed to create user in the db: no reachable servers"),
		},
		"db error, multitenant: general, compensate error": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			withTenantVerification:     true,
			propagate:                  true,
			shouldVerifyTenant:         true,
			shouldCompensateTenantUser: true,
			tenantDeleteUserErr:        errors.New("delate user error"),

			dbErr:  errors.New("no reachable servers"),
			outErr: errors.New("useradm: failed to create user in the db: faield to delete tenant user: delate user error: no reachable servers"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("CreateUser",
				ContextMatcher(),
				mock.AnythingOfType("*model.User")).
				Return(tc.dbErr)

			db.On("GetUserByEmail", ContextMatcher(), mock.AnythingOfType("model.Email")).
				Return(tc.dbUser, tc.dbGetUserErr)

			useradm := NewUserAdm(nil, db, Config{})
			cTenant := &mct.ClientRunner{}

			id := &identity.Identity{
				Tenant: "foo",
			}
			ctx = identity.WithContext(ctx, id)

			if tc.shouldVerifyTenant {
				cTenant.On("CreateUser",
					ContextMatcher(),
					mock.AnythingOfType("*tenant.User"),
					&apiclient.HttpApi{}).
					Return(tc.tenantCreateUserErr)

				if tc.shouldCompensateTenantUser {
					cTenant.On("DeleteUser",
						ContextMatcher(),
						mock.AnythingOfType("string"), mock.AnythingOfType("string"),
						&apiclient.HttpApi{}).
						Return(tc.tenantDeleteUserErr)
				}
			}
			if tc.withTenantVerification {
				useradm = useradm.WithTenantVerification(cTenant)
			}

			err := useradm.doCreateUser(ctx, &tc.inUser, tc.propagate)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}

			cTenant.AssertExpectations(t)
		})
	}

}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestUserAdmUpdateUser(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		inUserUpdate   model.UserUpdate
		getUserById    *model.User
		getUserByIdErr error

		verifyTenant bool
		tenantErr    error

		dbErr error

		outErr error
	}{
		"ok": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			dbErr:  nil,
			outErr: nil,
		},
		"ok email with current token": {
			inUserUpdate: model.UserUpdate{
				Email:           "foofoo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
				Token:           &jwt.Token{Claims: jwt.Claims{ID: oid.NewUUIDv5("token-1")}},
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			dbErr:  nil,
			outErr: nil,
		},
		"ok with current token": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
				Token:           &jwt.Token{Claims: jwt.Claims{ID: oid.NewUUIDv5("token-1")}},
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			verifyTenant: true,
			tenantErr:    nil,

			dbErr:  nil,
			outErr: nil,
		},
		"ok, multitenant": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			verifyTenant: true,
			tenantErr:    nil,

			dbErr:  nil,
			outErr: nil,
		},
		"error, multitenant: duplicate user": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			verifyTenant: true,
			tenantErr:    ct.ErrDuplicateUser,

			dbErr:  nil,
			outErr: errors.New("user with a given email already exists"),
		},
		"error, multitenant: not found": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			verifyTenant: true,
			tenantErr:    ct.ErrUserNotFound,

			dbErr:  nil,
			outErr: errors.New("user not found"),
		},
		"error, multitenant: generic": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			verifyTenant: true,
			tenantErr:    errors.New("http 500"),

			dbErr:  nil,
			outErr: errors.New("useradm: failed to update user in tenantadm: http 500"),
		},
		"db error: duplicate email": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Email:    "foo@bar.com",
				Password: hashPassword("current"),
			},

			dbErr:  store.ErrDuplicateEmail,
			outErr: store.ErrDuplicateEmail,
		},
		"db error: general": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			dbErr:  errors.New("no reachable servers"),
			outErr: errors.New("useradm: failed to update user information: no reachable servers"),
		},
		"error: getUserById": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},
			getUserByIdErr: errors.New("error"),

			dbErr:  nil,
			outErr: errors.New("useradm: failed to get user: error"),
		},
		"error: getUserById not found": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "current",
			},

			dbErr:  nil,
			outErr: store.ErrUserNotFound,
		},
		"error: password mismatch": {
			inUserUpdate: model.UserUpdate{
				Email:           "foo@bar.com",
				Password:        "correcthorsebatterystaple",
				CurrentPassword: "wrong",
			},
			getUserById: &model.User{
				Password: hashPassword("current"),
			},

			dbErr:  nil,
			outErr: ErrCurrentPasswordMismatch,
		},
		"error: email without current password": {
			inUserUpdate: model.UserUpdate{
				Email: "foobar@bar.com",
			},
			getUserById: &model.User{
				Email:    "foo@bar.com",
				Password: hashPassword("current"),
			},

			dbErr:  nil,
			outErr: ErrCurrentPasswordMismatch,
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			const userID = "8258b2c3-38c2-4ffd-97c6-19e43f1ea2cf"

			ctx := identity.WithContext(context.Background(), &identity.Identity{Subject: userID})

			db := &mstore.DataStore{}
			defer db.AssertExpectations(t)

			db.On("GetUserAndPasswordById",
				ContextMatcher(),
				mock.AnythingOfType("string"),
			).Return(tc.getUserById, tc.getUserByIdErr)

			if tc.getUserByIdErr == nil && tc.outErr != ErrCurrentPasswordMismatch &&
				(len(tc.inUserUpdate.Password) == 0 || tc.getUserById != nil) &&
				(!tc.verifyTenant || tc.tenantErr == nil) {
				db.On("UpdateUser",
					ContextMatcher(),
					userID,
					mock.AnythingOfType("*model.UserUpdate")).
					Return(&model.User{
						Email:    tc.inUserUpdate.Email,
						Password: tc.inUserUpdate.Password,
					}, tc.dbErr)

				if tc.dbErr == nil && tc.inUserUpdate.Token == nil {
					db.On("DeleteTokensByUserId",
						ContextMatcher(),
						mock.AnythingOfType("string"),
					).Return(nil)
				} else if tc.dbErr == nil {
					db.On("DeleteTokensByUserIdExceptCurrentOne",
						ContextMatcher(),
						mock.AnythingOfType("string"),
						tc.inUserUpdate.Token.ID,
					).Return(nil)
				}
			}

			useradm := NewUserAdm(nil, db, Config{})

			if tc.verifyTenant {
				id := &identity.Identity{
					Tenant:  "foo",
					Subject: userID,
				}
				ctx = identity.WithContext(ctx, id)

				cTenant := &mct.ClientRunner{}
				defer cTenant.AssertExpectations(t)

				cTenant.On("UpdateUser",
					ContextMatcher(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.UserUpdate"),
					&apiclient.HttpApi{}).
					Return(tc.tenantErr)
				useradm = useradm.WithTenantVerification(cTenant)
			}

			err := useradm.UpdateUser(ctx, userID, &tc.inUserUpdate)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}

	etagPtr := func(etag model.ETag) *model.ETag { return &etag }
	type testCase struct {
		Name string

		CTX        context.Context
		ID         string
		UserUpdate *model.UserUpdate

		DataStore       func(t *testing.T, self *testCase) *mstore.DataStore
		TenantadmClient func(t *testing.T, self *testCase) *mct.ClientRunner

		Error error
	}
	tcs := []testCase{{
		Name: "changing password of other user not allowed",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant:  "000000000000000000000000",
			Subject: "36481319-7986-4bd9-9621-f143fc42bcca",
		}),
		ID: "0db11a0e-afac-4d73-aa6b-ccd857019553",
		UserUpdate: &model.UserUpdate{
			Password: "foobar",
			ETag:     etagPtr(model.ETag{0}),
		},
		DataStore: func(t *testing.T, self *testCase) *mstore.DataStore {
			ds := new(mstore.DataStore)
			ds.On("GetUserAndPasswordById", self.CTX, self.ID).
				Return(&model.User{ID: self.ID, ETag: &model.ETag{0}}, nil)
			return ds
		},
		Error: ErrCannotModifyPassword,
	}, {
		Name: "entity tag mismatch/on user lookup",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant:  "000000000000000000000000",
			Subject: "36481319-7986-4bd9-9621-f143fc42bcca",
		}),
		ID: "0db11a0e-afac-4d73-aa6b-ccd857019553",
		UserUpdate: &model.UserUpdate{
			Email: model.Email("test@mender.io"),
			ETag:  etagPtr(model.ETag{0}),
		},
		DataStore: func(t *testing.T, self *testCase) *mstore.DataStore {
			ds := new(mstore.DataStore)
			ds.On("GetUserAndPasswordById", self.CTX, self.ID).
				Return(&model.User{ID: self.ID, ETag: &model.ETag{1}}, nil)
			return ds
		},
		Error: ErrETagMismatch,
	}, {
		Name: "entity tag mismatch/on user update",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant:  "000000000000000000000000",
			Subject: "36481319-7986-4bd9-9621-f143fc42bcca",
		}),
		ID: "0db11a0e-afac-4d73-aa6b-ccd857019553",
		UserUpdate: &model.UserUpdate{
			Email: model.Email("test@mender.io"),
			ETag:  etagPtr(model.ETag{0}),
		},
		DataStore: func(t *testing.T, self *testCase) *mstore.DataStore {
			ds := new(mstore.DataStore)
			ds.On("GetUserAndPasswordById", self.CTX, self.ID).
				Return(&model.User{ID: self.ID, ETag: &model.ETag{0}}, nil).
				On("UpdateUser", self.CTX, self.ID, self.UserUpdate).
				Return(nil, store.ErrUserNotFound)

			return ds
		},
		TenantadmClient: func(t *testing.T, self *testCase) *mct.ClientRunner {
			tnc := new(mct.ClientRunner)
			tnc.On("UpdateUser",
				self.CTX,
				"000000000000000000000000",
				self.ID,
				&tenant.UserUpdate{
					Name: string(self.UserUpdate.Email),
				},
				mock.AnythingOfType("*http.Client")).
				Return(nil)
			return tnc
		},
		Error: ErrETagMismatch,
	}}
	for i := range tcs {
		tc := tcs[i]
		t.Run(tc.Name, func(t *testing.T) {
			var (
				ds  *mstore.DataStore
				tnc *mct.ClientRunner
			)
			if tc.DataStore != nil {
				ds = tc.DataStore(t, &tc)
			} else {
				ds = new(mstore.DataStore)
			}
			defer ds.AssertExpectations(t)
			if tc.TenantadmClient != nil {
				tnc = tc.TenantadmClient(t, &tc)
			} else {
				tnc = new(mct.ClientRunner)
			}
			defer tnc.AssertExpectations(t)

			app := UserAdm{
				verifyTenant: true,
				cTenant:      tnc,
				db:           ds,
				clientGetter: func() apiclient.HttpRunner {
					return new(http.Client)
				},
			}

			err := app.UpdateUser(tc.CTX, tc.ID, tc.UserUpdate)
			if tc.Error != nil {
				assert.ErrorIs(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAdmVerify(t *testing.T) {
	testCases := map[string]struct {
		token *jwt.Token

		dbUser    *model.User
		dbUserErr error

		dbToken    *jwt.Token
		dbTokenErr error

		err error
	}{
		"ok": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
			dbUser: &model.User{
				ID: oid.NewUUIDv5("1234").String(),
			},
			dbToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
		},
		"error: invalid token issuer": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "foo",
					User:    true,
				},
			},
			err: ErrUnauthorized,
		},
		"error: not a user token": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
				},
			},
			err: ErrUnauthorized,
		},
		"error: user not found": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
			err: ErrUnauthorized,
		},
		"error: token not found": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
			dbUser: &model.User{
				ID: oid.NewUUIDv5("1234").String(),
			},

			dbToken:    nil,
			dbTokenErr: nil,

			err: ErrUnauthorized,
		},
		"error: db user": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
			dbUserErr: errors.New("db internal error"),

			err: errors.New("useradm: failed to get user: db internal error"),
		},
		"error: db token": {
			token: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("token-1"),
					Subject: oid.NewUUIDv5("1234"),
					Issuer:  "mender",
					User:    true,
				},
			},
			dbUser: &model.User{
				ID: "1234",
			},

			dbToken:    nil,
			dbTokenErr: errors.New("db failed"),

			err: errors.New("useradm: failed to get token: db failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			config := Config{Issuer: "mender"}

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("GetUserById", ctx,
				tc.token.Claims.Subject.String()).
				Return(tc.dbUser, tc.dbUserErr)
			db.On("GetTokenById", ctx, tc.token.ID).
				Return(tc.dbToken, tc.dbTokenErr)

			useradm := NewUserAdm(nil, db, config)

			err := useradm.Verify(ctx, tc.token)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAdmGetUsers(t *testing.T) {
	t.Parallel()
	ts := time.Now()
	testCases := map[string]struct {
		dbUsers []model.User
		dbErr   error

		err error
	}{
		"ok: some users": {
			dbUsers: []model.User{
				{
					ID:        "1",
					Email:     "foo",
					CreatedTs: &ts,
				},
				{
					ID:        "2",
					Email:     "bar",
					UpdatedTs: &ts,
				},
			},
			dbErr: nil,
			err:   nil,
		},
		"ok: no users": {
			dbUsers: []model.User{},
			dbErr:   nil,
			err:     nil,
		},
		"error: db": {
			dbUsers: nil,
			dbErr:   errors.New("db connection failed"),
			err:     errors.New("useradm: failed to get users: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("GetUsers", ctx, model.UserFilter{}).
				Return(tc.dbUsers, tc.dbErr)

			useradm := NewUserAdm(nil, db, Config{})

			users, err := useradm.GetUsers(ctx, model.UserFilter{})

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.dbUsers, users)
			}
		})
	}
}

func TestUserAdmGetUser(t *testing.T) {
	t.Parallel()

	ts := time.Now()

	testCases := map[string]struct {
		dbUser *model.User
		dbErr  error

		err error
	}{
		"ok 1": {
			dbUser: &model.User{
				ID:        "1",
				Email:     "foo",
				UpdatedTs: &ts,
				CreatedTs: &ts,
			},
			dbErr: nil,
			err:   nil,
		},
		"ok: no user": {
			dbUser: nil,
			dbErr:  nil,
			err:    nil,
		},
		"error: generic db error": {
			dbUser: nil,
			dbErr:  errors.New("db connection failed"),
			err:    errors.New("useradm: failed to get user: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("GetUserById", ctx, "foo").Return(tc.dbUser, tc.dbErr)

			useradm := NewUserAdm(nil, db, Config{})

			user, err := useradm.GetUser(ctx, "foo")

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.dbUser, user)
			}
		})
	}
}

func TestUserAdmDeleteUser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		verifyTenant      bool
		tenantErr         error
		dbDeleteUserErr   error
		dbDeleteTokensErr error
		err               error
	}{
		"ok": {
			err: nil,
		},
		"ok, multitenant": {
			verifyTenant: true,
			err:          nil,
		},
		"multitenant, tenantadm error": {
			verifyTenant: true,
			tenantErr:    errors.New("http 500"),
			err:          errors.New("useradm: failed to delete user in tenantadm: http 500"),
		},
		"error deleting user": {
			dbDeleteUserErr: errors.New("db connection failed"),
			err:             errors.New("useradm: failed to delete user: db connection failed"),
		},
		"error deleting user tokens": {
			dbDeleteTokensErr: errors.New("db connection failed"),
			err:               errors.New("useradm: failed to delete user tokens: db connection failed"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("DeleteUser", ContextMatcher(), "foo").Return(tc.dbDeleteUserErr)
			db.On("DeleteTokensByUserId", ContextMatcher(), "foo").Return(tc.dbDeleteTokensErr)

			useradm := NewUserAdm(nil, db, Config{})
			if tc.verifyTenant {
				id := &identity.Identity{
					Tenant: "bar",
				}
				ctx = identity.WithContext(ctx, id)

				cTenant := &mct.ClientRunner{}
				cTenant.On("DeleteUser",
					ContextMatcher(),
					"bar", "foo",
					&apiclient.HttpApi{}).
					Return(tc.tenantErr)
				useradm = useradm.WithTenantVerification(cTenant)
			}

			err := useradm.DeleteUser(ctx, "foo")

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAdmCreateTenant(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tenant    string
		tenantErr error
		err       error
	}{
		"ok": {
			tenant: "foobar",
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			useradm := NewUserAdm(nil, nil, Config{})

			err := useradm.CreateTenant(ctx, model.NewTenant{ID: tc.tenant})
			assert.NoError(t, err)
		})
	}
}

func TestUserAdmSetPassword(t *testing.T) {
	testCases := map[string]struct {
		inUser       model.User
		currentToken *jwt.Token
		dbGetErr     error
		dbUpdateErr  error
		outErr       error
		foundUser    *model.User
	}{
		"ok": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			dbGetErr:  nil,
			outErr:    nil,
			foundUser: &model.User{ID: "test_id"},
		},

		"ok with current token": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},
			currentToken: &jwt.Token{Claims: jwt.Claims{ID: oid.NewUUIDv5("token-1")}},
			dbGetErr:     nil,
			outErr:       nil,
			foundUser:    &model.User{ID: "test_id"},
		},

		"error, user not found": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			dbGetErr:  nil,
			outErr:    errors.New("user not found"),
			foundUser: nil,
		},

		"error, get from db": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			dbGetErr:  errors.New("db failed"),
			outErr:    errors.New("useradm: failed to get user by email: db failed"),
			foundUser: nil,
		},
		"error, update db": {
			inUser: model.User{
				Email:    "foo@bar.com",
				Password: "correcthorsebatterystaple",
			},

			dbUpdateErr: errors.New("db failed"),
			outErr:      errors.New("useradm: failed to update user information: db failed"),
			foundUser:   &model.User{ID: "test_id"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}

			db.On("GetUserByEmail",
				ContextMatcher(),
				tc.inUser.Email).
				Return(tc.foundUser, tc.dbGetErr)

			if tc.foundUser != nil {
				db.On("UpdateUser",
					ContextMatcher(),
					tc.foundUser.ID,
					mock.AnythingOfType("*model.UserUpdate")).
					Return(&tc.inUser, tc.dbUpdateErr)
			}

			if tc.foundUser != nil && tc.dbUpdateErr == nil {
				if tc.currentToken == nil {
					db.On("DeleteTokensByUserId",
						ContextMatcher(),
						mock.AnythingOfType("string"),
					).Return(nil)
				} else {
					db.On("DeleteTokensByUserIdExceptCurrentOne",
						ContextMatcher(),
						mock.AnythingOfType("string"),
						tc.currentToken.ID,
					).Return(nil)
				}
			}

			useradm := NewUserAdm(nil, db, Config{})
			cTenant := &mct.ClientRunner{}

			err := useradm.SetPassword(ctx, model.UserUpdate{Email: tc.inUser.Email, Password: "new-password", Token: tc.currentToken})

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}

			cTenant.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}

}

func ContextMatcher() interface{} {
	return mock.MatchedBy(func(c context.Context) bool {
		return true
	})
}

func TestUserAdmDeleteTokens(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		user   string
		tenant string
		dbErr  error

		outErr error
	}{
		"ok": {},
		"ok, tenant": {
			tenant: "foo",
		},
		"ok, tenant and user": {
			tenant: "foo",
			user:   "foo",
		},
		"db error": {
			user:   "foo",
			tenant: "foo",
			dbErr:  errors.New("db connection failed"),
			outErr: errors.New("failed to delete tokens for tenant: foo, user id: foo: db connection failed"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("DeleteTokens", ContextMatcher(), mock.AnythingOfType("string")).Return(tc.dbErr)
			db.On("DeleteTokensByUserId", ContextMatcher(), mock.AnythingOfType("string")).Return(tc.dbErr)

			useradm := NewUserAdm(nil, db, Config{})

			err := useradm.DeleteTokens(ctx, tc.tenant, tc.user)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestUserAdmIssuePersonalAccessToken(t *testing.T) {
	testCases := map[string]struct {
		tokenRequest model.TokenRequest

		callDbSaveToken   bool
		dbSaveTokenErr    error
		callDbCountTokens bool
		dbCountTokens     int64
		dbCountTokensErr  error

		config Config

		outErr error
	}{
		"ok": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbSaveToken:   true,
			callDbCountTokens: true,
			dbCountTokens:     9,
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitTokensPerUser:    10,
			},
		},
		"ok, no limit": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbSaveToken: true,
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
			},
		},
		"error: too many tokens": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbCountTokens: true,
			dbCountTokens:     10,
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitTokensPerUser:    10,
			},
			outErr: ErrTooManyTokens,
		},
		"error: count tokens error": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbCountTokens: true,
			dbCountTokens:     0,
			dbCountTokensErr:  errors.New("count tokens error"),
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitTokensPerUser:    10,
			},
			outErr: errors.New(
				"useradm: failed to count personal access tokens: count tokens error"),
		},
		"error: duplicate token name": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbSaveToken:   true,
			dbSaveTokenErr:    store.ErrDuplicateTokenName,
			callDbCountTokens: true,
			dbCountTokens:     9,
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitTokensPerUser:    10,
			},
			outErr: ErrDuplicateTokenName,
		},
		"error: save token error": {
			tokenRequest: model.TokenRequest{
				Name:      stringPtr("foo"),
				ExpiresIn: 3600,
			},
			callDbSaveToken:   true,
			dbSaveTokenErr:    errors.New("save token error"),
			callDbCountTokens: true,
			dbCountTokens:     9,
			config: Config{
				Issuer:                "foobar",
				ExpirationTimeSeconds: 10,
				LimitTokensPerUser:    10,
			},
			outErr: errors.New("useradm: failed to save token: save token error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			defer db.AssertExpectations(t)
			if tc.callDbSaveToken {
				db.On("SaveToken",
					ContextMatcher(),
					mock.AnythingOfType("*jwt.Token")).
					Return(tc.dbSaveTokenErr)
			}
			if tc.callDbCountTokens {
				db.On("CountPersonalAccessTokens",
					ContextMatcher(),
					"foo").
					Return(tc.dbCountTokens, tc.dbCountTokensErr)
			}

			mockJWTHandler := mjwt.Handler{}
			mockJWTHandler.On("ToJWT",
				mock.AnythingOfType("*jwt.Token"),
			).Return("signed", nil)

			useradm := NewUserAdm(map[int]jwt.Handler{0: &mockJWTHandler}, db, tc.config)

			id := &identity.Identity{
				Subject: "foo",
				Tenant:  "bar",
			}
			ctx = identity.WithContext(ctx, id)

			_, err := useradm.IssuePersonalAccessToken(ctx, &tc.tokenRequest)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestUserAdmGetPlans(t *testing.T) {
	t.Parallel()

	planList := []model.Plan{
		{
			Name: "plan1",
		},
		{
			Name: "plan2",
		},
		{
			Name: "plan3",
		},
		{
			Name: "plan4",
		},
		{
			Name: "plan5",
		},
		{
			Name: "plan6",
		},
		{
			Name: "plan7",
		},
		{
			Name: "plan8",
		},
		{
			Name: "plan9",
		},
		{
			Name: "plan10",
		},
	}

	testCases := map[string]struct {
		skip  int
		limit int
		plans []model.Plan
	}{
		"ok, empty": {
			skip:  0,
			limit: 0,
			plans: []model.Plan{},
		},
		"ok": {
			skip:  0,
			limit: 10,
			plans: planList,
		},
		"ok 1": {
			skip:  3,
			limit: 4,
			plans: planList[3:7],
		},
		"ok 2": {
			skip:  7,
			limit: 10,
			plans: planList[7:],
		},
		"ok 3": {
			skip:  20,
			limit: 10,
			plans: []model.Plan{},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			useradm := NewUserAdm(nil, nil, Config{})

			model.PlanList = planList
			plans := useradm.GetPlans(ctx, tc.skip, tc.limit)

			assert.Equal(t, plans, tc.plans)
		})
	}
}

func TestUserAdmGetPlanBinding(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		plans       []model.Plan
		planBinding *model.PlanBindingDetails
	}{
		"ok": {
			plans: []model.Plan{
				{
					Name: "plan1",
				},
			},
			planBinding: &model.PlanBindingDetails{
				Plan: model.Plan{
					Name: "plan1",
				},
			},
		},
		"ok, no plans": {
			planBinding: &model.PlanBindingDetails{},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			useradm := NewUserAdm(nil, nil, Config{})
			model.PlanList = tc.plans

			pB, err := useradm.GetPlanBinding(ctx)
			assert.NoError(t, err)
			assert.Equal(t, pB, tc.planBinding)
		})
	}
}

var (
	// tokens* maps hold tokens gined by keys of given id and carrying kid, KeyIdZero, or no kid respectively
	tokensEdKeyNoKid = map[int]string{
		22899: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJhNzAxNTljYy1jZDE1LTQ0ZGMtODRmZi0wZmZiMjMzZTg1YzMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzNjYsImlhdCI6MTcwMTg1MDM2NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzY2fQ.jM3hZ2tRHEWXIWpKvmfKtQ_1HLlOt5wWc9mLVgp8TThT-Mmo2OrTfOZJXMBo00txPcMPJyRSC6ktb7uWNUiOBA",
		14211: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI4Mzc5YWE4YS05MzU2LTRmODgtYWUyNS1kYmYwMDcwMDdjODMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzNzgsImlhdCI6MTcwMTg1MDM3OCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzc4fQ.TqbH29njljgnrDNZYQgi2wlfSfZTAsfcUQpI5nS4e6iURBI57uTecovGsb0kOMAGQbyhFLWNG16etvve8-HoCw",
		5539:  "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJiNDcxOGE2NC05NWMxLTQ3MjUtODczNS04NDNhMDQyOWNiZWUiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzOTMsImlhdCI6MTcwMTg1MDM5MywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzkzfQ.UixZKn4b7pOBTYD1IksEE0I-memH7HWYIR_wNsVcSrgz8jA15veLhWwT4Jo7mn0NFMiBYu-Jm01h2VGncLr_CQ",
		826:   "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI4OWZjNWUxOC0xZmMyLTRiNzctYWZmMC03ZDdiMzMwODBjZDciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzOTksImlhdCI6MTcwMTg1MDM5OSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzk5fQ.VGVjdK4ytH-bVwZQ63xx8U9Md2T3RXJkM4e9EbVKuae1k82h_N4F1152Hd8H8dfBmGeUdAj5fbv54DXsFnr7Cw",
	}
	tokensRSAKeyNoKid = map[int]string{
		21172: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJkNzI0YjhmMS0xOGYyLTQ5NzItODA1YS1mNGZiZGVjNWNmMzEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyODAsImlhdCI6MTcwMTg1MDI4MCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjgwfQ.joeE0nOrOKw1vv7BHP4kRyZ8xTmufP881tuiGfGjTRaDRKE9kUNkYlJP5WNnvYfnkQY2SYvBiC0nJTLbZtvLTWD36wQUv6e7QR3p28ayrjSrm4ebkEfzeWQW_17oElBRLAzEEnpTENzCKXSGOydkOUS611TIlCtMpzp5EC5dO7VSjAyAD7luFMXZGp_Pc30K8VSD7Rxrg5s1BfbiqhfrmgwHeONr-ScGyvP5czqopY5-gJAkFJ54WwV6ZFqdW3-ykEn7CeNM1BHvN7veMHDsIk_NAqJfeUsOg7EMwZNmSnAKnjB1-xnL4qdqtORTcDrY0fWSeOO9pWO7MsxEQ8b9ag",
		13102: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjNTI1YjUzZC1lNjc0LTQ0YTktYjM0Mi00YmNmMTY3N2QxMzIiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyOTIsImlhdCI6MTcwMTg1MDI5MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjkyfQ.G4OemVKEY2f0qtUfZYWyo_wFem5O9jpABJ2JtT_BlvYAa4RFLVn_AF9-nyZY0aPDmuev5BSidzIC0d9ioTmyCUy6HOODGRosA_YxiI_NlZUoJk54Ml5kbFXzpy_b4TrCqoB7fk0qn8V7XZ7ujdmnaqE_mGdT2BAnyRS3xjxLabY42r9RL365Cu1LFqRPfmg2he6LBIb8KUUV3A-4gsrtvsUb321o_HlrD9sy8Qp2HMazEpC8Kl3b5k3F-rS6heQGyByMXcP-tvKNZCALSwIVvu00IWF8yEKBPd8XYQqPYZd8LYWyDamwrn2q138HjJpoTqWwZHGdVDEFxhQDkKgXJw",
		9478:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwNDlkYzEwMC01MDE3LTQ4ZjQtODdmMC00MDhiYjI1MGQyYjciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyOTgsImlhdCI6MTcwMTg1MDI5OCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjk4fQ.Pf3R-sMfimWokFUxlBLYL4yQjp_VTurG7Gccp9w7kmUb80Kad67ua0l1qe_Fq2PsYd5MSFdDxCC0z6o5biegHMKQiCzgdwWHfTvIDndJds6GZ5khPVwGUIVsY748mMdTwk082VCECRTgPygyOq5dBadC22PlCuYFGOeZGqIp4Fs6hl_NY3qgpPYEQBDG6JtCLa6wuSR-zxo7gVVz1RAVojVAWXfJN99pQT4OYc6DoXjQ5BC6mzKd4Jkao7pS9FnLCr40wjQgkknmjjpRZyVs7nPDuf3iedMuU7N5YH0LYHSflUKk2BB-ZkreL_BTxlXJ_OQKPHiC00mxANbY2cf9cA",
		20433: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI2OGVjNzBjOC1lZjMxLTRjNGItYjMyYS00M2YyOTJmMzRjNjciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzMDksImlhdCI6MTcwMTg1MDMwOSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzA5fQ.Enq2eJVcpCqkVTncTAlC0GXzzLAfIuMHVqptx_Lsx_3jlo8OViM72p_Kq2D-abrrBCTYiwHqqv_1Dz5quS9YAhh1Ck4X6zkSKOtKh1hZuofqqYCzDrpnbi8jSoxNy_LTUo2hTIskIjlNQsw9KdwWV4DmtSzj_kPdCH6NW4gH5mA0ndNYwW-ujKtFUESnG_7HIRwqNPkxF6498JaxBz8_z8pZNdmjAj5hEokTSUewmh54gzCaqKeayaTLdK-hU-7g9yW4ie8XialkrNGnoEEa1RuY9Gf6McKkzntjXLQtX59E1WcYoRHYM0VfY7MUcISCAk1hP6x2oPrO5oqigmFA1g",
	}
	tokensByKeyIdNoKid = map[int]string{
		22899: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJhNzAxNTljYy1jZDE1LTQ0ZGMtODRmZi0wZmZiMjMzZTg1YzMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzNjYsImlhdCI6MTcwMTg1MDM2NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzY2fQ.jM3hZ2tRHEWXIWpKvmfKtQ_1HLlOt5wWc9mLVgp8TThT-Mmo2OrTfOZJXMBo00txPcMPJyRSC6ktb7uWNUiOBA",
		14211: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI4Mzc5YWE4YS05MzU2LTRmODgtYWUyNS1kYmYwMDcwMDdjODMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzNzgsImlhdCI6MTcwMTg1MDM3OCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzc4fQ.TqbH29njljgnrDNZYQgi2wlfSfZTAsfcUQpI5nS4e6iURBI57uTecovGsb0kOMAGQbyhFLWNG16etvve8-HoCw",
		5539:  "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJiNDcxOGE2NC05NWMxLTQ3MjUtODczNS04NDNhMDQyOWNiZWUiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzOTMsImlhdCI6MTcwMTg1MDM5MywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzkzfQ.UixZKn4b7pOBTYD1IksEE0I-memH7HWYIR_wNsVcSrgz8jA15veLhWwT4Jo7mn0NFMiBYu-Jm01h2VGncLr_CQ",
		826:   "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI4OWZjNWUxOC0xZmMyLTRiNzctYWZmMC03ZDdiMzMwODBjZDciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzOTksImlhdCI6MTcwMTg1MDM5OSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzk5fQ.VGVjdK4ytH-bVwZQ63xx8U9Md2T3RXJkM4e9EbVKuae1k82h_N4F1152Hd8H8dfBmGeUdAj5fbv54DXsFnr7Cw",
		21172: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJkNzI0YjhmMS0xOGYyLTQ5NzItODA1YS1mNGZiZGVjNWNmMzEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyODAsImlhdCI6MTcwMTg1MDI4MCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjgwfQ.joeE0nOrOKw1vv7BHP4kRyZ8xTmufP881tuiGfGjTRaDRKE9kUNkYlJP5WNnvYfnkQY2SYvBiC0nJTLbZtvLTWD36wQUv6e7QR3p28ayrjSrm4ebkEfzeWQW_17oElBRLAzEEnpTENzCKXSGOydkOUS611TIlCtMpzp5EC5dO7VSjAyAD7luFMXZGp_Pc30K8VSD7Rxrg5s1BfbiqhfrmgwHeONr-ScGyvP5czqopY5-gJAkFJ54WwV6ZFqdW3-ykEn7CeNM1BHvN7veMHDsIk_NAqJfeUsOg7EMwZNmSnAKnjB1-xnL4qdqtORTcDrY0fWSeOO9pWO7MsxEQ8b9ag",
		13102: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjNTI1YjUzZC1lNjc0LTQ0YTktYjM0Mi00YmNmMTY3N2QxMzIiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyOTIsImlhdCI6MTcwMTg1MDI5MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjkyfQ.G4OemVKEY2f0qtUfZYWyo_wFem5O9jpABJ2JtT_BlvYAa4RFLVn_AF9-nyZY0aPDmuev5BSidzIC0d9ioTmyCUy6HOODGRosA_YxiI_NlZUoJk54Ml5kbFXzpy_b4TrCqoB7fk0qn8V7XZ7ujdmnaqE_mGdT2BAnyRS3xjxLabY42r9RL365Cu1LFqRPfmg2he6LBIb8KUUV3A-4gsrtvsUb321o_HlrD9sy8Qp2HMazEpC8Kl3b5k3F-rS6heQGyByMXcP-tvKNZCALSwIVvu00IWF8yEKBPd8XYQqPYZd8LYWyDamwrn2q138HjJpoTqWwZHGdVDEFxhQDkKgXJw",
		9478:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwNDlkYzEwMC01MDE3LTQ4ZjQtODdmMC00MDhiYjI1MGQyYjciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgyOTgsImlhdCI6MTcwMTg1MDI5OCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMjk4fQ.Pf3R-sMfimWokFUxlBLYL4yQjp_VTurG7Gccp9w7kmUb80Kad67ua0l1qe_Fq2PsYd5MSFdDxCC0z6o5biegHMKQiCzgdwWHfTvIDndJds6GZ5khPVwGUIVsY748mMdTwk082VCECRTgPygyOq5dBadC22PlCuYFGOeZGqIp4Fs6hl_NY3qgpPYEQBDG6JtCLa6wuSR-zxo7gVVz1RAVojVAWXfJN99pQT4OYc6DoXjQ5BC6mzKd4Jkao7pS9FnLCr40wjQgkknmjjpRZyVs7nPDuf3iedMuU7N5YH0LYHSflUKk2BB-ZkreL_BTxlXJ_OQKPHiC00mxANbY2cf9cA",
		20433: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI2OGVjNzBjOC1lZjMxLTRjNGItYjMyYS00M2YyOTJmMzRjNjciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgzMDksImlhdCI6MTcwMTg1MDMwOSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMzA5fQ.Enq2eJVcpCqkVTncTAlC0GXzzLAfIuMHVqptx_Lsx_3jlo8OViM72p_Kq2D-abrrBCTYiwHqqv_1Dz5quS9YAhh1Ck4X6zkSKOtKh1hZuofqqYCzDrpnbi8jSoxNy_LTUo2hTIskIjlNQsw9KdwWV4DmtSzj_kPdCH6NW4gH5mA0ndNYwW-ujKtFUESnG_7HIRwqNPkxF6498JaxBz8_z8pZNdmjAj5hEokTSUewmh54gzCaqKeayaTLdK-hU-7g9yW4ie8XialkrNGnoEEa1RuY9Gf6McKkzntjXLQtX59E1WcYoRHYM0VfY7MUcISCAk1hP6x2oPrO5oqigmFA1g",
	}
	tokensByKid = []map[int]string{
		22899: {
			22899: "eyJhbGciOiJFZERTQSIsImtpZCI6MjI4OTksInR5cCI6IkpXVCJ9.eyJqdGkiOiJmNmRiMTAxMy1jM2I4LTRhMDQtOTMzNS04ZmEzNjdjYTBiNTMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1MTcsImlhdCI6MTcwMTg0OTUxNywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTE3fQ.J8KAOd6gFcpusxJoREnrxyqcQ68IvWa8haGaAwExnUbJBL7QJj8i5KC4jhZqq097QB3yMWICGH02yl4o0WUnCQ",
		},
		14211: {
			14211: "eyJhbGciOiJFZERTQSIsImtpZCI6MTQyMTEsInR5cCI6IkpXVCJ9.eyJqdGkiOiJmY2M0MThiMi0zNDBiLTRiZmYtOGJiYi03NDczOWMxMjU5MDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NDEsImlhdCI6MTcwMTg0OTU0MSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTQxfQ.I6_FN8Ei5niyIQq6IKxf4XhHxWSxXWYaOmYSkSlZlDy-8-8rA5hm3c3Djn9ROP9VcKKM7UMbP3DMOK3wCg8xCg",
		},
		5539: {
			5539: "eyJhbGciOiJFZERTQSIsImtpZCI6NTUzOSwidHlwIjoiSldUIn0.eyJqdGkiOiJkMzQ4NTJiNy1lZDQzLTQ5MjItODU3MS01YTljZDg3ZTEzYmMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NTIsImlhdCI6MTcwMTg0OTU1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTUyfQ.EkLoISjqnq-sgQOqtY4J58g0fiyVZHEI_7Gl9FNcffsKVLWbRyOzdV1pedu3DXNmG4nU8cJ4_eyT_M64D4Z_BQ",
		},
		826: {
			826: "eyJhbGciOiJFZERTQSIsImtpZCI6ODI2LCJ0eXAiOiJKV1QifQ.eyJqdGkiOiJlNzU0YWRmOS0wZDYxLTRlYTMtOTdjNS1jMDhhODVlOTQwMDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NzYsImlhdCI6MTcwMTg0OTU3NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTc2fQ._z9Aqj-OxsRvK-ugQ-1s_fNElN0-A21Xm2NW5vGeBNEsxLvnOhr4fngQaa19JIZKBhn1smO1RVRJY94xYU7aBQ",
		},
		21172: {
			21172: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjExNzIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIyOGJjZGU4Ny00NzBmLTRlNTEtYjJlYS1mZTk5YTlkZTEyOGMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NTIsImlhdCI6MTcwMTg0OTc1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzUyfQ.jguN8vthU4Gjkd0OOeEgVFvHuhOihowq91-jXBXSYcS6wBofvNIVh5qqMVlj1HOd4sfV6BZpg0TU1SI4v92ApBjOJMZOw2Ah9TyExOBJYQ0iSw8Aa8fTByMhtBBNj_w2HyV9TwJfkKsMSZ6a4oJB4ZPGZDizXvEiMWYWkpTrEBTHbzjEIBUnIEZm2QQEVlODdIuKYyTlRTnVoBXkVSBUIaEOrpfnRr0AWKISQGToU3mMVNHkny9LbW47pKv0U3VK8p9x-5yqlzMDL4P7GgSdHIb_hL1BahUAEPcnDaOjzhw6WVUTrp77poXnaPLF1BOEirdb6hfsw8fAlmQAm4113Q",
		},
		13102: {
			13102: "eyJhbGciOiJSUzI1NiIsImtpZCI6MTMxMDIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI5MDFhMzA1Mi1mMmRjLTQwZjYtYmM4Ny1lZWY1ZmFlYjVmNTciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NzcsImlhdCI6MTcwMTg0OTc3NywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5Nzc3fQ.WU0Cd_wtwwyZJn-cyASqKe1yY_BYrRIltHTiZAL5zLo2e7gtLvp2uEKGI93oQGONGBKhd-_oTM_o8M6VqK0-4l_Sx600DOYkJlVzUngJ1GvIcWwL6AMHClPtA-t_lvpLKXqJwznD311bOEWciOrVXlK973M1stWuIdm-UTPr5xeP-4hRj_wLVfpMlAjQQIoJCQIq1tfg5GsUxO6YXT0wC0tEg5CeZHgCx7_bvqnbECKvQqOzb4NkfEqqc16TLNbjZvvKfvP8FrLbFlKtKe7RPI_oC2pzh6iAe3CG1ZLexArosfKKza0QqjTm7OvNA1Q0kRhDorzUotRSE3JHch6tvA",
		},
		9478: {
			9478: "eyJhbGciOiJSUzI1NiIsImtpZCI6OTQ3OCwidHlwIjoiSldUIn0.eyJqdGkiOiJmNGIwZGM2YS1mMWJjLTRkN2UtYTA0NS1iNzEzMmFkZTZlMGEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3OTIsImlhdCI6MTcwMTg0OTc5MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzkyfQ.aBP5XNAYbKFmx5DRAwTauglNd9A-JiimEb-EO_kqgHkjZA2-Jfte69TJAp1otFA-dBVrzcR4Kq1k5U0p7oJP4Q7Z1eS9ViXsSWxJAw8DjzLBztD89UNrIeSgEG7TRzL8CdC_51gudHbgaBYM0oyT_ul1qintktTNicLYjVp0a2x68ZvpJxiyq1jY2kTC_0lBk-ONAPxfSF6wU_wWY0c-XFLmgjto99NCPxemnCV33sZ7z2Fg533u04DaXGKYmX0dP0gGaoFQRQ6AG5H-C4pr8v7oe830zoiAOderv93l7AQz-R9n8FYw0YlGQocU05ZzP7aFgsG2uxR7BYOOF09xPg",
		},
		20433: {
			20433: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjA0MzMsInR5cCI6IkpXVCJ9.eyJqdGkiOiIxYmJmODI5MC03NDIzLTQ1NjEtYTY3My02NWY4NmJiYjlkMzMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc4MTAsImlhdCI6MTcwMTg0OTgxMCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5ODEwfQ.TJ3mEjhnWYHlPRAjvcH9rTMpjVKIWL_i-hQtubyCkH5WPuOo_DdLQKeiiuJD7jIp7AtzvkX21zTNQWwMsWaExWB3yhNbgHp92KWguUB1EBAfg7t0dDa0uBl6hgbCgV97Z4gM53BzbNySt9Vjm88meiobXUnpBy345B-XUZsLTaHdBRwrnueijkVhlALjfj96Xlm-QAPyN0JFNaLCAnWGGmO_dd9t0vlecMUuxkwfjnpF3zGVfsMSA5xwAPkbFKrt5kjEE0Wyngbe8UMhl5SWORb9yAiK7ubIXU-IjYghYBKFj0VjgoG2bmpZXHJuUx8Dr9HlFPwNl0IGZgIJMeri2Q",
		},
	}
	tokensByKidAll = []map[int]string{
		0: {
			22899: "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIzY2E3YWExZC05NjViLTRjZTQtODljZS1lYjEwY2I4YzdlYTkiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNDMsImlhdCI6MTcwMTg1MDA0MywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDQzfQ.WAKMeS8VYZ-v9FEzjs2yRo7ZddtdDCw9yuZBEVddca8dzCM6sNyWbjKzFQXFf4S6Dgu1stH3ZjKtOfdZ_9usBQ",
			14211: "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIxMTY4MDcwMi01ZGIzLTQyOTUtYjVjOS1iZjc4NWFlMzk2MWEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNTUsImlhdCI6MTcwMTg1MDA1NSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDU1fQ.Q6mdarlwtBdBfXIu-YbNCAZEsrkxjbwUgs4XBczrB1xa7V3Pj0QLL5gLxXsOOt_rPs10nPoYLKoXexdnh54jDQ",
			5539:  "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI3YjY1OTgzMS1jZTVkLTRjOTYtYWVmZC02NjBlNmIxMjhkZDEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNjUsImlhdCI6MTcwMTg1MDA2NSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDY1fQ.4E94wv9NALpuA2sTaduw5_WJ2aGtqqig5TyXzzxYEiVMtj9lcxu5PWMsdRQLG49MMEmpUyKcjXrNwHUabzIICg",
			826:   "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI0ZDVhNjViOS0wZmExLTQyZTYtYTY4MC1lZWZmNWJjOTFlNWIiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNzYsImlhdCI6MTcwMTg1MDA3NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDc2fQ.5W3W9XIcL8_k8Uc40SLckUIRux7WqsVBRXqYzyVepsXIBvrJ6AXixU9t_EwSW_SYsjD_r0lpUiVx7hG6EQOGDQ",
			21172: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI3ODY4YjExNy02MGM1LTQ5MmQtOWJiNC05ZTE3NjgzOTlhMDkiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNTYsImlhdCI6MTcwMTg1MDE1NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTU2fQ.IeaA4Kso34l_wOp-43BKQgq4r-iFdS1EGcux6eQxxR0ZVn6WuneCeFvIGS-2FeeCFr9AKg6fO-5IxKR8aM0SMYwLthTQWrAdpRhGlsISO12U-Zn1Y8GYTadLPlEIp-9B6cFf3fyRzPmxYbb700pw6mtzjxr0_k-hRl7f3pVGieI4qQNlPoL1D76CwnkJtaxrkJ4ieu65_PWmk-qQpLq-td03rx91pbXiQvW48y-ra6wKvZVzviiRi8GI_AY89QZk0TBtyzasf5glqQxG8PC6fnRBFQviQV2fsH4VrBmCZxuBOIks647JEDDMg2HGAfSZfl-qA2A0nLta7BAx0K6aBQ",
			13102: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiJmZjkyNzE3ZS1lOThiLTQ0ZTgtYTUxZi1lMjg5MDAzMmM1OWMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNjYsImlhdCI6MTcwMTg1MDE2NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTY2fQ.FZHZXPcHEDqTQ4mGxGbz1KXNFWuD2OYYZiJbSTyjidU1XAacx1qTMQvKF9aqaYjPEjNDw5LpLXHHIsBPQ6VCkYnXXsCFC4tCaQ57DLNauqGO4O7grqM3-ZspZzzpCkwkGrXbmolw0pPzz7DjaYLlU2ZVfgNsiOh0nEFgFqj7FryTzLHvfuD8K1RdZ28v5HTzXDF9mIUCvzcDQharV4w35ShN4VzsN_vtjM7fuq9eU643YPlf96CJcM8gUj7op_BQuUNpaPqthfyRAEzHTK7AIRWlQQ_h4dWF5jDuRDkH5pyyPb82yqM7975nupvneGxGBqin9KklgDVcNGZxgogeNA",
			9478:  "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIwZmIwNDQ5NS1mMDA1LTQ5MGItOGMyZC04OTFkY2IyYzEwMGQiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNzQsImlhdCI6MTcwMTg1MDE3NCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTc0fQ.MJuiU7p0aXjponWLzPLzyoLlBh1Z5cyB7vBFkFecT3o9LP2HRkAwtsLo1d799gAnIvf7wE6cuK-Mq9SmC5OfnTWhPldbFa4W2WkiqJ6ptTxnIL_V0puxwrQx3k2sibTznsNsSL0FrFUMnHI4IxhtED4XRV9B9oHor-s3Fo_6m9xtiFwqTbqxdze_QBHXV8R4ixu_w_rGnwKncVEKqJ6E6aduIbX22JsAs-5EprmNldviFFglZoEzwlThmCl2W0WBO1EJ66PvwwM9i0iMnDgdZiPWPF4D861hFHCesds7uEi9ZYZUt_d0hEmI91O34PSpJdxFFsqNMpX1hpxiI8nN4Q",
			20433: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI0NGZlZGRkZi00MzI1LTQxOGUtYmNjNS00MjQyYTE0MDk0NjEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxODEsImlhdCI6MTcwMTg1MDE4MSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTgxfQ.c9p-WRp36LPRqt1nzUQ27n3FCR5ue0XGZtrnGDIsSC3KtlkzUhSDJUlxn4moBAZA9R3Yp-neap7pN-SWvb9xnq--ui9b3DPxo3tn-CAs9_mZ6UJ2y51NzjyLpJe4jG1mkmZVxsCxsSIjvIFoIPtnbXEx4zZ2LxB6Mu0eSf5_j_S4ZO-1Iw93e15o1_v0IZKAcTLiBJR41X09MEBRbdOU-sRXmQvjM2UB6Kv47lliofkYinQ_T3LXrRWxYzW5iKeeZR2AO2olRLPiwWO1zR3tfZzow67XfNIHrmirUnghDrzSFdUGdl7JHAIkUWWske7ct1PEhXMuYepXQXqFcEPz-Q",
		},
		22899: {
			22899: "eyJhbGciOiJFZERTQSIsImtpZCI6MjI4OTksInR5cCI6IkpXVCJ9.eyJqdGkiOiJmNmRiMTAxMy1jM2I4LTRhMDQtOTMzNS04ZmEzNjdjYTBiNTMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1MTcsImlhdCI6MTcwMTg0OTUxNywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTE3fQ.J8KAOd6gFcpusxJoREnrxyqcQ68IvWa8haGaAwExnUbJBL7QJj8i5KC4jhZqq097QB3yMWICGH02yl4o0WUnCQ",
		},
		14211: {
			14211: "eyJhbGciOiJFZERTQSIsImtpZCI6MTQyMTEsInR5cCI6IkpXVCJ9.eyJqdGkiOiJmY2M0MThiMi0zNDBiLTRiZmYtOGJiYi03NDczOWMxMjU5MDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NDEsImlhdCI6MTcwMTg0OTU0MSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTQxfQ.I6_FN8Ei5niyIQq6IKxf4XhHxWSxXWYaOmYSkSlZlDy-8-8rA5hm3c3Djn9ROP9VcKKM7UMbP3DMOK3wCg8xCg",
		},
		5539: {
			5539: "eyJhbGciOiJFZERTQSIsImtpZCI6NTUzOSwidHlwIjoiSldUIn0.eyJqdGkiOiJkMzQ4NTJiNy1lZDQzLTQ5MjItODU3MS01YTljZDg3ZTEzYmMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NTIsImlhdCI6MTcwMTg0OTU1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTUyfQ.EkLoISjqnq-sgQOqtY4J58g0fiyVZHEI_7Gl9FNcffsKVLWbRyOzdV1pedu3DXNmG4nU8cJ4_eyT_M64D4Z_BQ",
		},
		826: {
			826: "eyJhbGciOiJFZERTQSIsImtpZCI6ODI2LCJ0eXAiOiJKV1QifQ.eyJqdGkiOiJlNzU0YWRmOS0wZDYxLTRlYTMtOTdjNS1jMDhhODVlOTQwMDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NzYsImlhdCI6MTcwMTg0OTU3NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTc2fQ._z9Aqj-OxsRvK-ugQ-1s_fNElN0-A21Xm2NW5vGeBNEsxLvnOhr4fngQaa19JIZKBhn1smO1RVRJY94xYU7aBQ",
		},
		21172: {
			21172: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjExNzIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIyOGJjZGU4Ny00NzBmLTRlNTEtYjJlYS1mZTk5YTlkZTEyOGMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NTIsImlhdCI6MTcwMTg0OTc1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzUyfQ.jguN8vthU4Gjkd0OOeEgVFvHuhOihowq91-jXBXSYcS6wBofvNIVh5qqMVlj1HOd4sfV6BZpg0TU1SI4v92ApBjOJMZOw2Ah9TyExOBJYQ0iSw8Aa8fTByMhtBBNj_w2HyV9TwJfkKsMSZ6a4oJB4ZPGZDizXvEiMWYWkpTrEBTHbzjEIBUnIEZm2QQEVlODdIuKYyTlRTnVoBXkVSBUIaEOrpfnRr0AWKISQGToU3mMVNHkny9LbW47pKv0U3VK8p9x-5yqlzMDL4P7GgSdHIb_hL1BahUAEPcnDaOjzhw6WVUTrp77poXnaPLF1BOEirdb6hfsw8fAlmQAm4113Q",
		},
		13102: {
			13102: "eyJhbGciOiJSUzI1NiIsImtpZCI6MTMxMDIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI5MDFhMzA1Mi1mMmRjLTQwZjYtYmM4Ny1lZWY1ZmFlYjVmNTciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NzcsImlhdCI6MTcwMTg0OTc3NywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5Nzc3fQ.WU0Cd_wtwwyZJn-cyASqKe1yY_BYrRIltHTiZAL5zLo2e7gtLvp2uEKGI93oQGONGBKhd-_oTM_o8M6VqK0-4l_Sx600DOYkJlVzUngJ1GvIcWwL6AMHClPtA-t_lvpLKXqJwznD311bOEWciOrVXlK973M1stWuIdm-UTPr5xeP-4hRj_wLVfpMlAjQQIoJCQIq1tfg5GsUxO6YXT0wC0tEg5CeZHgCx7_bvqnbECKvQqOzb4NkfEqqc16TLNbjZvvKfvP8FrLbFlKtKe7RPI_oC2pzh6iAe3CG1ZLexArosfKKza0QqjTm7OvNA1Q0kRhDorzUotRSE3JHch6tvA",
		},
		9478: {
			9478: "eyJhbGciOiJSUzI1NiIsImtpZCI6OTQ3OCwidHlwIjoiSldUIn0.eyJqdGkiOiJmNGIwZGM2YS1mMWJjLTRkN2UtYTA0NS1iNzEzMmFkZTZlMGEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3OTIsImlhdCI6MTcwMTg0OTc5MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzkyfQ.aBP5XNAYbKFmx5DRAwTauglNd9A-JiimEb-EO_kqgHkjZA2-Jfte69TJAp1otFA-dBVrzcR4Kq1k5U0p7oJP4Q7Z1eS9ViXsSWxJAw8DjzLBztD89UNrIeSgEG7TRzL8CdC_51gudHbgaBYM0oyT_ul1qintktTNicLYjVp0a2x68ZvpJxiyq1jY2kTC_0lBk-ONAPxfSF6wU_wWY0c-XFLmgjto99NCPxemnCV33sZ7z2Fg533u04DaXGKYmX0dP0gGaoFQRQ6AG5H-C4pr8v7oe830zoiAOderv93l7AQz-R9n8FYw0YlGQocU05ZzP7aFgsG2uxR7BYOOF09xPg",
		},
		20433: {
			20433: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjA0MzMsInR5cCI6IkpXVCJ9.eyJqdGkiOiIxYmJmODI5MC03NDIzLTQ1NjEtYTY3My02NWY4NmJiYjlkMzMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc4MTAsImlhdCI6MTcwMTg0OTgxMCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5ODEwfQ.TJ3mEjhnWYHlPRAjvcH9rTMpjVKIWL_i-hQtubyCkH5WPuOo_DdLQKeiiuJD7jIp7AtzvkX21zTNQWwMsWaExWB3yhNbgHp92KWguUB1EBAfg7t0dDa0uBl6hgbCgV97Z4gM53BzbNySt9Vjm88meiobXUnpBy345B-XUZsLTaHdBRwrnueijkVhlALjfj96Xlm-QAPyN0JFNaLCAnWGGmO_dd9t0vlecMUuxkwfjnpF3zGVfsMSA5xwAPkbFKrt5kjEE0Wyngbe8UMhl5SWORb9yAiK7ubIXU-IjYghYBKFj0VjgoG2bmpZXHJuUx8Dr9HlFPwNl0IGZgIJMeri2Q",
		},
	}
	tokensEdByKid = []map[int]string{
		0: {
			22899: "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIzY2E3YWExZC05NjViLTRjZTQtODljZS1lYjEwY2I4YzdlYTkiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNDMsImlhdCI6MTcwMTg1MDA0MywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDQzfQ.WAKMeS8VYZ-v9FEzjs2yRo7ZddtdDCw9yuZBEVddca8dzCM6sNyWbjKzFQXFf4S6Dgu1stH3ZjKtOfdZ_9usBQ",
			14211: "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIxMTY4MDcwMi01ZGIzLTQyOTUtYjVjOS1iZjc4NWFlMzk2MWEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNTUsImlhdCI6MTcwMTg1MDA1NSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDU1fQ.Q6mdarlwtBdBfXIu-YbNCAZEsrkxjbwUgs4XBczrB1xa7V3Pj0QLL5gLxXsOOt_rPs10nPoYLKoXexdnh54jDQ",
			5539:  "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI3YjY1OTgzMS1jZTVkLTRjOTYtYWVmZC02NjBlNmIxMjhkZDEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNjUsImlhdCI6MTcwMTg1MDA2NSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDY1fQ.4E94wv9NALpuA2sTaduw5_WJ2aGtqqig5TyXzzxYEiVMtj9lcxu5PWMsdRQLG49MMEmpUyKcjXrNwHUabzIICg",
			826:   "eyJhbGciOiJFZERTQSIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI0ZDVhNjViOS0wZmExLTQyZTYtYTY4MC1lZWZmNWJjOTFlNWIiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgwNzYsImlhdCI6MTcwMTg1MDA3NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMDc2fQ.5W3W9XIcL8_k8Uc40SLckUIRux7WqsVBRXqYzyVepsXIBvrJ6AXixU9t_EwSW_SYsjD_r0lpUiVx7hG6EQOGDQ",
		},
		22899: {
			22899: "eyJhbGciOiJFZERTQSIsImtpZCI6MjI4OTksInR5cCI6IkpXVCJ9.eyJqdGkiOiJmNmRiMTAxMy1jM2I4LTRhMDQtOTMzNS04ZmEzNjdjYTBiNTMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1MTcsImlhdCI6MTcwMTg0OTUxNywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTE3fQ.J8KAOd6gFcpusxJoREnrxyqcQ68IvWa8haGaAwExnUbJBL7QJj8i5KC4jhZqq097QB3yMWICGH02yl4o0WUnCQ",
		},
		14211: {
			14211: "eyJhbGciOiJFZERTQSIsImtpZCI6MTQyMTEsInR5cCI6IkpXVCJ9.eyJqdGkiOiJmY2M0MThiMi0zNDBiLTRiZmYtOGJiYi03NDczOWMxMjU5MDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NDEsImlhdCI6MTcwMTg0OTU0MSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTQxfQ.I6_FN8Ei5niyIQq6IKxf4XhHxWSxXWYaOmYSkSlZlDy-8-8rA5hm3c3Djn9ROP9VcKKM7UMbP3DMOK3wCg8xCg",
		},
		5539: {
			5539: "eyJhbGciOiJFZERTQSIsImtpZCI6NTUzOSwidHlwIjoiSldUIn0.eyJqdGkiOiJkMzQ4NTJiNy1lZDQzLTQ5MjItODU3MS01YTljZDg3ZTEzYmMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NTIsImlhdCI6MTcwMTg0OTU1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTUyfQ.EkLoISjqnq-sgQOqtY4J58g0fiyVZHEI_7Gl9FNcffsKVLWbRyOzdV1pedu3DXNmG4nU8cJ4_eyT_M64D4Z_BQ",
		},
		826: {
			826: "eyJhbGciOiJFZERTQSIsImtpZCI6ODI2LCJ0eXAiOiJKV1QifQ.eyJqdGkiOiJlNzU0YWRmOS0wZDYxLTRlYTMtOTdjNS1jMDhhODVlOTQwMDMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc1NzYsImlhdCI6MTcwMTg0OTU3NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NTc2fQ._z9Aqj-OxsRvK-ugQ-1s_fNElN0-A21Xm2NW5vGeBNEsxLvnOhr4fngQaa19JIZKBhn1smO1RVRJY94xYU7aBQ",
		},
	}
	tokensRSAByKid = []map[int]string{
		0: {
			21172: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI3ODY4YjExNy02MGM1LTQ5MmQtOWJiNC05ZTE3NjgzOTlhMDkiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNTYsImlhdCI6MTcwMTg1MDE1NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTU2fQ.IeaA4Kso34l_wOp-43BKQgq4r-iFdS1EGcux6eQxxR0ZVn6WuneCeFvIGS-2FeeCFr9AKg6fO-5IxKR8aM0SMYwLthTQWrAdpRhGlsISO12U-Zn1Y8GYTadLPlEIp-9B6cFf3fyRzPmxYbb700pw6mtzjxr0_k-hRl7f3pVGieI4qQNlPoL1D76CwnkJtaxrkJ4ieu65_PWmk-qQpLq-td03rx91pbXiQvW48y-ra6wKvZVzviiRi8GI_AY89QZk0TBtyzasf5glqQxG8PC6fnRBFQviQV2fsH4VrBmCZxuBOIks647JEDDMg2HGAfSZfl-qA2A0nLta7BAx0K6aBQ",
			13102: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiJmZjkyNzE3ZS1lOThiLTQ0ZTgtYTUxZi1lMjg5MDAzMmM1OWMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNjYsImlhdCI6MTcwMTg1MDE2NiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTY2fQ.FZHZXPcHEDqTQ4mGxGbz1KXNFWuD2OYYZiJbSTyjidU1XAacx1qTMQvKF9aqaYjPEjNDw5LpLXHHIsBPQ6VCkYnXXsCFC4tCaQ57DLNauqGO4O7grqM3-ZspZzzpCkwkGrXbmolw0pPzz7DjaYLlU2ZVfgNsiOh0nEFgFqj7FryTzLHvfuD8K1RdZ28v5HTzXDF9mIUCvzcDQharV4w35ShN4VzsN_vtjM7fuq9eU643YPlf96CJcM8gUj7op_BQuUNpaPqthfyRAEzHTK7AIRWlQQ_h4dWF5jDuRDkH5pyyPb82yqM7975nupvneGxGBqin9KklgDVcNGZxgogeNA",
			9478:  "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiIwZmIwNDQ5NS1mMDA1LTQ5MGItOGMyZC04OTFkY2IyYzEwMGQiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxNzQsImlhdCI6MTcwMTg1MDE3NCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTc0fQ.MJuiU7p0aXjponWLzPLzyoLlBh1Z5cyB7vBFkFecT3o9LP2HRkAwtsLo1d799gAnIvf7wE6cuK-Mq9SmC5OfnTWhPldbFa4W2WkiqJ6ptTxnIL_V0puxwrQx3k2sibTznsNsSL0FrFUMnHI4IxhtED4XRV9B9oHor-s3Fo_6m9xtiFwqTbqxdze_QBHXV8R4ixu_w_rGnwKncVEKqJ6E6aduIbX22JsAs-5EprmNldviFFglZoEzwlThmCl2W0WBO1EJ66PvwwM9i0iMnDgdZiPWPF4D861hFHCesds7uEi9ZYZUt_d0hEmI91O34PSpJdxFFsqNMpX1hpxiI8nN4Q",
			20433: "eyJhbGciOiJSUzI1NiIsImtpZCI6MCwidHlwIjoiSldUIn0.eyJqdGkiOiI0NGZlZGRkZi00MzI1LTQxOGUtYmNjNS00MjQyYTE0MDk0NjEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTgxODEsImlhdCI6MTcwMTg1MDE4MSwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODUwMTgxfQ.c9p-WRp36LPRqt1nzUQ27n3FCR5ue0XGZtrnGDIsSC3KtlkzUhSDJUlxn4moBAZA9R3Yp-neap7pN-SWvb9xnq--ui9b3DPxo3tn-CAs9_mZ6UJ2y51NzjyLpJe4jG1mkmZVxsCxsSIjvIFoIPtnbXEx4zZ2LxB6Mu0eSf5_j_S4ZO-1Iw93e15o1_v0IZKAcTLiBJR41X09MEBRbdOU-sRXmQvjM2UB6Kv47lliofkYinQ_T3LXrRWxYzW5iKeeZR2AO2olRLPiwWO1zR3tfZzow67XfNIHrmirUnghDrzSFdUGdl7JHAIkUWWske7ct1PEhXMuYepXQXqFcEPz-Q",
		},
		21172: {
			21172: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjExNzIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIyOGJjZGU4Ny00NzBmLTRlNTEtYjJlYS1mZTk5YTlkZTEyOGMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NTIsImlhdCI6MTcwMTg0OTc1MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzUyfQ.jguN8vthU4Gjkd0OOeEgVFvHuhOihowq91-jXBXSYcS6wBofvNIVh5qqMVlj1HOd4sfV6BZpg0TU1SI4v92ApBjOJMZOw2Ah9TyExOBJYQ0iSw8Aa8fTByMhtBBNj_w2HyV9TwJfkKsMSZ6a4oJB4ZPGZDizXvEiMWYWkpTrEBTHbzjEIBUnIEZm2QQEVlODdIuKYyTlRTnVoBXkVSBUIaEOrpfnRr0AWKISQGToU3mMVNHkny9LbW47pKv0U3VK8p9x-5yqlzMDL4P7GgSdHIb_hL1BahUAEPcnDaOjzhw6WVUTrp77poXnaPLF1BOEirdb6hfsw8fAlmQAm4113Q",
		},
		13102: {
			13102: "eyJhbGciOiJSUzI1NiIsImtpZCI6MTMxMDIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI5MDFhMzA1Mi1mMmRjLTQwZjYtYmM4Ny1lZWY1ZmFlYjVmNTciLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3NzcsImlhdCI6MTcwMTg0OTc3NywibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5Nzc3fQ.WU0Cd_wtwwyZJn-cyASqKe1yY_BYrRIltHTiZAL5zLo2e7gtLvp2uEKGI93oQGONGBKhd-_oTM_o8M6VqK0-4l_Sx600DOYkJlVzUngJ1GvIcWwL6AMHClPtA-t_lvpLKXqJwznD311bOEWciOrVXlK973M1stWuIdm-UTPr5xeP-4hRj_wLVfpMlAjQQIoJCQIq1tfg5GsUxO6YXT0wC0tEg5CeZHgCx7_bvqnbECKvQqOzb4NkfEqqc16TLNbjZvvKfvP8FrLbFlKtKe7RPI_oC2pzh6iAe3CG1ZLexArosfKKza0QqjTm7OvNA1Q0kRhDorzUotRSE3JHch6tvA",
		},
		9478: {
			9478: "eyJhbGciOiJSUzI1NiIsImtpZCI6OTQ3OCwidHlwIjoiSldUIn0.eyJqdGkiOiJmNGIwZGM2YS1mMWJjLTRkN2UtYTA0NS1iNzEzMmFkZTZlMGEiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc3OTIsImlhdCI6MTcwMTg0OTc5MiwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5NzkyfQ.aBP5XNAYbKFmx5DRAwTauglNd9A-JiimEb-EO_kqgHkjZA2-Jfte69TJAp1otFA-dBVrzcR4Kq1k5U0p7oJP4Q7Z1eS9ViXsSWxJAw8DjzLBztD89UNrIeSgEG7TRzL8CdC_51gudHbgaBYM0oyT_ul1qintktTNicLYjVp0a2x68ZvpJxiyq1jY2kTC_0lBk-ONAPxfSF6wU_wWY0c-XFLmgjto99NCPxemnCV33sZ7z2Fg533u04DaXGKYmX0dP0gGaoFQRQ6AG5H-C4pr8v7oe830zoiAOderv93l7AQz-R9n8FYw0YlGQocU05ZzP7aFgsG2uxR7BYOOF09xPg",
		},
		20433: {
			20433: "eyJhbGciOiJSUzI1NiIsImtpZCI6MjA0MzMsInR5cCI6IkpXVCJ9.eyJqdGkiOiIxYmJmODI5MC03NDIzLTQ1NjEtYTY3My02NWY4NmJiYjlkMzMiLCJzdWIiOiI3OGQyN2ViMS02Y2FiLTQ0ZGMtODc5Yi1jZTdlZTYxMzg1ZmUiLCJleHAiOjU3Mzg0NTc4MTAsImlhdCI6MTcwMTg0OTgxMCwibWVuZGVyLnRlbmFudCI6IjVhYmNiNmRlN2E2NzNhMDAwMTI4N2M3MSIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJtZW5kZXIudXNlcmFkbSIsInNjcCI6Im1lbmRlci4qIiwibmJmIjoxNzAxODQ5ODEwfQ.TJ3mEjhnWYHlPRAjvcH9rTMpjVKIWL_i-hQtubyCkH5WPuOo_DdLQKeiiuJD7jIp7AtzvkX21zTNQWwMsWaExWB3yhNbgHp92KWguUB1EBAfg7t0dDa0uBl6hgbCgV97Z4gM53BzbNySt9Vjm88meiobXUnpBy345B-XUZsLTaHdBRwrnueijkVhlALjfj96Xlm-QAPyN0JFNaLCAnWGGmO_dd9t0vlecMUuxkwfjnpF3zGVfsMSA5xwAPkbFKrt5kjEE0Wyngbe8UMhl5SWORb9yAiK7ubIXU-IjYghYBKFj0VjgoG2bmpZXHJuUx8Dr9HlFPwNl0IGZgIJMeri2Q",
		},
	}
)

func TestUserAdmMultipleKeys(t *testing.T) {
	testCases := map[string]struct {
		inEmail      model.Email
		inPassword   string
		dbUser       *model.User
		keyIds       []int
		defaultKeyId int
	}{
		"ok one key": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			keyIds:       []int{21172},
			defaultKeyId: 13102,
		},
		"ok all keys": {
			inEmail:    "foo@bar.com",
			inPassword: "correcthorsebatterystaple",

			dbUser: &model.User{
				ID:       oid.NewUUIDv5("1234").String(),
				Email:    "foo@bar.com",
				Password: `$2a$10$wMW4kC6o1fY87DokgO.lDektJO7hBXydf4B.yIWmE8hR9jOiO8way`,
			},
			keyIds: []int{
				22899,
				14211,
				5539,
				826,
				21172,
				13102,
				9478,
				20433,
			},
			defaultKeyId: 826,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			config := Config{
				Issuer:                         "mender-units",
				ExpirationTimeSeconds:          32,
				LimitSessionsPerUser:           10,
				LimitTokensPerUser:             10,
				TokenLastUsedUpdateFreqMinutes: 10,
				PrivateKeyPath:                 "testdata/private.id." + strconv.Itoa(tc.keyIds[0]) + ".pem",
				PrivateKeyFileNamePattern:      "private\\.id\\.([0-9]*)\\.pem",
			}
			handlersByKeyId := make(map[int]jwt.Handler, len(tc.keyIds)+1)
			for i := 0; i < len(tc.keyIds); i++ {
				handlersByKeyId[tc.keyIds[i]], _ = jwt.NewJWTHandler(
					"testdata/private.id."+strconv.Itoa(tc.keyIds[i])+".pem",
					"private\\.id\\.([0-9]*)\\.pem",
				)
			}
			handlersByKeyId[common.KeyIdZero], _ = jwt.NewJWTHandler(
				"testdata/private-"+strconv.Itoa(tc.defaultKeyId)+".pem",
				"private\\.id\\.([0-9]*)\\.pem",
			)

			db := &mstore.DataStore{}
			db.On("GetUserByEmail", ContextMatcher(), tc.inEmail).Return(tc.dbUser, nil)
			db.On("GetUserById", ContextMatcher(), mock.AnythingOfType("string")).Return(tc.dbUser, nil)

			db.On("SaveToken", ContextMatcher(), mock.AnythingOfType("*jwt.Token")).Return(nil)
			db.On("EnsureSessionTokensLimit", ContextMatcher(), mock.AnythingOfType("oid.ObjectID"),
				mock.AnythingOfType("int")).Return(nil)
			db.On("UpdateLoginTs", ContextMatcher(), tc.dbUser.ID).
				Return(nil)

			useradm := NewUserAdm(handlersByKeyId, db, config)
			cTenant := &mct.ClientRunner{}
			cTenant.On("GetTenant", ContextMatcher(), string(tc.inEmail), &apiclient.HttpApi{}).
				Return(&ct.Tenant{
					ID:     "5abcb6de7a673a0001287c71",
					Name:   "tenant1",
					Status: "active",
				}, nil)
			useradm = useradm.WithTenantVerification(cTenant)
			loginToken, err := useradm.Login(ctx, tc.inEmail, tc.inPassword, &LoginOptions{})
			db.On("GetTokenById", ContextMatcher(), mock.AnythingOfType("oid.ObjectID")).
				Return(loginToken, nil)

			signed, err := useradm.SignToken(ctx, loginToken)

			assert.NoError(t, err)
			token, err := handlersByKeyId[tc.keyIds[0]].FromJWT(signed)
			err = useradm.Verify(ctx, token)
			assert.NoError(t, err)

			// the default key (as loaded from testdata/private-tc.defaultKeyId.pem above
			// and assigned to KeyIdZero) is the one used to verify
			// the tokens with no kid or with kid equal 0
			token, err = handlersByKeyId[0].FromJWT(tokensByKeyIdNoKid[tc.defaultKeyId])
			err = useradm.Verify(ctx, token)
			assert.NoError(t, err)

			token, err = handlersByKeyId[0].FromJWT(tokensByKidAll[0][tc.defaultKeyId])
			err = useradm.Verify(ctx, token)
			assert.NoError(t, err)

			for i := range tc.keyIds {
				// all the tokens signed by given key and carrying the kid should be valid
				token, err = handlersByKeyId[tc.keyIds[i]].FromJWT(tokensByKid[tc.keyIds[i]][tc.keyIds[i]])
				err = useradm.Verify(ctx, token)
				assert.NoError(t, err)
			}

			for i := range tc.keyIds {
				for j := range tc.keyIds {
					if i == j {
						continue
					}
					// tokens signed by different keys cant verify
					token, err = handlersByKeyId[tc.keyIds[i]].FromJWT(tokensByKid[tc.keyIds[j]][tc.keyIds[j]])
					err = useradm.Verify(ctx, token)
					assert.Error(t, err)
				}
			}

			for i := range tc.keyIds {
				// tokens with no kid are assumed to be signed by the default key hence cant verify
				token, err = handlersByKeyId[tc.keyIds[i]].FromJWT(tokensByKeyIdNoKid[tc.keyIds[i]])
				err = useradm.Verify(ctx, token)
				assert.Error(t, err)
			}
		})
	}
}
