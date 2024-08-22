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

	"github.com/mendersoftware/mender-server/services/iot-manager/model"

	"github.com/google/uuid"
)

type deviceGetter interface {
	GetDevice(context.Context, string) (*model.Device, error)
}

// device provides an interface to lazily load the device from the database
// only when required.
type device struct {
	m            map[uuid.UUID]struct{}
	err          error
	DeviceID     string
	DeviceGetter deviceGetter
}

func newDevice(deviceID string, deviceGetter deviceGetter) *device {
	return &device{
		DeviceID:     deviceID,
		DeviceGetter: deviceGetter,
	}
}

func (m *device) HasIntegration(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.m == nil {
		dev, err := m.DeviceGetter.GetDevice(ctx, m.DeviceID)
		if err != nil {
			m.err = err
			return false, m.err
		}
		m.m = make(map[uuid.UUID]struct{}, len(dev.IntegrationIDs))
		for _, iid := range dev.IntegrationIDs {
			m.m[iid] = struct{}{}
		}
	}
	_, ret := m.m[id]
	return ret, nil
}
