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

package indexer

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/client/deployments"
	"github.com/mendersoftware/mender-server/services/reporting/client/deviceauth"
	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	rconfig "github.com/mendersoftware/mender-server/services/reporting/config"
	"github.com/mendersoftware/mender-server/services/reporting/model"
)

const undefinedCoordinateIdx = -1

type IDs map[string]bool
type ActionIDs map[string]IDs
type TenantActionIDs map[string]ActionIDs

func (i *indexer) GetJobs(ctx context.Context, jobs chan model.Job) error {
	streamName := config.Config.GetString(rconfig.SettingNatsStreamName)

	topic := config.Config.GetString(rconfig.SettingNatsSubscriberTopic)
	subject := streamName + "." + topic
	durableName := config.Config.GetString(rconfig.SettingNatsSubscriberDurable)

	err := i.nats.JetStreamSubscribe(ctx, subject, durableName, jobs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to the nats JetStream")
	}

	return nil
}

func (i *indexer) ProcessJobs(ctx context.Context, jobs []model.Job) {
	l := log.FromContext(ctx)
	l.Debugf("Processing %d jobs", len(jobs))
	tenantsActionIDs := groupJobsIntoTenantActionIDs(jobs)
	for tenant, actionIDs := range tenantsActionIDs {
		for action, IDs := range actionIDs {
			if action == model.ActionReindex {
				i.processJobDevices(ctx, tenant, IDs)
			} else if action == model.ActionReindexDeployment {
				i.processJobDeployments(ctx, tenant, IDs)
			} else {
				l.Warnf("ignoring unknown job action: %v", action)
			}
		}
	}
}

func (i *indexer) processJobDevices(
	ctx context.Context,
	tenant string,
	IDs IDs,
) {
	l := log.FromContext(ctx)
	devices := make([]*model.Device, 0, len(IDs))
	removedDevices := make([]*model.Device, 0, len(IDs))

	deviceIDs := make([]string, 0, len(IDs))
	for deviceID := range IDs {
		deviceIDs = append(deviceIDs, deviceID)
	}
	// get devices from deviceauth
	deviceAuthDevices, err := i.devClient.GetDevices(ctx, tenant, deviceIDs)
	if err != nil {
		l.Error(errors.Wrap(err, "failed to get devices from deviceauth"))
		return
	}
	// get devices from inventory
	inventoryDevices, err := i.invClient.GetDevices(ctx, tenant, deviceIDs)
	if err != nil {
		l.Error(errors.Wrap(err, "failed to get devices from inventory"))
		return
	}
	// get last deployment statuses from deployment
	deploymentsDevices, err := i.deplClient.GetLatestFinishedDeployment(ctx, tenant, deviceIDs)
	if err != nil {
		l.Error(errors.Wrap(err, "failed to get last device deployments from deployments"))
		return
	}

	// process the results
	devices = devices[:0]
	removedDevices = removedDevices[:0]
	for _, deviceID := range deviceIDs {
		var deviceAuthDevice *deviceauth.DeviceAuthDevice
		var inventoryDevice *inventory.Device
		var deploymentsDevice *deployments.LastDeviceDeployment
		if d, ok := deviceAuthDevices[deviceID]; ok {
			deviceAuthDevice = &d
		}
		for _, d := range inventoryDevices {
			if d.ID == inventory.DeviceID(deviceID) {
				inventoryDevice = &d
				break
			}
		}
		for _, d := range deploymentsDevices {
			if d.DeviceID == deviceID {
				deploymentsDevice = &d
				break
			}
		}
		if deviceAuthDevice == nil || inventoryDevice == nil {
			removedDevices = append(removedDevices, &model.Device{
				ID:       &deviceID,
				TenantID: &tenant,
			})
			continue
		}
		device := i.processJobDevice(
			ctx,
			tenant,
			deviceAuthDevice,
			inventoryDevice,
			deploymentsDevice,
		)
		if device != nil {
			devices = append(devices, device)
		}
	}
	// bulk index the device
	if len(devices) > 0 || len(removedDevices) > 0 {
		err = i.store.BulkIndexDevices(ctx, devices, removedDevices)
		if err != nil {
			err = errors.Wrap(err, "failed to bulk index the devices")
			l.Error(err)
		}
	}
}

func (i *indexer) processJobDevice(
	ctx context.Context,
	tenant string,
	deviceAuthDevice *deviceauth.DeviceAuthDevice,
	inventoryDevice *inventory.Device,
	deploymentsDevice *deployments.LastDeviceDeployment,
) *model.Device {
	l := log.FromContext(ctx)
	//
	device := model.NewDevice(tenant, string(inventoryDevice.ID))
	// data from inventory
	device.SetUpdatedAt(inventoryDevice.UpdatedTs)
	// last checkin date
	if deviceAuthDevice != nil {
		device.SetLastCheckIn(deviceAuthDevice.LastCheckinDate)
	}
	// extract location from attributes
	ok, location := extractLocation(inventoryDevice.Attributes)
	if ok {
		device.Location = &location
	}
	attributes, err := i.mapper.MapInventoryAttributes(ctx, tenant,
		inventoryDevice.Attributes, true, false)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to map device data for tenant %s, "+
				"device %s", tenant, inventoryDevice.ID)
		l.Warn(err)
	} else {
		for _, invattr := range attributes {
			attr := model.NewInventoryAttribute(invattr.Scope).
				SetName(invattr.Name).
				SetVal(invattr.Value)
			if err := device.AppendAttr(attr); err != nil {
				err = errors.Wrapf(err,
					"failed to convert device data for tenant %s, "+
						"device %s", tenant, inventoryDevice.ID)
				l.Warn(err)
			}
		}
	}
	// data from device auth
	_ = device.AppendAttr(&model.InventoryAttribute{
		Scope:  model.ScopeIdentity,
		Name:   model.AttrNameStatus,
		String: []string{deviceAuthDevice.Status},
	})
	for name, value := range deviceAuthDevice.IdDataStruct {
		attr := model.NewInventoryAttribute(model.ScopeIdentity).
			SetName(name).
			SetVal(value)
		if err := device.AppendAttr(attr); err != nil {
			err = errors.Wrapf(err,
				"failed to convert identity data for tenant %s, "+
					"device %s", tenant, inventoryDevice.ID)
			l.Warn(err)
		}
	}

	// data from deployments
	if deploymentsDevice != nil {
		_ = device.AppendAttr(&model.InventoryAttribute{
			Scope:  model.ScopeSystem,
			Name:   model.AttrNameLatestDeploymentStatus,
			String: []string{deploymentsDevice.DeviceDeploymentStatus},
		})
	}

	// return the device
	return device
}

func extractLocation(
	attrs inventory.DeviceAttributes,
) (bool, string) {
	latIdx := undefinedCoordinateIdx
	lonIdx := undefinedCoordinateIdx

	for i, attr := range attrs {
		if attr.Name == model.AttrNameGeoLatitude {
			latIdx = i
		} else if attr.Name == model.AttrNameGeoLongitude {
			lonIdx = i
		}
		if latIdx != undefinedCoordinateIdx && lonIdx != undefinedCoordinateIdx {
			break
		}
	}
	if latIdx != undefinedCoordinateIdx && lonIdx != undefinedCoordinateIdx {
		latStr, ok := attrs[latIdx].Value.(string)
		if !ok {
			return false, ""
		}
		lonStr, ok := attrs[lonIdx].Value.(string)
		if !ok {
			return false, ""
		}
		if validLocation(latStr, lonStr) {
			return true, latStr + "," + lonStr
		}
	}
	return false, ""
}

func validLocation(latStr, lonStr string) bool {
	lat, err := strconv.ParseFloat(latStr, 32)
	if err != nil {
		return false
	}
	if lat < -90 || lat > 90 {
		return false
	}
	lon, err := strconv.ParseFloat(lonStr, 32)
	if err != nil {
		return false
	}
	if lon < -180 || lon > 180 {
		return false
	}
	return true
}

func (i *indexer) processJobDeployments(
	ctx context.Context,
	tenant string,
	IDs IDs,
) {
	l := log.FromContext(ctx)
	depls := make([]*model.Deployment, 0, len(IDs))
	deploymentIDs := make([]string, 0, len(IDs))
	for deploymentID := range IDs {
		deploymentIDs = append(deploymentIDs, deploymentID)
	}
	// get device deployments from deployments
	deviceDeployments, err := i.deplClient.GetDeployments(ctx, tenant, deploymentIDs)
	if err != nil {
		l.Error(errors.Wrap(err, "failed to get device deployments from device deployments"))
		return
	}
	// process the results
	for deploymentID := range IDs {
		for _, d := range deviceDeployments {
			if d.ID == deploymentID {
				depl := i.processJobDeployment(ctx, tenant, d)
				if depl != nil {
					depls = append(depls, depl)
				}
				break
			}
		}
	}
	// bulk index the device
	if len(depls) > 0 {
		err = i.store.BulkIndexDeployments(ctx, depls)
		if err != nil {
			err = errors.Wrap(err, "failed to bulk index the deployments")
			l.Error(err)
		}
	}
}

func (i *indexer) processJobDeployment(
	ctx context.Context,
	tenant string,
	deployment *deployments.DeviceDeployment,
) *model.Deployment {
	deviceElapsedSeconds := uint(0)
	if deployment.Device == nil ||
		deployment.Deployment == nil {
		return nil
	} else if deployment.Device.Finished != nil && deployment.Device.Created != nil {
		deviceElapsedSeconds = uint(deployment.Device.Finished.Sub(
			*deployment.Device.Created).Seconds())
	}
	res := &model.Deployment{
		ID:                          deployment.ID,
		DeviceID:                    deployment.Device.Id,
		DeploymentID:                deployment.Deployment.Id,
		TenantID:                    tenant,
		DeploymentName:              deployment.Deployment.Name,
		DeploymentArtifactName:      deployment.Deployment.ArtifactName,
		DeploymentType:              deployment.Deployment.Type,
		DeploymentCreated:           deployment.Deployment.Created,
		DeploymentFilterID:          deployment.Deployment.FilterId,
		DeploymentAllDevices:        deployment.Deployment.AllDevices,
		DeploymentForceInstallation: deployment.Deployment.ForceInstallation,
		DeploymentGroups:            deployment.Deployment.Groups,
		DeploymentPhased:            deployment.Deployment.PhaseId != "",
		DeploymentPhaseId:           deployment.Deployment.PhaseId,
		DeploymentRetries:           deployment.Deployment.Retries,
		DeploymentMaxDevices:        uint(deployment.Deployment.MaxDevices),
		DeploymentAutogenerateDelta: deployment.Deployment.AutogenerateDelta,
		DeviceCreated:               deployment.Device.Created,
		DeviceFinished:              deployment.Device.Finished,
		DeviceElapsedSeconds:        deviceElapsedSeconds,
		DeviceDeleted:               deployment.Device.Deleted,
		DeviceStatus:                deployment.Device.Status,
		DeviceIsLogAvailable:        deployment.Device.IsLogAvailable,
		DeviceRetries:               deployment.Device.Retries,
		DeviceAttempts:              deployment.Device.Attempts,
	}
	if deployment.Device.Image != nil {
		res.ImageID = deployment.Device.Image.Id
		res.ImageDescription = deployment.Device.Image.Description
		res.ImageArtifactName = deployment.Device.Image.Name
		res.ImageDeviceTypes = deployment.Device.Image.DeviceTypesCompatible
		res.ImageSigned = deployment.Device.Image.Signed
		res.ImageProvides = deployment.Device.Image.Provides
		res.ImageDepends = deployment.Device.Image.Depends
		res.ImageClearsProvides = deployment.Device.Image.ClearsProvides
		res.ImageSize = deployment.Device.Image.Size
		if deployment.Device.Image.Info != nil {
			res.ImageArtifactInfoFormat = deployment.Device.Image.Info.Format
			res.ImageArtifactInfoVersion = deployment.Device.Image.Info.Version
		}
	}
	return res
}
