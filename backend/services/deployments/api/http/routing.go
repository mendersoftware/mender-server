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
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"

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
		"/#id/complete"
	ApiUrlManagementArtifactsId         = ApiUrlManagement + "/artifacts/#id"
	ApiUrlManagementArtifactsIdDownload = ApiUrlManagement + "/artifacts/#id/download"

	ApiUrlManagementDeployments                   = ApiUrlManagement + "/deployments"
	ApiUrlManagementMultipleDeploymentsStatistics = ApiUrlManagement +
		"/deployments/statistics/list"
	ApiUrlManagementDeploymentsGroup       = ApiUrlManagement + "/deployments/group/#name"
	ApiUrlManagementDeploymentsId          = ApiUrlManagement + "/deployments/#id"
	ApiUrlManagementDeploymentsStatistics  = ApiUrlManagement + "/deployments/#id/statistics"
	ApiUrlManagementDeploymentsStatus      = ApiUrlManagement + "/deployments/#id/status"
	ApiUrlManagementDeploymentsDevices     = ApiUrlManagement + "/deployments/#id/devices"
	ApiUrlManagementDeploymentsDevicesList = ApiUrlManagement + "/deployments/#id/devices/list"
	ApiUrlManagementDeploymentsLog         = ApiUrlManagement +
		"/deployments/#id/devices/#devid/log"
	ApiUrlManagementDeploymentsDeviceId      = ApiUrlManagement + "/deployments/devices/#id"
	ApiUrlManagementDeploymentsDeviceHistory = ApiUrlManagement + "/deployments/devices/#id/history"
	ApiUrlManagementDeploymentsDeviceList    = ApiUrlManagement + "/deployments/#id/device_list"

	ApiUrlManagementReleases     = ApiUrlManagement + "/deployments/releases"
	ApiUrlManagementReleasesList = ApiUrlManagement + "/deployments/releases/list"

	ApiUrlManagementLimitsName = ApiUrlManagement + "/limits/#name"

	ApiUrlManagementV2                      = "/api/management/v2/deployments"
	ApiUrlManagementV2Releases              = ApiUrlManagementV2 + "/deployments/releases"
	ApiUrlManagementV2ReleasesName          = ApiUrlManagementV2Releases + "/#name"
	ApiUrlManagementV2ReleaseTags           = ApiUrlManagementV2Releases + "/#name/tags"
	ApiUrlManagementV2ReleaseAllTags        = ApiUrlManagementV2 + "/releases/all/tags"
	ApiUrlManagementV2ReleaseAllUpdateTypes = ApiUrlManagementV2 + "/releases/all/types"

	ApiUrlDevicesDeploymentsNext  = ApiUrlDevices + "/device/deployments/next"
	ApiUrlDevicesDeploymentStatus = ApiUrlDevices + "/device/deployments/#id/status"
	ApiUrlDevicesDeploymentsLog   = ApiUrlDevices + "/device/deployments/#id/log"
	ApiUrlDevicesDownloadConfig   = ApiUrlDevices +
		"/download/configuration/#deployment_id/#device_type/#device_id"

	ApiUrlInternalAlive                    = ApiUrlInternal + "/alive"
	ApiUrlInternalHealth                   = ApiUrlInternal + "/health"
	ApiUrlInternalTenants                  = ApiUrlInternal + "/tenants"
	ApiUrlInternalTenantDeployments        = ApiUrlInternal + "/tenants/#tenant/deployments"
	ApiUrlInternalTenantDeploymentsDevices = ApiUrlInternal + "/tenants/#tenant/deployments/devices"
	ApiUrlInternalTenantDeploymentsDevice  = ApiUrlInternal +
		"/tenants/#tenant/deployments/devices/#id"
	ApiUrlInternalTenantArtifacts       = ApiUrlInternal + "/tenants/#tenant/artifacts"
	ApiUrlInternalTenantStorageSettings = ApiUrlInternal +
		"/tenants/#tenant/storage/settings"
	ApiUrlInternalDeviceConfigurationDeployments = ApiUrlInternal +
		"/tenants/#tenant/configuration/deployments/#deployment_id/devices/#device_id"
	ApiUrlInternalDeviceDeploymentLastStatusDeployments = ApiUrlInternal +
		"/tenants/#tenant/devices/deployments/last"
)

func contentTypeMiddleware(h rest.HandlerFunc) rest.HandlerFunc {
	checkJSON := (&rest.ContentTypeCheckerMiddleware{}).
		MiddlewareFunc(h)
	checkMultipart := func(w rest.ResponseWriter, r *rest.Request) {
		mediatype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if r.ContentLength > 0 && !(mediatype == "multipart/form-data") {
			rest.Error(w,
				"Bad Content-Type, expected 'multipart/form-data'",
				http.StatusUnsupportedMediaType)
			return
		}
		h(w, r)
	}
	return func(w rest.ResponseWriter, r *rest.Request) {
		if r.Method == http.MethodPost &&
			(r.URL.Path == ApiUrlManagementArtifacts ||
				r.URL.Path == ApiUrlManagementArtifactsGenerate) {
			checkMultipart(w, r)
		} else {
			checkJSON(w, r)
		}
	}
}

func wrapMiddleware(middleware rest.Middleware, routes ...*rest.Route) []*rest.Route {
	for _, route := range routes {
		route.Func = middleware.MiddlewareFunc(route.Func)
	}
	return routes
}

// NewRouter defines all REST API routes.
func NewHandler(
	ctx context.Context,
	app app.App,
	ds store.DataStore,
	cfg *Config,
) (http.Handler, error) {
	api := rest.NewApi()
	api.Use(
		// logging
		&requestlog.RequestLogMiddleware{},
		&requestid.RequestIdMiddleware{},
		&accesslog.AccessLogMiddleware{
			Format: accesslog.SimpleLogFormat,
			DisableLog: func(statusCode int, r *rest.Request) bool {
				if statusCode < 300 {
					if r.URL.Path == ApiUrlInternalHealth ||
						r.URL.Path == ApiUrlInternalAlive {
						return true
					}
				}
				return false
			},
		},
	)

	// Create and configure API handlers
	//
	// Encode base64 secret in either std or URL encoding ignoring padding.
	deploymentsHandlers := NewDeploymentsApiHandlers(
		ds, new(view.RESTView), app, cfg,
	)

	// Routing
	imageRoutes := NewImagesResourceRoutes(deploymentsHandlers, cfg)
	deploymentsRoutes := NewDeploymentsResourceRoutes(deploymentsHandlers)
	limitsRoutes := NewLimitsResourceRoutes(deploymentsHandlers)
	internalRoutes := InternalRoutes(deploymentsHandlers)
	releasesRoutes := ReleasesRoutes(deploymentsHandlers)

	publicRoutes := append(releasesRoutes, deploymentsRoutes...)
	publicRoutes = append(publicRoutes, limitsRoutes...)
	publicRoutes = append(publicRoutes, imageRoutes...)
	publicRoutes = restutil.AutogenOptionsRoutes(
		restutil.NewOptionsHandler,
		publicRoutes...,
	)
	publicRoutes = wrapMiddleware(&identity.IdentityMiddleware{
		UpdateLogger: true,
	}, publicRoutes...)
	publicRoutes = wrapMiddleware(
		rest.MiddlewareSimple(contentTypeMiddleware),
		publicRoutes...,
	)
	routes := append(publicRoutes, internalRoutes...)

	restApp, err := rest.MakeRouter(routes...)
	if err != nil {
		return nil, err
	}

	api.SetApp(restApp)
	return api.MakeHandler(), nil
}

func NewImagesResourceRoutes(controller *DeploymentsApiHandlers, cfg *Config) []*rest.Route {

	if controller == nil {
		return []*rest.Route{}
	}

	routes := []*rest.Route{
		rest.Get(ApiUrlManagementArtifacts, controller.GetImages),
		rest.Get(ApiUrlManagementArtifactsList, controller.ListImages),
		rest.Get(ApiUrlManagementArtifactsId, controller.GetImage),
		rest.Get(ApiUrlManagementArtifactsIdDownload, controller.DownloadLink),
	}
	if !controller.config.DisableNewReleasesFeature {
		routes = append(routes,
			rest.Post(ApiUrlManagementArtifacts, controller.NewImage),
			rest.Post(ApiUrlManagementArtifactsGenerate, controller.GenerateImage),
			rest.Delete(ApiUrlManagementArtifactsId, controller.DeleteImage),
			rest.Put(ApiUrlManagementArtifactsId, controller.EditImage),
		)
	} else {
		routes = append(routes,
			rest.Post(ApiUrlManagementArtifacts, ServiceUnavailable),
			rest.Post(ApiUrlManagementArtifactsGenerate, ServiceUnavailable),
			rest.Delete(ApiUrlManagementArtifactsId, ServiceUnavailable),
			rest.Put(ApiUrlManagementArtifactsId, ServiceUnavailable),
		)
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
		routes = append(routes, rest.Post(
			ApiUrlManagementArtifactsDirectUpload,
			controller.UploadLink,
		))
		routes = append(routes, rest.Post(
			ApiUrlManagementArtifactsCompleteUpload,
			controller.CompleteUpload,
		))
	}
	return routes
}

func NewDeploymentsResourceRoutes(controller *DeploymentsApiHandlers) []*rest.Route {

	if controller == nil {
		return []*rest.Route{}
	}

	return []*rest.Route{
		// Deployments
		rest.Post(ApiUrlManagementDeployments, controller.PostDeployment),
		rest.Post(ApiUrlManagementDeploymentsGroup, controller.DeployToGroup),
		rest.Get(ApiUrlManagementDeployments, controller.LookupDeployment),
		rest.Get(ApiUrlManagementDeploymentsId, controller.GetDeployment),
		rest.Post(ApiUrlManagementMultipleDeploymentsStatistics,
			controller.GetDeploymentsStats),
		rest.Get(ApiUrlManagementDeploymentsStatistics, controller.GetDeploymentStats),
		rest.Put(ApiUrlManagementDeploymentsStatus, controller.AbortDeployment),
		rest.Get(ApiUrlManagementDeploymentsDevices,
			controller.GetDeviceStatusesForDeployment),
		rest.Get(ApiUrlManagementDeploymentsDevicesList,
			controller.GetDevicesListForDeployment),
		rest.Get(ApiUrlManagementDeploymentsLog,
			controller.GetDeploymentLogForDevice),
		rest.Delete(ApiUrlManagementDeploymentsDeviceId,
			controller.AbortDeviceDeployments),
		rest.Delete(ApiUrlManagementDeploymentsDeviceHistory,
			controller.DeleteDeviceDeploymentsHistory),
		rest.Get(ApiUrlManagementDeploymentsDeviceId,
			controller.ListDeviceDeployments),
		rest.Get(ApiUrlManagementDeploymentsDeviceList,
			controller.GetDeploymentDeviceList),

		// Devices
		rest.Get(ApiUrlDevicesDeploymentsNext, controller.GetDeploymentForDevice),
		rest.Post(ApiUrlDevicesDeploymentsNext, controller.GetDeploymentForDevice),
		rest.Put(ApiUrlDevicesDeploymentStatus,
			controller.PutDeploymentStatusForDevice),
		rest.Put(ApiUrlDevicesDeploymentsLog,
			controller.PutDeploymentLogForDevice),
		rest.Get(ApiUrlDevicesDownloadConfig,
			controller.DownloadConfiguration),
	}
}

func NewLimitsResourceRoutes(controller *DeploymentsApiHandlers) []*rest.Route {

	if controller == nil {
		return []*rest.Route{}
	}

	return []*rest.Route{
		// limits
		rest.Get(ApiUrlManagementLimitsName, controller.GetLimit),
	}
}

func InternalRoutes(controller *DeploymentsApiHandlers) []*rest.Route {
	if controller == nil {
		return []*rest.Route{}
	}

	routes := []*rest.Route{
		rest.Post(ApiUrlInternalTenants, controller.ProvisionTenantsHandler),
		rest.Get(ApiUrlInternalTenantDeployments, controller.DeploymentsPerTenantHandler),
		rest.Get(ApiUrlInternalTenantDeploymentsDevices,
			controller.ListDeviceDeploymentsByIDsInternal),
		rest.Get(ApiUrlInternalTenantDeploymentsDevice,
			controller.ListDeviceDeploymentsInternal),
		rest.Delete(ApiUrlInternalTenantDeploymentsDevice,
			controller.AbortDeviceDeploymentsInternal),
		// per-tenant storage settings
		rest.Get(ApiUrlInternalTenantStorageSettings, controller.GetTenantStorageSettingsHandler),
		rest.Put(ApiUrlInternalTenantStorageSettings, controller.PutTenantStorageSettingsHandler),

		// Configuration deployments (internal)
		rest.Post(ApiUrlInternalDeviceConfigurationDeployments,
			controller.PostDeviceConfigurationDeployment),

		// Last device deployment status deployments (internal)
		rest.Post(ApiUrlInternalDeviceDeploymentLastStatusDeployments,
			controller.GetDeviceDeploymentLastStatus),

		// Health Check
		rest.Get(ApiUrlInternalAlive, controller.AliveHandler),
		rest.Get(ApiUrlInternalHealth, controller.HealthHandler),
	}

	if !controller.config.DisableNewReleasesFeature {
		routes = append(routes,
			rest.Post(ApiUrlInternalTenantArtifacts, controller.NewImageForTenantHandler),
		)
	} else {
		routes = append(routes,
			rest.Post(ApiUrlInternalTenantArtifacts, ServiceUnavailable),
		)
	}

	return routes
}

func ReleasesRoutes(controller *DeploymentsApiHandlers) []*rest.Route {
	if controller == nil {
		return []*rest.Route{}
	}

	if controller.config.DisableNewReleasesFeature {
		return []*rest.Route{
			rest.Get(ApiUrlManagementReleases, controller.GetReleases),
			rest.Get(ApiUrlManagementReleasesList, controller.ListReleases),
		}
	} else {
		return []*rest.Route{
			rest.Get(ApiUrlManagementReleases, controller.GetReleases),
			rest.Get(ApiUrlManagementReleasesList, controller.ListReleases),
			rest.Get(ApiUrlManagementV2Releases, controller.ListReleasesV2),
			rest.Put(ApiUrlManagementV2ReleaseTags, controller.PutReleaseTags),
			rest.Get(ApiUrlManagementV2ReleaseAllTags, controller.GetReleaseTagKeys),
			rest.Get(ApiUrlManagementV2ReleaseAllUpdateTypes, controller.GetReleasesUpdateTypes),
			rest.Patch(ApiUrlManagementV2ReleasesName, controller.PatchRelease),
			rest.Delete(ApiUrlManagementV2Releases, controller.DeleteReleases),
		}
	}
}

func FMTConfigURL(scheme, hostname, deploymentID, deviceType, deviceID string) string {
	repl := strings.NewReplacer(
		"#"+ParamDeploymentID, url.PathEscape(deploymentID),
		"#"+ParamDeviceType, url.PathEscape(deviceType),
		"#"+ParamDeviceID, url.PathEscape(deviceID),
	)
	return scheme + "://" + hostname + repl.Replace(ApiUrlDevicesDownloadConfig)
}

func ServiceUnavailable(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
}
