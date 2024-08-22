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

package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func (mc *InternalController) SearchDevices(c *gin.Context) {
	tid := c.Param("tenant_id")

	ctx := c.Request.Context()
	ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})

	params, err := parseSearchDevicesParams(ctx, c)

	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	res, total, err := mc.reporting.SearchDevices(ctx, params)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	pageLinkHdrs(c, params.Page, params.PerPage, total)

	c.Header(hdrTotalCount, strconv.Itoa(total))
	c.JSON(http.StatusOK, res)
}
