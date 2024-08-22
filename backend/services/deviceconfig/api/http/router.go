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
	"os"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"

	"github.com/mendersoftware/mender-server/services/deviceconfig/app"
)

// API URL used by the HTTP router
const (
	pathParamDeviceID = "device_id"
	pathParamTenantID = "tenant_id"

	URIDevices    = "/api/devices/v1/deviceconfig"
	URIInternal   = "/api/internal/v1/deviceconfig"
	URIManagement = "/api/management/v1/deviceconfig"

	URITenants       = "/tenants"
	URITenant        = "/tenants/:tenant_id"
	URITenantDevices = "/tenants/:tenant_id/devices"
	URITenantDevice  = "/tenants/:tenant_id/devices/:device_id"

	URIConfiguration       = "/configurations/device/:device_id"
	URIDeployConfiguration = "/configurations/device/:device_id/deploy"
	URIDeviceConfiguration = "/configuration"

	URIAlive  = "/alive"
	URIHealth = "/health"
)

func init() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()
}

type APIHandler struct {
	App app.App
}

func NewAPIHandler(app app.App) *APIHandler {
	return &APIHandler{
		App: app,
	}
}

// NewRouter initializes a new gin.Engine as a http.Handler
func NewRouter(app app.App) http.Handler {
	router := gin.New()
	// accesslog provides logging of http responses and recovery on panic.
	router.Use(accesslog.Middleware())
	// requestid attaches X-Men-Requestid header to context
	router.Use(requestid.Middleware())

	apiHandler := NewAPIHandler(app)

	intrnlAPI := (*InternalAPI)(apiHandler)
	intrnlGrp := router.Group(URIInternal)

	intrnlGrp.GET(URIAlive, intrnlAPI.Alive)
	intrnlGrp.GET(URIHealth, intrnlAPI.Health)

	intrnlGrp.POST(URITenants, intrnlAPI.ProvisionTenant)
	intrnlGrp.DELETE(URITenant, intrnlAPI.DeleteTenant)
	intrnlGrp.POST(URITenantDevices, intrnlAPI.ProvisionDevice)
	intrnlGrp.DELETE(URITenantDevice, intrnlAPI.DecommissionDevice)

	intrnlGrp.PATCH(URITenant+URIConfiguration, intrnlAPI.UpdateConfiguration)
	intrnlGrp.POST(URITenant+URIDeployConfiguration, intrnlAPI.DeployConfiguration)

	mgmtAPI := (*ManagementAPI)(apiHandler)
	mgmtGrp := router.Group(URIManagement)

	// identity middleware for collecting JWT claims into request Context.
	mgmtGrp.Use(identity.Middleware())
	mgmtGrp.GET(URIConfiguration, mgmtAPI.GetConfiguration)
	mgmtGrp.PUT(URIConfiguration, mgmtAPI.SetConfiguration)
	mgmtGrp.POST(URIDeployConfiguration, mgmtAPI.DeployConfiguration)

	devAPI := (*DevicesAPI)(apiHandler)
	devGrp := router.Group(URIDevices)
	devGrp.Use(identity.Middleware())
	devGrp.GET(URIDeviceConfiguration, devAPI.GetConfiguration)
	devGrp.PUT(URIDeviceConfiguration, devAPI.SetConfiguration)

	return router
}
