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
	"github.com/mendersoftware/mender-server/pkg/requestsize"
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

	ApiUrlManagementArtifacts               = ApiUrlManagement + "/artifacts"
	ApiUrlManagementArtifactsList           = ApiUrlManagement + "/artifacts/list"
	ApiUrlManagementArtifactsGenerate       = ApiUrlManagement + "/artifacts/generate"
	ApiUrlManagementArtifactsDirectUpload   = ApiUrlManagement + "/artifacts/directupload"
	ApiUrlManagementArtifactsCompleteUpload = ApiUrlManagementArtifactsDirectUpload +
		"/:id/complete"
	ApiUrlManagementArtifactsId         = ApiUrlManagement + "/artifacts/:id"
	ApiUrlManagementArtifactsIdDownload = ApiUrlManagement + "/artifacts/:id/download"

	ApiUrlManagementDeployments                   = ApiUrlManagement + "/deployments"
	ApiUrlManagementMultipleDeploymentsStatistics = ApiUrlManagement +
		"/deployments/statistics/list"
	ApiUrlManagementDeploymentsGroup       = ApiUrlManagement + "/deployments/group/:name"
	ApiUrlManagementDeploymentsId          = ApiUrlManagement + "/deployments/:id"
	ApiUrlManagementDeploymentsStatistics  = ApiUrlManagement + "/deployments/:id/statistics"
	ApiUrlManagementDeploymentsStatus      = ApiUrlManagement + "/deployments/:id/status"
	ApiUrlManagementDeploymentsDevices     = ApiUrlManagement + "/deployments/:id/devices"
	ApiUrlManagementDeploymentsDevicesList = ApiUrlManagement + "/deployments/:id/devices/list"
	ApiUrlManagementDeploymentsLog         = ApiUrlManagement +
		"/deployments/:id/devices/:devid/log"
	ApiUrlManagementDeploymentsDeviceId      = ApiUrlManagement + "/deployments/devices/:id"
	ApiUrlManagementDeploymentsDeviceHistory = ApiUrlManagement + "/deployments/devices/:id/history"
	ApiUrlManagementDeploymentsDeviceList    = ApiUrlManagement + "/deployments/:id/device_list"

	ApiUrlManagementReleases     = ApiUrlManagement + "/deployments/releases"
	ApiUrlManagementReleasesList = ApiUrlManagement + "/deployments/releases/list"

	ApiUrlManagementLimitsName = ApiUrlManagement + "/limits/:name"

	ApiUrlManagementV2                      = "/api/management/v2/deployments"
	ApiUrlManagementV2Artifacts             = ApiUrlManagementV2 + "/artifacts"
	ApiUrlManagementV2Releases              = ApiUrlManagementV2 + "/deployments/releases"
	ApiUrlManagementV2ReleasesName          = ApiUrlManagementV2Releases + "/:name"
	ApiUrlManagementV2ReleaseTags           = ApiUrlManagementV2Releases + "/:name/tags"
	ApiUrlManagementV2ReleaseAllTags        = ApiUrlManagementV2 + "/releases/all/tags"
	ApiUrlManagementV2ReleaseAllUpdateTypes = ApiUrlManagementV2 + "/releases/all/types"
	ApiUrlManagementV2Deployments           = ApiUrlManagementV2 + "/deployments"

	ApiUrlDevicesDeploymentsNext  = ApiUrlDevices + "/device/deployments/next"
	ApiUrlDevicesDeploymentStatus = ApiUrlDevices + "/device/deployments/:id/status"
	ApiUrlDevicesDeploymentsLog   = ApiUrlDevices + "/device/deployments/:id/log"
	ApiUrlDevicesDownloadConfig   = ApiUrlDevices +
		"/download/configuration/:deployment_id/:device_type/:device_id"

	ApiUrlInternalAlive                    = ApiUrlInternal + "/alive"
	ApiUrlInternalHealth                   = ApiUrlInternal + "/health"
	ApiUrlInternalTenants                  = ApiUrlInternal + "/tenants"
	ApiUrlInternalTenantDeployments        = ApiUrlInternal + "/tenants/:tenant/deployments"
	ApiUrlInternalTenantDeploymentsDevices = ApiUrlInternal + "/tenants/:tenant/deployments/devices"
	ApiUrlInternalTenantDeploymentsDevice  = ApiUrlInternal +
		"/tenants/:tenant/deployments/devices/:id"
	ApiUrlInternalTenantArtifacts       = ApiUrlInternal + "/tenants/:tenant/artifacts"
	ApiUrlInternalTenantStorageSettings = ApiUrlInternal +
		"/tenants/:tenant/storage/settings"
	ApiUrlInternalDeviceConfigurationDeployments = ApiUrlInternal +
		"/tenants/:tenant/configuration/deployments/:deployment_id/devices/:device_id"
	ApiUrlInternalDeviceDeploymentLastStatusDeployments = ApiUrlInternal +
		"/tenants/:tenant/devices/deployments/last"
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
	if cfg == nil {
		cfg = NewConfig()
	}
	router := routing.NewMinimalGinRouter()
	// Create and configure API handlers
	//
	// Encode base64 secret in either std or URL encoding ignoring padding.
	deploymentsHandlers := NewDeploymentsApiHandlers(
		ds, new(view.RESTView), app, cfg,
	)
	router.Use(requestid.Middleware())
	requestLimit := requestsize.Middleware(cfg.MaxRequestSize)
	accesslogDefault := accesslog.Middleware()
	accesslogSupressed := accesslog.AccessLogger{
		DisableLog: func(c *gin.Context) bool {
			return c.Writer.Status() < 300
		}}.Middleware

	// Setup routing groups
	routerHealthz := router.Group("", accesslogSupressed, requestLimit)
	routerOpen := router.Group("", accesslogDefault, requestLimit)
	routerAuth := router.Group("", accesslogDefault, identity.Middleware(), requestLimit)
	routerAuthNoLimit := router.Group("", accesslogDefault, identity.Middleware())

	NewImagesResourceRoutes(routerAuth, routerAuthNoLimit, deploymentsHandlers, cfg)
	NewDeploymentsResourceRoutes(routerAuth, routerOpen, deploymentsHandlers)
	NewLimitsResourceRoutes(routerAuth, deploymentsHandlers)
	InternalRoutes(routerOpen, routerHealthz, deploymentsHandlers)
	ReleasesRoutes(routerAuth, deploymentsHandlers)

	restutil.AutogenOptionsRoutes(
		restutil.NewOptionsHandler,
		router)

	return router
}

func NewImagesResourceRoutes(router, routerNoLimit *gin.RouterGroup,
	controller *DeploymentsApiHandlers, cfg *Config) {

	if controller == nil {
		return
	}

	artifactType := contenttype.Middleware("multipart/form-data", "multipart/mixed")

	router.GET(ApiUrlManagementArtifacts, controller.GetImages)
	router.GET(ApiUrlManagementV2Artifacts, controller.ListImagesV2)
	router.GET(ApiUrlManagementArtifactsList, controller.ListImages)
	router.GET(ApiUrlManagementArtifactsId, controller.GetImage)
	router.GET(ApiUrlManagementArtifactsIdDownload, controller.DownloadLink)
	if !controller.config.DisableNewReleasesFeature {
		router.DELETE(ApiUrlManagementArtifactsId, controller.DeleteImage)
		routerNoLimit.Group("", artifactType).
			POST(ApiUrlManagementArtifacts,
				requestsize.Middleware(cfg.MaxImageSize),
				controller.NewImage).
			POST(ApiUrlManagementArtifactsGenerate,
				requestsize.Middleware(cfg.MaxGenerateDataSize),
				controller.GenerateImage)
		router.Group("", contenttype.CheckJSON()).
			PUT(ApiUrlManagementArtifactsId, controller.EditImage)

	} else {
		router.DELETE(ApiUrlManagementArtifactsId, ServiceUnavailable)

		router.Group("", artifactType).
			POST(ApiUrlManagementArtifacts, ServiceUnavailable).
			POST(ApiUrlManagementArtifactsGenerate, ServiceUnavailable)
		router.PUT(ApiUrlManagementArtifactsId, ServiceUnavailable)

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
		router.Group("", contenttype.CheckJSON()).
			POST(ApiUrlManagementArtifactsDirectUpload,
				controller.UploadLink).
			POST(ApiUrlManagementArtifactsCompleteUpload,
				controller.CompleteUpload)
	}
}

func NewDeploymentsResourceRoutes(
	groupAuthz *gin.RouterGroup,
	groupOpen *gin.RouterGroup,
	controller *DeploymentsApiHandlers,
) {

	if controller == nil {
		return
	}

	// nolint:lll
	groupAuthz.
		GET(ApiUrlManagementDeployments, controller.LookupDeployment).
		GET(ApiUrlManagementV2Deployments, controller.LookupDeploymentV2).
		GET(ApiUrlManagementDeploymentsId, controller.GetDeployment).
		GET(ApiUrlManagementDeploymentsStatistics, controller.GetDeploymentStats).
		GET(ApiUrlManagementDeploymentsDevices, controller.GetDeviceStatusesForDeployment).
		GET(ApiUrlManagementDeploymentsDevicesList, controller.GetDevicesListForDeployment).
		GET(ApiUrlManagementDeploymentsLog, controller.GetDeploymentLogForDevice).
		GET(ApiUrlManagementDeploymentsDeviceId, controller.ListDeviceDeployments).
		GET(ApiUrlManagementDeploymentsDeviceList, controller.GetDeploymentDeviceList).
		DELETE(ApiUrlManagementDeploymentsDeviceId, controller.AbortDeviceDeployments).
		DELETE(ApiUrlManagementDeploymentsDeviceHistory, controller.DeleteDeviceDeploymentsHistory)

	// nolint:lll
	groupAuthz.Group("", contenttype.CheckJSON()).
		POST(ApiUrlManagementDeployments, controller.PostDeployment).
		POST(ApiUrlManagementDeploymentsGroup, controller.DeployToGroup).
		POST(ApiUrlManagementMultipleDeploymentsStatistics, controller.GetDeploymentsStats).
		PUT(ApiUrlManagementDeploymentsStatus, controller.AbortDeployment)

	groupOpen.GET(ApiUrlDevicesDownloadConfig, controller.DownloadConfiguration)

	groupAuthz.GET(ApiUrlDevicesDeploymentsNext, controller.GetDeploymentForDevice)
	groupAuthz.Group("", contenttype.CheckJSON()).
		POST(ApiUrlDevicesDeploymentsNext, controller.GetDeploymentForDevice).
		PUT(ApiUrlDevicesDeploymentStatus, controller.PutDeploymentStatusForDevice).
		PUT(ApiUrlDevicesDeploymentsLog, controller.PutDeploymentLogForDevice)

}

func NewLimitsResourceRoutes(router *gin.RouterGroup, controller *DeploymentsApiHandlers) {

	if controller == nil {
		return
	}
	mgmtV1 := router.Group(ApiUrlManagement)

	mgmtV1.GET(ApiUrlManagementLimitsName, controller.GetLimit)

}

func InternalRoutes(
	routerOpen *gin.RouterGroup,
	routerHealthz *gin.RouterGroup,
	controller *DeploymentsApiHandlers,
) {
	if controller == nil {
		return
	}
	// Health Check
	// Skiping logging 2XX status code requests to decrease	noise
	routerHealthz.GET(ApiUrlInternalAlive, controller.AliveHandler)
	routerHealthz.GET(ApiUrlInternalHealth, controller.HealthHandler)

	routerOpen.POST(ApiUrlInternalTenants, controller.ProvisionTenantsHandler)
	routerOpen.GET(ApiUrlInternalTenantDeployments, controller.DeploymentsPerTenantHandler)
	routerOpen.GET(ApiUrlInternalTenantDeploymentsDevices,
		controller.ListDeviceDeploymentsByIDsInternal)
	routerOpen.GET(ApiUrlInternalTenantDeploymentsDevice,
		controller.ListDeviceDeploymentsInternal)
	routerOpen.DELETE(ApiUrlInternalTenantDeploymentsDevice,
		controller.AbortDeviceDeploymentsInternal)
	// per-tenant storage settings
	routerOpen.GET(ApiUrlInternalTenantStorageSettings, controller.GetTenantStorageSettingsHandler)
	routerOpen.PUT(ApiUrlInternalTenantStorageSettings, controller.PutTenantStorageSettingsHandler)

	// Configuration deployments (internal)
	routerOpen.POST(ApiUrlInternalDeviceConfigurationDeployments,
		controller.PostDeviceConfigurationDeployment)

	// Last device deployment status deployments (internal)
	routerOpen.POST(ApiUrlInternalDeviceDeploymentLastStatusDeployments,
		controller.GetDeviceDeploymentLastStatus)

	if !controller.config.DisableNewReleasesFeature {
		routerOpen.POST(ApiUrlInternalTenantArtifacts, controller.NewImageForTenantHandler)
	} else {
		routerOpen.POST(ApiUrlInternalTenantArtifacts, ServiceUnavailable)
	}
}

func ReleasesRoutes(routerAuth *gin.RouterGroup, controller *DeploymentsApiHandlers) {
	if controller == nil {
		return
	}

	if controller.config.DisableNewReleasesFeature {
		routerAuth.GET(ApiUrlManagementReleases, controller.GetReleases)
		routerAuth.GET(ApiUrlManagementReleasesList, controller.ListReleases)

	} else {

		routerAuth.GET(ApiUrlManagementReleases, controller.GetReleases)
		routerAuth.GET(ApiUrlManagementReleasesList, controller.ListReleases)
		routerAuth.GET(ApiUrlManagementV2Releases, controller.ListReleasesV2)
		routerAuth.GET(ApiUrlManagementV2ReleasesName, controller.GetRelease)
		routerAuth.GET(ApiUrlManagementV2ReleaseAllTags, controller.GetReleaseTagKeys)
		routerAuth.GET(ApiUrlManagementV2ReleaseAllUpdateTypes, controller.GetReleasesUpdateTypes)
		routerAuth.DELETE(ApiUrlManagementV2Releases, controller.DeleteReleases)
		routerAuth.Group("", contenttype.CheckJSON()).
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
	return scheme + "://" + hostname + repl.Replace(ApiUrlDevicesDownloadConfig)
}

func ServiceUnavailable(c *gin.Context) {
	c.Status(http.StatusServiceUnavailable)
}
