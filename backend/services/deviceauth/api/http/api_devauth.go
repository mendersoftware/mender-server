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

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deviceauth/access"
	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

const (
	defaultTimeout = time.Second * 5
)

var (
	ErrIncorrectStatus = errors.New("incorrect device status")
	ErrNoAuthHeader    = errors.New("Authorization not present in header")
)

type DevAuthApiHandlers struct {
	app devauth.App
	db  store.DataStore
}

type DevAuthApiStatus struct {
	Status string `json:"status"`
}

func NewDevAuthApiHandlers(devAuth devauth.App, db store.DataStore) *DevAuthApiHandlers {
	return &DevAuthApiHandlers{
		app: devAuth,
		db:  db,
	}
}

func (i *DevAuthApiHandlers) AliveHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) HealthCheckHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := i.app.HealthCheck(ctx)
	if err != nil {
		rest.RenderError(c, http.StatusServiceUnavailable, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) SubmitAuthRequestHandler(c *gin.Context) {
	var authreq model.AuthReq

	ctx := c.Request.Context()

	//validate req body by reading raw content manually
	//(raw body will be needed later, DecodeJsonPayload would
	//unmarshal and close it)
	body, err := utils.ReadBodyRaw(c.Request)
	if err != nil {
		err = errors.Wrap(err, "failed to decode auth request")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	err = json.Unmarshal(body, &authreq)
	if err != nil {
		err = errors.Wrap(err, "failed to decode auth request")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	err = authreq.Validate()
	if err != nil {
		err = errors.Wrap(err, "invalid auth request")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	//verify signature
	signature := c.GetHeader(HdrAuthReqSign)
	if signature == "" {
		rest.RenderError(c, http.StatusBadRequest,
			errors.New("missing request signature header"),
		)
		return
	}

	token, err := i.app.SubmitAuthRequest(ctx, &authreq)
	if err != nil {
		if devauth.IsErrDevAuthUnauthorized(err) {
			rest.RenderError(c,
				http.StatusUnauthorized,
				errors.Cause(err),
			)
			return
		} else if devauth.IsErrDevAuthBadRequest(err) {
			rest.RenderError(c,
				http.StatusBadRequest,
				errors.Cause(err),
			)
			return
		}
	}

	switch err {
	case nil:
		err = utils.VerifyAuthReqSign(signature, authreq.PubKeyStruct, body)
		if err != nil {
			rest.RenderErrorWithMessage(c,
				http.StatusUnauthorized,
				errors.Cause(err),
				"signature verification failed",
			)
			return
		}
		c.Header("Content-Type", "application/jwt")
		_, _ = c.Writer.Write([]byte(token))
		return
	case devauth.ErrDevIdAuthIdMismatch, devauth.ErrMaxDeviceCountReached:
		// error is always set to unauthorized, client does not need to
		// know why
		rest.RenderErrorWithMessage(c,
			http.StatusUnauthorized,
			errors.Cause(err),
			"unauthorized",
		)
		return
	default:
		rest.RenderInternalError(c, err)
		return
	}
}

func (i *DevAuthApiHandlers) PostDevicesV2Handler(c *gin.Context) {
	ctx := c.Request.Context()

	req, err := parsePreAuthReq(c.Request.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to decode preauth request")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	reqDbModel, err := req.getDbModel()
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	device, err := i.app.PreauthorizeDevice(ctx, reqDbModel)
	switch err {
	case nil:
		c.Header("Location", "devices/"+reqDbModel.DeviceId)
		c.Status(http.StatusCreated)
	case devauth.ErrDeviceExists:
		c.JSON(http.StatusConflict, device)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (i *DevAuthApiHandlers) SearchDevicesV2Handler(c *gin.Context) {
	ctx := c.Request.Context()

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}
	fltr := model.DeviceFilter{}

	switch strings.ToLower(c.GetHeader("Content-Type")) {
	case "application/json", "":
		err := c.ShouldBindJSON(&fltr)
		if err != nil {
			err = errors.Wrap(err, "api: malformed request body")
			rest.RenderError(c, http.StatusBadRequest, err)
			return
		}
	case "application/x-www-form-urlencoded":
		if err = c.Request.ParseForm(); err != nil {
			err = errors.Wrap(err, "api: malformed query parameters")
			rest.RenderError(c, http.StatusBadRequest, err)
			return
		}
		if err = fltr.ParseForm(c.Request.Form); err != nil {
			rest.RenderError(c, http.StatusBadRequest, err)
			return
		}

	default:
		rest.RenderError(c,
			http.StatusUnsupportedMediaType,
			errors.Errorf(
				"Content-Type '%s' not supported",
				c.GetHeader("Content-Type"),
			))
		return
	}

	skip := (page - 1) * perPage
	limit := perPage + 1
	devs, err := i.app.GetDevices(ctx, uint(skip), uint(limit), fltr)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	numDevs := len(devs)
	hasNext := false
	if int64(numDevs) > perPage {
		hasNext = true
		numDevs = int(perPage)
	}
	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
		SetHasNext(hasNext).
		SetTotalCount(int64(numDevs))
	links, err := rest.MakePagingHeaders(c.Request, hints)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
	}
	for _, l := range links {
		c.Writer.Header().Add("Link", l)
	}
	c.JSON(http.StatusOK, devs[:numDevs])
}

func (i *DevAuthApiHandlers) GetDevicesV2Handler(c *gin.Context) {
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	i.SearchDevicesV2Handler(c)
}

func (i *DevAuthApiHandlers) GetDevicesCountHandler(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Query("status")

	switch status {
	case model.DevStatusAccepted,
		model.DevStatusRejected,
		model.DevStatusPending,
		model.DevStatusPreauth,
		model.DevStatusNoAuth,
		"":
	default:
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("status must be one of: pending, accepted, rejected, preauthorized, noauth"),
		)
		return
	}

	count, err := i.app.GetDevCountByStatus(ctx, status)

	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.Count{Count: count})
}

func (i *DevAuthApiHandlers) GetDeviceV2Handler(c *gin.Context) {

	ctx := c.Request.Context()

	devId := c.Param("id")

	dev, err := i.app.GetDevice(ctx, devId)
	switch {
	case err == store.ErrDevNotFound:
		rest.RenderError(c, http.StatusNotFound, err)
	case dev != nil:
		c.JSON(http.StatusOK, dev)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (i *DevAuthApiHandlers) DecommissionDeviceHandler(c *gin.Context) {

	ctx := c.Request.Context()

	devId := c.Param("id")

	if err := i.app.DecommissionDevice(ctx, devId); err != nil {
		if err == store.ErrDevNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) DeleteDeviceAuthSetHandler(c *gin.Context) {

	ctx := c.Request.Context()

	devId := c.Param("id")
	authId := c.Param("aid")

	if err := i.app.DeleteAuthSet(ctx, devId, authId); err != nil {
		if err == store.ErrAuthSetNotFound {
			c.Status(http.StatusNotFound)
			return
		}
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) DeleteTokenHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tokenID := c.Param("id")
	err := i.app.RevokeToken(ctx, tokenID)
	if err != nil {
		if err == store.ErrTokenNotFound ||
			err == devauth.ErrInvalidAuthSetID {
			c.Status(http.StatusNotFound)
			return
		}
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) VerifyTokenHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tokenStr, err := extractToken(c.Request.Header)
	if err != nil {
		rest.RenderError(c, http.StatusUnauthorized, ErrNoAuthHeader)
		return
	}

	ctx = ctxhttpheader.WithContext(ctx,
		c.Request.Header,
		"X-Forwarded-Method",
		"X-Forwarded-Uri")

	// verify token
	err = i.app.VerifyToken(ctx, tokenStr)
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
				rest.RenderError(c, http.StatusForbidden, e)
			} else {
				rest.RenderInternalError(c, err)
			}
			return
		}
	}

	c.Status(code)
}

func (i *DevAuthApiHandlers) UpdateDeviceStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()

	devid := c.Param("id")
	authid := c.Param("aid")

	var status DevAuthApiStatus
	err := c.ShouldBindJSON(&status)
	if err != nil {
		err = errors.Wrap(err, "failed to decode status data")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if err := statusValidate(&status); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if status.Status == model.DevStatusAccepted {
		err = i.app.AcceptDeviceAuth(ctx, devid, authid)
	} else if status.Status == model.DevStatusRejected {
		err = i.app.RejectDeviceAuth(ctx, devid, authid)
	} else if status.Status == model.DevStatusPending {
		err = i.app.ResetDeviceAuth(ctx, devid, authid)
	}
	if err != nil {
		switch err {
		case store.ErrDevNotFound, store.ErrAuthSetNotFound:
			rest.RenderError(c, http.StatusNotFound, err)
		case devauth.ErrDevIdAuthIdMismatch, devauth.ErrDevAuthBadRequest:
			rest.RenderError(c, http.StatusBadRequest, err)
		case devauth.ErrMaxDeviceCountReached:
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		default:
			rest.RenderInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

type LimitValue struct {
	Limit uint64 `json:"limit"`
}

func (i *DevAuthApiHandlers) PutTenantLimitHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("id")
	reqLimitName := c.Param("name")

	if !model.IsValidLimit(reqLimitName) {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Errorf("unsupported limit %v", reqLimitName))
		return
	}

	var value LimitValue
	err := c.ShouldBindJSON(&value)
	if err != nil {
		err = errors.Wrap(err, "failed to decode limit request")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	limit := model.Limit{
		Value: value.Limit,
		Name:  reqLimitName,
	}

	if err := i.app.SetTenantLimit(ctx, tenantId, limit); err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) DeleteTenantLimitHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("id")
	reqLimitName := c.Param("name")

	if !model.IsValidLimit(reqLimitName) {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Errorf("unsupported limit %v", reqLimitName))
		return
	}

	if err := i.app.DeleteTenantLimit(ctx, tenantId, reqLimitName); err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *DevAuthApiHandlers) GetTenantLimitHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("id")
	limitName := c.Param("name")

	if !model.IsValidLimit(limitName) {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Errorf("unsupported limit %v", limitName))
		return
	}

	lim, err := i.app.GetTenantLimit(ctx, limitName, tenantId)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, LimitValue{lim.Value})
}

func (i *DevAuthApiHandlers) GetLimitHandler(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.Param("name")

	if !model.IsValidLimit(name) {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Errorf("unsupported limit %v", name))
		return
	}

	lim, err := i.app.GetLimit(ctx, name)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, LimitValue{lim.Value})
}

func (i *DevAuthApiHandlers) DeleteTokensHandler(c *gin.Context) {

	ctx := c.Request.Context()

	tenantId := c.Query("tenant_id")
	if tenantId == "" {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("tenant_id must be provided"))
		return
	}
	devId := c.Query("device_id")

	err := i.app.DeleteTokens(ctx, tenantId, devId)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (i *DevAuthApiHandlers) GetAuthSetStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()

	devid := c.Param("id")
	authid := c.Param("aid")

	// get authset directly from store
	aset, err := i.db.GetAuthSetById(ctx, authid)
	switch err {
	case nil:
		c.JSON(http.StatusOK, &model.Status{Status: aset.Status})
	case store.ErrDevNotFound, store.ErrAuthSetNotFound:
		rest.RenderError(c, http.StatusNotFound, store.ErrAuthSetNotFound)
	default:
		rest.RenderInternalError(c,
			errors.Wrapf(err,
				"failed to fetch auth set %s for device %s",
				authid, devid))
	}
}

func (i *DevAuthApiHandlers) ProvisionTenantHandler(c *gin.Context) {
	// NOTE: This handler was used to initialize database collections. This is no longer
	//       needed after migration 2.0.0.
	c.Status(http.StatusCreated)
}

func (i *DevAuthApiHandlers) GetTenantDeviceStatus(c *gin.Context) {
	ctx := c.Request.Context()

	tid := c.Param("tid")
	did := c.Param("did")

	if did == "" {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("device id (did) cannot be empty"))
		return
	}

	status, err := i.app.GetTenantDeviceStatus(ctx, tid, did)
	switch err {
	case nil:
		c.JSON(http.StatusOK, status)
	case devauth.ErrDeviceNotFound:
		rest.RenderError(c, http.StatusNotFound, err)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (i *DevAuthApiHandlers) GetTenantDevicesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if tid := c.Param("tid"); tid != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})
	}
	c.Request = c.Request.WithContext(ctx)

	i.GetDevicesV2Handler(c)
}

func (i *DevAuthApiHandlers) GetTenantDevicesCountHandler(c *gin.Context) {
	ctx := c.Request.Context()
	if tid := c.Param("tid"); tid != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{Tenant: tid})
	}
	c.Request = c.Request.WithContext(ctx)

	i.GetDevicesCountHandler(c)
}

func (i *DevAuthApiHandlers) DeleteDeviceHandler(c *gin.Context) {
	ctx := c.Request.Context()
	did := c.Param("did")

	err := i.app.DeleteDevice(ctx, did)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	case devauth.ErrInvalidDeviceID:
		didErr := errors.New("device id (did) cannot be empty")
		rest.RenderError(c, http.StatusBadRequest, didErr)
	case store.ErrDevNotFound:
		c.Status(http.StatusNotFound)
	default:
		rest.RenderInternalError(c, err)
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
