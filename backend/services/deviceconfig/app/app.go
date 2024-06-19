// Copyright 2021 Northern.tech AS
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
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deviceconfig/client/workflows"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

// App errors
var (
	ErrDeviceNotFound     = errors.New("device not found")
	ErrDeviceNotConnected = errors.New("device not connected")
)

// App interface describes app objects
//
//nolint:lll
//go:generate ../x/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error

	ProvisionTenant(ctx context.Context, tenant model.NewTenant) error

	ProvisionDevice(ctx context.Context, dev model.NewDevice) error
	DecommissionDevice(ctx context.Context, devID string) error

	SetConfiguration(ctx context.Context, devID string, configuration model.Attributes) error
	UpdateConfiguration(ctx context.Context, devID string, attrs model.Attributes) error
	SetReportedConfiguration(ctx context.Context, devID string, configuration model.Attributes) error
	GetDevice(ctx context.Context, devID string) (model.Device, error)
	DeployConfiguration(ctx context.Context, device model.Device, request model.DeployConfigurationRequest) (model.DeployConfigurationResponse, error)
}

// app is an app object
type app struct {
	store     store.DataStore
	workflows workflows.Client
	Config
}

type Config struct {
	HaveAuditLogs bool
}

// NewApp initialize a new deviceconfig App
func New(ds store.DataStore, wf workflows.Client, config ...Config) App {
	conf := Config{}
	for _, cfgIn := range config {
		if cfgIn.HaveAuditLogs {
			conf.HaveAuditLogs = true
		}
	}
	return &app{
		store:     ds,
		workflows: wf,
		Config:    conf,
	}
}

// HealthCheck performs a health check and returns an error if it fails
func (a *app) HealthCheck(ctx context.Context) error {
	return a.store.Ping(ctx)
}

func (a *app) ProvisionTenant(ctx context.Context, tenant model.NewTenant) error {
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant.TenantID,
	})
	return a.store.MigrateLatest(ctx)
}

func (a *app) ProvisionDevice(ctx context.Context, dev model.NewDevice) error {
	now := time.Now()
	return a.store.InsertDevice(ctx, model.Device{
		ID:        dev.ID,
		UpdatedTS: &now,
	})
}

func (a *app) DecommissionDevice(ctx context.Context, devID string) error {
	return a.store.DeleteDevice(ctx, devID)
}

func (a *app) SetConfiguration(ctx context.Context,
	devID string,
	configuration model.Attributes) error {
	now := time.Now()
	err := a.store.ReplaceConfiguration(ctx, model.Device{
		ID:                   devID,
		ConfiguredAttributes: configuration,
		UpdatedTS:            &now,
	})
	if err != nil {
		return err
	}
	if identity := identity.FromContext(ctx); identity != nil &&
		identity.IsUser && a.HaveAuditLogs {
		userID := identity.Subject
		configuration, err := configuration.MarshalJSON()
		if err == nil {
			err = a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
				Action: workflows.ActionSetConfiguration,
				Actor: workflows.Actor{
					ID:   userID,
					Type: workflows.ActorUser,
				},
				Object: workflows.Object{
					ID:   devID,
					Type: workflows.ObjectDevice,
				},
				Change:  string(configuration),
				EventTS: time.Now(),
			})
		}
		if err != nil {
			return errors.Wrap(err,
				"failed to submit audit log for setting the device configuration",
			)
		}
	}

	return nil
}

func (a *app) UpdateConfiguration(
	ctx context.Context,
	devID string,
	attrs model.Attributes,
) error {
	err := a.store.UpdateConfiguration(ctx, devID, attrs)
	if err != nil {
		return err
	}
	if identity := identity.FromContext(ctx); identity != nil &&
		identity.IsUser && a.HaveAuditLogs {
		userID := identity.Subject
		configuration, err := attrs.MarshalJSON()
		if err == nil {
			err = a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
				Action: workflows.ActionSetConfiguration,
				Actor: workflows.Actor{
					ID:   userID,
					Type: workflows.ActorUser,
				},
				Object: workflows.Object{
					ID:   devID,
					Type: workflows.ObjectDevice,
				},
				Change:  string(configuration),
				EventTS: time.Now(),
			})
		}
		if err != nil {
			return errors.Wrap(err,
				"failed to submit audit log for updating the device configuration",
			)
		}
	}
	return nil
}

func (a *app) SetReportedConfiguration(ctx context.Context,
	devID string,
	configuration model.Attributes) error {
	now := time.Now()
	return a.store.ReplaceReportedConfiguration(ctx, model.Device{
		ID:                 devID,
		ReportedAttributes: configuration,
		ReportTS:           &now,
	})
}

func (a *app) GetDevice(ctx context.Context, devID string) (model.Device, error) {
	return a.store.GetDevice(ctx, devID)
}

func (a *app) DeployConfiguration(ctx context.Context, device model.Device,
	request model.DeployConfigurationRequest) (model.DeployConfigurationResponse, error) {
	response := model.DeployConfigurationResponse{}
	configuration, err := device.ConfiguredAttributes.MarshalJSON()
	if err != nil {
		return response, err
	}
	identity := identity.FromContext(ctx)
	if identity == nil {
		return response, errors.New("identity missing from the context")
	}
	deploymentID := uuid.New()
	err = a.store.SetDeploymentID(ctx, device.ID, deploymentID)
	if err != nil {
		return response, nil
	}
	response.DeploymentID = deploymentID
	err = a.workflows.DeployConfiguration(ctx, identity.Tenant, device.ID,
		response.DeploymentID, configuration, request.Retries, request.UpdateControlMap)
	if err != nil {
		return response, err
	}
	if a.HaveAuditLogs {
		userID := identity.Subject
		err = a.workflows.SubmitAuditLog(ctx, workflows.AuditLog{
			Action: workflows.ActionDeployConfiguration,
			Actor: workflows.Actor{
				ID:   userID,
				Type: workflows.ActorUser,
			},
			Object: workflows.Object{
				ID:   device.ID,
				Type: workflows.ObjectDevice,
			},
			Change:  string(configuration),
			EventTS: time.Now(),
		})
		if err != nil {
			return response, errors.Wrap(err,
				"failed to submit audit log for deploying the device configuration",
			)
		}
	}
	return response, nil
}
