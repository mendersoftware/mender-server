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

package iotcore

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

func init() {
	model.SetTrustedHostnames([]string{"*.iot.*.amazonaws.com", "localhost"})
}

var (
	accessKeyID         string
	secretAccessKey     string
	awsRegion           string
	awsDevicePolicyName string
	awsCredentials      = model.AWSCredentials{}
)

func init() {
	flag.StringVar(&accessKeyID,
		"test.aws-access-key-id",
		"",
		"AWS Access Key ID (overwrite with env var TEST_AWS_ACCESS_KEY_ID).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_ACCESS_KEY_ID"); ok && val != "" {
		accessKeyID = val
	}
	flag.StringVar(&secretAccessKey,
		"test.aws-secret-access-key",
		"",
		"AWS Secret Access Key (overwrite with env var TEST_AWS_SECRET_ACCESS_KEY).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_SECRET_ACCESS_KEY"); ok && val != "" {
		secretAccessKey = val
	}
	flag.StringVar(&awsRegion,
		"test.aws-region",
		"",
		"AWS IoT Core region (overwrite with env var TEST_AWS_REGION).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_REGION"); ok && val != "" {
		awsRegion = val
	}

	flag.StringVar(&awsDevicePolicyName,
		"test.aws-device-policy-name",
		"",
		"AWS IoT Core device policy name (overwrite with env var TEST_AWS_DEVICE_POLICY_NAME).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_DEVICE_POLICY_NAME"); ok && val != "" {
		awsDevicePolicyName = val
	}

	testing.Init()

	awsCredentials.AccessKeyID = &accessKeyID
	awsCredentials.SecretAccessKey = (*crypto.String)(&secretAccessKey)
	awsCredentials.Region = &awsRegion
	awsCredentials.DevicePolicyName = &awsDevicePolicyName
}

func validAWSSettings(t *testing.T) bool {
	if accessKeyID == "" || secretAccessKey == "" || awsRegion == "" || awsDevicePolicyName == "" {
		t.Skip("AWS settings not provided or invalid")
		return false
	}
	return true
}

func TestGetDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()
	_, err := client.UpsertDevice(ctx, awsCredentials, deviceID, &Device{}, awsDevicePolicyName)
	assert.NoError(t, err)

	device, err := client.GetDevice(ctx, awsCredentials, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, device)

	assert.Equal(t, deviceID, device.Name)

	_, err = client.GetDevice(ctx, awsCredentials, "dummy")
	assert.EqualError(t, err, ErrDeviceNotFound.Error())

	err = client.DeleteDevice(ctx, awsCredentials, device.Name)
	assert.NoError(t, err)

	device, err = client.GetDevice(ctx, awsCredentials, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, device)
}

func TestDeleteDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()

	device, err := client.UpsertDevice(ctx, awsCredentials, deviceID, &Device{}, awsDevicePolicyName)
	assert.NoError(t, err)

	err = client.DeleteDevice(ctx, awsCredentials, device.Name)
	assert.NoError(t, err)

	err = client.DeleteDevice(ctx, awsCredentials, device.Name)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())

	device, err = client.GetDevice(ctx, awsCredentials, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, device)
}

func TestUpsertDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()
	device, err := client.UpsertDevice(ctx, awsCredentials, deviceID, &Device{
		Status: StatusDisabled,
	}, awsDevicePolicyName)
	assert.NoError(t, err)
	assert.Equal(t, StatusDisabled, device.Status)

	assert.NotEmpty(t, device.ID)
	assert.NotEmpty(t, device.PrivateKey)
	assert.NotEmpty(t, device.Certificate)
	assert.NotEmpty(t, device.Endpoint)

	device.Status = StatusEnabled
	device, err = client.UpsertDevice(ctx, awsCredentials, deviceID, device, awsDevicePolicyName)
	assert.NoError(t, err)
	assert.Equal(t, StatusEnabled, device.Status)

	err = client.DeleteDevice(ctx, awsCredentials, deviceID)
	assert.NoError(t, err)
}

func TestIoTCoreExternal(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()

	// no device
	shadow, err := client.GetDeviceShadow(ctx, awsCredentials, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, shadow)

	_, err = client.UpsertDevice(ctx, awsCredentials, deviceID, &Device{}, awsDevicePolicyName)
	assert.NoError(t, err)

	device, err := client.GetDevice(ctx, awsCredentials, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, device)

	// no shadow set in IoT Core, it returns an empty shadow
	shadow, err = client.GetDeviceShadow(ctx, awsCredentials, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, shadow, &DeviceShadow{
		Payload: model.DeviceState{
			Desired:  map[string]interface{}{},
			Reported: map[string]interface{}{},
		},
	})

	// update shadow
	update := DeviceShadowUpdate{
		State: DesiredState{
			Desired: map[string]interface{}{
				"foo": "bar",
			},
		},
	}
	updatedShadow, err := client.UpdateDeviceShadow(ctx, awsCredentials, deviceID, update)
	assert.NoError(t, err)
	assert.NotNil(t, updatedShadow)

	// get shadow and compare with update result
	shadow, err = client.GetDeviceShadow(ctx, awsCredentials, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, updatedShadow, shadow)

	err = client.DeleteDevice(ctx, awsCredentials, device.Name)
	assert.NoError(t, err)

	device, err = client.GetDevice(ctx, awsCredentials, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, device)
}
