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

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func (d *Deployments) isAlreadyInstalled(
	request *model.DeploymentNextRequest,
	deviceDeployment *model.DeviceDeployment,
) bool {
	if request == nil ||
		request.DeviceProvides == nil ||
		deviceDeployment == nil ||
		deviceDeployment.Image == nil ||
		deviceDeployment.Image.ArtifactMeta == nil {
		return false
	}

	// check if the device reported same artifact name as the one
	// in the artifact selected for a given device deployment
	return request.DeviceProvides.ArtifactName == deviceDeployment.Image.ArtifactMeta.Name
}

func (d *Deployments) handleAlreadyInstalled(
	ctx context.Context,
	deviceDeployment *model.DeviceDeployment,
) error {
	l := log.FromContext(ctx)
	if err := d.UpdateDeviceDeploymentStatus(
		ctx,
		deviceDeployment.DeploymentId,
		deviceDeployment.DeviceId,
		model.DeviceDeploymentState{
			Status: model.DeviceDeploymentStatusAlreadyInst,
		}); err != nil {
		return errors.Wrap(err, "Failed to update deployment status")
	}
	if err := d.reindexDevice(ctx, deviceDeployment.DeviceId); err != nil {
		l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
	}
	if err := d.reindexDeployment(ctx, deviceDeployment.DeviceId,
		deviceDeployment.DeploymentId, deviceDeployment.Id); err != nil {
		l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
	}

	return nil
}

// assignArtifact assigns artifact to the device deployment
func (d *Deployments) assignArtifact(
	ctx context.Context,
	deployment *model.Deployment,
	deviceDeployment *model.DeviceDeployment,
	installed *model.InstalledDeviceDeployment) error {

	// Assign artifact to the device deployment.
	var artifact *model.Image
	var err error

	if err = installed.Validate(); err != nil {
		return err
	}

	if deviceDeployment.DeploymentId == "" || deviceDeployment.DeviceId == "" {
		return ErrModelInternal
	}

	// Clear device deployment image
	// New artifact will be selected for the device deployment
	deviceDeployment.Image = nil

	// First case is for backward compatibility.
	// It is possible that there is old deployment structure in the system.
	// In such case we need to select artifact using name and device type.
	if deployment.Artifacts == nil || len(deployment.Artifacts) == 0 {
		artifact, err = d.db.ImageByNameAndDeviceType(
			ctx,
			installed.ArtifactName,
			installed.DeviceType,
		)
		if err != nil {
			return errors.Wrap(err, "assigning artifact to device deployment")
		}
	} else {
		// Select artifact for the device deployment from artifacts assigned to the deployment.
		artifact, err = d.db.ImageByIdsAndDeviceType(
			ctx,
			deployment.Artifacts,
			installed.DeviceType,
		)
		if err != nil {
			return errors.Wrap(err, "assigning artifact to device deployment")
		}
	}

	// If not having appropriate image, set noartifact status
	if artifact == nil {
		return d.assignNoArtifact(ctx, deviceDeployment)
	}

	if err := d.db.AssignArtifact(
		ctx,
		deviceDeployment.DeviceId,
		deviceDeployment.DeploymentId,
		artifact,
	); err != nil {
		return errors.Wrap(err, "Assigning artifact to the device deployment")
	}

	deviceDeployment.Image = artifact

	return nil
}

func (d *Deployments) assignNoArtifact(
	ctx context.Context,
	deviceDeployment *model.DeviceDeployment,
) error {
	l := log.FromContext(ctx)
	if err := d.UpdateDeviceDeploymentStatus(ctx, deviceDeployment.DeploymentId,
		deviceDeployment.DeviceId,
		model.DeviceDeploymentState{
			Status: model.DeviceDeploymentStatusNoArtifact,
		}); err != nil {
		return errors.Wrap(err, "Failed to update deployment status")
	}
	if err := d.reindexDevice(ctx, deviceDeployment.DeviceId); err != nil {
		l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
	}
	if err := d.reindexDeployment(ctx, deviceDeployment.DeviceId,
		deviceDeployment.DeploymentId, deviceDeployment.Id); err != nil {
		l := log.FromContext(ctx)
		l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
	}
	return nil
}
