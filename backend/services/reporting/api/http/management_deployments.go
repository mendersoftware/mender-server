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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rbac"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

func (mc *ManagementController) AggregateDeployments(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := parseAggregateDeploymentsParams(ctx, c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	res, err := mc.reporting.AggregateDeployments(ctx, params)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.JSON(http.StatusOK, res)
}

func parseAggregateDeploymentsParams(ctx context.Context, c *gin.Context) (
	*model.AggregateDeploymentsParams, error) {
	var aggregateParams model.AggregateDeploymentsParams

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
		aggregateParams.DeploymentGroups = scope.DeviceGroups
	}

	if err := aggregateParams.Validate(); err != nil {
		return nil, err
	}

	return &aggregateParams, nil
}

func (mc *ManagementController) SearchDeployments(c *gin.Context) {
	ctx := c.Request.Context()
	params, err := parseDeploymentsSearchParams(ctx, c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	res, total, err := mc.reporting.SearchDeployments(ctx, params)
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

func parseDeploymentsSearchParams(ctx context.Context, c *gin.Context) (
	*model.DeploymentsSearchParams, error) {
	var searchParams model.DeploymentsSearchParams

	err := c.ShouldBindJSON(&searchParams)
	if err != nil {
		return nil, err
	}

	if id := identity.FromContext(ctx); id != nil {
		searchParams.TenantID = id.Tenant
	} else {
		return nil, errors.New("missing tenant ID from the context")
	}

	if searchParams.PerPage <= 0 {
		searchParams.PerPage = ParamPerPageDefault
	}
	if searchParams.Page <= 0 {
		searchParams.Page = ParamPageDefault
	}

	if scope := rbac.ExtractScopeFromHeader(c.Request); scope != nil {
		searchParams.DeploymentGroups = scope.DeviceGroups
	}

	if err := searchParams.Validate(); err != nil {
		return nil, err
	}

	return &searchParams, nil
}
