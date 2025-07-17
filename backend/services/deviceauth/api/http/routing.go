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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/contenttype"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestsize"
	"github.com/mendersoftware/mender-server/pkg/routing"
	dconfig "github.com/mendersoftware/mender-server/services/deviceauth/config"
	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	apiUrlDevicesV1 = "/api/devices/v1/authentication"
	uriAuthReqs     = "/auth_requests"

	// internal API
	apiUrlInternalV1      = "/api/internal/v1/devauth"
	uriAlive              = "/alive"
	uriHealth             = "/health"
	uriTokenVerify        = "/tokens/verify"
	uriTenantLimit        = "/tenant/:id/limits/:name"
	uriTokens             = "/tokens"
	uriTenants            = "/tenants"
	uriTenantDevice       = "/tenants/:tid/devices/:did"
	uriTenantDeviceStatus = "/tenants/:tid/devices/:did/status"
	uriTenantDevices      = "/tenants/:tid/devices"
	uriTenantDevicesCount = "/tenants/:tid/devices/count"

	// management API v2
	apiUrlManagementV2       = "/api/management/v2/devauth"
	v2uriDevices             = "/devices"
	v2uriDevicesCount        = "/devices/count"
	v2uriDevicesSearch       = "/devices/search"
	v2uriDevice              = "/devices/:id"
	v2uriDeviceAuthSet       = "/devices/:id/auth/:aid"
	v2uriDeviceAuthSetStatus = "/devices/:id/auth/:aid/status"
	v2uriToken               = "/tokens/:id"
	v2uriDevicesLimit        = "/limits/:name"

	HdrAuthReqSign = "X-MEN-Signature"
)

type HttpOptionsGenerator func(methods []string) gin.HandlerFunc

func AllowHeaderOptionsGenerator(methods []string) gin.HandlerFunc {
	// return a dummy handler for now
	return func(c *gin.Context) {
		for _, m := range methods {
			c.Writer.Header().Add("Allow", m)
		}
	}
}

func supportsMethod(method string, methods []string) bool {
	return utils.ContainsString(method, methods)
}

// Automatically add OPTIONS method support for each defined route,
// only if there's no OPTIONS handler for that route yet
func AutogenOptionsRoutes(router *gin.Engine, gen HttpOptionsGenerator) {

	routes := router.Routes()
	methodGroups := make(map[string][]string, len(routes))

	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/api/internal") {
			continue
		}
		methods, ok := methodGroups[route.Path]
		if !ok {
			methods = make([]string, 0)
		}

		methodGroups[route.Path] = append(methods, route.Method)
	}

	for route, methods := range methodGroups {
		// skip if there's a handler for OPTIONS already
		if !supportsMethod(http.MethodOptions, methods) {
			router.OPTIONS(route, gen(methods))
		}
	}

}

type Config struct {
	MaxRequestSize int64
}

func NewConfig() *Config {
	return &Config{
		MaxRequestSize: dconfig.SettingMaxRequestSizeDefault,
	}
}

type Option func(c *Config)

func SetMaxRequestSize(size int64) Option {
	return func(c *Config) {
		c.MaxRequestSize = size
	}
}

func NewRouter(app devauth.App, db store.DataStore, options ...Option) http.Handler {
	config := NewConfig()
	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	router := routing.NewGinRouter()
	router.Use(requestsize.Middleware(config.MaxRequestSize))

	d := NewDevAuthApiHandlers(app, db)

	publicAPIs := router.Group(".")
	publicAPIs.Use(identity.Middleware())

	mgmtAPIV2 := publicAPIs.Group(apiUrlManagementV2)
	devicesAPIs := router.Group(apiUrlDevicesV1)

	// Devices API
	devicesAPIs.Group(".").Use(contenttype.CheckJSON()).
		POST(uriAuthReqs, d.SubmitAuthRequestHandler)

	// API v2
	mgmtAPIV2.GET(v2uriDevicesCount, d.GetDevicesCountHandler)
	mgmtAPIV2.GET(v2uriDevices, d.GetDevicesV2Handler)
	mgmtAPIV2.GET(v2uriDevice, d.GetDeviceV2Handler)
	mgmtAPIV2.GET(v2uriDeviceAuthSetStatus, d.GetAuthSetStatusHandler)
	mgmtAPIV2.GET(v2uriDevicesLimit, d.GetLimitHandler)
	mgmtAPIV2.DELETE(v2uriDevice, d.DecommissionDeviceHandler)
	mgmtAPIV2.DELETE(v2uriDeviceAuthSet, d.DeleteDeviceAuthSetHandler)
	mgmtAPIV2.DELETE(v2uriToken, d.DeleteTokenHandler)
	mgmtAPIV2.Group(".").Use(contenttype.CheckJSON()).
		POST(v2uriDevices, d.PostDevicesV2Handler).
		PUT(v2uriDeviceAuthSetStatus, d.UpdateDeviceStatusHandler).
		POST(v2uriDevicesSearch, d.SearchDevicesV2Handler)

	// automatically add Option routes for public endpoints
	AutogenOptionsRoutes(router, AllowHeaderOptionsGenerator)

	intrnlAPIV1 := router.Group(apiUrlInternalV1)

	intrnlAPIV1.GET(uriAlive, d.AliveHandler)
	intrnlAPIV1.GET(uriHealth, d.HealthCheckHandler)
	intrnlAPIV1.GET(uriTokenVerify,
		identity.Middleware(),
		d.VerifyTokenHandler)
	intrnlAPIV1.POST(uriTokenVerify,
		identity.Middleware(),
		d.VerifyTokenHandler)
	intrnlAPIV1.DELETE(uriTokens, d.DeleteTokensHandler)
	intrnlAPIV1.PUT(uriTenantLimit, d.PutTenantLimitHandler)
	intrnlAPIV1.GET(uriTenantLimit, d.GetTenantLimitHandler)
	intrnlAPIV1.DELETE(uriTenantLimit, d.DeleteTenantLimitHandler)
	intrnlAPIV1.POST(uriTenants, d.ProvisionTenantHandler)
	intrnlAPIV1.GET(uriTenantDeviceStatus, d.GetTenantDeviceStatus)
	intrnlAPIV1.GET(uriTenantDevices, d.GetTenantDevicesHandler)
	intrnlAPIV1.GET(uriTenantDevicesCount, d.GetTenantDevicesCountHandler)
	intrnlAPIV1.DELETE(uriTenantDevice, d.DeleteDeviceHandler)

	return router
}
