// Copyright 2024 Northern.tech AS
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
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/contenttype"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/routing"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
)

const (
	ApiUrlInternal   = "/api/internal/v1/deployments"
	ApiUrlManagement = "/api/management/v1/deployments"
	ApiUrlDevices    = "/api/devices/v1/deployments"

	ApiUrlManagementArtifacts               = "/artifacts"
	ApiUrlManagementArtifactsList           = "/artifacts/list"
	ApiUrlManagementArtifactsGenerate       = "/artifacts/generate"
	ApiUrlManagementArtifactsDirectUpload   = "/artifacts/directupload"
	ApiUrlManagementArtifactsCompleteUpload = ApiUrlManagementArtifactsDirectUpload +
		"/:id/complete"
	ApiUrlManagementArtifactsId         = "/artifacts/:id"
	ApiUrlManagementArtifactsIdDownload = "/artifacts/:id/download"

	ApiUrlManagementDeployments                   = "/deployments"
	ApiUrlManagementMultipleDeploymentsStatistics = "/deployments/statistics/list"
	ApiUrlManagementDeploymentsGroup              = "/deployments/group/:name"
	ApiUrlManagementDeploymentsId                 = "/deployments/:id"
	ApiUrlManagementDeploymentsStatistics         = "/deployments/:id/statistics"
	ApiUrlManagementDeploymentsStatus             = "/deployments/:id/status"
	ApiUrlManagementDeploymentsDevices            = "/deployments/:id/devices"
	ApiUrlManagementDeploymentsDevicesList        = "/deployments/:id/devices/list"
	ApiUrlManagementDeploymentsLog                = "/deployments/:id/devices/:devid/log"
	ApiUrlManagementDeploymentsDeviceId           = "/deployments/devices/:id"
	ApiUrlManagementDeploymentsDeviceHistory      = "/deployments/devices/:id/history"
	ApiUrlManagementDeploymentsDeviceList         = "/deployments/:id/device_list"

	ApiUrlManagementReleases     = "/deployments/releases"
	ApiUrlManagementReleasesList = "/deployments/releases/list"

	ApiUrlManagementLimitsName = "/limits/:name"

	ApiUrlManagementV2                      = "/api/management/v2/deployments"
	ApiUrlManagementV2Releases              = "/deployments/releases"
	ApiUrlManagementV2ReleasesName          = ApiUrlManagementV2Releases + "/:name"
	ApiUrlManagementV2ReleaseTags           = ApiUrlManagementV2Releases + "/:name/tags"
	ApiUrlManagementV2ReleaseAllTags        = "/releases/all/tags"
	ApiUrlManagementV2ReleaseAllUpdateTypes = "/releases/all/types"
	ApiUrlManagementV2Deployments           = "/deployments"

	ApiUrlDevicesDeploymentsNext  = "/device/deployments/next"
	ApiUrlDevicesDeploymentStatus = "/device/deployments/:id/status"
	ApiUrlDevicesDeploymentsLog   = "/device/deployments/:id/log"
	ApiUrlDevicesDownloadConfig   = "/download/configuration" +
		"/:deployment_id/:device_type/:device_id"

	ApiUrlInternalAlive                          = "/alive"
	ApiUrlInternalHealth                         = "/health"
	ApiUrlInternalTenants                        = "/tenants"
	ApiUrlInternalTenantDeployments              = "/tenants/:tenant/deployments"
	ApiUrlInternalTenantDeploymentsDevices       = "/tenants/:tenant/deployments/devices"
	ApiUrlInternalTenantDeploymentsDevice        = "/tenants/:tenant/deployments/devices/:id"
	ApiUrlInternalTenantArtifacts                = "/tenants/:tenant/artifacts"
	ApiUrlInternalTenantStorageSettings          = "/tenants/:tenant/storage/settings"
	ApiUrlInternalDeviceConfigurationDeployments = "/tenants/:tenant/configuration/deployments" +
		"/:deployment_id/devices/:device_id"
	ApiUrlInternalDeviceDeploymentLastStatusDeployments = "/tenants/:tenant/devices/deployments" +
		"/last"
)

func init() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()
}

// NewRouter defines all REST API routes.
func NewRouter(
	ctx context.Context,
	app app.App,
	ds store.DataStore,
	cfg *Config,
) http.Handler {
	router := routing.NewMinimalGinRouter()
	// Create and configure API handlers
	//
	// Encode base64 secret in either std or URL encoding ignoring padding.
	deploymentsHandlers := NewDeploymentsApiHandlers(
		ds, new(view.RESTView), app, cfg,
	)

	// Routing
	internalAPIs := router.Group(ApiUrlInternal)

	publicAPIs := router.Group(".")
	publicAPIs.Use(accesslog.Middleware())
	publicAPIs.Use(requestid.Middleware())

	withAuth := publicAPIs.Group(".")
	withAuth.Use(identity.Middleware())

	NewImagesResourceRoutes(withAuth, deploymentsHandlers, cfg)
	NewDeploymentsResourceRoutes(publicAPIs, deploymentsHandlers)
	NewLimitsResourceRoutes(withAuth, deploymentsHandlers)
	InternalRoutes(internalAPIs, deploymentsHandlers)
	ReleasesRoutes(withAuth, deploymentsHandlers)

	restutil.AutogenOptionsRoutes(
		restutil.NewOptionsHandler,
		router)

	return router
}

func NewImagesResourceRoutes(router *gin.RouterGroup,
	controller *DeploymentsApiHandlers, cfg *Config) {

	if controller == nil {
		return
	}
	mgmtV1 := router.Group(ApiUrlManagement)

	artifcatType := contenttype.Middleware("multipart/form-data", "multipart/mixed")

	mgmtV1.GET(ApiUrlManagementArtifacts, controller.GetImages)
	mgmtV1.GET(ApiUrlManagementArtifactsList, controller.ListImages)
	mgmtV1.GET(ApiUrlManagementArtifactsId, controller.GetImage)
	mgmtV1.GET(ApiUrlManagementArtifactsIdDownload, controller.DownloadLink)
	if !controller.config.DisableNewReleasesFeature {
		mgmtV1.DELETE(ApiUrlManagementArtifactsId, controller.DeleteImage)
		mgmtV1.Group(".").Use(artifcatType).
			POST(ApiUrlManagementArtifacts, controller.NewImage).
			POST(ApiUrlManagementArtifactsGenerate, controller.GenerateImage)
		mgmtV1.Group(".").Use(contenttype.CheckJSON()).
			PUT(ApiUrlManagementArtifactsId, controller.EditImage)

	} else {
		mgmtV1.DELETE(ApiUrlManagementArtifactsId, ServiceUnavailable)

		mgmtV1.Group(".").Use(artifcatType).
			POST(ApiUrlManagementArtifacts, ServiceUnavailable).
			POST(ApiUrlManagementArtifactsGenerate, ServiceUnavailable)
		mgmtV1.PUT(ApiUrlManagementArtifactsId, ServiceUnavailable)

	}
	if !controller.config.DisableNewReleasesFeature && cfg.EnableDirectUpload {
		log.NewEmpty().Infof(
			"direct upload enabled: POST %s",
			ApiUrlManagementArtifactsDirectUpload,
		)
		if cfg.EnableDirectUploadSkipVerify {
			log.NewEmpty().Info(
				"direct upload enabled SkipVerify",
			)
		}
		mgmtV1.Group(".").Use(contenttype.CheckJSON()).
			POST(ApiUrlManagementArtifactsDirectUpload,
				controller.UploadLink).
			POST(ApiUrlManagementArtifactsCompleteUpload,
				controller.CompleteUpload)
	}
}

func NewDeploymentsResourceRoutes(router *gin.RouterGroup, controller *DeploymentsApiHandlers) {

	if controller == nil {
		return
	}
	mgmtV1 := router.Group(ApiUrlManagement)
	mgmtV1.Use(identity.Middleware())
	mgmtV2 := router.Group(ApiUrlManagementV2)
	mgmtV2.Use(identity.Middleware())

	mgmtV1.GET(ApiUrlManagementDeployments, controller.LookupDeployment)
	mgmtV2.GET(ApiUrlManagementV2Deployments, controller.LookupDeploymentV2)
	mgmtV1.GET(ApiUrlManagementDeploymentsId, controller.GetDeployment)
	mgmtV1.GET(ApiUrlManagementDeploymentsStatistics, controller.GetDeploymentStats)
	mgmtV1.GET(ApiUrlManagementDeploymentsDevices,
		controller.GetDeviceStatusesForDeployment)
	mgmtV1.GET(ApiUrlManagementDeploymentsDevicesList,
		controller.GetDevicesListForDeployment)
	mgmtV1.GET(ApiUrlManagementDeploymentsLog,
		controller.GetDeploymentLogForDevice)
	mgmtV1.GET(ApiUrlManagementDeploymentsDeviceId,
		controller.ListDeviceDeployments)
	mgmtV1.GET(ApiUrlManagementDeploymentsDeviceList,
		controller.GetDeploymentDeviceList)

	mgmtV1.DELETE(ApiUrlManagementDeploymentsDeviceId,
		controller.AbortDeviceDeployments)
	mgmtV1.DELETE(ApiUrlManagementDeploymentsDeviceHistory,
		controller.DeleteDeviceDeploymentsHistory)

	mgmtV1.Group(".").Use(contenttype.CheckJSON()).
		POST(ApiUrlManagementDeployments, controller.PostDeployment).
		POST(ApiUrlManagementDeploymentsGroup, controller.DeployToGroup).
		POST(ApiUrlManagementMultipleDeploymentsStatistics,
			controller.GetDeploymentsStats).
		PUT(ApiUrlManagementDeploymentsStatus, controller.AbortDeployment)

	// Devices
	devices := router.Group(ApiUrlDevices)

	devices.GET(ApiUrlDevicesDownloadConfig,
		controller.DownloadConfiguration)

	devices.Use(identity.Middleware())

	devices.GET(ApiUrlDevicesDeploymentsNext, controller.GetDeploymentForDevice)
	devices.Group(".").Use(contenttype.CheckJSON()).
		POST(ApiUrlDevicesDeploymentsNext,
			controller.GetDeploymentForDevice).
		PUT(ApiUrlDevicesDeploymentStatus,
			controller.PutDeploymentStatusForDevice).
		PUT(ApiUrlDevicesDeploymentsLog,
			controller.PutDeploymentLogForDevice)

}

func NewLimitsResourceRoutes(router *gin.RouterGroup, controller *DeploymentsApiHandlers) {

	if controller == nil {
		return
	}
	mgmtV1 := router.Group(ApiUrlManagement)

	mgmtV1.GET(ApiUrlManagementLimitsName, controller.GetLimit)

}

func InternalRoutes(router *gin.RouterGroup, controller *DeploymentsApiHandlers) {
	if controller == nil {
		return
	}
	accesslogErrorsOnly := accesslog.AccessLogger{
		DisableLog: func(c *gin.Context) bool {
			if c.Writer.Status()/100 == 2 {
				// 2XX
				return true
			}
			return false
		},
	}
	// Health Check
	// Skiping logging 2XX status code requests to decrease	noise
	router.GET(ApiUrlInternalAlive, accesslogErrorsOnly.Middleware,
		requestid.Middleware(),
		controller.AliveHandler)
	router.GET(ApiUrlInternalHealth, accesslogErrorsOnly.Middleware,
		requestid.Middleware(),
		controller.HealthHandler)

	router.Use(accesslog.Middleware())
	router.Use(requestid.Middleware())

	router.POST(ApiUrlInternalTenants, controller.ProvisionTenantsHandler)
	router.GET(ApiUrlInternalTenantDeployments, controller.DeploymentsPerTenantHandler)
	router.GET(ApiUrlInternalTenantDeploymentsDevices,
		controller.ListDeviceDeploymentsByIDsInternal)
	router.GET(ApiUrlInternalTenantDeploymentsDevice,
		controller.ListDeviceDeploymentsInternal)
	router.DELETE(ApiUrlInternalTenantDeploymentsDevice,
		controller.AbortDeviceDeploymentsInternal)
	// per-tenant storage settings
	router.GET(ApiUrlInternalTenantStorageSettings, controller.GetTenantStorageSettingsHandler)
	router.PUT(ApiUrlInternalTenantStorageSettings, controller.PutTenantStorageSettingsHandler)

	// Configuration deployments (internal)
	router.POST(ApiUrlInternalDeviceConfigurationDeployments,
		controller.PostDeviceConfigurationDeployment)

	// Last device deployment status deployments (internal)
	router.POST(ApiUrlInternalDeviceDeploymentLastStatusDeployments,
		controller.GetDeviceDeploymentLastStatus)

	if !controller.config.DisableNewReleasesFeature {
		router.POST(ApiUrlInternalTenantArtifacts, controller.NewImageForTenantHandler)
	} else {
		router.POST(ApiUrlInternalTenantArtifacts, ServiceUnavailable)
	}
}

func ReleasesRoutes(router *gin.RouterGroup, controller *DeploymentsApiHandlers) {
	if controller == nil {
		return
	}
	mgmtV1 := router.Group(ApiUrlManagement)
	mgmtV2 := router.Group(ApiUrlManagementV2)

	if controller.config.DisableNewReleasesFeature {
		mgmtV1.GET(ApiUrlManagementReleases, controller.GetReleases)
		mgmtV1.GET(ApiUrlManagementReleasesList, controller.ListReleases)

	} else {

		mgmtV1.GET(ApiUrlManagementReleases, controller.GetReleases)
		mgmtV1.GET(ApiUrlManagementReleasesList, controller.ListReleases)
		mgmtV2.GET(ApiUrlManagementV2Releases, controller.ListReleasesV2)
		mgmtV2.GET(ApiUrlManagementV2ReleasesName, controller.GetRelease)
		mgmtV2.GET(ApiUrlManagementV2ReleaseAllTags, controller.GetReleaseTagKeys)
		mgmtV2.GET(ApiUrlManagementV2ReleaseAllUpdateTypes, controller.GetReleasesUpdateTypes)
		mgmtV2.DELETE(ApiUrlManagementV2Releases, controller.DeleteReleases)
		mgmtV2.Group(".").Use(contenttype.CheckJSON()).
			PUT(ApiUrlManagementV2ReleaseTags, controller.PutReleaseTags).
			PATCH(ApiUrlManagementV2ReleasesName, controller.PatchRelease)

	}
}

func FMTConfigURL(scheme, hostname, deploymentID, deviceType, deviceID string) string {
	repl := strings.NewReplacer(
		":"+ParamDeploymentID, url.PathEscape(deploymentID),
		":"+ParamDeviceType, url.PathEscape(deviceType),
		":"+ParamDeviceID, url.PathEscape(deviceID),
	)
	return scheme + "://" + hostname + ApiUrlDevices + repl.Replace(ApiUrlDevicesDownloadConfig)
}

func ServiceUnavailable(c *gin.Context) {
	c.Status(http.StatusServiceUnavailable)
}
