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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
)

// API URL used by the HTTP router
const (
	APIURLInternal = "/api/internal/v1/iot-manager"

	APIURLAlive             = "/alive"
	APIURLHealth            = "/health"
	APIURLTenants           = "/tenants"
	APIURLTenant            = APIURLTenants + "/:tenant_id"
	APIURLTenantAuth        = APIURLTenant + "/auth"
	APIURLTenantDevices     = APIURLTenant + "/devices"
	APIURLTenantDevice      = APIURLTenantDevices + "/:device_id"
	APIURLTenantBulkDevices = APIURLTenant + "/bulk/devices"
	APIURLTenantBulkStatus  = APIURLTenantBulkDevices + "/status/:status"

	APIURLManagement = "/api/management/v1/iot-manager"

	APIURLIntegrations           = "/integrations"
	APIURLIntegration            = "/integrations/:id"
	APIURLIntegrationCredentials = APIURLIntegration + "/credentials"

	APIURLDevice                 = "/devices/:id"
	APIURLDeviceState            = APIURLDevice + "/state"
	APIURLDeviceStateIntegration = APIURLDevice + "/state/:integrationId"

	APIURLEvents = "/events"
)

const (
	defaultTimeout = time.Second * 10
)

type Config struct {
	Client *http.Client
}

// NewConfig initializes a new empty config and optionally merges the
// configurations provided as argument.
func NewConfig(configs ...*Config) *Config {
	var config = new(Config)
	for _, conf := range configs {
		if conf == nil {
			continue
		}
		if conf.Client != nil {
			config.Client = conf.Client
		}
	}
	return config
}

func (conf *Config) SetClient(client *http.Client) *Config {
	conf.Client = client
	return conf
}

// NewRouter returns the gin router
func NewRouter(
	app app.App,
	config ...*Config,
) *gin.Engine {
	conf := NewConfig(config...)
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	handler := NewAPIHandler(app, conf)
	internal := (*InternalHandler)(handler)
	management := (*ManagementHandler)(handler)

	router := gin.New()
	router.Use(accesslog.Middleware())
	router.Use(requestid.Middleware())

	router.NoRoute(handler.NoRoute)

	internalAPI := router.Group(APIURLInternal)
	internalAPI.GET(APIURLAlive, handler.Alive)
	internalAPI.GET(APIURLHealth, handler.Health)

	internalAPI.POST(APIURLTenantDevices, internal.ProvisionDevice)
	internalAPI.DELETE(APIURLTenantDevice, internal.DecommissionDevice)
	internalAPI.PUT(APIURLTenantBulkStatus, internal.BulkSetDeviceStatus)

	internalAPI.POST(APIURLTenantAuth, internal.PreauthorizeHandler)

	managementAPI := router.Group(APIURLManagement, identity.Middleware())
	managementAPI.GET(APIURLIntegrations, management.GetIntegrations)
	managementAPI.GET(APIURLIntegration, management.GetIntegrationById)
	managementAPI.POST(APIURLIntegrations, management.CreateIntegration)
	managementAPI.PUT(APIURLIntegrationCredentials, management.SetIntegrationCredentials)
	managementAPI.DELETE(APIURLIntegration, management.RemoveIntegration)

	managementAPI.GET(APIURLDeviceState, management.GetDeviceState)
	managementAPI.GET(APIURLDeviceStateIntegration, management.GetDeviceStateIntegration)
	managementAPI.PUT(APIURLDeviceStateIntegration, management.SetDeviceStateIntegration)

	managementAPI.GET(APIURLEvents, management.GetEvents)

	return router
}

type APIHandler struct {
	*http.Client
	app app.App
}

func NewAPIHandler(app app.App, config ...*Config) *APIHandler {
	conf := NewConfig(config...)
	if conf.Client == nil {
		conf.Client = new(http.Client)
	}
	return &APIHandler{
		Client: conf.Client,
		app:    app,
	}
}

// Alive responds to GET /alive
func (h *APIHandler) Alive(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Health responds to GET /health
func (h *APIHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := h.app.HealthCheck(ctx)
	if err != nil {
		l.Error(errors.Wrap(err, "health check failed"))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) NoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, rest.Error{
		Err:       "not found",
		RequestID: requestid.FromContext(c.Request.Context()),
	})
}

// Make gin-gonic use validatable structs instead of relying on go-playground
// validator interface.
type validateValidatableValidator struct{}

func (validateValidatableValidator) ValidateStruct(obj interface{}) error {
	if v, ok := obj.(interface{ Validate() error }); ok {
		return v.Validate()
	}
	return nil
}

func (validateValidatableValidator) Engine() interface{} {
	return nil
}

func init() {
	binding.Validator = validateValidatableValidator{}
}
