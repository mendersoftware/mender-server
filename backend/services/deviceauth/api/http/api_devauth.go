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
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/accesslog"
	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deviceauth/access"
	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	uriAuthReqs = "/api/devices/v1/authentication/auth_requests"

	// internal API
	uriAlive              = "/api/internal/v1/devauth/alive"
	uriHealth             = "/api/internal/v1/devauth/health"
	uriTokenVerify        = "/api/internal/v1/devauth/tokens/verify"
	uriTenantLimit        = "/api/internal/v1/devauth/tenant/#id/limits/#name"
	uriTokens             = "/api/internal/v1/devauth/tokens"
	uriTenants            = "/api/internal/v1/devauth/tenants"
	uriTenantDevice       = "/api/internal/v1/devauth/tenants/#tid/devices/#did"
	uriTenantDeviceStatus = "/api/internal/v1/devauth/tenants/#tid/devices/#did/status"
	uriTenantDevices      = "/api/internal/v1/devauth/tenants/#tid/devices"
	uriTenantDevicesCount = "/api/internal/v1/devauth/tenants/#tid/devices/count"

	// management API v2
	v2uriDevices             = "/api/management/v2/devauth/devices"
	v2uriDevicesCount        = "/api/management/v2/devauth/devices/count"
	v2uriDevicesSearch       = "/api/management/v2/devauth/devices/search"
	v2uriDevice              = "/api/management/v2/devauth/devices/#id"
	v2uriDeviceAuthSet       = "/api/management/v2/devauth/devices/#id/auth/#aid"
	v2uriDeviceAuthSetStatus = "/api/management/v2/devauth/devices/#id/auth/#aid/status"
	v2uriToken               = "/api/management/v2/devauth/tokens/#id"
	v2uriDevicesLimit        = "/api/management/v2/devauth/limits/#name"

	HdrAuthReqSign = "X-MEN-Signature"
)

func init() {
	rest.ErrorFieldName = "error"
}

const (
	defaultTimeout = time.Second * 5
)

var (
	ErrIncorrectStatus = errors.New("incorrect device status")
	ErrNoAuthHeader    = errors.New("no authorization header")
)

type DevAuthApiHandlers struct {
	devAuth devauth.App
	db      store.DataStore
}

type DevAuthApiStatus struct {
	Status string `json:"status"`
}

func NewDevAuthApiHandlers(devAuth devauth.App, db store.DataStore) ApiHandler {
	return &DevAuthApiHandlers{
		devAuth: devAuth,
		db:      db,
	}
}

func wrapMiddleware(middleware rest.Middleware, routes ...*rest.Route) []*rest.Route {
	for _, route := range routes {
		route.Func = middleware.MiddlewareFunc(route.Func)
	}
	return routes
}

func (d *DevAuthApiHandlers) Build() (http.Handler, error) {
	identityMiddleware := &identity.IdentityMiddleware{
		UpdateLogger: true,
	}
	internalRoutes := []*rest.Route{
		rest.Get(uriAlive, d.AliveHandler),
		rest.Get(uriHealth, d.HealthCheckHandler),
		rest.Get(uriTokenVerify, identityMiddleware.MiddlewareFunc(
			d.VerifyTokenHandler,
		)),
		rest.Post(uriTokenVerify, identityMiddleware.MiddlewareFunc(
			d.VerifyTokenHandler,
		)),
		rest.Delete(uriTokens, d.DeleteTokensHandler),

		rest.Put(uriTenantLimit, d.PutTenantLimitHandler),
		rest.Get(uriTenantLimit, d.GetTenantLimitHandler),
		rest.Delete(uriTenantLimit, d.DeleteTenantLimitHandler),

		rest.Post(uriTenants, d.ProvisionTenantHandler),
		rest.Get(uriTenantDeviceStatus, d.GetTenantDeviceStatus),
		rest.Get(uriTenantDevices, d.GetTenantDevicesHandler),
		rest.Get(uriTenantDevicesCount, d.GetTenantDevicesCountHandler),
		rest.Delete(uriTenantDevice, d.DeleteDeviceHandler),
	}
	publicRoutes := []*rest.Route{
		// Devices API
		rest.Post(uriAuthReqs, d.SubmitAuthRequestHandler),

		// API v2
		rest.Get(v2uriDevicesCount, d.GetDevicesCountHandler),
		rest.Get(v2uriDevices, d.GetDevicesV2Handler),
		rest.Post(v2uriDevicesSearch, d.SearchDevicesV2Handler),
		rest.Post(v2uriDevices, d.PostDevicesV2Handler),
		rest.Get(v2uriDevice, d.GetDeviceV2Handler),
		rest.Delete(v2uriDevice, d.DecommissionDeviceHandler),
		rest.Delete(v2uriDeviceAuthSet, d.DeleteDeviceAuthSetHandler),
		rest.Put(v2uriDeviceAuthSetStatus, d.UpdateDeviceStatusHandler),
		rest.Get(v2uriDeviceAuthSetStatus, d.GetAuthSetStatusHandler),
		rest.Delete(v2uriToken, d.DeleteTokenHandler),
		rest.Get(v2uriDevicesLimit, d.GetLimitHandler),
	}
	publicRoutes = wrapMiddleware(identityMiddleware, publicRoutes...)

	routes := append(publicRoutes, internalRoutes...)

	app, err := rest.MakeRouter(
		// augment routes with OPTIONS handler
		AutogenOptionsRoutes(routes, AllowHeaderOptionsGenerator)...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	api := rest.NewApi()
	api.SetApp(app)
	api.Use(
		&requestlog.RequestLogMiddleware{},
		&requestid.RequestIdMiddleware{},
		&accesslog.AccessLogMiddleware{
			Format: accesslog.SimpleLogFormat,
			DisableLog: func(statusCode int, r *rest.Request) bool {
				if statusCode < 300 &&
					(r.Request.URL.Path == uriAlive ||
						r.Request.URL.Path == uriHealth ||
						r.Request.URL.Path == uriTokenVerify) {
					return true
				}
				return false
			},
		},
		// verifies the request Content-Type header
		// The expected Content-Type is 'application/json'
		// if the content is non-null
		&rest.ContentTypeCheckerMiddleware{},
	)

	return api.MakeHandler(), nil
}

func (d *DevAuthApiHandlers) AliveHandler(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) HealthCheckHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := d.devAuth.HealthCheck(ctx)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) SubmitAuthRequestHandler(w rest.ResponseWriter, r *rest.Request) {
	var authreq model.AuthReq

	ctx := r.Context()

	l := log.FromContext(ctx)

	//validate req body by reading raw content manually
	//(raw body will be needed later, DecodeJsonPayload would
	//unmarshal and close it)
	body, err := utils.ReadBodyRaw(r)
	if err != nil {
		err = errors.Wrap(err, "failed to decode auth request")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &authreq)
	if err != nil {
		err = errors.Wrap(err, "failed to decode auth request")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = authreq.Validate()
	if err != nil {
		err = errors.Wrap(err, "invalid auth request")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	//verify signature
	signature := r.Header.Get(HdrAuthReqSign)
	if signature == "" {
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.New("missing request signature header"),
			http.StatusBadRequest,
		)
		return
	}

	token, err := d.devAuth.SubmitAuthRequest(ctx, &authreq)
	if err != nil {
		if devauth.IsErrDevAuthUnauthorized(err) {
			rest_utils.RestErrWithWarningMsg(w, r, l, err,
				http.StatusUnauthorized, errors.Cause(err).Error())
			return
		} else if devauth.IsErrDevAuthBadRequest(err) {
			rest_utils.RestErrWithWarningMsg(w, r, l, err,
				http.StatusBadRequest, errors.Cause(err).Error())
			return
		}
	}

	switch err {
	case nil:
		err = utils.VerifyAuthReqSign(signature, authreq.PubKeyStruct, body)
		if err != nil {
			rest_utils.RestErrWithLogMsg(
				w,
				r,
				l,
				err,
				http.StatusUnauthorized,
				"signature verification failed",
			)
			return
		}
		w.Header().Set("Content-Type", "application/jwt")
		_, _ = w.(http.ResponseWriter).Write([]byte(token))
		return
	case devauth.ErrDevIdAuthIdMismatch, devauth.ErrMaxDeviceCountReached:
		// error is always set to unauthorized, client does not need to
		// know why
		rest_utils.RestErrWithWarningMsg(w, r, l, devauth.ErrDevAuthUnauthorized,
			http.StatusUnauthorized, "unauthorized")
		return
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}
}

func (d *DevAuthApiHandlers) PostDevicesV2Handler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	req, err := parsePreAuthReq(r.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to decode preauth request")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	reqDbModel, err := req.getDbModel()
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	device, err := d.devAuth.PreauthorizeDevice(ctx, reqDbModel)
	switch err {
	case nil:
		w.Header().Set("Location", "devices/"+reqDbModel.DeviceId)
		w.WriteHeader(http.StatusCreated)
	case devauth.ErrDeviceExists:
		l.Error(err)
		w.WriteHeader(http.StatusConflict)
		_ = w.WriteJson(device)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
	}
}

func (d *DevAuthApiHandlers) SearchDevicesV2Handler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	page, perPage, err := rest_utils.ParsePagination(r)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}
	fltr := model.DeviceFilter{}

	switch strings.ToLower(r.Header.Get("Content-Type")) {
	case "application/json", "":
		err := r.DecodeJsonPayload(&fltr)
		if err != nil {
			err = errors.Wrap(err, "api: malformed request body")
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
			return
		}

	case "application/x-www-form-urlencoded":
		if err = r.ParseForm(); err != nil {
			err = errors.Wrap(err, "api: malformed query parameters")
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
			return
		}
		if err = fltr.ParseForm(r.Form); err != nil {
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
			return
		}

	default:
		rest_utils.RestErrWithLog(w, r, l, errors.Errorf(
			"Content-Type '%s' not supported",
			r.Header.Get("Content-Type"),
		), http.StatusUnsupportedMediaType)
		return
	}

	skip := (page - 1) * perPage
	limit := perPage + 1
	devs, err := d.devAuth.GetDevices(ctx, uint(skip), uint(limit), fltr)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	numDevs := len(devs)
	hasNext := false
	if uint64(numDevs) > perPage {
		hasNext = true
		numDevs = int(perPage)
	}

	links := rest_utils.MakePageLinkHdrs(r, page, perPage, hasNext)

	for _, l := range links {
		w.Header().Add("Link", l)
	}

	_ = w.WriteJson(devs[:numDevs])
}

func (d *DevAuthApiHandlers) GetDevicesV2Handler(w rest.ResponseWriter, r *rest.Request) {
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	d.SearchDevicesV2Handler(w, r)
}

func (d *DevAuthApiHandlers) GetDevicesCountHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	status := r.URL.Query().Get("status")

	switch status {
	case model.DevStatusAccepted,
		model.DevStatusRejected,
		model.DevStatusPending,
		model.DevStatusPreauth,
		model.DevStatusNoAuth,
		"":
	default:
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.New("status must be one of: pending, accepted, rejected, preauthorized, noauth"),
			http.StatusBadRequest,
		)
		return
	}

	count, err := d.devAuth.GetDevCountByStatus(ctx, status)

	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	_ = w.WriteJson(model.Count{Count: count})
}

func (d *DevAuthApiHandlers) GetDeviceV2Handler(w rest.ResponseWriter, r *rest.Request) {

	ctx := r.Context()

	l := log.FromContext(ctx)

	devId := r.PathParam("id")

	dev, err := d.devAuth.GetDevice(ctx, devId)
	switch {
	case err == store.ErrDevNotFound:
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
	case dev != nil:
		_ = w.WriteJson(dev)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
	}
}

func (d *DevAuthApiHandlers) DecommissionDeviceHandler(w rest.ResponseWriter, r *rest.Request) {

	ctx := r.Context()

	l := log.FromContext(ctx)

	devId := r.PathParam("id")

	if err := d.devAuth.DecommissionDevice(ctx, devId); err != nil {
		if err == store.ErrDevNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) DeleteDeviceAuthSetHandler(w rest.ResponseWriter, r *rest.Request) {

	ctx := r.Context()

	l := log.FromContext(ctx)

	devId := r.PathParam("id")
	authId := r.PathParam("aid")

	if err := d.devAuth.DeleteAuthSet(ctx, devId, authId); err != nil {
		if err == store.ErrAuthSetNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) DeleteTokenHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tokenID := r.PathParam("id")
	err := d.devAuth.RevokeToken(ctx, tokenID)
	if err != nil {
		if err == store.ErrTokenNotFound ||
			err == devauth.ErrInvalidAuthSetID {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) VerifyTokenHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tokenStr, err := extractToken(r.Header)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, ErrNoAuthHeader, http.StatusUnauthorized)
		return
	}

	ctx = ctxhttpheader.WithContext(ctx,
		r.Header,
		"X-Forwarded-Method",
		"X-Forwarded-Uri")

	// verify token
	err = d.devAuth.VerifyToken(ctx, tokenStr)
	code := http.StatusOK
	if err != nil {
		switch e := errors.Cause(err); e {
		case jwt.ErrTokenExpired:
			code = http.StatusForbidden
		case store.ErrTokenNotFound, store.ErrAuthSetNotFound, jwt.ErrTokenInvalid:
			code = http.StatusUnauthorized
		case cache.ErrTooManyRequests:
			code = http.StatusTooManyRequests
		default:
			if _, ok := e.(access.PermissionError); ok {
				rest_utils.RestErrWithLog(w, r, l, e, http.StatusForbidden)
			} else {
				rest_utils.RestErrWithLogInternal(w, r, l, err)
			}
			return
		}
		l.Error(err)
	}

	w.WriteHeader(code)
}

func (d *DevAuthApiHandlers) UpdateDeviceStatusHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	devid := r.PathParam("id")
	authid := r.PathParam("aid")

	var status DevAuthApiStatus
	err := r.DecodeJsonPayload(&status)
	if err != nil {
		err = errors.Wrap(err, "failed to decode status data")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	if err := statusValidate(&status); err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	if status.Status == model.DevStatusAccepted {
		err = d.devAuth.AcceptDeviceAuth(ctx, devid, authid)
	} else if status.Status == model.DevStatusRejected {
		err = d.devAuth.RejectDeviceAuth(ctx, devid, authid)
	} else if status.Status == model.DevStatusPending {
		err = d.devAuth.ResetDeviceAuth(ctx, devid, authid)
	}
	if err != nil {
		switch err {
		case store.ErrDevNotFound, store.ErrAuthSetNotFound:
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
		case devauth.ErrDevIdAuthIdMismatch, devauth.ErrDevAuthBadRequest:
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		case devauth.ErrMaxDeviceCountReached:
			rest_utils.RestErrWithLog(w, r, l, err, http.StatusUnprocessableEntity)
		default:
			rest_utils.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type LimitValue struct {
	Limit uint64 `json:"limit"`
}

func (d *DevAuthApiHandlers) PutTenantLimitHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.PathParam("id")
	reqLimitName := r.PathParam("name")

	if !model.IsValidLimit(reqLimitName) {
		rest_utils.RestErrWithLog(w, r, l,
			errors.Errorf("unsupported limit %v", reqLimitName),
			http.StatusBadRequest)
		return
	}

	var value LimitValue
	err := r.DecodeJsonPayload(&value)
	if err != nil {
		err = errors.Wrap(err, "failed to decode limit request")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	limit := model.Limit{
		Value: value.Limit,
		Name:  reqLimitName,
	}

	if err := d.devAuth.SetTenantLimit(ctx, tenantId, limit); err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) DeleteTenantLimitHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.PathParam("id")
	reqLimitName := r.PathParam("name")

	if !model.IsValidLimit(reqLimitName) {
		rest_utils.RestErrWithLog(w, r, l,
			errors.Errorf("unsupported limit %v", reqLimitName),
			http.StatusBadRequest)
		return
	}

	if err := d.devAuth.DeleteTenantLimit(ctx, tenantId, reqLimitName); err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DevAuthApiHandlers) GetTenantLimitHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.PathParam("id")
	limitName := r.PathParam("name")

	if !model.IsValidLimit(limitName) {
		rest_utils.RestErrWithLog(w, r, l,
			errors.Errorf("unsupported limit %v", limitName),
			http.StatusBadRequest)
		return
	}

	lim, err := d.devAuth.GetTenantLimit(ctx, limitName, tenantId)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	_ = w.WriteJson(LimitValue{lim.Value})
}

func (d *DevAuthApiHandlers) GetLimitHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	name := r.PathParam("name")

	if !model.IsValidLimit(name) {
		rest_utils.RestErrWithLog(w, r, l,
			errors.Errorf("unsupported limit %v", name),
			http.StatusBadRequest)
		return
	}

	lim, err := d.devAuth.GetLimit(ctx, name)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	_ = w.WriteJson(LimitValue{lim.Value})
}

func (d *DevAuthApiHandlers) DeleteTokensHandler(w rest.ResponseWriter, r *rest.Request) {

	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.URL.Query().Get("tenant_id")
	if tenantId == "" {
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.New("tenant_id must be provided"),
			http.StatusBadRequest,
		)
		return
	}
	devId := r.URL.Query().Get("device_id")

	err := d.devAuth.DeleteTokens(ctx, tenantId, devId)
	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
	}
}

func (d *DevAuthApiHandlers) GetAuthSetStatusHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	devid := r.PathParam("id")
	authid := r.PathParam("aid")

	// get authset directly from store
	aset, err := d.db.GetAuthSetById(ctx, authid)
	switch err {
	case nil:
		_ = w.WriteJson(&model.Status{Status: aset.Status})
	case store.ErrDevNotFound, store.ErrAuthSetNotFound:
		rest_utils.RestErrWithLog(w, r, l, store.ErrAuthSetNotFound, http.StatusNotFound)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l,
			errors.Wrapf(err,
				"failed to fetch auth set %s for device %s",
				authid, devid))
	}
}

func (d *DevAuthApiHandlers) ProvisionTenantHandler(w rest.ResponseWriter, r *rest.Request) {
	// NOTE: This handler was used to initialize database collections. This is no longer
	//       needed after migration 2.0.0.
	w.WriteHeader(http.StatusCreated)
}

func (d *DevAuthApiHandlers) GetTenantDeviceStatus(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tid := r.PathParam("tid")
	did := r.PathParam("did")

	if did == "" {
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.New("device id (did) cannot be empty"),
			http.StatusBadRequest,
		)
		return
	}

	status, err := d.devAuth.GetTenantDeviceStatus(ctx, tid, did)
	switch err {
	case nil:
		_ = w.WriteJson(status)
	case devauth.ErrDeviceNotFound:
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
	}
}

func (d *DevAuthApiHandlers) GetTenantDevicesHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	if tid := r.PathParam("tid"); tid != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})
	}
	r.Request = r.WithContext(ctx)

	d.GetDevicesV2Handler(w, r)
}

func (d *DevAuthApiHandlers) GetTenantDevicesCountHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	if tid := r.PathParam("tid"); tid != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})
	}
	r.Request = r.WithContext(ctx)

	d.GetDevicesCountHandler(w, r)
}

func (d *DevAuthApiHandlers) DeleteDeviceHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)
	did := r.PathParam("did")

	err := d.devAuth.DeleteDevice(ctx, did)
	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	case devauth.ErrInvalidDeviceID:
		didErr := errors.New("device id (did) cannot be empty")
		rest_utils.RestErrWithLog(w, r, l, didErr, http.StatusBadRequest)
	case store.ErrDevNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		rest_utils.RestErrWithLogInternal(w, r, l, err)
	}
}

// Validate status.
// Expected statuses:
// - "accepted"
// - "rejected"
// - "pending"
func statusValidate(status *DevAuthApiStatus) error {
	if status.Status != model.DevStatusAccepted &&
		status.Status != model.DevStatusRejected &&
		status.Status != model.DevStatusPending {
		return ErrIncorrectStatus
	} else {
		return nil
	}
}

// extracts JWT from authorization header
func extractToken(header http.Header) (string, error) {
	const authHeaderName = "Authorization"
	authHeader := header.Get(authHeaderName)
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}
	tokenStr := strings.Replace(authHeader, "Bearer", "", 1)
	tokenStr = strings.Replace(tokenStr, "bearer", "", 1)
	return strings.TrimSpace(tokenStr), nil
}
