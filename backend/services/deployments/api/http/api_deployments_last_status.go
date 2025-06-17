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

package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// device deployments last status handler
func (d *DeploymentsApiHandlers) GetDeviceDeploymentLastStatus(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)

	l.Debugf("starting")

	tenantId := c.Param("tenant")
	var req model.DeviceDeploymentLastStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Errorf("error during DecodeJsonPayload: %s.", err.Error())
		d.view.RenderError(
			c,
			errors.Wrap(err, "cannot parse device ids array"),
			http.StatusBadRequest,
		)
		return
	} else if len(req.DeviceIds) == 0 {
		d.view.RenderError(
			c,
			errors.Wrap(err, "device ids array cannot be empty"),
			http.StatusBadRequest,
		)
	}

	l.Debugf("querying %d devices ids", len(req.DeviceIds))
	if tenantId != "" {
		ctx = identity.WithContext(
			ctx,
			&identity.Identity{
				Tenant: tenantId,
			},
		)
	}
	lastDeployments, err := d.app.GetDeviceDeploymentLastStatus(ctx, req.DeviceIds)
	switch err {
	default:
		d.view.RenderInternalError(c, err)
	case nil:
		l.Infof("outputting: %+v", lastDeployments)
		c.JSON(http.StatusOK, lastDeployments)
	}
}
