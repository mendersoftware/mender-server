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
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/contenttype"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/services/inventory/inv"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
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

func init() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()
}

func NewRouter(app inv.InventoryApp) http.Handler {
	router := gin.New()
	router.Use(accesslog.Middleware())
	router.Use(requestid.Middleware())

	mgmtHandler := NewManagementHandler(app)

	publicAPIs := router.Group(".")

	publicAPIs.Use(identity.Middleware())

	mgmtAPIV1 := publicAPIs.Group(apiUrlManagementV1)
	mgmtAPIV1Legacy := publicAPIs.Group(apiUrlLegacy)
	mgmtAPIV2 := publicAPIs.Group(apiUrlManagementV2)
	devicesAPIs := publicAPIs.Group(apiUrlDevicesV1)
	devicesAPILegacy := publicAPIs.Group(apiUrlLegacy)

	mgmtAPIV1.GET(uriDevices, mgmtHandler.GetDevicesHandler)
	mgmtAPIV1.GET(uriDevice, mgmtHandler.GetDeviceHandler)
	mgmtAPIV1.GET(uriDeviceGroups, mgmtHandler.GetDeviceGroupHandler)
	mgmtAPIV1.GET(uriGroups, mgmtHandler.GetGroupsHandler)
	mgmtAPIV1.GET(uriGroupsDevices, mgmtHandler.GetDevicesByGroupHandler)
	mgmtAPIV1.DELETE(uriDevice, mgmtHandler.DeleteDeviceInventoryHandler)
	mgmtAPIV1.DELETE(uriDeviceGroup, mgmtHandler.DeleteDeviceGroupHandler)
	mgmtAPIV1.DELETE(uriGroupsName, mgmtHandler.DeleteGroupHandler)
	mgmtAPIV1.DELETE(uriGroupsDevices, mgmtHandler.ClearDevicesGroupHandler)
	mgmtAPIV1.Group(".").Use(contenttype.CheckJSON()).
		PUT(uriDeviceGroups, mgmtHandler.AddDeviceToGroupHandler).
		PATCH(uriGroupsDevices, mgmtHandler.AppendDevicesToGroup).
		PUT(uriDeviceTags, mgmtHandler.UpdateDeviceTagsHandler).
		PATCH(uriDeviceTags, mgmtHandler.UpdateDeviceTagsHandler)

	mgmtAPIV2.GET(urlFiltersAttributes, mgmtHandler.FiltersAttributesHandler)
	mgmtAPIV2.Group(".").Use(contenttype.CheckJSON()).
		POST(urlFiltersSearch, mgmtHandler.FiltersSearchHandler)

	mgmtAPIV1Legacy.GET(uriDevices, mgmtHandler.GetDevicesHandler)
	mgmtAPIV1Legacy.GET(uriDevice, mgmtHandler.GetDeviceHandler)
	mgmtAPIV1Legacy.GET(uriDeviceGroups, mgmtHandler.GetDeviceGroupHandler)
	mgmtAPIV1Legacy.GET(uriGroups, mgmtHandler.GetGroupsHandler)
	mgmtAPIV1Legacy.GET(uriGroupsDevices, mgmtHandler.GetDevicesByGroupHandler)
	mgmtAPIV1Legacy.DELETE(uriDevice, mgmtHandler.DeleteDeviceInventoryHandler)
	mgmtAPIV1Legacy.DELETE(uriDeviceGroup, mgmtHandler.DeleteDeviceGroupHandler)
	mgmtAPIV1Legacy.DELETE(uriGroupsName, mgmtHandler.DeleteGroupHandler)
	mgmtAPIV1Legacy.DELETE(uriGroupsDevices, mgmtHandler.ClearDevicesGroupHandler)

	mgmtAPIV1Legacy.Group(".").Use(contenttype.CheckJSON()).
		PUT(uriDeviceGroups, mgmtHandler.AddDeviceToGroupHandler).
		PATCH(uriGroupsDevices, mgmtHandler.AppendDevicesToGroup).
		PUT(uriDeviceTags, mgmtHandler.UpdateDeviceTagsHandler).
		PATCH(uriDeviceTags, mgmtHandler.UpdateDeviceTagsHandler)

	devicesAPIs.Group(".").Use(contenttype.CheckJSON()).
		PATCH(uriDeviceAttributes, mgmtHandler.UpdateDeviceAttributesHandler).
		PUT(uriDeviceAttributes, mgmtHandler.UpdateDeviceAttributesHandler)

	devicesAPILegacy.Group(".").Use(contenttype.CheckJSON()).
		PATCH(uriAttributes, mgmtHandler.UpdateDeviceAttributesHandler).
		PUT(uriAttributes, mgmtHandler.UpdateDeviceAttributesHandler)

	// automatically add Option routes for public endpoints
	AutogenOptionsRoutes(router, AllowHeaderOptionsGenerator)

	// internal endpoints
	intrnlHandler := NewInternalHandler(app)
	intrnlAPIV1 := router.Group(apiUrlInternalV1)

	intrnlAPIV1.GET(uriInternalHealth, intrnlHandler.HealthCheckHandler)
	intrnlAPIV1.GET(uriInternalAlive, intrnlHandler.LivelinessHandler)
	intrnlAPIV1.PATCH(urlInternalAttributes, intrnlHandler.PatchDeviceAttributesInternalHandler)
	intrnlAPIV1.POST(urlInternalReindex, intrnlHandler.ReindexDeviceDataHandler)
	intrnlAPIV1.POST(uriInternalTenants, intrnlHandler.CreateTenantHandler)

	intrnlAPIV1.POST(uriInternalDevices, intrnlHandler.AddDeviceHandler)
	intrnlAPIV1.DELETE(uriInternalDeviceDetails, intrnlHandler.DeleteDeviceHandler)
	intrnlAPIV1.POST(urlInternalDevicesStatus, intrnlHandler.InternalDevicesStatusHandler)
	intrnlAPIV1.GET(uriInternalDeviceGroups, intrnlHandler.GetDeviceGroupsInternalHandler)

	internalAPIV2 := router.Group(apiUrlInternalV2)
	internalAPIV2.POST(urlInternalFiltersSearch, intrnlHandler.InternalFiltersSearchHandler)

	return router
}
