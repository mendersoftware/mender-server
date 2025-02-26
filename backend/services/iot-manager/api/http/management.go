// Copyright 2025 Northern.tech AS
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
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

var (
	ErrMissingUserAuthentication = errors.New(
		"user identity missing from authorization token",
	)
	ErrIntegrationNotFound = errors.New("integration not found")
)

const hdrLocation = "Location"

func getContextAndIdentity(c *gin.Context) (context.Context, *identity.Identity, error) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)
	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return nil, nil, ErrMissingUserAuthentication
	}
	return ctx, id, nil
}

// ManagementHandler is the namespace for management API handlers.
type ManagementHandler APIHandler

// GET /integrations
func (h *ManagementHandler) GetIntegrations(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	integrations, err := h.app.GetIntegrations(ctx)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, integrations)
}

// POST /integrations
func (h *ManagementHandler) CreateIntegration(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	integration := model.Integration{}
	if err := c.ShouldBindJSON(&integration); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	// TODO verify that Azure connectionstring / AWS equivalent has correct permissions
	//      - service
	//      - registry read/write

	inserted, err := h.app.CreateIntegration(ctx, integration)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationExists:
			// NOTE: temporary limitation
			rest.RenderError(c, http.StatusConflict, cause)
		default:
			_ = c.Error(err)
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}
		return
	}

	path := strings.Replace(APIURLIntegration, ":id", inserted.ID.String(), 1)
	c.Header(hdrLocation, APIURLManagement+path)
	c.Status(http.StatusCreated)

}

// GET /integrations/{id}
func (h *ManagementHandler) GetIntegrationById(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "integration ID must be a valid UUID"),
		)
		return
	}

	integration, err := h.app.GetIntegrationById(ctx, integrationID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationNotFound:
			rest.RenderError(c, http.StatusNotFound, ErrIntegrationNotFound)
		default:
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}
		return
	}

	c.JSON(http.StatusOK, integration)
}

// PUT /integrations/{id}/credentials
func (h *ManagementHandler) SetIntegrationCredentials(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "integration ID must be a valid UUID"),
		)
		return
	}

	credentials := model.Credentials{}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	err = h.app.SetIntegrationCredentials(ctx, integrationID, credentials)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationNotFound:
			rest.RenderError(c, http.StatusNotFound, ErrIntegrationNotFound)
		default:
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /integrations/{id}
func (h *ManagementHandler) RemoveIntegration(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "integration ID must be a valid UUID"),
		)
		return
	}
	if err = h.app.RemoveIntegration(ctx, integrationID); err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationNotFound:
			rest.RenderError(c, http.StatusNotFound, ErrIntegrationNotFound)
		case app.ErrCannotRemoveIntegration:
			rest.RenderError(c, http.StatusConflict, app.ErrCannotRemoveIntegration)
		default:
			rest.RenderError(c,
				http.StatusInternalServerError,
				err,
			)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /events
func (h *ManagementHandler) GetEvents(c *gin.Context) {
	ctx, _, err := getContextAndIdentity(c)
	if err != nil {
		return
	}

	filter, err := getEventsFilterFromQuery(c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			err,
		)
		return
	}
	integrationID := c.Request.URL.Query().Get(paramQueryIntegrationID)
	if len(integrationID) > 0 {
		if err := uuid.Validate(integrationID); err == nil {
			filter.IntegrationID = &integrationID
		} else {
			rest.RenderError(c,
				http.StatusBadRequest,
				ErrInvalidIntegrationID,
			)
			return
		}
	}

	events, err := h.app.GetEvents(ctx, *filter)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, events)
}

// get events filter from query params
func getEventsFilterFromQuery(c *gin.Context) (*model.EventsFilter, error) {
	filter := model.EventsFilter{}
	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		return nil, err
	}
	filter.Skip = (page - 1) * perPage
	filter.Limit = perPage
	return &filter, err
}
