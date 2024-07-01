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

package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/iot-manager/client"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

const (
	iotHubMetadata = "$metadata"
	iotHubVersion  = "$version"
)

func removeIoTHubMetadata(values map[string]interface{}) map[string]interface{} {
	for key := range values {
		switch key {
		case iotHubMetadata, iotHubVersion:
			delete(values, key)
		}
	}
	return values
}

func (a *app) provisionIoTHubDevice(
	ctx context.Context,
	deviceID string,
	integration model.Integration,
	deviceUpdate ...*iothub.Device,
) error {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return ErrNoCredentials
	}

	dev, err := a.iothubClient.UpsertDevice(ctx, cs, deviceID, deviceUpdate...)
	if err != nil {
		if htErr, ok := err.(client.HTTPError); ok {
			switch htErr.Code() {
			case http.StatusUnauthorized:
				return ErrNoCredentials
			case http.StatusConflict:
				return ErrDeviceAlreadyExists
			}
		}
		return errors.Wrap(err, "failed to update iothub devices")
	}
	if dev.Auth == nil || dev.Auth.SymmetricKey == nil {
		return ErrNoDeviceConnectionString
	}
	primKey := &model.ConnectionString{
		Key:      crypto.String(dev.Auth.SymmetricKey.Primary),
		DeviceID: dev.DeviceID,
		HostName: cs.HostName,
	}

	err = a.wf.ProvisionExternalDevice(ctx, dev.DeviceID, map[string]string{
		confKeyPrimaryKey: primKey.String(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to submit iothub authn to deviceconfig")
	}
	err = a.iothubClient.UpdateDeviceTwin(ctx, cs, dev.DeviceID, &iothub.DeviceTwinUpdate{
		Tags: map[string]interface{}{
			"mender": true,
		},
	})
	return errors.Wrap(err, "failed to tag provisioned iothub device")
}

func (a *app) setDeviceStatusIoTHub(ctx context.Context, deviceID string, status model.Status,
	integration model.Integration) error {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return ErrNoCredentials
	}
	azureStatus := iothub.NewStatusFromMenderStatus(status)
	dev, err := a.iothubClient.GetDevice(ctx, cs, deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve device from IoT Hub")
	} else if dev.Status == azureStatus {
		// We're done...
		return nil
	}

	dev.Status = azureStatus
	_, err = a.iothubClient.UpsertDevice(ctx, cs, deviceID, dev)
	return err
}

func (a *app) decommissionIoTHubDevice(ctx context.Context, deviceID string,
	integration model.Integration) error {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return ErrNoCredentials
	}
	err := a.iothubClient.DeleteDevice(ctx, cs, deviceID)
	if err != nil {
		if htErr, ok := err.(client.HTTPError); ok &&
			htErr.Code() == http.StatusNotFound {
			return nil
		}
		return errors.Wrap(err, "failed to delete IoT Hub device")
	}
	return nil
}

func (a *app) syncIoTHubDevices(
	ctx context.Context,
	deviceIDs []string,
	integration model.Integration,
	failEarly bool,
) error {
	l := log.FromContext(ctx)
	cs := integration.Credentials.ConnectionString

	// Get device authentication
	devAuths, err := a.devauth.GetDevices(ctx, deviceIDs)
	if err != nil {
		return errors.Wrap(err, "app: failed to lookup device authentication")
	}

	statuses := make(map[string]iothub.Status, len(deviceIDs))
	for _, auth := range devAuths {
		statuses[auth.ID] = iothub.NewStatusFromMenderStatus(auth.Status)
	}
	// Find devices that shouldn't exist
	var (
		i int
		j int = len(deviceIDs)
	)
	for i < j {
		id := deviceIDs[i]
		if _, ok := statuses[id]; !ok {
			l.Warnf("Device '%s' does not have an auth set: deleting device", id)
			err := a.decommissionDevice(ctx, id)
			if err != nil && err != ErrDeviceNotFound {
				err = errors.Wrap(err, "app: failed to decommission device")
				if failEarly {
					return err
				}
				l.Error(err)
			}
			// swap(deviceIDs[i], deviceIDs[j])
			j--
			tmp := deviceIDs[i]
			deviceIDs[i] = deviceIDs[j]
			deviceIDs[j] = tmp
		} else {
			i++
		}
	}

	// Fetch IoT Hub device twins
	hubDevs, err := a.iothubClient.GetDeviceTwins(ctx, cs, deviceIDs[:j])
	if err != nil {
		return errors.Wrap(err, "app: failed to get devices from IoT Hub")
	}

	// Set of device IDs in iot hub
	devicesInHub := make(map[string]struct{}, len(hubDevs))

	// Check if devices (statuses) are in sync
	for _, twin := range hubDevs {
		devicesInHub[twin.DeviceID] = struct{}{}
		if stat, ok := statuses[twin.DeviceID]; ok {
			if stat == twin.Status {
				continue
			}
			l.Warnf("Device '%s' status does not match Mender auth status, updating status",
				twin.DeviceID)
			// Update the device's status
			// NOTE need to fetch device identity first
			dev, err := a.iothubClient.GetDevice(ctx, cs, twin.DeviceID)
			if err != nil {
				err = errors.Wrap(err, "failed to retrieve IoT Hub device identity")
				if failEarly {
					return err
				}
				l.Error(err)
				continue
			}
			dev.Status = stat
			_, err = a.iothubClient.UpsertDevice(ctx, cs, twin.DeviceID, dev)
			if err != nil {
				err = errors.Wrap(err, "failed to update IoT Hub device identity")
				if failEarly {
					return err
				}
				l.Error(err)
			}
		}
	}

	// Find devices not present in IoT Hub
	for id, status := range statuses {
		if _, ok := devicesInHub[id]; !ok {
			l.Warnf("Found device not existing in IoT Hub '%s': provisioning device", id)
			// Device inconsistency
			// Device exist in Mender but not in IoT Hub
			err := a.provisionIoTHubDevice(ctx, id, integration, &iothub.Device{
				DeviceID: id,
				Status:   status,
			})
			if err != nil {
				if failEarly {
					return err
				}
				l.Error(err)
				continue
			}
		}
	}
	return nil
}

func (a *app) GetDeviceStateIoTHub(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
) (*model.DeviceState, error) {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return nil, ErrNoCredentials
	}
	twin, err := a.iothubClient.GetDeviceTwin(ctx, cs, deviceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the device twin")
	}
	return &model.DeviceState{
		Desired:  removeIoTHubMetadata(twin.Properties.Desired),
		Reported: removeIoTHubMetadata(twin.Properties.Reported),
	}, nil
}

func (a *app) SetDeviceStateIoTHub(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
	state *model.DeviceState,
) (*model.DeviceState, error) {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return nil, ErrNoCredentials
	}
	twin, err := a.iothubClient.GetDeviceTwin(ctx, cs, deviceID)
	if err == nil {
		update := &iothub.DeviceTwinUpdate{
			Tags: twin.Tags,
			Properties: iothub.UpdateProperties{
				Desired: state.Desired,
			},
			ETag:    twin.ETag,
			Replace: true,
		}
		err = a.iothubClient.UpdateDeviceTwin(ctx, cs, deviceID, update)
	}
	if errHTTP, ok := err.(client.HTTPError); ok &&
		errHTTP.Code() == http.StatusPreconditionFailed {
		return nil, ErrDeviceStateConflict
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to update the device twin")
	}
	return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
}

func (app *app) VerifyDeviceTwin(ctx context.Context, req model.PreauthRequest) error {
	integrations, err := app.GetIntegrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve integration: %w", err)
	}
	var integration model.Integration
	for _, integration = range integrations {
		if integration.Provider == model.ProviderIoTHub {
			break
		}
	}
	deviceModule := strings.SplitN(req.DeviceID, "/", 2)
	for i := range deviceModule {
		deviceModule[i] = url.PathEscape(deviceModule[i])
	}
	id := strings.Join(deviceModule, "/modules/")
	log.FromContext(ctx).Debugf("getting twin: %s", id)
	twin, err := app.iothubClient.GetDeviceTwin(
		ctx, integration.Credentials.ConnectionString, id,
	)
	if err != nil {
		return fmt.Errorf("failed to get module twin from integration: %w", err)
	}
	idData, ok := twin.Properties.Reported["id_data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing identity data")
	}
	if !reflect.DeepEqual(idData, req.IdentityData) {
		return fmt.Errorf(`reported "id_data" does not match request`)
	}
	pubkeyPEM, ok := twin.Properties.Reported["pubkey"].(string)
	if !ok {
		return fmt.Errorf("missing pubkey")
	}
	var pubkey model.PublicKey
	err = pubkey.UnmarshalText([]byte(pubkeyPEM))
	if err != nil {
		return fmt.Errorf(
			"invalid public key from twin: %w",
			err,
		)
	}
	if !pubkey.PublicKey.Equal(req.PublicKey.PublicKey) {
		return fmt.Errorf("key does not match")
	}
	return nil
}
