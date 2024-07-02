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

package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/devauth"
	mdevauth "github.com/mendersoftware/mender-server/services/iot-manager/client/devauth/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore"
	coreMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore/mocks"
	wfMocks "github.com/mendersoftware/mender-server/services/iot-manager/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	storeMocks "github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
)

var (
	awsAccessKeyID      = "dummy"
	awsSecretAccessKey  = crypto.String("dummy")
	awsRegion           = "us-east-1"
	awsDevicePolicyName = `device-policy-name`
)

func statusPtr(s model.Status) *model.Status {
	return &s
}

func TestProvisionDeviceIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name     string
		DeviceID string

		Integration model.Integration

		Core func(t *testing.T, self *testCase) *coreMocks.Client
		Wf   func(t *testing.T, self *testCase) *wfMocks.Client

		Error error
	}
	awsEndpoint := "test_aws_endpoint"
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusEnabled,
					},
					awsDevicePolicyName).
					Return(&iotcore.Device{
						ID:          self.DeviceID,
						PrivateKey:  "private_key",
						Certificate: "certificate",
						Endpoint:    &awsEndpoint,
					}, nil)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyAWSPrivateKey:  "private_key",
						confKeyAWSCertificate: "certificate",
						confKeyAWSEndpoint:    awsEndpoint,
					}).Return(nil)
				return wf
			},
		},
		{
			Name:     "error, no credentials",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: ErrNoCredentials,
		},
		{
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusEnabled,
					},
					awsDevicePolicyName).
					Return(nil, errors.New("internal error"))
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: errors.New("failed to update iotcore devices: internal error"),
		},
		{
			Name:     "error, deviceconfig",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusEnabled,
					},
					awsDevicePolicyName).
					Return(&iotcore.Device{
						ID:          self.DeviceID,
						PrivateKey:  "private_key",
						Certificate: "certificate",
						Endpoint:    &awsEndpoint,
					}, nil)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyAWSPrivateKey:  "private_key",
						confKeyAWSCertificate: "certificate",
						confKeyAWSEndpoint:    awsEndpoint,
					}).Return(errors.New("internal error"))
				return wf
			},
			Error: errors.New("failed to submit iotcore credentials to deviceconfig: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			wf := tc.Wf(t, &tc)
			defer wf.AssertExpectations(t)

			a := New(ds, wf, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			a = a.WithIoTCore(core)

			err := a.(*app).provisionIoTCoreDevice(ctx,
				tc.DeviceID,
				tc.Integration,
				&iotcore.Device{
					Status: iotcore.StatusEnabled,
				},
			)

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

func TestDecommissionDeviceIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name        string
		DeviceID    string
		Integration model.Integration

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok, iot core",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("DeleteDevice", contextMatcher, mock.AnythingOfType("model.AWSCredentials"), self.DeviceID).
					Return(nil)
				return core
			},
		},
		{
			Name:     "error, no credentials",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
			},
			Error: ErrNoCredentials,
		},
		{
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("DeleteDevice", contextMatcher, mock.AnythingOfType("model.AWSCredentials"), self.DeviceID).
					Return(errors.New("failed to delete IoT Core device: store: unexpected error"))
				return core
			},

			Error: errors.New("failed to delete IoT Core device: store: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			a := New(ds, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			a = a.WithIoTCore(core)

			err := a.(*app).decommissionIoTCoreDevice(ctx, tc.DeviceID, tc.Integration)

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

func TestSetDeviceStatusIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name string

		DeviceID    string
		Status      model.Status
		Integration model.Integration

		Core func(t *testing.T, self *testCase) *coreMocks.Client

		Error error
	}
	awsEndpoint := "aws_endpoint"
	testCases := []testCase{
		{
			Name: "ok",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				dev := &iotcore.Device{
					ID:       "foobar",
					Status:   iotcore.StatusDisabled,
					Endpoint: &awsEndpoint,
				}
				core.On("UpsertDevice", contextMatcher, mock.AnythingOfType("model.AWSCredentials"), self.DeviceID,
					mock.MatchedBy(func(dev *iotcore.Device) bool {
						return dev.Status == iotcore.StatusEnabled
					}), awsDevicePolicyName).
					Return(dev, nil)
				return core
			},
		},
		{
			Name: "error, missing credentials",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
			},
			Error: ErrNoCredentials,
		},
		{
			Name: "error, fail to update device",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",
			Integration: model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice", contextMatcher, mock.AnythingOfType("model.AWSCredentials"), self.DeviceID,
					mock.MatchedBy(func(dev *iotcore.Device) bool {
						return dev.Status == iotcore.StatusEnabled
					}), awsDevicePolicyName).
					Return(nil, errors.New("failed to update IoT Hub device: hub: unexpected error"))
				return core
			},
			Error: errors.New("failed to update IoT Hub device: hub: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := new(storeMocks.DataStore)
			defer ds.AssertExpectations(t)

			a := New(ds, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			a = a.WithIoTCore(core)

			err := a.(*app).setDeviceStatusIoTCore(ctx, tc.DeviceID, tc.Status, tc.Integration)

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

func TestGetDeviceStateIoTCore(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name string

		DeviceID     string
		DeviceState  *model.DeviceState
		DeviceShadow *iotcore.DeviceShadow
		Integration  *model.Integration

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "1",
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
			DeviceShadow: &iotcore.DeviceShadow{
				Payload: model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"GetDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
				).Return(self.DeviceShadow, nil)
				return core
			},
		},
		{
			Name:     "ok, not found",
			DeviceID: "1",
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"GetDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
				).Return(self.DeviceShadow, iotcore.ErrDeviceNotFound)
				return core
			},
		},
		{
			Name:     "ko, some error",
			DeviceID: "1",
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"GetDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
				).Return(self.DeviceShadow, errors.New("get shadow error"))
				return core
			},
			Error: errors.New("get shadow error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()

			app := New(nil, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			app = app.WithIoTCore(core)

			state, err := app.GetDeviceStateIoTCore(ctx, tc.DeviceID, tc.Integration)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.DeviceState, state)
			}
		})
	}
}

func TestSetDeviceStateIoTCore(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name string

		DeviceID     string
		DeviceUpdate *model.DeviceState
		DeviceState  *model.DeviceState
		DeviceShadow *iotcore.DeviceShadow
		Integration  *model.Integration

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "1",
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			DeviceUpdate: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
			DeviceShadow: &iotcore.DeviceShadow{
				Payload: model.DeviceState{
					Desired: map[string]interface{}{
						"key": "value",
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"UpdateDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					iotcore.DeviceShadowUpdate{
						State: iotcore.DesiredState{
							Desired: self.DeviceUpdate.Desired,
						},
					},
				).Return(self.DeviceShadow, nil)
				return core
			},
		},
		{
			Name:     "ok, not found",
			DeviceID: "1",
			DeviceUpdate: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"UpdateDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					iotcore.DeviceShadowUpdate{
						State: iotcore.DesiredState{
							Desired: self.DeviceUpdate.Desired,
						},
					},
				).Return(nil, iotcore.ErrDeviceNotFound)
				return core
			},
		},
		{
			Name:     "ko, some error",
			DeviceID: "1",
			DeviceUpdate: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},
			Integration: &model.Integration{
				ID:       integrationID,
				Provider: model.ProviderIoTCore,
				Credentials: model.Credentials{
					Type: model.CredentialTypeAWS,
					AWSCredentials: &model.AWSCredentials{
						AccessKeyID:      &awsAccessKeyID,
						SecretAccessKey:  &awsSecretAccessKey,
						Region:           &awsRegion,
						DevicePolicyName: &awsDevicePolicyName,
					},
				},
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On(
					"UpdateDeviceShadow",
					contextMatcher,
					mock.AnythingOfType("model.AWSCredentials"),
					self.DeviceID,
					iotcore.DeviceShadowUpdate{
						State: iotcore.DesiredState{
							Desired: self.DeviceUpdate.Desired,
						},
					},
				).Return(self.DeviceShadow, errors.New("set shadow error"))
				return core
			},
			Error: errors.New("set shadow error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()

			app := New(nil, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			app = app.WithIoTCore(core)

			state, err := app.SetDeviceStateIoTCore(
				ctx,
				tc.DeviceID,
				tc.Integration,
				tc.DeviceUpdate,
			)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.DeviceState, state)
			}
		})
	}
}

func createSelfSignedCertificate(deviceID string) (cert []byte, private []byte) {
	pkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{
		NotAfter: time.Now().Add(time.Hour),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"TestingMcTestFace inc."},
			CommonName:   deviceID,
		},
		SerialNumber:       big.NewInt(1234),
		PublicKeyAlgorithm: x509.RSA,
		PublicKey:          &pkey.PublicKey,
	}
	cert, err = x509.CreateCertificate(rand.Reader, template, template, &pkey.PublicKey, pkey)
	if err != nil {
		panic(err)
	}
	private, _ = x509.MarshalPKCS8PrivateKey(pkey)
	return cert, private
}

func TestSyncIoTCoreDevices(t *testing.T) {
	t.Parallel()
	noLogger := log.NewEmpty()
	noLogger.Logger.Out = io.Discard
	type testDevice struct {
		ID            string
		CoreStatus    *iotcore.Status
		DevauthStatus *model.Status

		DeleteDeviceError error // decommisssionIoTCoreDevice
		UpsertDeviceError error
		GetDeviceError    error
	}
	type testCase struct {
		Name string

		Devices     []testDevice
		Integration model.Integration
		FailEarly   bool

		DataStore func(t *testing.T, self *testCase) *storeMocks.DataStore
		Devauth   func(t *testing.T, self *testCase) *mdevauth.Client
		Core      func(t *testing.T, self *testCase) *coreMocks.Client
		Wf        func(t *testing.T, self *testCase) *wfMocks.Client

		GetDevicesError error
		SaveEventError  error

		Error error
	}
	iotStatusPtr := func(s iotcore.Status) *iotcore.Status {
		return &s
	}
	testCases := []testCase{{
		Name: "ok/10 devices in all cases",

		Devices: []testDevice{{
			ID:            "38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    nil,
		}, {
			ID:            "72334767-ff25-48ef-ae10-9dcf4f98587d",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),
		}, {
			ID:            "1280cb45-e941-47fb-922e-8dc55006d127",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),
		}, {
			ID:            "6b7ed385-91ca-4499-a118-3e6b863a9082",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),
		}, {
			ID:            "4e8e5b20-5558-486c-891c-41e3a4d309a4",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusDisabled),
		}, {
			ID:            "49900bc3-9f2b-4b84-ad0d-bec7313b866b",
			DevauthStatus: statusPtr(model.StatusRejected),
			CoreStatus:    iotStatusPtr(iotcore.StatusDisabled),
		}, {
			ID:            "3146cc4d-21eb-4f67-bdb8-96e3222b1b4b",
			DevauthStatus: statusPtr(model.StatusRejected),
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),
		}, {
			ID:            "02d9ab3e-ca1c-4a61-bf06-b23a224935d4",
			DevauthStatus: statusPtr(model.StatusNoAuth),
			CoreStatus:    iotStatusPtr(iotcore.StatusDisabled),
		}, {
			ID:            "a4a32db1-047d-4b4b-9f4a-b86a6c16ab90",
			DevauthStatus: nil,
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),
		}, {
			ID:            "1434a240-e556-4acf-b96d-ac66a20f82de",
			DevauthStatus: nil,
			CoreStatus:    nil,
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
	}, {
		Name: "error/invalid credentials",

		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeSAS, // NOTE Invalid for provider
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
		Error: ErrNoCredentials,
	}, {
		Name: "error/from device auth",

		Devices: []testDevice{},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
		GetDevicesError: errors.New("internal error"),
		Error:           errors.New("internal error"),
	}, {
		Name: "error/deleting device from IoT Core",

		FailEarly: true,

		Devices: []testDevice{{
			ID:            "a4a32db1-047d-4b4b-9f4a-b86a6c16ab90",
			DevauthStatus: nil,
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),

			DeleteDeviceError: errors.New("internal error"),
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
	}, {
		Name: "error/SaveEvent/deleting device from IoT Core",

		FailEarly: true,

		Devices: []testDevice{{
			ID:            "a4a32db1-047d-4b4b-9f4a-b86a6c16ab90",
			DevauthStatus: nil,
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),

			DeleteDeviceError: errors.New("internal error"),
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},

		SaveEventError: errors.New("internal error"),
		Error:          errors.New("internal error"),
	}, {
		Name: "error/provisioning device to IoT Core",

		FailEarly: true,

		Devices: []testDevice{{
			ID:            "38e5ebfb-963d-4ac2-8f5e-d51b2df1fa6e",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    nil,

			UpsertDeviceError: errors.New("internal error"),
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
		Error: errors.New("internal error"),
	}, {
		Name: "error/retrieving device from IoT Core",

		FailEarly: true,

		Devices: []testDevice{{
			ID:            "72334767-ff25-48ef-ae10-9dcf4f98587d",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusEnabled),

			GetDeviceError: errors.New("internal error"),
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
		Error: errors.New("internal error"),
	}, {
		Name: "error/updating IoT Core device status",

		FailEarly: true,

		Devices: []testDevice{{
			ID:            "72334767-ff25-48ef-ae10-9dcf4f98587d",
			DevauthStatus: statusPtr(model.StatusAccepted),
			CoreStatus:    iotStatusPtr(iotcore.StatusDisabled),

			UpsertDeviceError: errors.New("internal error"),
		}},
		Integration: model.Integration{
			ID:       uuid.New(),
			Provider: model.ProviderIoTCore,
			Credentials: model.Credentials{
				Type: model.CredentialTypeAWS,
				AWSCredentials: &model.AWSCredentials{
					AccessKeyID:      &awsAccessKeyID,
					SecretAccessKey:  &awsSecretAccessKey,
					Region:           &awsRegion,
					DevicePolicyName: &awsDevicePolicyName,
				},
			},
		},
		Error: errors.New("internal error"),
	}}
	matchConf := func(cert, pkey, endpoint string) func(map[string]string) bool {
		return func(m map[string]string) bool {
			return assert.Equal(t, map[string]string{
				confKeyAWSCertificate: cert,
				confKeyAWSPrivateKey:  pkey,
				confKeyAWSEndpoint:    endpoint,
			}, m)
		}
	}
	matchDev := func(expected iotcore.Device) func(*iotcore.Device) bool {
		return func(actual *iotcore.Device) bool {
			return actual != nil && reflect.DeepEqual(expected, *actual)
		}
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := log.WithContext(context.Background(), noLogger)

			ds := new(storeMocks.DataStore)
			da := new(mdevauth.Client)
			wf := new(wfMocks.Client)
			core := new(coreMocks.Client)

			defer da.AssertExpectations(t)
			defer ds.AssertExpectations(t)
			defer wf.AssertExpectations(t)
			defer core.AssertExpectations(t)

			authSets := make([]devauth.Device, 0, len(tc.Devices))

			// Initialize mock assertions
			for _, dev := range tc.Devices {
				if tc.GetDevicesError != nil {
					break
				}
				if dev.DevauthStatus != nil {
					authSets = append(authSets, devauth.Device{
						ID:     dev.ID,
						Status: model.Status(*dev.DevauthStatus),
					})
					awsEndpoint := "test_aws_endpoint"
					// Generate a random "Thing" identity
					cert, pkey := createSelfSignedCertificate(dev.ID)
					iotDev := iotcore.Device{
						ID:            dev.ID,
						Name:          dev.ID,
						CertificateID: uuid.NewString(),
						Certificate:   string(cert),
						PrivateKey:    string(pkey),
						Endpoint:      &awsEndpoint,
					}
					if dev.CoreStatus != nil {
						iotDev.Status = *dev.CoreStatus
						core.On("GetDevice",
							contextMatcher,
							*tc.Integration.Credentials.AWSCredentials,
							dev.ID).
							Return(&iotDev, dev.GetDeviceError).
							Once()
						if dev.GetDeviceError != nil {
							break
						}
						desiredStatus := iotcore.NewStatusFromMenderStatus(*dev.DevauthStatus)
						desiredDev := iotDev
						desiredDev.Status = desiredStatus
						desiredDev.Endpoint = &awsEndpoint
						if *dev.CoreStatus != desiredStatus {
							// Status mismatch
							core.On("UpsertDevice",
								contextMatcher,
								*tc.Integration.Credentials.
									AWSCredentials,
								dev.ID,
								mock.MatchedBy(matchDev(iotcore.Device{
									Status: desiredStatus,
								})),
								*tc.Integration.Credentials.
									AWSCredentials.
									DevicePolicyName).
								Return(&desiredDev, dev.UpsertDeviceError).
								Once()
						}
					} else {
						iotDev.Status = iotcore.NewStatusFromMenderStatus(*dev.DevauthStatus)
						iotDev.Endpoint = &awsEndpoint
						// Provision device
						core.On("GetDevice",
							contextMatcher,
							*tc.Integration.Credentials.AWSCredentials,
							dev.ID).
							Return(nil, iotcore.ErrDeviceNotFound).
							Once()
						core.On("UpsertDevice",
							contextMatcher,
							*tc.Integration.Credentials.AWSCredentials,
							dev.ID,
							mock.AnythingOfType("*iotcore.Device"),
							mock.AnythingOfType("string")).
							Return(&iotDev, dev.UpsertDeviceError).
							Once()
						if dev.UpsertDeviceError == nil {
							wf.On("ProvisionExternalDevice",
								contextMatcher,
								dev.ID,
								mock.MatchedBy(matchConf(iotDev.Certificate, iotDev.PrivateKey, *iotDev.Endpoint))).
								Return(nil).
								Once()
						}
					}
				} else {
					// Decommission device
					ds.On("GetDevice",
						contextMatcher,
						dev.ID).
						Return(&model.Device{
							ID:             dev.ID,
							IntegrationIDs: []uuid.UUID{tc.Integration.ID},
						}, nil).
						Once()

					ds.On("GetIntegrations",
						contextMatcher,
						model.IntegrationFilter{}).
						Return([]model.Integration{tc.Integration}, nil).
						Once()

					deviceID := dev.ID
					// mock.MatchedBy function is executed twice for some reason
					ds.On("SaveEvent",
						contextMatcher,
						mock.MatchedBy(func(actual model.Event) bool {
							ret := model.EventTypeDeviceDecommissioned == actual.Type
							if ret {
								_, ret = actual.Data.(model.DeviceEvent)
							}
							if ret {
								ret = deviceID == actual.Data.(model.DeviceEvent).ID
							}
							return ret
						})).
						Return(tc.SaveEventError).
						Once()

					var mockErr error = dev.DeleteDeviceError
					if dev.CoreStatus == nil {
						mockErr = iotcore.ErrDeviceNotFound
					}
					core.On("DeleteDevice",
						contextMatcher,
						*tc.Integration.Credentials.AWSCredentials,
						dev.ID).
						Return(mockErr).
						Once()

					ds.On("DeleteDevice",
						contextMatcher,
						dev.ID).
						Return(nil).
						Once()

				}
				if tc.SaveEventError != nil {
					break
				}
			}

			deviceIDs := make([]string, len(tc.Devices))
			for i, dev := range tc.Devices {
				deviceIDs[i] = dev.ID
			}

			if tc.Devices != nil {
				da.On("GetDevices", contextMatcher, deviceIDs).
					Return(authSets, tc.GetDevicesError)
			}
			app := New(ds, wf, da).WithIoTCore(core).(*app)
			err := app.syncIoTCoreDevices(ctx, deviceIDs, tc.Integration, tc.FailEarly)
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
