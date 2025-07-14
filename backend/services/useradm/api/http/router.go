package http

import (
	"net/http"

	"github.com/mendersoftware/mender-server/pkg/contenttype"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/routing"
)

const (
	apiUrlManagementV1       = "/api/management/v1/useradm"
	uriManagementAuthLogin   = "/auth/login"
	uriManagementAuthLogout  = "/auth/logout"
	uriManagementUser        = "/users/:id"
	uriManagementUsers       = "/users"
	uriManagementSettings    = "/settings"
	uriManagementSettingsMe  = "/settings/me"
	uriManagementTokens      = "/settings/tokens"
	uriManagementToken       = "/settings/tokens/:id"
	uriManagementPlans       = "/plans"
	uriManagementPlanBinding = "/plan_binding"

	apiUrlInternalV1  = "/api/internal/v1/useradm"
	uriInternalAlive  = "/alive"
	uriInternalHealth = "/health"

	uriInternalAuthVerify  = "/auth/verify"
	uriInternalTenants     = "/tenants"
	uriInternalTenantUsers = "/tenants/:id/users"
	uriInternalTenantUser  = "/tenants/:id/users/:userid"
	uriInternalTokens      = "/tokens"
)

func MakeRouter(i *UserAdmApiHandlers) http.Handler {
	router := routing.NewGinRouter()

	mgmt := router.Group(apiUrlManagementV1)

	mgmt.Group(".").Use(contenttype.CheckJSON()).
		POST(uriManagementAuthLogin, i.AuthLoginHandler)

	mgmt.Use(identity.Middleware())

	mgmt.GET(uriManagementUsers, i.GetUsersHandler)
	mgmt.GET(uriManagementUser, i.GetUserHandler)
	mgmt.GET(uriManagementSettings, i.GetSettingsHandler)
	mgmt.GET(uriManagementSettingsMe, i.GetSettingsMeHandler)
	mgmt.GET(uriManagementTokens, i.GetTokensHandler)
	mgmt.GET(uriManagementPlans, i.GetPlansHandler)
	mgmt.GET(uriManagementPlanBinding, i.GetPlanBindingHandler)
	mgmt.DELETE(uriManagementUser, i.DeleteUserHandler)
	mgmt.DELETE(uriManagementToken, i.DeleteTokenHandler)

	mgmt.Group(".").Use(contenttype.CheckJSON()).
		POST(uriManagementAuthLogout, i.AuthLogoutHandler).
		POST(uriManagementUsers, i.AddUserHandler).
		PUT(uriManagementUser, i.UpdateUserHandler).
		POST(uriManagementSettings, i.SaveSettingsHandler).
		POST(uriManagementSettingsMe, i.SaveSettingsMeHandler).
		POST(uriManagementTokens, i.IssueTokenHandler)

	routing.AutogenOptionsRoutes(router,
		routing.AllowHeaderOptionsGenerator)

	internal := router.Group(apiUrlInternalV1)

	internal.GET(uriInternalAlive, i.AliveHandler)
	internal.GET(uriInternalHealth, i.HealthHandler)

	internal.GET(uriInternalAuthVerify, identity.Middleware(),
		i.AuthVerifyHandler)
	internal.POST(uriInternalAuthVerify, identity.Middleware(),
		i.AuthVerifyHandler)

	internal.POST(uriInternalTenants, i.CreateTenantHandler)
	internal.POST(uriInternalTenantUsers, i.CreateTenantUserHandler)
	internal.DELETE(uriInternalTenantUser, i.DeleteTenantUserHandler)
	internal.GET(uriInternalTenantUsers, i.GetTenantUsersHandler)
	internal.DELETE(uriInternalTokens, i.DeleteTokensHandler)

	return router
}
