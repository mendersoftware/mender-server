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
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deviceconfig/app"
	"github.com/mendersoftware/mender-server/services/deviceconfig/model"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

// DevicesAPI is a namespace for the APIHandler
type DevicesAPI APIHandler

var errInvalidIdentity = errors.New("forbidden: invalid identity data")

func NewDevicesAPI(app app.App) *DevicesAPI {
	return &DevicesAPI{
		App: app,
	}
}

func (api *DevicesAPI) SetConfiguration(c *gin.Context) {
	var configuration model.Attributes

	ctx := c.Request.Context()
	identity := identity.FromContext(ctx)
	if identity == nil || !identity.IsDevice {
		rest.RenderError(c, http.StatusForbidden, errInvalidIdentity)
		return
	}

	devID := identity.Subject
	err := c.ShouldBindJSON(&configuration)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	for _, a := range configuration {
		if err := a.Validate(); err != nil {
			rest.RenderError(c,
				http.StatusBadRequest,
				errors.Wrap(err, "invalid request body"),
			)
			return
		}
	}

	err = api.App.SetReportedConfiguration(ctx, devID, configuration)
	if err != nil {
		c.Error(err) //nolint:errcheck
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}
	c.Status(http.StatusNoContent)
}

func (api *DevicesAPI) GetConfiguration(c *gin.Context) {
	ctx := c.Request.Context()
	identity := identity.FromContext(ctx)
	if identity == nil || !identity.IsDevice {
		rest.RenderError(c, http.StatusForbidden, errInvalidIdentity)
		return
	}

	devID := identity.Subject
	device, err := api.App.GetDevice(ctx, devID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrDeviceNoExist:
			c.Error(err) //nolint:errcheck
			rest.RenderError(c,
				http.StatusNotFound,
				cause,
			)
			return
		default:
			c.Error(err) //nolint:errcheck
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New(http.StatusText(http.StatusInternalServerError)),
			)
			return
		}
	}

	c.JSON(http.StatusOK, device.ConfiguredAttributes)
}
