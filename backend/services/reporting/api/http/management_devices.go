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

package http

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rbac"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

type attributes struct {
	Limit      int         `json:"limit"`
	Count      int         `json:"count"`
	Attributes []attribute `json:"attributes"`
}

type attribute struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

func (mc *ManagementController) AggregateDevices(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := parseAggregateDevicesParams(ctx, c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	res, err := mc.reporting.AggregateDevices(ctx, params)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, res)
}

func parseAggregateDevicesParams(ctx context.Context, c *gin.Context) (
	*model.AggregateParams, error) {
	var aggregateParams model.AggregateParams

	err := c.ShouldBindJSON(&aggregateParams)
	if err != nil {
		return nil, err
	}

	if id := identity.FromContext(ctx); id != nil {
		aggregateParams.TenantID = id.Tenant
	} else {
		return nil, errors.New("missing tenant ID from the context")
	}

	if scope := rbac.ExtractScopeFromHeader(c.Request); scope != nil {
		aggregateParams.Groups = scope.DeviceGroups
	}

	if err := aggregateParams.Validate(); err != nil {
		return nil, err
	}

	return &aggregateParams, nil
}

func (mc *ManagementController) DeviceAttrs(c *gin.Context) {
	ctx := c.Request.Context()

	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	} else {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("missing tenant ID from the context"),
		)
		return
	}

	mapping, err := mc.reporting.GetMapping(ctx, tenantID)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to retrieve the mapping"),
		)
		return
	} else if mapping == nil {
		mapping = &model.Mapping{}
	}

	n := model.MaxMappingInventoryAttributes
	if n > len(mapping.Inventory) {
		n = len(mapping.Inventory)
	}
	attributesList := make([]attribute, 0, n)
	if mapping.Inventory != nil {
		for _, attr := range mapping.Inventory[:n] {
			parts := strings.SplitN(attr, string(os.PathSeparator), 2)
			attributesList = append(attributesList, attribute{
				Name:  parts[1],
				Scope: parts[0],
			})
		}
	}
	res := &attributes{
		Limit:      model.MaxMappingInventoryAttributes,
		Count:      len(attributesList),
		Attributes: attributesList,
	}

	c.JSON(http.StatusOK, res)
}

func (mc *ManagementController) SearchDevices(c *gin.Context) {
	ctx := c.Request.Context()
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

func parseSearchDevicesParams(ctx context.Context, c *gin.Context) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	err := c.ShouldBindJSON(&searchParams)
	if err != nil {
		return nil, err
	}

	if id := identity.FromContext(ctx); id != nil {
		searchParams.TenantID = id.Tenant
	} else {
		return nil, errors.New("missing tenant ID from the context")
	}

	if scope := rbac.ExtractScopeFromHeader(c.Request); scope != nil {
		searchParams.Groups = scope.DeviceGroups
	}

	if searchParams.PerPage <= 0 {
		searchParams.PerPage = ParamPerPageDefault
	}
	if searchParams.Page <= 0 {
		searchParams.Page = ParamPageDefault
	}

	if err := searchParams.Validate(); err != nil {
		return nil, err
	}

	return &searchParams, nil
}

func (mc *ManagementController) SearchDeviceAttrs(c *gin.Context) {
	ctx := c.Request.Context()

	id := identity.FromContext(ctx)
	res, err := mc.reporting.GetSearchableInvAttrs(ctx, id.Tenant)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, res)
}
