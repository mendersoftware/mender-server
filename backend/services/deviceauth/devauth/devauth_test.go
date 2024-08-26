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

package devauth

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/mendersoftware/mender-server/pkg/ratelimits"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	mcache "github.com/mendersoftware/mender-server/services/deviceauth/cache/mocks"
	minv "github.com/mendersoftware/mender-server/services/deviceauth/client/inventory/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator"
	morchestrator "github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/tenant"
	mtenant "github.com/mendersoftware/mender-server/services/deviceauth/client/tenant/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	mjwt "github.com/mendersoftware/mender-server/services/deviceauth/jwt/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	mstore "github.com/mendersoftware/mender-server/services/deviceauth/store/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
	mtesting "github.com/mendersoftware/mender-server/services/deviceauth/utils/testing"
	uto "github.com/mendersoftware/mender-server/services/deviceauth/utils/to"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		MultiTenant bool

		DataStoreError error
		InventoryError error
		WorkflowsError error
		TenantAdmError error
	}{{
		Name: "ok",
	}, {
		Name:        "ok, multitenant",
		MultiTenant: true,
	}, {
		Name:           "error, datastore",
		DataStoreError: errors.New("connection error"),
	}, {
		Name:           "error, inventory",
		InventoryError: errors.New("connection error"),
	}, {
		Name:           "error, workflows",
		WorkflowsError: errors.New("connection error"),
	}, {
		Name:           "error, tenantadm",
		MultiTenant:    true,
		TenantAdmError: errors.New("connection error"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
			defer cancel()

			db := &mstore.DataStore{}
			ta := &mtenant.ClientRunner{}
			wf := &morchestrator.ClientRunner{}
			inv := &minv.Client{}
			devauth := NewDevAuth(db, wf, nil, Config{})
			devauth.invClient = inv
			switch {
			default:
				fallthrough
			case tc.TenantAdmError != nil:
				if tc.MultiTenant {
					ta.On("CheckHealth", ctx).
						Return(tc.TenantAdmError)
					devauth.WithTenantVerification(ta)
				}
				fallthrough
			case tc.WorkflowsError != nil:
				wf.On("CheckHealth", ctx).
					Return(tc.WorkflowsError)
				fallthrough
			case tc.InventoryError != nil:
				inv.On("CheckHealth", ctx).
					Return(tc.InventoryError)
				fallthrough
			case tc.DataStoreError != nil:
				db.On("Ping", ctx).
					Return(tc.DataStoreError)
			}

			err := devauth.HealthCheck(ctx)
			switch {
			case tc.DataStoreError != nil:
				assert.EqualError(t, err,
					"error reaching MongoDB: "+
						tc.DataStoreError.Error(),
				)
			case tc.InventoryError != nil:
				assert.EqualError(t, err,
					"Inventory service unhealthy: "+
						tc.InventoryError.Error(),
				)
			case tc.WorkflowsError != nil:
				assert.EqualError(t, err,
					"Workflows service unhealthy: "+
						tc.WorkflowsError.Error(),
				)
			case tc.TenantAdmError != nil:
				assert.EqualError(t, err,
					"Tenantadm service unhealthy: "+
						tc.TenantAdmError.Error(),
				)
			default:
				assert.NoError(t, err)
			}
			db.AssertExpectations(t)
			inv.AssertExpectations(t)
			wf.AssertExpectations(t)
			ta.AssertExpectations(t)
		})
	}
}

func TestDevAuthSubmitAuthRequest(t *testing.T) {
	t.Parallel()

	pubKey := "dummy_pubkey"
	idData := "{\"mac\":\"00:00:00:01\"}"
	devId := oid.NewUUIDv4().String()
	authId := oid.NewUUIDv4().String()

	idDataStruct, idDataHash, err := parseIdData(idData)
	assert.NoError(t, err)

	req := model.AuthReq{
		IdData:      idData,
		TenantToken: "tenant",
		PubKey:      pubKey,
	}

	badReq := model.AuthReq{
		IdData:      "a",
		TenantToken: "tenant",
		PubKey:      pubKey,
	}

	testCases := []struct {
		desc string

		inReq model.AuthReq

		trial bool

		devStatus string

		// key of returned device
		getDevByIdKey string
		getDevByIdErr error

		//id of returned device
		getDevByKeyId string
		getAuthSetErr error

		addDeviceErr  error
		addAuthSetErr error

		devAdmErr error

		tenantVerify                 bool
		tenantVerificationErr        error
		tenantVerificationDefaultErr error

		config Config

		updateDeviceInventory bool
		updateDeviceStatus    bool

		res string
		err error
	}{
		{
			// pretend we failed to add device to DB
			desc: "db add fail",

			inReq:        req,
			addDeviceErr: errors.New("failed"),
			err:          errors.New("failed"),
		},
		{
			//existing device, existing auth set, auth set accepted,
			//so we should get a token right away
			desc: "known, accepted, give out token",

			inReq: req,

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			devStatus:     model.DevStatusAccepted,
			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			updateDeviceStatus:    true,
			updateDeviceInventory: true,

			res: "dummytoken",
		},
		{
			//existing device, existing auth set, auth set rejected,
			//no token
			desc: "known, rejected",

			inReq: req,

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			devStatus:     model.DevStatusRejected,
			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			updateDeviceStatus: true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//existing, pending device
			desc: "known, pending",

			inReq: req,

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			devStatus:     model.DevStatusPending,
			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			updateDeviceStatus: true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//new device
			desc: "new device",

			inReq: req,

			devStatus: model.DevStatusPending,

			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			updateDeviceInventory: true,
			updateDeviceStatus:    true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//known device, adding returns that device exists, but
			//trying to fetch it fails
			desc: "known, device fetch data fail",

			inReq: req,

			addDeviceErr: store.ErrObjectExists,

			getDevByIdErr: store.ErrDevNotFound,

			err: errors.New("failed to locate device"),
		},
		{
			//known device and auth set, but fetching auth set fails
			desc: "known, auth set fetch fail",

			inReq: req,

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			getAuthSetErr: store.ErrAuthSetNotFound,

			updateDeviceStatus: true,

			err: errors.New("failed to locate device auth set"),
		},
		{
			//new device - tenant token verification failed
			desc: "new device, tenant token verification fail",

			inReq: req,

			err: errors.New("dev auth: unauthorized: tenant token verification failed: account suspended"),

			tenantVerify:          true,
			tenantVerificationErr: errors.New("tenant token verification failed: account suspended"),
		},
		{
			//new device - valid default tenant token present, but given tenant token verification failed
			desc: "new device, valid default tenant token present, but given tenant token verification fail",

			inReq: req,

			config: Config{
				// token with the following claims:
				//   {
				//      "sub": "bogusdevice",
				//      "mender.tenant": "foobar"
				//   }
				DefaultTenantToken: "fake.eyJzdWIiOiJib2d1c2RldmljZSIsIm1lbmRlci50ZW5hbnQiOiJmb29iYXIifQ.fake",
			},

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			tenantVerify:          true,
			tenantVerificationErr: errors.New("tenant not found"),

			updateDeviceStatus: true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//new device - both given and default tenant token verification fail
			desc: "new device, both given and default tenant token verification fail",

			inReq: req,

			config: Config{
				DefaultTenantToken: "bogustoken",
			},

			err: errors.New("dev auth: unauthorized: tenant token verification failed: account suspended"),

			tenantVerify:                 true,
			tenantVerificationErr:        errors.New("tenant token verification failed: account suspended"),
			tenantVerificationDefaultErr: errors.New("tenant token verification failed: account suspended"),
		},
		{
			//new device - tenant token verification failed because of other reasons
			desc: "new device, tenant token other fail",

			inReq: req,

			err: errors.New("request to verify tenant token failed: something something failed"),

			tenantVerify:          true,
			tenantVerificationErr: errors.New("something something failed"),
		},
		{
			//new device - tenant token required but not provided
			desc: "new device, missing but required tenant token",

			inReq: model.AuthReq{
				IdData:      idData,
				TenantToken: "",
				PubKey:      pubKey,
			},

			err: MakeErrDevAuthUnauthorized(errors.New("tenant token missing")),

			tenantVerify:          true,
			tenantVerificationErr: errors.New("should not be called"),
		},
		{
			// a known device with a correct tenant token
			desc: "known, correct tenant token",

			inReq: model.AuthReq{
				IdData: idData,
				// token with the following claims:
				//   {
				//      "sub": "bogusdevice",
				//      "mender.tenant": "foobar"
				//   }
				TenantToken: "fake.eyJzdWIiOiJib2d1c2RldmljZSIsIm1lbmRlci50ZW5hbnQiOiJmb29iYXIifQ.fake",
				PubKey:      pubKey,
			},

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			tenantVerify:       true,
			updateDeviceStatus: true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//new device - tenant token optional, and not provided
			desc: "new device, missing and optional tenant token",

			inReq: model.AuthReq{
				IdData: idData,
				PubKey: pubKey,
			},

			config: Config{
				// token with the following claims:
				//   {
				//      "sub": "bogusdevice",
				//      "mender.tenant": "foobar"
				//   }
				DefaultTenantToken: "fake.eyJzdWIiOiJib2d1c2RldmljZSIsIm1lbmRlci50ZW5hbnQiOiJmb29iYXIifQ.fake",
			},

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			tenantVerify:       true,
			updateDeviceStatus: true,

			err: ErrDevAuthUnauthorized,
		},
		{
			//identity data malformed
			desc: "identity datat malformed",

			inReq: badReq,

			devStatus: model.DevStatusPending,

			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			err: errors.New("dev auth: bad request: failed to parse identity data: a: invalid character 'a' looking for beginning of value"),
		},
		{
			// a known device of with a correct tenant token, hand
			// out a token with tenant claim in it
			desc: "known, accepted, tenant, give out token with tenant claim",

			inReq: model.AuthReq{
				IdData: idData,
				// token with the following claims:
				//   {
				//      "sub": "bogusdevice",
				//      "mender.tenant": "foobar"
				//   }
				TenantToken: "fake.eyJzdWIiOiJib2d1c2RldmljZSIsIm1lbmRlci50ZW5hbnQiOiJmb29iYXIifQ.fake",
				PubKey:      pubKey,
			},

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			tenantVerify:          true,
			updateDeviceStatus:    true,
			updateDeviceInventory: true,

			devStatus: model.DevStatusAccepted,

			res: "dummytoken",
		},
		{
			// a known device of with a correct tenant token, hand
			// out a token with tenant claim in it
			desc: "known, accepted, tenant, give out token with tenant claim, trial",

			inReq: model.AuthReq{
				IdData: idData,
				// token with the following claims:
				//   {
				//      "sub": "bogusdevice",
				//      "mender.tenant": "foobar"
				//   }
				TenantToken: "fake.eyJzdWIiOiJib2d1c2RldmljZSIsIm1lbmRlci50ZW5hbnQiOiJmb29iYXIifQ.fake",
				PubKey:      pubKey,
			},

			trial: true,

			addDeviceErr:  store.ErrObjectExists,
			addAuthSetErr: store.ErrObjectExists,

			getDevByIdKey: pubKey,
			getDevByKeyId: devId,

			tenantVerify:          true,
			updateDeviceStatus:    true,
			updateDeviceInventory: true,

			devStatus: model.DevStatusAccepted,

			res: "dummytoken",
		},
	}

	for tcidx := range testCases {
		tc := testCases[tcidx]
		t.Run(fmt.Sprintf("tc: %s", tc.desc), func(t *testing.T) {
			t.Parallel()

			// match context in mocks
			ctxMatcher := mtesting.ContextMatcher()

			if tc.tenantVerify {
				// context must carry identity information if
				// tenant verification is enabled, also it must
				// be set up with http Authorization header to
				// use in outgoing requests (via
				// go-lib-micro/context/httpheader packaage)
				ctxMatcher = mock.MatchedBy(func(c context.Context) bool {
					return assert.NotNil(t, identity.FromContext(c))
				})
			}

			db := mstore.DataStore{}
			db.On("AddDevice",
				ctxMatcher,
				mock.MatchedBy(
					func(d model.Device) bool {
						return d.IdData == idData
					})).Return(tc.addDeviceErr)

			db.On("GetDeviceStatus", ctxMatcher,
				mock.AnythingOfType("string")).Return(
				"pending", nil)

			db.On("GetDeviceByIdentityDataHash",
				ctxMatcher,
				idDataHash).Return(
				func(ctx context.Context, idDataHash []byte) *model.Device {
					if tc.getDevByIdErr == nil {
						return &model.Device{
							IdDataSha256: idDataHash,
							IdDataStruct: idDataStruct,
							Id:           devId,
						}
					}
					return nil
				},
				tc.getDevByIdErr)
			db.On("AddAuthSet",
				ctxMatcher,
				mock.MatchedBy(
					func(m model.AuthSet) bool {
						return m.DeviceId == devId
					})).Return(tc.addAuthSetErr)
			db.On("RejectAuthSetsForDevice",
				ctxMatcher,
				tc.getDevByKeyId).
				Return(nil)
			db.On("GetAuthSetByIdDataHashKey",
				ctxMatcher,
				idDataHash,
				pubKey,
			).Return(
				func(ctx context.Context, idDataHash []byte, key string) *model.AuthSet {
					if tc.getAuthSetErr == nil {
						return &model.AuthSet{
							Id:           authId,
							DeviceId:     tc.getDevByKeyId,
							IdDataSha256: idDataHash,
							PubKey:       key,
							Status:       tc.devStatus,
						}
					}
					return nil
				},
				tc.getAuthSetErr)
			db.On("GetAuthSetByIdDataHashKeyByStatus",
				ctxMatcher,
				idDataHash,
				pubKey,
				model.DevStatusPreauth,
			).Return(
				func(ctx context.Context, idDataHash []byte, key string, status string) *model.AuthSet {
					if tc.getAuthSetErr == nil {
						return &model.AuthSet{
							Id:           authId,
							DeviceId:     tc.getDevByKeyId,
							IdDataSha256: idDataHash,
							PubKey:       key,
							Status:       tc.devStatus,
						}
					}
					return nil
				},
				tc.getAuthSetErr)
			db.On("AddToken",
				ctxMatcher,
				mock.AnythingOfType("*jwt.Token")).Return(nil)
			db.On("GetDeviceStatus", ctxMatcher,
				mock.AnythingOfType("string")).Return(
				"pending", nil)
			db.On("UpdateDevice", ctxMatcher,
				devId,
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)
			db.On("GetDeviceById", ctxMatcher,
				mock.AnythingOfType("string")).Return(&model.Device{}, nil)

			jwth := mjwt.Handler{}
			jwth.On("ToJWT",
				mock.MatchedBy(func(jt *jwt.Token) bool {
					t.Logf("token: %v", jt)
					devUUID := oid.FromString(devId)
					return assert.NotNil(t, jt) &&
						assert.Equal(t, devUUID, jt.Claims.Subject) &&
						(tc.tenantVerify == false ||
							assert.Equal(t, "foobar", jt.Claims.Tenant)) &&
						(tc.tenantVerify == false ||
							assert.Equal(t, tc.trial, jt.Claims.Trial))
				})).
				Return("dummytoken", nil)

			ctx := context.Background()
			id := &identity.Identity{
				Tenant: "foobar",
			}
			ctx = identity.WithContext(ctx, id)
			co := morchestrator.ClientRunner{}
			if tc.updateDeviceStatus {
				co.On("SubmitUpdateDeviceStatusJob", mock.Anything,
					mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
					Return(nil)
			}
			if tc.updateDeviceInventory {
				co.On("SubmitUpdateDeviceInventoryJob", ctxMatcher,
					mock.AnythingOfType("orchestrator.UpdateDeviceInventoryReq")).
					Return(nil)
			}

			devauth := NewDevAuth(&db, &co, &jwth, tc.config)

			if tc.tenantVerify {
				ct := mtenant.ClientRunner{}
				if tc.inReq.TenantToken != "" {
					ct.On("VerifyToken",
						mtesting.ContextMatcher(),
						tc.inReq.TenantToken).
						Return(
							&tenant.Tenant{ID: "foobar", Trial: tc.trial},
							tc.tenantVerificationErr)
				}
				if tc.config.DefaultTenantToken != "" {
					ct.On("VerifyToken",
						mtesting.ContextMatcher(),
						tc.config.DefaultTenantToken).
						Return(
							&tenant.Tenant{},
							tc.tenantVerificationDefaultErr)
				}
				devauth = devauth.WithTenantVerification(&ct)
			}

			res, err := devauth.SubmitAuthRequest(ctx, &tc.inReq)

			t.Logf("error: %v", err)
			assert.Equal(t, tc.res, res)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
			co.AssertExpectations(t)
		})
	}
}

// still a Submit... test, but focuses on preauth
func TestDevAuthSubmitAuthRequestPreauth(t *testing.T) {
	idData := "{\"mac\":\"00:00:00:01\"}"
	_, idDataSha256, err := parseIdData(idData)
	assert.NoError(t, err)

	inReq := model.AuthReq{
		IdData:      idData,
		PubKey:      "foo-pubkey",
		TenantToken: "foo-tenant",
	}

	dummyDevId := oid.NewUUIDv5("dummy_devid").String()
	dummyAuthID := oid.NewUUIDv5("dummy_aid").String()
	dummyToken := "dummytoken"

	testCases := []struct {
		desc string

		dbGetAuthSetByDataKeyRes *model.AuthSet
		dbGetAuthSetByDataKeyErr error

		dbGetLimitRes *model.Limit
		dbGetLimitErr error

		dbGetDevCountByStatusRes int
		dbGetDevCountByStatusErr error

		dev                *model.Device
		dbGetDeviceByIdErr error

		coSubmitProvisionDeviceJobErr error

		res string
		err error
	}{
		{
			desc: "ok: preauthorized set is auto-accepted",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 0,
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusPending,
			},
			res: dummyToken,
		},
		{
			desc: "ok: device during decommissioning",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 0,
			dev: &model.Device{
				Id:              dummyDevId,
				Status:          model.DevStatusPending,
				Decommissioning: true,
			},
			err: ErrDevAuthUnauthorized,
		},
		{
			desc:                     "error: can't get an existing authset",
			dbGetAuthSetByDataKeyErr: errors.New("db error"),
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusPending,
			},
			err: errors.New("failed to fetch auth set: db error"),
		},
		{
			desc: "error: preauthorized set would exceed limit",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 5,
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusPending,
			},
			err: ErrMaxDeviceCountReached,
		},
		{
			desc: "error: can't get device limit",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitErr: errors.New("db error"),
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusPending,
			},
			err: errors.New("can't get current device limit: db error"),
		},
		{
			desc: "error: failed to submit job to workflows",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 0,
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusPending,
			},
			coSubmitProvisionDeviceJobErr: errors.New("workflows failed"),
			err:                           errors.New("submit device provisioning job error: workflows failed"),
		},
		{
			desc: "ok: preauthorized set is auto-accepted, device was already accepted",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 0,
			dev: &model.Device{
				Id:     dummyDevId,
				Status: model.DevStatusAccepted,
			},
			//coSubmitProvisionDeviceJobErr: errors.New("workflows shouldn't be called"), // MEN-6961: we accept preauth at all times
			res: "dummytoken",
		},
		{
			desc: "error: cannot get device status",
			dbGetAuthSetByDataKeyRes: &model.AuthSet{
				Id:           dummyAuthID,
				IdDataSha256: idDataSha256,
				DeviceId:     dummyDevId,
				PubKey:       inReq.PubKey,
				Status:       model.DevStatusPreauth,
			},
			dbGetLimitRes: &model.Limit{
				Value: 5,
			},
			dbGetDevCountByStatusRes: 0,
			dbGetDeviceByIdErr:       errors.New("Get device failed"),
			err:                      errors.New("Get device failed"),
		},
	}

	for tcidx := range testCases {
		tc := testCases[tcidx]
		t.Run(fmt.Sprintf("tc: %s", tc.desc), func(t *testing.T) {
			t.Parallel()

			ctxMatcher := mtesting.ContextMatcher()

			// setup mocks
			db := mstore.DataStore{}

			// get the auth set to check if preauthorized
			db.On("GetAuthSetByIdDataHashKeyByStatus",
				ctxMatcher,
				idDataSha256,
				inReq.PubKey,
				model.DevStatusPreauth,
			).Return(
				tc.dbGetAuthSetByDataKeyRes,
				tc.dbGetAuthSetByDataKeyErr,
			)

			// for a preauthorized set - check if we're not over the limit
			db.On("GetLimit",
				ctxMatcher,
				model.LimitMaxDeviceCount,
			).Return(
				tc.dbGetLimitRes,
				tc.dbGetLimitErr,
			)

			// takes part in limit checking
			db.On("GetDevCountByStatus",
				ctxMatcher,
				model.DevStatusAccepted,
			).Return(
				tc.dbGetDevCountByStatusRes,
				tc.dbGetDevCountByStatusErr,
			)

			db.On("RejectAuthSetsForDevice",
				ctxMatcher,
				dummyDevId).
				Return(nil)

			// at the end of processing, updates the preauthorized set to 'accepted'
			// just happy path, errors tested elsewhere
			db.On("UpdateAuthSetById",
				ctxMatcher,
				mock.AnythingOfType("string"),
				mock.MatchedBy(
					func(u model.AuthSetUpdate) bool {
						return u.Status == model.DevStatusAccepted
					}),
			).Return(nil)

			// at the end of processing, updates the device status to 'accepted'
			db.On("UpdateDevice",
				ctxMatcher,
				dummyDevId,
				mock.MatchedBy(
					func(u model.DeviceUpdate) bool {
						return u.Status == model.DevStatusAccepted || !u.CheckInTime.IsZero()
					}),
			).Return(nil)

			// at the end of processing, saves the issued token
			// only happy path, errors tested elsewhere
			db.On("AddToken",
				ctxMatcher,
				mock.AnythingOfType("*jwt.Token"),
			).Return(nil)

			db.On("GetDeviceById",
				ctxMatcher, dummyDevId).Return(tc.dev, tc.dbGetDeviceByIdErr)

			db.On("GetDeviceStatus", ctxMatcher,
				mock.AnythingOfType("string")).Return(
				"pending", nil)
			// token serialization - happy path only, errors tested elsewhere
			jwth := mjwt.Handler{}
			jwth.On("ToJWT",
				mock.AnythingOfType("*jwt.Token"),
			).Return(dummyToken, nil)

			co := morchestrator.ClientRunner{}
			co.On("SubmitProvisionDeviceJob", ctxMatcher,
				mock.AnythingOfType("orchestrator.ProvisionDeviceReq")).
				Return(tc.coSubmitProvisionDeviceJobErr)
			co.On("SubmitUpdateDeviceStatusJob", ctxMatcher,
				mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
				Return(nil)
			co.On("SubmitUpdateDeviceInventoryJob", ctxMatcher,
				mock.AnythingOfType("orchestrator.UpdateDeviceInventoryReq")).
				Return(nil)

			// setup devauth
			devauth := NewDevAuth(&db, &co, &jwth, Config{})

			// test
			res, err := devauth.SubmitAuthRequest(context.Background(), &inReq)

			// verify
			assert.Equal(t, tc.res, res)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevAuthPreauthorizeDevice(t *testing.T) {
	t.Parallel()

	authsetID := oid.NewUUIDv5("aid").String()
	deviceID := oid.NewUUIDv5("did").String()
	idData := "{\"mac\":\"00:00:00:01\"}"
	pubKey := "pubkey"

	_, idDataSha256, err := parseIdData(idData)
	assert.NoError(t, err)

	req := &model.PreAuthReq{
		AuthSetId: authsetID,
		DeviceId:  deviceID,
		IdData:    idData,
		PubKey:    pubKey,
	}

	badReq := &model.PreAuthReq{
		AuthSetId: authsetID,
		DeviceId:  deviceID,
		IdData:    "a",
		PubKey:    pubKey,
	}

	testCases := []struct {
		desc string
		req  *model.PreAuthReq

		enableReporting bool

		addDeviceErr  error
		addAuthSetErr error
		getDevByIdErr error
		inventoryErr  error
		reportingErr  error

		updateDeviceStatus    bool
		updateDeviceInventory bool
		callDb                bool

		outDev *model.Device
		err    error
	}{
		{
			desc:                  "ok",
			req:                   req,
			updateDeviceStatus:    true,
			updateDeviceInventory: true,
			callDb:                true,
		},
		{
			desc:                  "ok with reporting",
			req:                   req,
			updateDeviceStatus:    true,
			updateDeviceInventory: true,
			enableReporting:       true,
			callDb:                true,
		},
		{
			desc:   "error: add device, exists",
			req:    req,
			callDb: true,

			addDeviceErr: store.ErrObjectExists,

			outDev: &model.Device{Id: deviceID},
			err:    ErrDeviceExists,
		},
		{
			desc:   "error: add device, generic",
			req:    req,
			callDb: true,

			addDeviceErr: errors.New("generic error"),

			err: errors.New("failed to add device: generic error"),
		},
		{
			desc:               "error: add auth set, exists",
			req:                req,
			updateDeviceStatus: true,
			callDb:             true,

			addAuthSetErr: store.ErrObjectExists,

			outDev: &model.Device{Id: deviceID},
			err:    ErrDeviceExists,
		},
		{
			desc:               "error: add auth set, exists",
			req:                req,
			updateDeviceStatus: true,
			callDb:             true,

			addAuthSetErr: errors.New("generic error"),

			err: errors.New("failed to add auth set: generic error"),
		},
		{
			desc: "error: identity data malformed",
			req:  badReq,

			err: errors.New("dev auth: bad request: failed to parse identity data: a: invalid character 'a' looking for beginning of value"),
		},
	}

	for tcidx := range testCases {
		tc := testCases[tcidx]
		t.Run(fmt.Sprintf("tc: %s", tc.desc), func(t *testing.T) {
			t.Parallel()

			ctxMatcher := mock.MatchedBy(func(c context.Context) bool {
				return true
			})

			db := mstore.DataStore{}
			if tc.callDb {
				db.On("AddDevice",
					ctxMatcher,
					mock.MatchedBy(
						func(d model.Device) bool {
							return (d.IdData == tc.req.IdData) &&
								(d.Id == tc.req.DeviceId)
						})).Return(tc.addDeviceErr)

				if tc.addDeviceErr == nil {
					db.On("AddAuthSet",
						ctxMatcher,
						mock.MatchedBy(
							func(m model.AuthSet) bool {
								return (m.Id == tc.req.AuthSetId) &&
									(m.DeviceId == tc.req.DeviceId) &&
									(m.IdData == tc.req.IdData) &&
									(m.PubKey == tc.req.PubKey)
							})).Return(tc.addAuthSetErr)
				}
			}

			co := morchestrator.ClientRunner{}
			defer co.AssertExpectations(t)

			if tc.updateDeviceStatus {
				co.On("SubmitUpdateDeviceStatusJob", ctxMatcher,
					mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
					Return(nil)
			}
			if tc.updateDeviceInventory {
				co.On("SubmitUpdateDeviceInventoryJob", ctxMatcher,
					mock.AnythingOfType("orchestrator.UpdateDeviceInventoryReq")).
					Return(nil)
			}
			if tc.enableReporting {
				co.On("SubmitReindexReporting",
					ctxMatcher,
					deviceID,
				).Return(nil).Once()
			}

			if tc.err == ErrDeviceExists {
				db.On("GetDeviceByIdentityDataHash",
					ctxMatcher,
					idDataSha256).Return(
					func(ctx context.Context, idDataHash []byte) *model.Device {
						if tc.getDevByIdErr == nil {
							return &model.Device{
								IdDataSha256: idDataSha256,
								Id:           deviceID,
							}
						}
						return nil
					},
					tc.getDevByIdErr)
			}
			devauth := NewDevAuth(&db, &co, nil, Config{
				EnableReporting: tc.enableReporting,
			})
			dev, err := devauth.PreauthorizeDevice(context.Background(), tc.req)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tc.outDev != nil {
				assert.Equal(t, tc.outDev.Id, dev.Id)
			} else {
				assert.Nil(t, dev)
			}
			co.AssertExpectations(t)
			db.AssertExpectations(t)
		})
	}
}

func TestDevAuthAcceptDevice(t *testing.T) {
	t.Parallel()

	dummyAuthID := oid.NewUUIDv5("dummy_aid").String()
	dummyDevID := oid.NewUUIDv5("dummy_devid").String()

	testCases := []struct {
		aset *model.AuthSet

		dbLimit    *model.Limit
		dbLimitErr error

		dbCount    int
		dbCountErr error

		dev                *model.Device
		dbGetDeviceByIdErr error

		dbUpdateErr               error
		dbUpdateRevokeAuthSetsErr error

		dbGetErr error

		coSubmitProvisionDeviceJobErr error

		outErr string
	}{
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit: &model.Limit{Value: 0},
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPreauth,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPreauth,
			},
			dbLimit: &model.Limit{Value: 0},
			outErr:  "dev auth: bad request",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit: &model.Limit{Value: 5},
			dbCount: 4,
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			coSubmitProvisionDeviceJobErr: errors.New("workflows shouldn't be called"),
			dbLimit:                       &model.Limit{Value: 5},
			dbCount:                       4,
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyAuthID,
				Status: model.DevStatusAccepted,
			},
			coSubmitProvisionDeviceJobErr: errors.New("workflows shouldn't be called"),
			dbLimit:                       &model.Limit{Value: 5},
			dbCount:                       4,
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit: &model.Limit{Value: 5},
			dbCount: 5,
			outErr:  "maximum number of accepted devices reached",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit: &model.Limit{Value: 5},
			dbCount: 6,
			outErr:  "maximum number of accepted devices reached",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit:    &model.Limit{Value: 5},
			dbLimitErr: errors.New("error"),
			outErr:     "can't get current device limit: error",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbLimit:    &model.Limit{Value: 5},
			dbCountErr: errors.New("error"),
			outErr:     "can't get current device count: error",
		},
		{
			dbLimit:  &model.Limit{Value: 0},
			dbGetErr: store.ErrAuthSetNotFound,
			outErr:   store.ErrAuthSetNotFound.Error(),
		},
		{
			dbLimit: &model.Limit{Value: 0},
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbUpdateErr: errors.New("failed to update device"),
			outErr:      "db update device auth set error: failed to update device",
		},
		{
			dbLimit: &model.Limit{Value: 0},
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			coSubmitProvisionDeviceJobErr: errors.New("workflows failed"),
			outErr:                        "submit device provisioning job error: workflows failed",
		},
		{
			dbLimit: &model.Limit{Value: 0},
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbUpdateRevokeAuthSetsErr: store.ErrAuthSetNotFound,
		},
		{
			dbLimit: &model.Limit{Value: 0},
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dev: &model.Device{
				Id:     dummyDevID,
				Status: model.DevStatusPending,
			},
			dbUpdateRevokeAuthSetsErr: errors.New("foobar"),
			outErr:                    "failed to reject auth sets: foobar",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPending,
			},
			dbLimit:            &model.Limit{Value: 5},
			dbCount:            4,
			dbGetDeviceByIdErr: errors.New("Get device failed"),
			outErr:             "Get device failed",
		},
	}

	for idx := range testCases {
		tc := testCases[idx]
		t.Run(fmt.Sprintf("tc %v", idx), func(t *testing.T) {
			//t.Parallel()
			defer func() { recover() }()

			db := mstore.DataStore{}
			db.On("GetAuthSetById", context.Background(),
				dummyAuthID).
				Return(tc.aset, tc.dbGetErr)
			db.On("GetLimit", context.Background(),
				model.LimitMaxDeviceCount).
				Return(tc.dbLimit, tc.dbLimitErr)
			db.On("GetDevCountByStatus", context.Background(),
				model.DevStatusAccepted).
				Return(tc.dbCount, tc.dbCountErr)
			db.On("GetDeviceById", context.Background(),
				dummyDevID).
				Return(tc.dev, tc.dbGetDeviceByIdErr)
			db.On("UpdateDevice", context.Background(),
				dummyDevID,
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)
			db.On("GetDeviceStatus", context.Background(),
				dummyDevID).Return(
				"accpted", nil)

			if tc.aset != nil {
				// for rejecting all auth sets
				db.On("RejectAuthSetsForDevice",
					context.Background(),
					dummyDevID).
					Return(tc.dbUpdateRevokeAuthSetsErr)
				// for accepting a single one
				db.On("UpdateAuthSetById", context.Background(),
					tc.aset.Id,
					model.AuthSetUpdate{
						Status: model.DevStatusAccepted,
					}).Return(tc.dbUpdateErr)
			}

			co := morchestrator.ClientRunner{}
			co.On("SubmitProvisionDeviceJob", context.Background(),
				mock.AnythingOfType("orchestrator.ProvisionDeviceReq")).
				Return(tc.coSubmitProvisionDeviceJobErr)
			co.On("SubmitUpdateDeviceStatusJob", context.Background(),
				mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
				Return(nil)

			devauth := NewDevAuth(&db, &co, nil, Config{})
			err := devauth.AcceptDeviceAuth(
				context.Background(), dummyDevID, dummyAuthID)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevAuthRejectDevice(t *testing.T) {
	t.Parallel()

	dummyAuthID := oid.NewUUIDv5("dummy_aid").String()
	dummyDevUUID := oid.NewUUIDv5("dummy_devid")
	dummyDevID := dummyDevUUID.String()

	testCases := []struct {
		aset *model.AuthSet

		tenant         string
		withCache      bool
		cacheDeleteErr error

		dbErr            error
		dbDelDevTokenErr error

		outErr string
	}{
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			withCache:      true,
			tenant:         "acme",
			cacheDeleteErr: errors.New("redis error"),
			outErr:         "failed to delete token for 9c5df658-26ff-55e1-87a1-6780ca473154 from cache: redis error",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			withCache: true,
			outErr:    "failed to delete token for 9c5df658-26ff-55e1-87a1-6780ca473154 from cache: can't unpack tenant identity data from context",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			withCache: true,
			tenant:    "acme",
		},
		{
			dbErr:            errors.New("failed"),
			dbDelDevTokenErr: nil,
			outErr:           "db get auth set error: failed",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			dbDelDevTokenErr: store.ErrTokenNotFound,
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusAccepted,
			},
			dbDelDevTokenErr: errors.New("some error"),
			outErr:           "db delete device token error: some error",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   model.DevStatusPreauth,
			},
			outErr: "dev auth: bad request",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			db := mstore.DataStore{}
			db.On("GetAuthSetById", ctx,
				dummyAuthID).
				Return(tc.aset, tc.dbErr)
			if tc.aset != nil {
				db.On("UpdateAuthSetById", ctx, tc.aset.Id,
					model.AuthSetUpdate{Status: model.DevStatusRejected}).Return(nil)
			}
			db.On("DeleteTokenByDevId", ctx,
				dummyDevUUID).
				Return(tc.dbDelDevTokenErr)
			db.On("GetDeviceStatus", ctx,
				dummyDevID).
				Return("accpted", nil)
			db.On("UpdateDevice", ctx,
				dummyDevID,
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)
			db.On("GetDeviceById", ctx,
				mock.AnythingOfType("string")).Return(&model.Device{}, nil)

			co := morchestrator.ClientRunner{}
			co.On("SubmitUpdateDeviceStatusJob", ctx,
				mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
				Return(nil)

			devauth := NewDevAuth(&db, &co, nil, Config{})

			c := &mcache.Cache{}
			if tc.withCache {
				devauth = devauth.WithCache(c)
				if tc.tenant != "" {
					c.On("DeleteToken",
						mock.MatchedBy(func(ctx context.Context) bool {
							ident := identity.FromContext(ctx)
							return assert.NotNil(t, ident) &&
								assert.Equal(t, tc.tenant, ident.Tenant)
						}),
						tc.tenant,
						tc.aset.DeviceId,
						cache.IdTypeDevice).
						Return(tc.cacheDeleteErr)
				} else {
					c.AssertNotCalled(t, "DeleteToken")
				}
			} else {
				c.AssertNotCalled(t, "DeleteToken")
			}

			err := devauth.RejectDeviceAuth(
				ctx, dummyDevID, dummyAuthID,
			)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}

			c.AssertExpectations(t)
		})
	}
}

func TestDevAuthRevokeToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tokenId string

		tenant         string
		withCache      bool
		cacheDeleteErr error

		dbToken  *jwt.Token
		dbGetErr error
		dbDelErr error

		outErr error
	}{
		{
			tokenId: "foo",
		},
		{
			tokenId:  "foo",
			dbDelErr: store.ErrTokenNotFound,
			outErr:   store.ErrTokenNotFound,
		},
		{
			tokenId:   "foo",
			withCache: true,
			tenant:    "acme",
			dbToken: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("device"),
				},
			},
		},
		{
			tokenId:   "foo",
			withCache: true,
			tenant:    "acme",
			dbGetErr:  store.ErrTokenNotFound,
			outErr:    store.ErrTokenNotFound,
		},
		{
			tokenId:   "foo",
			withCache: true,
			tenant:    "acme",
			dbToken: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("device"),
				},
			},
			cacheDeleteErr: errors.New("redis error"),
			outErr:         errors.New("failed to delete token for 884482f8-4b20-5b83-b674-6ca5cb3e7525 from cache: redis error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			tokenOID := oid.FromString(tc.tokenId)

			db := mstore.DataStore{}
			db.On("GetToken", ctx,
				tokenOID).
				Return(tc.dbToken, tc.dbGetErr)
			db.On("DeleteToken", ctx,
				tokenOID).
				Return(tc.dbDelErr)

			c := &mcache.Cache{}

			devauth := NewDevAuth(&db, nil, nil, Config{})

			if tc.withCache {
				devauth = devauth.WithCache(c)
				if tc.tenant != "" && tc.dbToken != nil {
					c.On("DeleteToken",
						mock.MatchedBy(func(ctx context.Context) bool {
							ident := identity.FromContext(ctx)
							return assert.NotNil(t, ident) &&
								assert.Equal(t, tc.tenant, ident.Tenant)
						}),
						tc.tenant,
						tc.dbToken.Claims.Subject.String(),
						cache.IdTypeDevice).
						Return(tc.cacheDeleteErr)
				} else {
					c.AssertNotCalled(t, "DeleteToken")
				}
			} else {
				c.AssertNotCalled(t, "DeleteToken")
			}

			err := devauth.RevokeToken(ctx, tc.tokenId)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}

			c.AssertExpectations(t)
		})
	}
}

func TestDevAuthResetDevice(t *testing.T) {
	t.Parallel()

	dummyDevUUID := oid.NewUUIDv5("dummy_devid")
	dummyDevID := dummyDevUUID.String()
	dummyAuthID := oid.NewUUIDv5("dummy_aid").String()

	testCases := []struct {
		aset             *model.AuthSet
		dbErr            error
		dbDelDevTokenErr error

		outErr string
	}{
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
			},
			dbDelDevTokenErr: nil,
		},
		{
			dbErr:  errors.New("failed"),
			outErr: "db get auth set error: failed",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   "accepted",
			},
			dbDelDevTokenErr: store.ErrTokenNotFound,
			outErr:           "db delete device token error: token not found",
		},
		{
			aset: &model.AuthSet{
				Id:       dummyAuthID,
				DeviceId: dummyDevID,
				Status:   "accepted",
			},
			dbDelDevTokenErr: errors.New("some error"),
			outErr:           "db delete device token error: some error",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			db := mstore.DataStore{}
			db.On("GetAuthSetById", context.Background(),
				dummyAuthID).
				Return(tc.aset, tc.dbErr)
			if tc.aset != nil {
				db.On("UpdateAuthSetById", context.Background(), tc.aset.Id,
					model.AuthSetUpdate{Status: model.DevStatusPending}).Return(nil)
			}
			db.On("DeleteTokenByDevId", context.Background(),
				dummyDevUUID).Return(
				tc.dbDelDevTokenErr)
			db.On("GetDeviceStatus", context.Background(),
				dummyDevID).Return(
				"accpted", nil)
			db.On("UpdateDevice", context.Background(),
				func() interface{} {
					if tc.aset != nil {
						return tc.aset.DeviceId
					}
					return mock.AnythingOfType("string")
				}(),
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)
			db.On("GetDeviceById", context.Background(),
				mock.AnythingOfType("string")).Return(&model.Device{}, nil)

			co := morchestrator.ClientRunner{}
			co.On("SubmitUpdateDeviceStatusJob", context.Background(),
				mock.AnythingOfType("orchestrator.UpdateDeviceStatusReq")).
				Return(nil)

			devauth := NewDevAuth(&db, &co, nil, Config{})
			err := devauth.ResetDeviceAuth(
				context.Background(), dummyDevID, dummyAuthID,
			)

			if tc.dbErr != nil ||
				(tc.dbDelDevTokenErr != nil &&
					tc.dbDelDevTokenErr != store.ErrTokenNotFound) {

				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevAuthVerifyToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tokenString      string
		tokenParseErr    error
		tokenValidateErr error
		tokenOtherError  error

		jwToken             *jwt.Token
		validateErr         error
		fallbackValidateErr error

		getToken    bool
		getTokenErr error

		auth       *model.AuthSet
		getAuthErr error

		dev          *model.Device
		getDeviceErr error

		updateDeviceErr error

		tenantVerify       bool
		willUpdateDevice   bool
		jwtHandlerFallback bool
	}{
		{
			tokenString:      "expired",
			tokenValidateErr: jwt.ErrTokenExpired,

			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("expired"),
					Subject: oid.NewUUIDv5("foo"),
					Device:  true,
					ExpiresAt: jwt.Time{
						Time: time.Now().
							Add(-time.Hour),
					},
				},
			},
			validateErr: jwt.ErrTokenExpired,
		},
		{
			tokenString:   "bad",
			tokenParseErr: jwt.ErrTokenInvalid,

			jwToken:     nil,
			validateErr: jwt.ErrTokenInvalid,
		},
		{
			tokenString: "good-accepted",
			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("good"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Issuer: "Tester",
					Device: true,
				},
			},
			getToken: true,
			auth: &model.AuthSet{
				Id:       oid.NewUUIDv5("good").String(),
				Status:   model.DevStatusAccepted,
				DeviceId: oid.NewUUIDv5("bar").String(),
			},
			dev: &model.Device{
				Id:              oid.NewUUIDv5("bar").String(),
				Decommissioning: false,
			},
			willUpdateDevice: true,
		},
		{
			tokenString:      "good-rejected",
			tokenValidateErr: jwt.ErrTokenInvalid,

			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:       oid.NewUUIDv5("good-rejected"),
					Subject:  oid.NewUUIDv5("baz"),
					Issuer:   "Tester",
					IssuedAt: jwt.Time{Time: time.Now()},
					ExpiresAt: jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Device: true,
				},
			},
			getToken: true,
			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("good-rejected").String(),
				Status: model.DevStatusRejected,
			},
		},
		{
			tokenString:      "good-accepted-decommissioning",
			tokenValidateErr: jwt.ErrTokenInvalid,

			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("good-decommissioning"),
					Subject: oid.NewUUIDv5("idk"),
					Device:  true,
					Issuer:  "Tester",
					ExpiresAt: jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
				},
			},
			getToken: true,
			auth: &model.AuthSet{
				Id: oid.NewUUIDv5("good-decommissioning").
					String(),
				Status:   model.DevStatusAccepted,
				DeviceId: oid.NewUUIDv5("idk").String(),
			},
			dev: &model.Device{
				Id:              oid.NewUUIDv5("idk").String(),
				Decommissioning: true,
			},
		},
		{
			tokenString:     "not-a-device-token",
			tokenOtherError: jwt.ErrTokenInvalid,

			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("missing-tenant"),
					Subject: oid.NewUUIDv5("foo"),
				},
			},

			tenantVerify: true,
		},
		{
			tokenString:     "missing-tenant-claim",
			tokenOtherError: jwt.ErrTokenInvalid,

			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("missing-tenant"),
					Subject: oid.NewUUIDv5("foo"),
					Device:  true,
				},
			},

			tenantVerify: true,
		},
		{
			tokenString: "with fallback",
			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("good"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Issuer: "Tester",
					Device: true,
				},
			},
			validateErr: jwt.ErrTokenInvalid,
			getToken:    true,
			auth: &model.AuthSet{
				Id:       oid.NewUUIDv5("good").String(),
				Status:   model.DevStatusAccepted,
				DeviceId: oid.NewUUIDv5("bar").String(),
			},
			dev: &model.Device{
				Id:              oid.NewUUIDv5("bar").String(),
				Decommissioning: false,
			},
			willUpdateDevice:   true,
			jwtHandlerFallback: true,
		},
		{
			tokenString:      "failed-validation-with-fallback",
			tokenValidateErr: jwt.ErrTokenInvalid,
			jwToken: &jwt.Token{
				Claims: jwt.Claims{
					ID:      oid.NewUUIDv5("good"),
					Subject: oid.NewUUIDv5("bar"),
					ExpiresAt: jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Issuer: "Tester",
					Device: true,
				},
			},
			validateErr:         jwt.ErrTokenInvalid,
			fallbackValidateErr: jwt.ErrTokenInvalid,
			jwtHandlerFallback:  true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", tc.tokenString), func(t *testing.T) {
			t.Parallel()

			db := &mstore.DataStore{}
			ja := &mjwt.Handler{}
			co := &morchestrator.ClientRunner{}
			co.On("SubmitUpdateDeviceInventoryJob", context.Background(),
				mock.AnythingOfType("orchestrator.UpdateDeviceInventoryReq")).
				Return(nil)

			devauth := NewDevAuth(db, co, ja, Config{})
			if tc.jwtHandlerFallback {
				jaFallback := &mjwt.Handler{}
				jaFallback.On("Validate", tc.tokenString).Return(tc.fallbackValidateErr)

				devauth = devauth.WithJWTFallbackHandler(jaFallback)
			}
			if tc.tenantVerify {
				// ok to pass nil tenantadm client here
				devauth = devauth.WithTenantVerification(nil)
			}

			ja.On("FromJWT", tc.tokenString).Return(
				func(s string) *jwt.Token {
					t.Logf("string: %v return %+v", s, tc.jwToken)
					return tc.jwToken
				}, tc.tokenParseErr)

			if tc.tokenParseErr == nil && tc.jwToken != nil && tc.tokenOtherError == nil &&
				tc.tokenString != "missing-tenant-claim" {
				ja.On("Validate", tc.tokenString).Return(tc.validateErr)
			}

			if tc.validateErr == jwt.ErrTokenExpired {
				db.On("DeleteToken",
					context.Background(),
					tc.jwToken.Claims.ID).Return(nil)
			}

			if tc.getToken {
				db.On("GetToken", context.Background(),
					tc.jwToken.Claims.ID).
					Return(tc.jwToken, tc.getTokenErr)
				db.On("GetAuthSetById", context.Background(),
					tc.jwToken.ID.String()).
					Return(tc.auth, tc.getAuthErr)
				// devauth will ask for a device if auth set is
				// found and accepted
				if tc.dev != nil {
					db.On("GetDeviceById", context.Background(),
						tc.auth.DeviceId).Return(tc.dev, tc.getDeviceErr)
				}
			}

			if tc.willUpdateDevice {
				db.On("UpdateDevice", context.Background(),
					tc.jwToken.Claims.Subject.String(),
					mock.AnythingOfType("model.DeviceUpdate"),
				).Return(tc.updateDeviceErr)
			} else {
				db.AssertNotCalled(t, "UpdateDevice")
			}

			err := devauth.VerifyToken(context.Background(), tc.tokenString)
			if tc.tokenParseErr != nil {
				assert.EqualError(t, err, tc.tokenParseErr.Error())
			} else if tc.tokenValidateErr != nil {
				assert.EqualError(t, err, tc.tokenValidateErr.Error())
			} else if tc.tokenOtherError != nil {
				assert.EqualError(t, err, tc.tokenOtherError.Error())
			} else {
				assert.NoError(t, err)
			}
			ja.AssertExpectations(t)
			db.AssertExpectations(t)

		})
	}
}

func TestDevAuthVerifyTokenWithCache(t *testing.T) {
	t.Parallel()

	nowUnix := int64(1590105600)
	mclock := utils.NewMockClock(nowUnix)
	token := &jwt.Token{
		Claims: jwt.Claims{
			Subject: oid.NewUUIDv5("device"),
			Tenant:  "tenant",
			Device:  true,
			ExpiresAt: jwt.Time{
				Time: time.Unix(nowUnix+1000, 0),
			},
		},
	}

	// assume valid input jwt token always
	testCases := map[string]struct {
		tokenString string

		cacheGetLimits    *ratelimits.ApiLimits
		cacheGetLimitsErr error

		cachedToken string
		throttleErr error

		getTokenErr error

		auth       *model.AuthSet
		getAuthErr error

		dev          *model.Device
		getDeviceErr error

		tenant       *tenant.Tenant
		getTenantErr error

		cacheTokenErr  error
		cacheLimitsErr error

		checkInTime       *time.Time
		getCheckInTimeErr error

		cacheCheckInTimeErr error

		updateDeviceErr error

		willCallThrottle     bool
		willVerifyDb         bool
		willCacheToken       bool
		willFetchLimits      bool
		willGetCheckInTime   bool
		willCacheCheckInTime bool
		willUpdateDevice     bool

		outErr error
	}{
		"token cached, no limiting err - db not called for verification, success": {
			tokenString: "valid",
			cachedToken: "valid",

			cacheGetLimits:    &ratelimits.ApiLimits{},
			cacheGetLimitsErr: nil,

			willCallThrottle:     true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"token cached, but limits exceeded - early return": {
			tokenString: "valid",
			cachedToken: "valid",
			throttleErr: cache.ErrTooManyRequests,

			cacheGetLimits:    &ratelimits.ApiLimits{},
			cacheGetLimitsErr: nil,

			willVerifyDb:   false,
			willCacheToken: false,

			outErr: cache.ErrTooManyRequests,

			willCallThrottle: true,
		},
		"throttle transient error - swallow error, proceed with standard db verification flow: success, cache token": {
			tokenString: "valid",
			cachedToken: "",
			throttleErr: errors.New("redis error"),

			cacheGetLimits:    &ratelimits.ApiLimits{},
			cacheGetLimitsErr: nil,

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				Id:              oid.NewUUIDv5("device").String(),
				Decommissioning: false,
			},

			willCallThrottle:     true,
			willVerifyDb:         true,
			willCacheToken:       true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"throttle transient error - swallow error, proceed with standard db verification flow: success, cache token: error (don't fail)": {
			tokenString: "valid",
			cachedToken: "",
			throttleErr: errors.New("redis error"),

			cacheGetLimits:    &ratelimits.ApiLimits{},
			cacheGetLimitsErr: nil,

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				Id:              oid.NewUUIDv5("device").String(),
				Decommissioning: false,
			},

			willCallThrottle:     true,
			willVerifyDb:         true,
			willCacheToken:       true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"token not cached - verify against db, failed": {
			tokenString: "valid",
			getTokenErr: store.ErrTokenNotFound,

			cacheGetLimits:    &ratelimits.ApiLimits{},
			cacheGetLimitsErr: nil,

			willVerifyDb:   true,
			willCacheToken: false,

			outErr: store.ErrTokenNotFound,

			willCallThrottle: true,
		},
		"limits not in cache, db/service hit for limits (success)": {
			tokenString: "valid",
			cachedToken: "valid",

			cacheGetLimits:    nil,
			cacheGetLimitsErr: nil,

			dev: &model.Device{
				ApiLimits: ratelimits.ApiLimits{
					ApiBursts: []ratelimits.ApiBurst{},
				},
			},
			tenant: &tenant.Tenant{
				ApiLimits: tenant.TenantApiLimits{
					DeviceLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
				},
			},

			willCallThrottle:     true,
			willVerifyDb:         false,
			willCacheToken:       false,
			willFetchLimits:      true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"limits mgmt errors won't stop processing - 'get cached limits' failed": {
			tokenString: "valid",

			cacheGetLimits:    nil,
			cacheGetLimitsErr: errors.New("internal"),

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				ApiLimits: ratelimits.ApiLimits{
					ApiBursts: []ratelimits.ApiBurst{},
				},
			},

			tenant: &tenant.Tenant{
				ApiLimits: tenant.TenantApiLimits{
					DeviceLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
				},
			},

			willCallThrottle:     false,
			willVerifyDb:         true,
			willCacheToken:       true,
			willFetchLimits:      false,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"limits mgmt errors won't stop processing - 'get tenant' failed": {
			tokenString: "valid",

			cacheGetLimits:    nil,
			cacheGetLimitsErr: nil,

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				ApiLimits: ratelimits.ApiLimits{
					ApiBursts: []ratelimits.ApiBurst{},
				},
			},

			getTenantErr: errors.New("internal error"),

			willCallThrottle:     false,
			willVerifyDb:         true,
			willCacheToken:       true,
			willFetchLimits:      true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"limits mgmt errors won't stop processing - tenant not found": {
			tokenString: "valid",

			cacheGetLimits:    nil,
			cacheGetLimitsErr: nil,

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				ApiLimits: ratelimits.ApiLimits{
					ApiBursts: []ratelimits.ApiBurst{},
				},
			},

			getTenantErr: errors.New("internal error"),

			checkInTime:       nil,
			getCheckInTimeErr: nil,

			willCallThrottle:     false,
			willVerifyDb:         true,
			willCacheToken:       true,
			willFetchLimits:      true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
		"limits mgmt errors won't stop processing - 'cache limits' failed": {
			tokenString: "valid",

			cacheGetLimits:    nil,
			cacheGetLimitsErr: nil,

			auth: &model.AuthSet{
				Id:     oid.NewUUIDv5("foo").String(),
				Status: model.DevStatusAccepted,
			},

			dev: &model.Device{
				ApiLimits: ratelimits.ApiLimits{
					ApiBursts: []ratelimits.ApiBurst{},
				},
			},

			tenant: &tenant.Tenant{
				ApiLimits: tenant.TenantApiLimits{
					DeviceLimits: ratelimits.ApiLimits{
						ApiBursts: []ratelimits.ApiBurst{},
					},
				},
			},

			cacheLimitsErr: errors.New("redis error"),

			willCallThrottle:     false,
			willVerifyDb:         true,
			willCacheToken:       true,
			willFetchLimits:      true,
			willGetCheckInTime:   true,
			willCacheCheckInTime: true,
			willUpdateDevice:     true,
		},
	}

	for n, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := &mstore.DataStore{}
			ja := &mjwt.Handler{}
			c := &mcache.Cache{}
			co := &morchestrator.ClientRunner{}

			devauth := NewDevAuth(db, co, ja, Config{})
			devauth = devauth.WithCache(c)
			tclient := &mtenant.ClientRunner{}

			devauth = devauth.WithTenantVerification(tclient)
			devauth = devauth.WithClock(mclock)

			co.On("SubmitUpdateDeviceInventoryJob", ctx,
				mock.AnythingOfType("orchestrator.UpdateDeviceInventoryReq")).
				Return(nil)

			ja.On("FromJWT", tc.tokenString).Return(
				func(s string) *jwt.Token {
					t.Logf("string: %v return %+v", s, token)
					return token
				}, nil)

			ja.On("Validate", tc.tokenString).Return(nil)

			c.On("GetLimits",
				ctx,
				token.Claims.Tenant,
				token.Claims.Subject.String(),
				cache.IdTypeDevice).Return(tc.cacheGetLimits, tc.cacheGetLimitsErr)

			if tc.willGetCheckInTime {
				c.On("GetCheckInTime",
					ctx,
					token.Claims.Tenant,
					token.Claims.Subject.String(),
				).Return(tc.checkInTime, tc.getCheckInTimeErr)
			} else {
				db.AssertNotCalled(t, "GetCheckInTime")
			}

			if tc.willCacheCheckInTime {
				c.On("CacheCheckInTime",
					ctx,
					mock.AnythingOfType("*time.Time"),
					token.Claims.Tenant,
					token.Claims.Subject.String(),
				).Return(tc.cacheCheckInTimeErr)
			} else {
				db.AssertNotCalled(t, "CacheCheckInTime")
			}

			if tc.willCallThrottle {
				c.On("Throttle",
					ctx,
					tc.tokenString,
					mock.AnythingOfType("ratelimits.ApiLimits"),
					token.Claims.Tenant,
					token.Claims.Subject.String(),
					cache.IdTypeDevice,
					"",
					"").Return(tc.cachedToken, tc.throttleErr)
			}

			if tc.willVerifyDb {
				db.On("GetToken", ctx,
					token.Claims.ID).
					Return(token, tc.getTokenErr)
				if tc.getTokenErr == nil {
					db.On("GetAuthSetById", ctx,
						token.ID.String()).
						Return(tc.auth, tc.getAuthErr)
					db.On("GetDeviceById", ctx,
						tc.auth.DeviceId).Return(tc.dev, tc.getDeviceErr)
				}
			} else {
				db.AssertNotCalled(t, "GetToken")
			}

			if tc.willCacheToken {
				expireIn := time.Duration(token.Claims.ExpiresAt.Unix()-nowUnix) * time.Second
				c.On("CacheToken",
					ctx,
					token.Claims.Tenant,
					token.Claims.Subject.String(),
					cache.IdTypeDevice,
					tc.tokenString,
					expireIn).Return(tc.cacheTokenErr)
			} else {
				c.AssertNotCalled(t, "CacheToken")
			}

			if tc.willFetchLimits {
				db.On("GetDeviceById", ctx,
					token.Claims.Subject.String()).Return(tc.dev, tc.getDeviceErr)

				if tc.getDeviceErr == nil {
					tclient.On("GetTenant",
						ctx,
						token.Claims.Tenant).
						Return(tc.tenant, tc.getTenantErr)
				}

				if tc.getDeviceErr == nil && tc.getTenantErr == nil {
					c.On("CacheLimits",
						ctx,
						apiLimitsOverride(tc.dev.ApiLimits, tc.tenant.ApiLimits.DeviceLimits),
						token.Claims.Tenant,
						token.Claims.Subject.String(),
						cache.IdTypeDevice).Return(tc.cacheLimitsErr)
				} else {
					c.AssertNotCalled(t, "CacheLimits")
				}
			} else {
				db.AssertNotCalled(t, "GetDeviceById")
				tclient.AssertNotCalled(t, "GetTenant")
				c.AssertNotCalled(t, "CacheLimits")
			}
			if tc.willUpdateDevice {
				db.On("UpdateDevice", ctx,
					token.Claims.Subject.String(),
					mock.AnythingOfType("model.DeviceUpdate"),
				).Return(tc.updateDeviceErr)
			} else {
				db.AssertNotCalled(t, "UpdateDevice")
			}

			err := devauth.VerifyToken(context.Background(), tc.tokenString)
			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}

			db.AssertExpectations(t)
			c.AssertExpectations(t)
		})
	}
}

func TestDevAuthDecommissionDevice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		devId  string
		tenant string

		dbUpdateDeviceErr error

		withCache      bool
		cacheDeleteErr error

		coSubmitDeviceDecommisioningJobErr error
		coAuthorization                    string

		outErr string
	}{
		{
			devId: oid.NewUUIDv5("devId1").String(),

			dbUpdateDeviceErr: errors.New("UpdateDevice Error"),
			outErr:            "UpdateDevice Error",
		},
		{
			devId: oid.NewUUIDv5("devId2").String(),

			dbUpdateDeviceErr: errors.New("DeleteDevice Error"),
			outErr:            "DeleteDevice Error",
		},
		{
			devId: oid.NewUUIDv5("devId3").String(),

			coSubmitDeviceDecommisioningJobErr: errors.New("SubmitDeviceDecommisioningJob Error"),
			outErr:                             "submit device decommissioning job error: SubmitDeviceDecommisioningJob Error",
		},
		{
			devId:           oid.NewUUIDv5("devId4").String(),
			coAuthorization: "Bearer foobar",
		},
		{
			devId:           oid.NewUUIDv5("devId4").String(),
			withCache:       true,
			tenant:          "acme",
			coAuthorization: "Bearer foobar",
		},
		{
			devId:          oid.NewUUIDv5("devId4").String(),
			withCache:      true,
			tenant:         "acme",
			cacheDeleteErr: errors.New("redis error"),
			outErr:         "failed to delete token for 725e564d-cde0-5b11-bace-b79393120089 from cache: redis error",
		},
		{
			devId:          oid.NewUUIDv5("devId4").String(),
			withCache:      true,
			tenant:         "acme",
			cacheDeleteErr: errors.New("redis error"),
			outErr:         "failed to delete token for 725e564d-cde0-5b11-bace-b79393120089 from cache: redis error",
		},
		{
			devId:     oid.NewUUIDv5("devId4").String(),
			withCache: true,
			outErr:    "failed to delete token for 725e564d-cde0-5b11-bace-b79393120089 from cache: can't unpack tenant identity data from context",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			if tc.coAuthorization != "" {
				ctx = ctxhttpheader.WithContext(ctx, http.Header{
					"Authorization": []string{tc.coAuthorization},
				}, "Authorization")
			}

			co := morchestrator.ClientRunner{}
			co.On("SubmitDeviceDecommisioningJob", ctx,
				orchestrator.DecommissioningReq{
					DeviceId: tc.devId,
					TenantID: tc.tenant,
				}).
				Return(tc.coSubmitDeviceDecommisioningJobErr)

			db := mstore.DataStore{}
			db.On("UpdateDevice", ctx,
				tc.devId,
				model.DeviceUpdate{
					Decommissioning: uto.BoolPtr(true),
				}).Return(
				tc.dbUpdateDeviceErr)
			db.On("UpdateDevice", ctx,
				mock.AnythingOfType("model.Device"),
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)

			devauth := NewDevAuth(&db, &co, nil, Config{})
			c := &mcache.Cache{}

			if tc.withCache {
				devauth = devauth.WithCache(c)
				if tc.tenant != "" {
					c.On("DeleteToken",
						mock.MatchedBy(func(ctx context.Context) bool {
							ident := identity.FromContext(ctx)
							return assert.NotNil(t, ident) &&
								assert.Equal(t, tc.tenant, ident.Tenant)
						}),
						tc.tenant,
						tc.devId,
						cache.IdTypeDevice).
						Return(tc.cacheDeleteErr)
				} else {
					c.AssertNotCalled(t, "DeleteToken")
				}
			} else {
				c.AssertNotCalled(t, "DeleteToken")
			}

			err := devauth.DecommissionDevice(ctx, tc.devId)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}

			c.AssertExpectations(t)
		})
	}
}

func TestTestDevAuthDeleteDevice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		devId  string
		tenant string

		dbDeleteAuthSetsForDeviceErr error
		dbDeleteTokenByDevIdErr      error
		dbDeleteDeviceErr            error

		coSubmitDeviceDecommisioningJobErr error

		outErr string
	}{
		{
			devId: oid.NewUUIDv5("devId1").String(),

			dbDeleteDeviceErr: store.ErrDevNotFound,
			outErr:            "device not found",
		},
		{
			devId: oid.NewUUIDv5("devId2").String(),

			dbDeleteDeviceErr: errors.New("failed to remove device"),
			outErr:            "failed to remove device",
		},
		{
			devId: oid.NewUUIDv5("devId3").String(),

			dbDeleteAuthSetsForDeviceErr: errors.New("DeleteAuthSetsForDevice Error"),
			outErr:                       "db delete device authorization sets error: DeleteAuthSetsForDevice Error",
		},
		{
			devId: oid.NewUUIDv5("devId4").String(),

			dbDeleteAuthSetsForDeviceErr: store.ErrAuthSetNotFound,
		},
		{
			devId: oid.NewUUIDv5("devId5").String(),

			dbDeleteTokenByDevIdErr: errors.New("DeleteTokenByDevId Error"),
			outErr:                  "db delete device tokens error: DeleteTokenByDevId Error",
		},
		{
			devId: "",

			outErr: "invalid device ID type",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := mstore.DataStore{}
			devUUID := oid.FromString(tc.devId)
			db.On("DeleteAuthSetsForDevice", ctx,
				tc.devId).Return(
				tc.dbDeleteAuthSetsForDeviceErr)
			db.On("DeleteTokenByDevId", ctx,
				devUUID).Return(
				tc.dbDeleteTokenByDevIdErr)
			db.On("DeleteDevice", ctx,
				tc.devId).Return(
				tc.dbDeleteDeviceErr)

			co := morchestrator.ClientRunner{}
			co.On("SubmitDeviceDecommisioningJob", ctx,
				orchestrator.DecommissioningReq{
					DeviceId: tc.devId,
					TenantID: tc.tenant,
				}).
				Return(tc.coSubmitDeviceDecommisioningJobErr)

			devauth := NewDevAuth(&db, &co, nil, Config{})

			err := devauth.DeleteDevice(ctx, tc.devId)
			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestDevAuthSetTenantLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tenantId string

		dbPutLimitErr error
		limit         model.Limit

		outErr string
	}{
		{
			tenantId:      "tenant1",
			dbPutLimitErr: errors.New("PutLimit error"),
			outErr:        "failed to save limit { foobar 1234 } for tenant tenant1 to database: PutLimit error",
			limit: model.Limit{
				Name:  "foobar",
				Value: 1234,
			},
		},
		{
			tenantId: "tenant2",
			limit: model.Limit{
				Name:  "foobar2",
				Value: 9999999,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := mstore.DataStore{}
			db.On("PutLimit",
				mock.MatchedBy(func(ctx context.Context) bool {
					ident := identity.FromContext(ctx)
					return assert.NotNil(t, ident) &&
						assert.Equal(t, tc.tenantId, ident.Tenant)
				}),
				tc.limit).
				Return(tc.dbPutLimitErr)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			err := devauth.SetTenantLimit(ctx, tc.tenantId, tc.limit)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevAuthDeleteTenantLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tenantId string

		dbPutLimitErr error
		limit         string

		outErr string
	}{
		{
			tenantId:      "tenant1",
			dbPutLimitErr: errors.New("DeleteLimit error"),
			outErr:        "failed to delete limit foobar for tenant tenant1 to database: DeleteLimit error",
			limit:         "foobar",
		},
		{
			tenantId: "tenant2",
			limit:    "foobar2",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := mstore.DataStore{}
			db.On("DeleteLimit",
				mock.MatchedBy(func(ctx context.Context) bool {
					ident := identity.FromContext(ctx)
					return assert.NotNil(t, ident) &&
						assert.Equal(t, tc.tenantId, ident.Tenant)
				}),
				tc.limit).
				Return(tc.dbPutLimitErr)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			err := devauth.DeleteTenantLimit(ctx, tc.tenantId, tc.limit)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevAuthGetLimit(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inName string

		dbLimit *model.Limit
		dbErr   error

		outLimit *model.Limit
		outErr   error
	}{
		"ok": {
			inName: "other_limit",

			dbLimit: &model.Limit{Name: "other_limit", Value: 123},
			dbErr:   nil,

			outLimit: &model.Limit{Name: "other_limit", Value: 123},
			outErr:   nil,
		},
		"ok max_devices": {
			inName: model.LimitMaxDeviceCount,

			dbLimit: &model.Limit{Name: model.LimitMaxDeviceCount, Value: 123},
			dbErr:   nil,

			outLimit: &model.Limit{Name: model.LimitMaxDeviceCount, Value: 123},
			outErr:   nil,
		},
		"limit not found": {
			inName: "other_limit",

			dbLimit: nil,
			dbErr:   store.ErrLimitNotFound,

			outLimit: &model.Limit{Name: "other_limit", Value: 0},
			outErr:   nil,
		},
		"generic error": {
			inName: "max_devices",

			dbLimit: nil,
			dbErr:   errors.New("db error"),

			outLimit: nil,
			outErr:   errors.New("db error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := mstore.DataStore{}
			db.On("GetLimit", ctx, tc.inName).Return(tc.dbLimit, tc.dbErr)
			db.On("UpdateDevice", ctx,
				mock.AnythingOfType("model.Device"),
				mock.AnythingOfType("model.DeviceUpdate")).Return(nil)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			limit, err := devauth.GetLimit(ctx, tc.inName)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tc.outLimit, *limit)
			}
		})
	}
}

func TestDevAuthGetTenantLimit(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inName   string
		inTenant string

		dbLimit *model.Limit
		dbErr   error

		outLimit *model.Limit
		outErr   error
	}{
		"ok": {
			inName:   "max_devices",
			inTenant: "tenant-foo",

			dbLimit: &model.Limit{Name: "max_devices", Value: 123},
			dbErr:   nil,

			outLimit: &model.Limit{Name: "max_devices", Value: 123},
			outErr:   nil,
		},
		"limit not found": {
			inName:   "max_devices",
			inTenant: "tenant-bar",

			dbLimit: nil,
			dbErr:   store.ErrLimitNotFound,

			outLimit: &model.Limit{Name: "max_devices", Value: 0},
			outErr:   nil,
		},
		"generic error": {
			inName:   "max_devices",
			inTenant: "tenant-baz",

			dbLimit: nil,
			dbErr:   errors.New("db error"),

			outLimit: nil,
			outErr:   errors.New("db error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			id := &identity.Identity{
				Tenant: "5f23456789cafddfe",
			}
			ctx = identity.WithContext(ctx, id)

			db := mstore.DataStore{}
			// in get limit, verify the correct db was set
			verifyCtx := func(args mock.Arguments) {
				ctx := args.Get(0).(context.Context)
				id := identity.FromContext(ctx)
				assert.Equal(t, tc.inTenant, id.Tenant)
			}

			ctxMatcher := mock.MatchedBy(func(c context.Context) bool {
				return assert.NotNil(t, identity.FromContext(c))
			})

			db.On("GetLimit", ctxMatcher, tc.inName).
				Run(verifyCtx).
				Return(tc.dbLimit, tc.dbErr)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			limit, err := devauth.GetTenantLimit(ctx, tc.inName, tc.inTenant)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tc.outLimit, *limit)
			}
		})
	}
}

func TestDevAuthGetDevCountByStatus(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		status string

		dbCnt int
		dbErr error

		cnt int
		err error
	}{
		"ok": {
			status: "pending",

			dbCnt: 5,
			dbErr: nil,

			cnt: 5,
			err: nil,
		},
		"ok 2": {
			status: "accepted",

			dbCnt: 0,
			dbErr: nil,

			cnt: 0,
			err: nil,
		},
		"generic error": {
			status: "accepted",

			dbCnt: 5,
			dbErr: errors.New("db error"),

			err: errors.New("db error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %s", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			db := mstore.DataStore{}
			db.On("GetDevCountByStatus", ctx, tc.status).Return(tc.dbCnt, tc.dbErr)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			cnt, err := devauth.GetDevCountByStatus(ctx, tc.status)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.cnt, cnt)
			}
		})
	}
}

func TestDevAuthDeleteAuthSet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		devId  string
		authId string

		tenant          string
		withCache       bool
		cacheDeleteErr  error
		enableReporting bool

		dbGetAuthSetByIdErr         error
		dbDeleteTokenByDevIdErr     error
		dbDeleteAuthSetForDeviceErr error
		dbGetAuthSetsForDeviceErr   error
		dbDeleteDeviceErr           error
		dbGetDeviceStatus           string
		dbGetDeviceStatusErr        error
		dbUpdateDeviceErr           error

		orchestratorErr          error
		orchestratorReportingErr error

		authSet *model.AuthSet

		outErr string
	}{
		{
			devId:               oid.NewUUIDv5("devId1").String(),
			authId:              oid.NewUUIDv5("authId1").String(),
			dbGetAuthSetByIdErr: errors.New("GetAuthSetById Error"),
			outErr:              "db get auth set error: GetAuthSetById Error",
		},
		{
			devId:               oid.NewUUIDv5("devId2").String(),
			authId:              oid.NewUUIDv5("authId2").String(),
			dbGetAuthSetByIdErr: store.ErrAuthSetNotFound,
			outErr:              store.ErrAuthSetNotFound.Error(),
		},
		{
			devId:                   oid.NewUUIDv5("devId3").String(),
			authId:                  oid.NewUUIDv5("authId3").String(),
			authSet:                 &model.AuthSet{Status: model.DevStatusAccepted},
			dbDeleteTokenByDevIdErr: errors.New("DeleteTokenByDevId Error"),
			outErr:                  "db delete device tokens error: DeleteTokenByDevId Error",
		},
		{
			devId:                   oid.NewUUIDv5("devId4").String(),
			authId:                  oid.NewUUIDv5("authId4").String(),
			authSet:                 &model.AuthSet{Status: model.DevStatusPending},
			dbDeleteTokenByDevIdErr: errors.New("DeleteTokenByDevId Error"),
		},
		{
			devId:                   oid.NewUUIDv5("devId5").String(),
			authId:                  oid.NewUUIDv5("authId5").String(),
			dbDeleteTokenByDevIdErr: store.ErrTokenNotFound,
		},
		{
			devId:                       oid.NewUUIDv5("devId6").String(),
			authId:                      oid.NewUUIDv5("authId6").String(),
			dbDeleteAuthSetForDeviceErr: errors.New("DeleteAuthSetsForDevice Error"),
			outErr:                      "DeleteAuthSetsForDevice Error",
		},
		{
			devId:             oid.NewUUIDv5("devId8").String(),
			authId:            oid.NewUUIDv5("authId8").String(),
			authSet:           &model.AuthSet{Status: model.DevStatusPreauth},
			dbGetDeviceStatus: "decommissioned",
			dbDeleteDeviceErr: errors.New("DeleteDevice Error"),
			outErr:            "DeleteDevice Error",
		},
		{
			devId:             oid.NewUUIDv5("devId8").String(),
			authId:            oid.NewUUIDv5("authId8").String(),
			authSet:           &model.AuthSet{Status: model.DevStatusPreauth},
			dbGetDeviceStatus: "decommissioned",
			orchestratorErr:   errors.New("orchestrator error"),
			outErr:            "update device status job error: orchestrator error",
		},

		{
			devId:             oid.NewUUIDv5("devId9").String(),
			authId:            oid.NewUUIDv5("authId9").String(),
			dbDeleteDeviceErr: errors.New("DeleteDevice Error"),
		},
		{
			devId:                oid.NewUUIDv5("devId10").String(),
			authId:               oid.NewUUIDv5("authId10").String(),
			dbGetDeviceStatusErr: errors.New("Get Device Status Error"),
			outErr:               "Cannot determine device status: Get Device Status Error",
		},
		{
			devId:             oid.NewUUIDv5("devId11").String(),
			authId:            oid.NewUUIDv5("authId11").String(),
			dbUpdateDeviceErr: errors.New("Update Device Error"),
			outErr:            "failed to update device status: Update Device Error",
		},
		{
			devId:             oid.NewUUIDv5("devId12").String(),
			authId:            oid.NewUUIDv5("authId12").String(),
			dbGetDeviceStatus: "accepted",
		},
		{
			devId:             oid.NewUUIDv5("devId12").String(),
			authId:            oid.NewUUIDv5("authId12").String(),
			withCache:         true,
			tenant:            "acme",
			dbGetDeviceStatus: "accepted",
		},
		{
			devId:             oid.NewUUIDv5("devId12").String(),
			authId:            oid.NewUUIDv5("authId12").String(),
			dbGetDeviceStatus: "accepted",
			enableReporting:   true,
		},
		{
			devId:                    oid.NewUUIDv5("devId8").String(),
			authId:                   oid.NewUUIDv5("authId8").String(),
			authSet:                  &model.AuthSet{Status: model.DevStatusPreauth},
			dbGetDeviceStatus:        "decommissioned",
			enableReporting:          true,
			orchestratorReportingErr: errors.New("orchestrator error"),
			outErr:                   "reindex reporting job error: orchestrator error",
		},
		{
			devId:                    oid.NewUUIDv5("devId12").String(),
			authId:                   oid.NewUUIDv5("authId12").String(),
			dbGetDeviceStatus:        "accepted",
			enableReporting:          true,
			orchestratorReportingErr: errors.New("orchestrator error"),
			outErr:                   "reindex reporting job error: orchestrator error",
		},
		{
			devId:                oid.NewUUIDv5("devId12").String(),
			authId:               oid.NewUUIDv5("authId12").String(),
			dbGetDeviceStatusErr: store.ErrAuthSetNotFound,
		},
		{
			devId:          oid.NewUUIDv5("devId12").String(),
			authId:         oid.NewUUIDv5("authId12").String(),
			withCache:      true,
			tenant:         "acme",
			cacheDeleteErr: errors.New("redis error"),
			outErr:         "failed to delete token for c410d383-c9cd-5c98-9aeb-87166c5920f2 from cache: redis error",
		},
		{
			devId:     oid.NewUUIDv5("devId12").String(),
			authId:    oid.NewUUIDv5("authId12").String(),
			withCache: true,
			outErr:    "failed to delete token for c410d383-c9cd-5c98-9aeb-87166c5920f2 from cache: can't unpack tenant identity data from context",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			authSet := &model.AuthSet{Status: model.DevStatusPending}
			if tc.authSet != nil {
				authSet = tc.authSet
			}

			devUUID := oid.FromString(tc.devId)
			db := mstore.DataStore{}
			db.On("GetAuthSetById", ctx,
				tc.authId).Return(
				authSet,
				tc.dbGetAuthSetByIdErr)
			db.On("DeleteAuthSetForDevice", ctx,
				tc.devId, tc.authId).Return(
				tc.dbDeleteAuthSetForDeviceErr)
			db.On("DeleteTokenByDevId", ctx,
				devUUID).Return(
				tc.dbDeleteTokenByDevIdErr)
			db.On("DeleteDevice", ctx,
				tc.devId).Return(
				tc.dbDeleteDeviceErr)
			db.On("GetDeviceStatus", ctx,
				tc.devId).Return(
				tc.dbGetDeviceStatus,
				tc.dbGetDeviceStatusErr)
			db.On("UpdateDevice", ctx,
				tc.devId,
				mock.AnythingOfType("model.DeviceUpdate")).Return(tc.dbUpdateDeviceErr)
			db.On("GetDeviceById", ctx,
				mock.AnythingOfType("string")).Return(&model.Device{Id: tc.devId}, nil)

			co := morchestrator.ClientRunner{}
			co.On("SubmitUpdateDeviceStatusJob", ctx,
				mock.MatchedBy(
					func(req orchestrator.UpdateDeviceStatusReq) bool {
						if tc.dbGetDeviceStatusErr == store.ErrAuthSetNotFound {
							assert.Equal(t, tc.devId, req.Devices[0].Id)
							assert.Equal(t, "noauth", req.Status)
							return true
						} else {
							assert.Equal(t, tc.devId, req.Devices[0].Id)
							assert.Equal(t, tc.dbGetDeviceStatus, req.Status)
							return true
						}
					})).Return(tc.orchestratorErr)

			if tc.enableReporting {
				co.On("SubmitReindexReporting",
					ctx,
					tc.devId,
				).Return(tc.orchestratorReportingErr).Once()
			}

			devauth := NewDevAuth(&db, &co, nil, Config{
				EnableReporting: tc.enableReporting,
			})

			c := &mcache.Cache{}
			if tc.withCache {
				devauth = devauth.WithCache(c)
				if tc.tenant != "" {
					c.On("DeleteToken",
						mock.MatchedBy(func(ctx context.Context) bool {
							ident := identity.FromContext(ctx)
							return assert.NotNil(t, ident) &&
								assert.Equal(t, tc.tenant, ident.Tenant)
						}),
						tc.tenant,
						tc.devId,
						cache.IdTypeDevice).
						Return(tc.cacheDeleteErr)
				} else {
					c.AssertNotCalled(t, "DeleteToken")
				}
			} else {
				c.AssertNotCalled(t, "DeleteToken")
			}

			err := devauth.DeleteAuthSet(ctx, tc.devId, tc.authId)

			if tc.outErr != "" {
				assert.EqualError(t, err, tc.outErr)
			} else {
				assert.NoError(t, err)
				if authSet.Status == model.DevStatusPreauth {
					db.AssertCalled(t, "DeleteDevice", tc.devId)
				} else {
					db.AssertNotCalled(t, "DeleteDevice", tc.devId)
				}
			}
			c.AssertExpectations(t)
		})
	}
}

func TestDeleteTokens(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tenantId string
		deviceId string

		cacheErr error

		dbErrDeleteTokenById error
		dbErrDeleteTokens    error

		outErr error
	}{
		"ok, all tenant's devs": {
			tenantId: "foo",
			deviceId: oid.NewUUIDv5("dev-foo").String(),
		},
		"ok, single dev": {
			tenantId: "foo",
		},
		"ok, single dev, token not found": {
			tenantId:             "foo",
			dbErrDeleteTokenById: store.ErrTokenNotFound,
		},
		"error, single dev": {
			tenantId:             "foo",
			deviceId:             oid.NewUUIDv5("dev-foo").String(),
			dbErrDeleteTokenById: errors.New("db error"),
			outErr: errors.Errorf(
				"failed to delete tokens for tenant: foo, "+
					"device id: %s: db error",
				oid.NewUUIDv5("dev-foo").String()),
		},
		"error, all tenant's devs": {
			tenantId:          "foo",
			dbErrDeleteTokens: errors.New("db error"),
			outErr:            errors.New("failed to delete tokens for tenant: foo, device id: : db error"),
		},
		"error(cache), all tenant's devs": {
			tenantId: "foo",
			cacheErr: errors.New("redis error"),
			outErr:   errors.New("failed to flush cache when cleaning tokens for tenant foo: redis error"),
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctxMatcher := mtesting.ContextMatcher()

			db := mstore.DataStore{}
			devUUID := oid.FromString(tc.deviceId)
			db.On("DeleteTokenByDevId", ctxMatcher, devUUID).
				Return(tc.dbErrDeleteTokenById)
			db.On("DeleteTokens", ctxMatcher).
				Return(tc.dbErrDeleteTokens)

			c := &mcache.Cache{}
			if tc.deviceId == "" {
				c.On("SuspendTenant", ctxMatcher, tc.tenantId).
					Return(tc.cacheErr)
			} else {
				c.AssertNotCalled(t, "FlushDB")
			}

			devauth := NewDevAuth(&db, nil, nil, Config{})
			devauth = devauth.WithCache(c)

			err := devauth.DeleteTokens(ctx, tc.tenantId, tc.deviceId)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTenantDeviceStatus(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tenantId string
		deviceId string

		dev                *model.Device
		dbGetDeviceByIdErr error

		outErr    error
		outStatus model.Status
	}{
		"ok": {
			tenantId: "foo",
			deviceId: "dev-foo",

			dev: &model.Device{
				Id:     "dev-foo",
				Status: model.DevStatusAccepted,
			},

			outStatus: model.Status{Status: "accepted"},
		},
		"error, not found": {
			tenantId: "foo",
			deviceId: "dev-foo",

			dbGetDeviceByIdErr: store.ErrDevNotFound,
			outErr:             ErrDeviceNotFound,
		},
		"error, generic": {
			tenantId: "foo",
			deviceId: "dev-foo",

			dbGetDeviceByIdErr: errors.New("get device error"),
			outErr:             errors.New("get device dev-foo failed: get device error"),
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(fmt.Sprintf("tc %s", n), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctxMatcher := mock.MatchedBy(func(c context.Context) bool {
				ident := identity.FromContext(c)
				assert.NotNil(t, ident)
				assert.Equal(t, tc.tenantId, ident.Tenant)

				return true
			})

			db := mstore.DataStore{}
			db.On("GetDeviceById",
				ctxMatcher,
				tc.deviceId,
			).Return(tc.dev, tc.dbGetDeviceByIdErr)

			devauth := NewDevAuth(&db, nil, nil, Config{})
			status, err := devauth.GetTenantDeviceStatus(ctx, tc.tenantId, tc.deviceId)

			if tc.outErr != nil {
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outStatus, *status)
			}
		})
	}
}

func TestApiLimitsOverride(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		base     ratelimits.ApiLimits
		override ratelimits.ApiLimits

		out ratelimits.ApiLimits
	}{
		{
			// some values over defaults - all overriden
			base: ratelimits.ApiLimits{
				ApiQuota:  ratelimits.ApiQuota{},
				ApiBursts: []ratelimits.ApiBurst{},
			},
			override: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
			out: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
		},
		{
			// defaults over some values - none overriden
			base: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
			override: ratelimits.ApiLimits{
				ApiQuota:  ratelimits.ApiQuota{},
				ApiBursts: []ratelimits.ApiBurst{},
			},
			out: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
		},
		{
			// override particular burst
			base: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    10,
					IntervalSec: 3600,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
			override: ratelimits.ApiLimits{
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 100,
					},
				},
			},
			out: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    10,
					IntervalSec: 3600,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 100,
					},
				},
			},
		},
		{
			// add burst
			base: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    10,
					IntervalSec: 3600,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
			override: ratelimits.ApiLimits{
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "PATCH",
						Uri:            "inventory/attributes",
						MinIntervalSec: 60,
					},
				},
			},
			out: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    10,
					IntervalSec: 3600,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
					{
						Action:         "PATCH",
						Uri:            "inventory/attributes",
						MinIntervalSec: 60,
					},
				},
			},
		},
		{
			// override and add burst
			base: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    10,
					IntervalSec: 3600,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 10,
					},
				},
			},
			override: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 100,
					},
					{
						Action:         "PATCH",
						Uri:            "inventory/attributes",
						MinIntervalSec: 60,
					},
				},
			},
			out: ratelimits.ApiLimits{
				ApiQuota: ratelimits.ApiQuota{
					MaxCalls:    100,
					IntervalSec: 60,
				},
				ApiBursts: []ratelimits.ApiBurst{
					{
						Action:         "POST",
						Uri:            "deployments/next",
						MinIntervalSec: 100,
					},
					{
						Action:         "PATCH",
						Uri:            "inventory/attributes",
						MinIntervalSec: 60,
					},
				},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(fmt.Sprintf("tc %d", n), func(t *testing.T) {
			t.Parallel()

			out := apiLimitsOverride(tc.base, tc.override)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestPurgeUriArgs(t *testing.T) {
	out := purgeUriArgs("/api/devices/v1/deployments/device/deployments/next?artifact_name=release-v1&device_type=foo")
	assert.Equal(t, "/api/devices/v1/deployments/device/deployments/next", out)

	out = purgeUriArgs("/api/devices/v1/deployments/device/deployments/next")
	assert.Equal(t, "/api/devices/v1/deployments/device/deployments/next", out)
}

func TestDevAuthGetDeviceWithCache(t *testing.T) {
	t.Parallel()

	devId := oid.NewUUIDv5("devId1").String()
	checkInTime := time.Now()

	testCases := []struct {
		devId  string
		tenant string

		dbGetDeviceById        *model.Device
		dbGetAuthSetsForDevice []model.AuthSet

		cacheGetCheckInTimeErr error
		cacheGetCheckInTime    *time.Time
	}{
		{
			devId:  devId,
			tenant: "foo",
			dbGetDeviceById: &model.Device{
				Id: devId,
			},
			dbGetAuthSetsForDevice: []model.AuthSet{
				{
					Id: "foo",
				},
			},
			cacheGetCheckInTimeErr: nil,
			cacheGetCheckInTime:    &checkInTime,
		},
		{
			devId:  devId,
			tenant: "foo",
			dbGetDeviceById: &model.Device{
				Id: devId,
			},
			dbGetAuthSetsForDevice: []model.AuthSet{
				{
					Id: "foo",
				},
			},
			cacheGetCheckInTimeErr: errors.New("foo"),
			cacheGetCheckInTime:    nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			db := mstore.DataStore{}
			db.On("GetDeviceById",
				ctx,
				tc.devId,
			).Return(tc.dbGetDeviceById, nil)
			db.On("GetAuthSetsForDevice",
				ctx,
				tc.devId,
			).Return(tc.dbGetAuthSetsForDevice, nil)
			defer db.AssertExpectations(t)

			devauth := NewDevAuth(&db, nil, nil, Config{})

			c := &mcache.Cache{}
			devauth = devauth.WithCache(c)
			c.On("GetCheckInTime",
				ctx,
				tc.tenant,
				tc.devId,
			).Return(tc.cacheGetCheckInTime, tc.cacheGetCheckInTimeErr)
			defer c.AssertExpectations(t)

			dev, err := devauth.GetDevice(ctx, tc.devId)
			assert.NoError(t, err)
			if tc.cacheGetCheckInTimeErr == nil {
				assert.WithinDuration(t, checkInTime, *(dev.CheckInTime), time.Nanosecond)
			} else {
				assert.Nil(t, dev.CheckInTime)
			}

		})
	}
}

func TestDevAuthGetDevicesWithCache(t *testing.T) {
	t.Parallel()

	devId := oid.NewUUIDv5("devId1").String()
	checkInTime := time.Now()

	testCases := []struct {
		devId  string
		tenant string

		dbGetDevices           []model.Device
		dbGetAuthSetsForDevice []model.AuthSet

		cacheGetCheckInTimesErr error
		cacheGetCheckInTimes    []*time.Time
	}{
		{
			devId:  devId,
			tenant: "foo",
			dbGetDevices: []model.Device{
				{
					Id: devId,
				},
			},
			dbGetAuthSetsForDevice: []model.AuthSet{
				{
					Id: "foo",
				},
			},
			cacheGetCheckInTimesErr: nil,
			cacheGetCheckInTimes:    []*time.Time{&checkInTime},
		},
		{
			devId:  devId,
			tenant: "foo",
			dbGetDevices: []model.Device{
				{
					Id: devId,
				},
			},
			dbGetAuthSetsForDevice: []model.AuthSet{
				{
					Id: "foo",
				},
			},
			cacheGetCheckInTimesErr: errors.New("foo"),
			cacheGetCheckInTimes:    nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			if tc.tenant != "" {
				id := &identity.Identity{
					Tenant: tc.tenant,
				}
				ctx = identity.WithContext(ctx, id)
			}

			db := mstore.DataStore{}
			db.On("GetDevices",
				ctx,
				uint(0),
				uint(10),
				model.DeviceFilter{
					IDs: []string{tc.devId},
				},
			).Return(tc.dbGetDevices, nil)
			db.On("GetAuthSetsForDevice",
				ctx,
				tc.devId,
			).Return(tc.dbGetAuthSetsForDevice, nil)
			defer db.AssertExpectations(t)

			devauth := NewDevAuth(&db, nil, nil, Config{})

			c := &mcache.Cache{}
			devauth = devauth.WithCache(c)
			c.On("GetCheckInTimes",
				ctx,
				tc.tenant,
				[]string{tc.devId},
			).Return(tc.cacheGetCheckInTimes, tc.cacheGetCheckInTimesErr)
			defer c.AssertExpectations(t)

			devs, err := devauth.GetDevices(
				ctx,
				uint(0),
				uint(10),
				model.DeviceFilter{IDs: []string{tc.devId}},
			)
			assert.NoError(t, err)
			assert.Len(t, devs, 1)
			if tc.cacheGetCheckInTimesErr == nil {
				assert.WithinDuration(t, checkInTime, *(devs[0].CheckInTime), time.Nanosecond)
			} else {
				assert.Nil(t, devs[0].CheckInTime)
			}

		})
	}
}
