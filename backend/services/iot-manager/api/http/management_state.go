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
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

const (
	paramDeviceID           = "id"
	paramIntegrationID      = "integrationId"
	paramQueryIntegrationID = "integration_id"
)

var (
	ErrEmptyDeviceID        = errors.New("device ID is empty")
	ErrInvalidIntegrationID = errors.New("integration ID is not a valid UUID")
)

// GET /devices/:id/state
func (h *ManagementHandler) GetDeviceState(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	deviceID := c.Param(paramDeviceID)
	if deviceID == "" {
		rest.RenderError(c, http.StatusBadRequest, ErrEmptyDeviceID)
		return
	}

	device, err := h.app.GetDevice(ctx, deviceID)
	if err == app.ErrDeviceNotFound {
		rest.RenderError(c, http.StatusNotFound, app.ErrDeviceNotFound)
		return
	} else if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
		return
	}

	states := make(model.DeviceStates)
	for _, integrationID := range device.IntegrationIDs {
		state, err := h.app.GetDeviceStateIntegration(ctx, deviceID, integrationID)
		if state == nil && (err == nil ||
			err == app.ErrIntegrationNotFound ||
			err == app.ErrUnknownIntegration) {
			continue
		} else if err != nil {
			rest.RenderError(c, http.StatusInternalServerError, err)
			return
		}
		states[integrationID.String()] = *state
	}

	c.JSON(http.StatusOK, states)
}

// GET /devices/:id/state/:integrationId
func (h *ManagementHandler) GetDeviceStateIntegration(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	deviceID := c.Param(paramDeviceID)
	if deviceID == "" {
		rest.RenderError(c, http.StatusBadRequest, ErrEmptyDeviceID)
		return
	}
	integrationID, err := uuid.Parse(c.Param(paramIntegrationID))
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, ErrInvalidIntegrationID)
		return
	}

	state, err := h.app.GetDeviceStateIntegration(ctx, deviceID, integrationID)
	if err == app.ErrIntegrationNotFound || err == app.ErrUnknownIntegration {
		rest.RenderError(c, http.StatusNotFound, err)
		return
	} else if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, state)
}

// PUT /devices/:id/state/:integrationId
func (h *ManagementHandler) SetDeviceStateIntegration(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	deviceID := c.Param(paramDeviceID)
	if deviceID == "" {
		rest.RenderError(c, http.StatusBadRequest, ErrEmptyDeviceID)
		return
	}
	integrationID, err := uuid.Parse(c.Param(paramIntegrationID))
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, ErrInvalidIntegrationID)
		return
	}

	state := &model.DeviceState{}
	if err := c.ShouldBindJSON(state); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	state, err = h.app.SetDeviceStateIntegration(ctx, deviceID, integrationID, state)
	if err == app.ErrIntegrationNotFound || err == app.ErrUnknownIntegration {
		rest.RenderError(c, http.StatusNotFound, err)
		return
	} else if err == app.ErrDeviceStateConflict {
		rest.RenderError(c, http.StatusConflict, err)
		return
	} else if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, state)
}
