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

package app

import (
	"context"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

const (
	MaxDeviceArrayLength = 1024
)

var (
	ErrNoIdsGiven  = errors.New("need at least one device id")
	ErrArrayTooBig = errors.New("too many device ids given")
)

// GetDeviceDeploymentLastStatus returns the array of last device deployment statuses.
func (d *Deployments) GetDeviceDeploymentLastStatus(
	ctx context.Context,
	devicesIds []string,
) (model.DeviceDeploymentLastStatuses, error) {
	length := len(devicesIds)
	if length < 1 {
		return model.DeviceDeploymentLastStatuses{
			DeviceDeploymentLastStatuses: []model.DeviceDeploymentLastStatus{},
		}, ErrNoIdsGiven
	}
	if length > MaxDeviceArrayLength {
		return model.DeviceDeploymentLastStatuses{
			DeviceDeploymentLastStatuses: []model.DeviceDeploymentLastStatus{},
		}, ErrArrayTooBig
	}

	statuses, err := d.db.GetLastDeviceDeploymentStatus(ctx, devicesIds)
	if len(statuses) < 1 {
		statuses = []model.DeviceDeploymentLastStatus{}
	}
	return model.DeviceDeploymentLastStatuses{
		DeviceDeploymentLastStatuses: statuses,
	}, err
}
