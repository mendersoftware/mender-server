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

package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deviceconfig/app"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

// InteralAPI is a namespace for the internal API handlers.
type InternalAPI APIHandler

func NewInternalAPI(app app.App) *InternalAPI {
	return &InternalAPI{
		App: app,
	}
}

func (api *InternalAPI) Alive(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (api *InternalAPI) Health(c *gin.Context) {
	err := api.App.HealthCheck(c.Request.Context())
	if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (api *InternalAPI) ProvisionTenant(c *gin.Context) {
	var tenant model.NewTenant
	ctx := c.Request.Context()

	err := c.ShouldBindJSON(&tenant)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	if err = tenant.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest,
			errors.Wrap(err, "invalid request body"),
		)
		return
	}
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant.TenantID,
	})
	c.Request = c.Request.WithContext(ctx)

	err = api.App.ProvisionTenant(ctx, tenant)
	if err != nil {
		c.Error(err) //nolint:errcheck
		rest.RenderError(c, http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}
	c.Status(http.StatusCreated)
}

func (api *InternalAPI) ProvisionDevice(c *gin.Context) {
	var dev model.NewDevice
	ctx := c.Request.Context()
	id := &identity.Identity{
		Tenant: c.Param("tenant_id"),
	}
	ctx = identity.WithContext(ctx, id)
	c.Request = c.Request.WithContext(ctx)
	err := c.ShouldBindJSON(&dev)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}
	if err = dev.Validate(); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "invalid request body"),
		)
		return
	}
	err = api.App.ProvisionDevice(ctx, dev)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrDeviceAlreadyExists:
			rest.RenderError(c, http.StatusConflict, cause)
		default:
			c.Error(err) //nolint:errcheck
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New(http.StatusText(http.StatusInternalServerError)),
			)
		}
		return
	}
	c.Status(http.StatusCreated)
}

// PATCH /tenants/:tenant_id/configurations/device/:device_id
func (api *InternalAPI) UpdateConfiguration(c *gin.Context) {
	deviceID := c.Param("device_id")
	tenantID := c.Param("tenant_id")
	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject: deviceID,
		Tenant:  tenantID,
	})

	var attrs model.Attributes
	if err := c.ShouldBindJSON(&attrs); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request parameters"),
		)
		return
	} else if err = attrs.Validate(); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "invalid request parameters"),
		)
		return
	}

	err := api.App.UpdateConfiguration(ctx, deviceID, attrs)
	if err != nil {
		_ = c.Error(err)
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}
	c.Status(http.StatusNoContent)
}

func (api *InternalAPI) DecommissionDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	ctx := identity.WithContext(c.Request.Context(),
		&identity.Identity{
			Tenant:  c.Param("tenant_id"),
			Subject: deviceID,
		},
	)
	c.Request = c.Request.WithContext(ctx)

	err := api.App.DecommissionDevice(ctx, deviceID)
	if err != nil {
		switch errors.Cause(err) {
		case store.ErrDeviceNoExist:
			rest.RenderError(c, http.StatusNotFound, store.ErrDeviceNoExist)
		default:
			c.Error(err) //nolint:errcheck
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New(http.StatusText(http.StatusInternalServerError)),
			)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (api *InternalAPI) DeployConfiguration(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = identity.WithContext(ctx, &identity.Identity{
		Subject: c.Param(pathParamDeviceID),
		Tenant:  c.Param(pathParamTenantID),
	})
	c.Request = c.Request.WithContext(ctx)
	mgmtAPI := (*ManagementAPI)(api)
	mgmtAPI.DeployConfiguration(c)
}
