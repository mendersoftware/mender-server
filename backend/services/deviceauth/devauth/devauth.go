// Copyright 2024 Northern.tech AS
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
package devauth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/addons"
	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/mendersoftware/mender-server/pkg/plan"
	"github.com/mendersoftware/mender-server/pkg/rate"
	"github.com/mendersoftware/mender-server/pkg/ratelimits"
	"github.com/mendersoftware/mender-server/pkg/requestid"

	"github.com/mendersoftware/mender-server/services/deviceauth/access"
	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/inventory"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/tenant"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
	uto "github.com/mendersoftware/mender-server/services/deviceauth/utils/to"
)

const (
	MsgErrDevAuthUnauthorized = "dev auth: unauthorized"
	MsgErrDevAuthBadRequest   = "dev auth: bad request"
	InventoryScopeSystem      = "system"
)

var (
	ErrDevAuthUnauthorized   = errors.New(MsgErrDevAuthUnauthorized)
	ErrDevIdAuthIdMismatch   = errors.New("dev auth: dev ID and auth ID mismatch")
	ErrMaxDeviceCountReached = errors.New("maximum number of accepted devices reached")
	ErrDeviceExists          = errors.New("device already exists")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrDevAuthBadRequest     = errors.New(MsgErrDevAuthBadRequest)

	ErrInvalidDeviceID  = errors.New("invalid device ID type")
	ErrInvalidAuthSetID = errors.New("auth set id is not a valid ID")
)

func IsErrDevAuthUnauthorized(e error) bool {
	return strings.HasPrefix(e.Error(), MsgErrDevAuthUnauthorized)
}

func MakeErrDevAuthUnauthorized(e error) error {
	return errors.Wrap(e, MsgErrDevAuthUnauthorized)
}

func IsErrDevAuthBadRequest(e error) bool {
	return strings.HasPrefix(e.Error(), MsgErrDevAuthBadRequest)
}

func MakeErrDevAuthBadRequest(e error) error {
	return errors.Wrap(e, MsgErrDevAuthBadRequest)
}

// this device auth service interface
//
//go:generate ../../../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	SubmitAuthRequest(ctx context.Context, r *model.AuthReq) (string, error)

	GetDevices(
		ctx context.Context,
		skip,
		limit uint,
		filter model.DeviceFilter,
	) ([]model.Device, error)
	GetDevice(ctx context.Context, dev_id string) (*model.Device, error)
	DecommissionDevice(ctx context.Context, dev_id string) error
	DeleteDevice(ctx context.Context, dev_id string) error
	DeleteAuthSet(ctx context.Context, dev_id string, auth_id string) error
	AcceptDeviceAuth(ctx context.Context, dev_id string, auth_id string) error
	RejectDeviceAuth(ctx context.Context, dev_id string, auth_id string) error
	ResetDeviceAuth(ctx context.Context, dev_id string, auth_id string) error
	PreauthorizeDevice(ctx context.Context, req *model.PreAuthReq) (*model.Device, error)

	RevokeToken(ctx context.Context, tokenID string) error
	VerifyToken(ctx context.Context, token string) error
	DeleteTokens(ctx context.Context, tenantID, deviceID string) error

	SetTenantLimit(ctx context.Context, tenant_id string, limit model.Limit) error
	DeleteTenantLimit(ctx context.Context, tenant_id string, limit string) error

	GetLimit(ctx context.Context, name string) (*model.Limit, error)
	GetTenantLimit(ctx context.Context, name, tenant_id string) (*model.Limit, error)

	GetDevCountByStatus(ctx context.Context, status string) (int, error)

	GetTenantDeviceStatus(ctx context.Context, tenantId, deviceId string) (*model.Status, error)
}

type DevAuth struct {
	db           store.DataStore
	invClient    inventory.Client
	cOrch        orchestrator.ClientRunner
	cTenant      tenant.ClientRunner
	jwt          jwt.Handler
	jwtFallback  jwt.Handler
	verifyTenant bool
	config       Config
	cache        cache.Cache
	clock        utils.Clock
	checker      access.Checker

	// deviceRatelimiter is used to limit authenticated device requests
	rateLimiter              rate.Limiter
	rateLimiterWeights       map[string]float64
	rateLimiterWeightDefault float64
}

type Config struct {
	// token issuer
	Issuer string
	// token expiration time
	ExpirationTime int64
	// Default tenant token to use when the client supplies none. Can be
	// empty
	DefaultTenantToken string
	InventoryAddr      string

	EnableReporting bool
	HaveAddons      bool
}

func NewDevAuth(d store.DataStore, co orchestrator.ClientRunner,
	jwt jwt.Handler, config Config,
) *DevAuth {
	// initialize checker using an empty merge (returns nil on validate)
	checker := access.Merge()
	if config.HaveAddons {
		checker = access.NewAddonChecker()
	}

	return &DevAuth{
		db:           d,
		invClient:    inventory.NewClient(config.InventoryAddr, false),
		cOrch:        co,
		jwt:          jwt,
		verifyTenant: false,
		config:       config,
		clock:        utils.NewClock(),
		checker:      checker,
	}
}

func (d *DevAuth) HealthCheck(ctx context.Context) error {
	err := d.db.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "error reaching MongoDB")
	}
	err = d.invClient.CheckHealth(ctx)
	if err != nil {
		return errors.Wrap(err, "Inventory service unhealthy")
	}
	err = d.cOrch.CheckHealth(ctx)
	if err != nil {
		return errors.Wrap(err, "Workflows service unhealthy")
	}
	if d.verifyTenant {
		err = d.cTenant.CheckHealth(ctx)
		if err != nil {
			return errors.Wrap(err, "Tenantadm service unhealthy")
		}
	}
	return nil
}

func (d *DevAuth) setDeviceIdentity(ctx context.Context, dev *model.Device, tenantId string) error {
	attributes := make([]model.DeviceAttribute, len(dev.IdDataStruct))
	i := 0
	for name, value := range dev.IdDataStruct {
		if name == "status" {
			// we have to forbid the client to override attribute status in identity scope
			// since it stands for status of a device (as in: accepted, rejected, preauthorized)
			continue
		}
		attribute := model.DeviceAttribute{
			Name:  name,
			Value: value,
			Scope: "identity",
		}
		attributes[i] = attribute
		i++
	}
	attrJson, err := json.Marshal(attributes)
	if err != nil {
		return errors.New("internal error: cannot marshal attributes into json")
	}
	if err := d.cOrch.SubmitUpdateDeviceInventoryJob(
		ctx,
		orchestrator.UpdateDeviceInventoryReq{
			RequestId:  requestid.FromContext(ctx),
			TenantId:   tenantId,
			DeviceId:   dev.Id,
			Scope:      "identity",
			Attributes: string(attrJson),
		}); err != nil {
		return errors.Wrap(err, "failed to start device inventory update job")
	}
	if d.config.EnableReporting {
		if err := d.cOrch.SubmitReindexReporting(ctx, string(dev.Id)); err != nil {
			return errors.Wrap(err, "reindex reporting job error")
		}
	}
	return nil
}

func (d *DevAuth) getDeviceFromAuthRequest(
	ctx context.Context,
	r *model.AuthReq,
) (*model.Device, error) {
	dev := model.NewDevice("", r.IdData, r.PubKey)

	l := log.FromContext(ctx)

	idDataStruct, idDataSha256, err := parseIdData(r.IdData)
	if err != nil {
		return nil, MakeErrDevAuthBadRequest(err)
	}

	dev.IdDataStruct = idDataStruct
	dev.IdDataSha256 = idDataSha256

	// record device
	err = d.db.AddDevice(ctx, *dev)
	addDeviceErr := err
	if err != nil && err != store.ErrObjectExists {
		l.Errorf("failed to add/find device: %v", err)
		return nil, err
	}

	// either the device was added or it was already present, in any case,
	// pull it from DB
	dev, err = d.db.GetDeviceByIdentityDataHash(ctx, idDataSha256)
	if err != nil {
		l.Error("failed to find device but could not add either")
		return nil, errors.New("failed to locate device")
	}

	idData := identity.FromContext(ctx)
	tenantId := ""
	if idData != nil {
		tenantId = idData.Tenant
	}
	if addDeviceErr != store.ErrObjectExists {
		if err := d.setDeviceIdentity(ctx, dev, tenantId); err != nil {
			return nil, err
		}
	}

	// check if the device is in the decommissioning state
	if dev.Decommissioning {
		l.Warnf("Device %s in the decommissioning state.", dev.Id)
		return nil, ErrDevAuthUnauthorized
	}

	return dev, nil
}

func (d *DevAuth) signToken(ctx context.Context) jwt.SignFunc {
	return func(t *jwt.Token) (string, error) {
		return d.jwt.ToJWT(t)
	}
}

func (d *DevAuth) doVerifyTenant(ctx context.Context, token string) (*tenant.Tenant, error) {
	t, err := d.cTenant.VerifyToken(ctx, token)
	if err != nil {
		if tenant.IsErrTokenVerificationFailed(err) {
			return nil, MakeErrDevAuthUnauthorized(err)
		}

		return nil, errors.Wrap(err, "request to verify tenant token failed")
	}

	return t, nil
}

func (d *DevAuth) getTenantWithDefault(
	ctx context.Context,
	tenantToken,
	defaultToken string,
) (context.Context, *tenant.Tenant, error) {
	l := log.FromContext(ctx)

	if tenantToken == "" && defaultToken == "" {
		return nil, nil, MakeErrDevAuthUnauthorized(errors.New("tenant token missing"))
	}

	var t *tenant.Tenant
	var err error

	// try the provided token
	// but continue on errors and maybe try the default token
	if tenantToken != "" {
		t, err = d.doVerifyTenant(ctx, tenantToken)
		if err != nil {
			l.Errorf("Failed to verify supplied tenant token: %s", err.Error())
		}
	}

	// if we still haven't selected a tenant - the token didn't work
	// try the default one
	if t == nil && defaultToken != "" {
		t, err = d.doVerifyTenant(ctx, defaultToken)
		if err != nil {
			l.Errorf("Failed to verify default tenant token: %s", err.Error())
		}
	}

	// none of the tokens worked
	if err != nil {
		if tenant.IsErrTokenVerificationFailed(err) {
			return ctx, nil, MakeErrDevAuthUnauthorized(err)
		}
		return ctx, nil, err
	}

	tCtx := identity.WithContext(ctx, &identity.Identity{
		Subject: "internal",
		Tenant:  t.ID,
	})

	return tCtx, t, nil
}

func (d *DevAuth) SubmitAuthRequest(ctx context.Context, r *model.AuthReq) (string, error) {
	l := log.FromContext(ctx)

	var tenant *tenant.Tenant
	var err error

	if d.verifyTenant {
		ctx, tenant, err = d.getTenantWithDefault(ctx, r.TenantToken, d.config.DefaultTenantToken)
		if err != nil {
			return "", err
		}
	} else {
		// ignore identity data when tenant verification is off
		// it's possible that the device will provide old auth token or old tenant token
		// in the authorization header;
		// in that case we need to wipe identity data from the context
		ctx = identity.WithContext(ctx, nil)
	}

	// first, try to handle preauthorization
	authSet, err := d.processPreAuthRequest(ctx, r)
	if err != nil {
		return "", err
	}

	// if not a preauth request, process with regular auth request handling
	if authSet == nil {
		authSet, err = d.processAuthRequest(ctx, r)
		if err != nil {
			return "", err
		}
	}

	// request was already present in DB, check its status
	if authSet.Status == model.DevStatusAccepted {
		jti := oid.FromString(authSet.Id)
		if jti.String() == "" {
			return "", ErrInvalidAuthSetID
		}
		sub := oid.FromString(authSet.DeviceId)
		if sub.String() == "" {
			return "", ErrInvalidDeviceID
		}
		now := time.Now()
		token := &jwt.Token{Claims: jwt.Claims{
			ID:      jti,
			Subject: sub,
			Issuer:  d.config.Issuer,
			ExpiresAt: jwt.Time{
				Time: now.Add(time.Second *
					time.Duration(d.config.ExpirationTime)),
			},
			IssuedAt: jwt.Time{Time: now},
			Device:   true,
		}}

		if d.verifyTenant {
			token.Claims.Tenant = tenant.ID
			token.Claims.Plan = tenant.Plan
			token.Claims.Addons = tenant.Addons
			token.Claims.Trial = tenant.Trial
		} else {
			token.Claims.Plan = plan.PlanEnterprise
			token.Addons = addons.AllAddonsEnabled
		}

		// sign and encode as JWT
		raw, err := token.MarshalJWT(d.signToken(ctx))
		if err != nil {
			return "", errors.Wrap(err, "generate token error")
		}

		if err := d.db.AddToken(ctx, token); err != nil {
			return "", errors.Wrap(err, "add token error")
		}

		l.Infof("Token %s assigned to device %s",
			token.Claims.ID, token.Claims.Subject)
		d.updateCheckInTime(ctx, authSet.DeviceId, token.Claims.Tenant, nil)
		return string(raw), nil
	}

	// no token, return device unauthorized
	return "", ErrDevAuthUnauthorized
}

func (d *DevAuth) handlePreAuthDevice(
	ctx context.Context,
	aset *model.AuthSet,
) (*model.AuthSet, error) {
	var deviceAlreadyAccepted bool
	// check the device status
	// if the device status is accepted then do not trigger provisioning workflow
	// this needs to be checked before changing authentication set status
	dev, err := d.db.GetDeviceById(ctx, aset.DeviceId)
	if err != nil {
		return nil, err
	}

	// check if the device is in the decommissioning state
	if dev.Decommissioning {
		l := log.FromContext(ctx)
		l.Warnf("Device %s in the decommissioning state.", dev.Id)
		return nil, ErrDevAuthUnauthorized
	}

	currentStatus := dev.Status
	if dev.Status != model.DevStatusAccepted {
		// auth set is ok for auto-accepting, check device limit
		allow, err := d.canAcceptDevice(ctx)
		if err != nil {
			return nil, err
		}

		if !allow {
			return nil, ErrMaxDeviceCountReached
		}
	}

	// Ensure that the old acceptable auth sets are rejected
	if err := d.db.RejectAuthSetsForDevice(ctx, aset.DeviceId); err != nil &&
		!errors.Is(err, store.ErrAuthSetNotFound) {
		return nil, errors.Wrap(err, "failed to reject auth sets")
	}
	update := model.AuthSetUpdate{
		Status: model.DevStatusAccepted,
	}
	// persist the 'accepted' status in both auth set, and device
	if err := d.db.UpdateAuthSetById(ctx, aset.Id, update); err != nil {
		return nil, errors.Wrap(err, "failed to update auth set status")
	}

	if err := d.updateDeviceStatus(
		ctx,
		aset.DeviceId,
		model.DevStatusAccepted,
		currentStatus,
	); err != nil {
		return nil, err
	}

	aset.Status = model.DevStatusAccepted
	dev.Status = model.DevStatusAccepted
	dev.AuthSets = append(dev.AuthSets, *aset)

	if !deviceAlreadyAccepted {
		reqId := requestid.FromContext(ctx)
		var tenantID string
		if idty := identity.FromContext(ctx); idty != nil {
			tenantID = idty.Tenant
		}

		// submit device accepted job
		if err := d.cOrch.SubmitProvisionDeviceJob(
			ctx,
			orchestrator.ProvisionDeviceReq{
				RequestId: reqId,
				DeviceID:  aset.DeviceId,
				TenantID:  tenantID,
				Device:    dev,
				Status:    dev.Status,
			}); err != nil {
			return nil, errors.Wrap(err, "submit device provisioning job error")
		}
	}
	return aset, nil
}

func (d *DevAuth) processPreAuthRequest(
	ctx context.Context,
	r *model.AuthReq,
) (*model.AuthSet, error) {
	_, idDataSha256, err := parseIdData(r.IdData)
	if err != nil {
		return nil, MakeErrDevAuthBadRequest(err)
	}

	// authset exists?
	aset, err := d.db.GetAuthSetByIdDataHashKeyByStatus(
		ctx,
		idDataSha256,
		r.PubKey,
		model.DevStatusPreauth,
	)
	switch err {
	case nil:
		break
	case store.ErrAuthSetNotFound:
		return nil, nil
	default:
		return nil, errors.Wrap(err, "failed to fetch auth set")
	}

	// if authset status is not 'preauthorized', nothing to do
	if aset.Status != model.DevStatusPreauth {
		return nil, nil
	}
	return d.handlePreAuthDevice(ctx, aset)
}

func (d *DevAuth) updateDeviceStatus(
	ctx context.Context,
	devId,
	status string,
	currentStatus string,
) error {
	newStatus, err := d.db.GetDeviceStatus(ctx, devId)
	if currentStatus == newStatus {
		return nil
	}
	if status == "" {
		switch err {
		case nil:
			status = newStatus
		case store.ErrAuthSetNotFound:
			status = model.DevStatusNoAuth
		default:
			return errors.Wrap(err, "Cannot determine device status")
		}
	}

	// submit device status change job
	dev, err := d.db.GetDeviceById(ctx, devId)
	if err != nil {
		return errors.Wrap(err, "db get device by id error")
	}

	tenantId := ""
	idData := identity.FromContext(ctx)
	if idData != nil {
		tenantId = idData.Tenant
	}
	req := orchestrator.UpdateDeviceStatusReq{
		RequestId: requestid.FromContext(ctx),
		Devices: []model.DeviceInventoryUpdate{{
			Id:       dev.Id,
			Revision: dev.Revision + 1,
		}},
		TenantId: tenantId,
		Status:   status,
	}
	if err := d.cOrch.SubmitUpdateDeviceStatusJob(ctx, req); err != nil {
		return errors.Wrap(err, "update device status job error")
	}

	if err := d.db.UpdateDevice(ctx,
		devId,
		model.DeviceUpdate{
			Status:    status,
			UpdatedTs: uto.TimePtr(time.Now().UTC()),
		}); err != nil {
		return errors.Wrap(err, "failed to update device status")
	}

	if d.config.EnableReporting {
		if err := d.cOrch.SubmitReindexReporting(ctx, devId); err != nil {
			return errors.Wrap(err, "reindex reporting job error")
		}
	}

	return nil
}

// processAuthRequest will process incoming auth request and record authentication
// data information it contains. Returns a tupe (auth set, error). If no errors were
// present, model.AuthSet.Status will indicate the status of device admission
func (d *DevAuth) processAuthRequest(
	ctx context.Context,
	r *model.AuthReq,
) (*model.AuthSet, error) {
	l := log.FromContext(ctx)

	// get device associated with given authorization request
	dev, err := d.getDeviceFromAuthRequest(ctx, r)
	if err != nil {
		return nil, err
	}

	idDataStruct, idDataSha256, err := parseIdData(r.IdData)
	if err != nil {
		return nil, MakeErrDevAuthBadRequest(err)
	}

	areq := &model.AuthSet{
		Id:           oid.NewUUIDv4().String(),
		IdData:       r.IdData,
		IdDataStruct: idDataStruct,
		IdDataSha256: idDataSha256,
		PubKey:       r.PubKey,
		DeviceId:     dev.Id,
		Status:       model.DevStatusPending,
		Timestamp:    uto.TimePtr(time.Now()),
	}

	// record authentication request
	err = d.db.AddAuthSet(ctx, *areq)
	if err != nil && err != store.ErrObjectExists {
		return nil, err
	}

	// update the device status
	if err := d.updateDeviceStatus(ctx, dev.Id, "", dev.Status); err != nil {
		return nil, err
	}

	// either the request was added or it was already present in the DB, get
	// it now
	areq, err = d.db.GetAuthSetByIdDataHashKey(ctx, idDataSha256, r.PubKey)
	if err != nil {
		l.Error("failed to find device auth set but could not add one either")
		return nil, errors.New("failed to locate device auth set")
	}

	return areq, nil
}

func (d *DevAuth) GetDevices(
	ctx context.Context,
	skip,
	limit uint,
	filter model.DeviceFilter,
) ([]model.Device, error) {
	devs, err := d.db.GetDevices(ctx, skip, limit, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list devices")
	}

	for i := range devs {
		devs[i].AuthSets, err = d.db.GetAuthSetsForDevice(ctx, devs[i].Id)
		if err != nil && err != store.ErrAuthSetNotFound {
			return nil, errors.Wrap(err, "db get auth sets error")
		}
	}

	// update check-in time
	if d.cache != nil {
		tenantID := ""
		idData := identity.FromContext(ctx)
		if idData != nil {
			tenantID = idData.Tenant
		}

		ids := make([]string, len(devs))
		for i := range devs {
			ids[i] = devs[i].Id
		}
		checkInTimes, err := d.cache.GetCheckInTimes(ctx, tenantID, ids)
		if err != nil {
			l := log.FromContext(ctx)
			l.Errorf("Failed to get check-in times for devices")
		} else {
			for i := range devs {
				if checkInTimes[i] != nil {
					devs[i].CheckInTime = checkInTimes[i]
				}
			}
		}
	}

	return devs, err
}

func (d *DevAuth) GetDevice(ctx context.Context, devId string) (*model.Device, error) {
	dev, err := d.db.GetDeviceById(ctx, devId)
	if err != nil {
		if err != store.ErrDevNotFound {
			return nil, errors.Wrap(err, "db get device by id error")
		}
		return nil, err
	}

	dev.AuthSets, err = d.db.GetAuthSetsForDevice(ctx, dev.Id)
	if err != nil {
		if err != store.ErrAuthSetNotFound {
			return nil, errors.Wrap(err, "db get auth sets error")
		}
		return dev, nil
	}

	if d.cache != nil {
		tenantID := ""
		idData := identity.FromContext(ctx)
		if idData != nil {
			tenantID = idData.Tenant
		}

		checkInTime, err := d.cache.GetCheckInTime(ctx, tenantID, devId)
		if err != nil {
			l := log.FromContext(ctx)
			l.Errorf("Failed to get check-in times for device")
		} else if checkInTime != nil {
			dev.CheckInTime = checkInTime
		}
	}

	return dev, err
}

// DecommissionDevice deletes device and all its tokens
func (d *DevAuth) DecommissionDevice(ctx context.Context, devID string) error {
	l := log.FromContext(ctx)

	l.Warnf("Decommission device with id: %s", devID)

	err := d.cacheDeleteToken(ctx, devID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete token for %s from cache", devID)
	}

	// set decommissioning flag on the device
	updev := model.DeviceUpdate{
		Decommissioning: uto.BoolPtr(true),
	}
	if err := d.db.UpdateDevice(
		ctx, devID, updev,
	); err != nil {
		return err
	}

	reqId := requestid.FromContext(ctx)

	tenantID := ""
	idData := identity.FromContext(ctx)
	if idData != nil {
		tenantID = idData.Tenant
	}

	// submit device decommissioning job
	if err := d.cOrch.SubmitDeviceDecommisioningJob(
		ctx,
		orchestrator.DecommissioningReq{
			DeviceId:  devID,
			RequestId: reqId,
			TenantID:  tenantID,
		}); err != nil {
		return errors.Wrap(err, "submit device decommissioning job error")
	}

	return err
}

// Delete a device and its tokens from deviceauth db
func (d *DevAuth) DeleteDevice(ctx context.Context, devID string) error {
	// delete device authorization sets
	if err := d.db.DeleteAuthSetsForDevice(ctx, devID); err != nil &&
		err != store.ErrAuthSetNotFound {
		return errors.Wrap(err, "db delete device authorization sets error")
	}

	devOID := oid.FromString(devID)
	// If the devID is not a valid string, there's no token.
	if devOID.String() == "" {
		return ErrInvalidDeviceID
	}
	// delete device tokens
	if err := d.db.DeleteTokenByDevId(
		ctx, devOID,
	); err != nil && err != store.ErrTokenNotFound {
		return errors.Wrap(err, "db delete device tokens error")
	}

	// delete device
	if err := d.db.DeleteDevice(ctx, devID); err != nil {
		return err
	}

	if d.config.EnableReporting {
		if err := d.cOrch.SubmitReindexReporting(ctx, devID); err != nil {
			return errors.Wrap(err, "reindex reporting job error")
		}
	}

	return nil
}

// Deletes device authentication set, and optionally the device.
func (d *DevAuth) DeleteAuthSet(ctx context.Context, devID string, authId string) error {
	l := log.FromContext(ctx)

	l.Warnf("Delete authentication set with id: "+
		"%s for the device with id: %s",
		authId, devID)

	err := d.cacheDeleteToken(ctx, devID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete token for %s from cache", devID)
	}

	// retrieve device authentication set to check its status
	authSet, err := d.db.GetAuthSetById(ctx, authId)
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			return err
		}
		return errors.Wrap(err, "db get auth set error")
	}

	// delete device authorization set
	if err := d.db.DeleteAuthSetForDevice(ctx, devID, authId); err != nil {
		return err
	}

	// if the device authentication set is accepted delete device tokens
	if authSet.Status == model.DevStatusAccepted {
		// If string is not a valid UUID there's no token.
		devOID := oid.FromString(devID)
		if err := d.db.DeleteTokenByDevId(
			ctx, devOID,
		); err != nil && err != store.ErrTokenNotFound {
			return errors.Wrap(err,
				"db delete device tokens error")
		}
	} else if authSet.Status == model.DevStatusPreauth {
		// if the auth set status is 'preauthorized', the device is deleted from
		// deviceauth. We cannot start the decommission_device workflow because
		// we don't provision devices until they are accepted. Still, we need to
		// remove the device from the inventory service because we index pre-authorized
		// devices for consumption via filtering APIs. To trigger the deletion
		// from the inventory service, we start the status update workflow with the
		// special value "decommissioned", which will cause the deletion of the
		// device from the inventory service's database
		tenantID := ""
		idData := identity.FromContext(ctx)
		if idData != nil {
			tenantID = idData.Tenant
		}
		req := orchestrator.UpdateDeviceStatusReq{
			RequestId: requestid.FromContext(ctx),
			Devices:   []model.DeviceInventoryUpdate{{Id: devID}},
			TenantId:  tenantID,
			Status:    "decommissioned",
		}
		if err = d.cOrch.SubmitUpdateDeviceStatusJob(ctx, req); err != nil {
			return errors.Wrap(err, "update device status job error")
		}

		// delete device
		if err := d.db.DeleteDevice(ctx, devID); err != nil {
			return err
		}

		if d.config.EnableReporting {
			if err := d.cOrch.SubmitReindexReporting(ctx, devID); err != nil {
				return errors.Wrap(err, "reindex reporting job error")
			}
		}

		return nil
	}

	return d.updateDeviceStatus(ctx, devID, "", authSet.Status)
}

func (d *DevAuth) AcceptDeviceAuth(ctx context.Context, device_id string, auth_id string) error {
	l := log.FromContext(ctx)

	aset, err := d.db.GetAuthSetById(ctx, auth_id)
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			return err
		}
		return errors.Wrap(err, "db get auth set error")
	}

	// device authentication set already accepted, nothing to do here
	if aset.Status == model.DevStatusAccepted {
		l.Debugf("Device %s already accepted", device_id)
		return nil
	} else if aset.Status != model.DevStatusRejected && aset.Status != model.DevStatusPending {
		// device authentication set can be accepted only from 'pending' or 'rejected' statuses
		return ErrDevAuthBadRequest
	}

	// check the device status
	// if the device status is accepted then do not trigger provisioning workflow
	// this needs to be checked before changing authentication set status
	dev, err := d.db.GetDeviceById(ctx, device_id)
	if err != nil {
		return err
	}

	// possible race, consider accept-count-unaccept pattern if that's problematic
	allow, err := d.canAcceptDevice(ctx)
	if err != nil {
		return err
	}

	if !allow {
		return ErrMaxDeviceCountReached
	}

	if err := d.setAuthSetStatus(ctx, device_id, auth_id, model.DevStatusAccepted); err != nil {
		return err
	}

	if dev.Status != model.DevStatusPending {
		// Device already exist in all services
		// We're done...
		return nil
	}

	dev.Status = model.DevStatusAccepted
	aset.Status = model.DevStatusAccepted
	dev.AuthSets = []model.AuthSet{*aset}

	reqId := requestid.FromContext(ctx)

	var tenantID string
	if idty := identity.FromContext(ctx); idty != nil {
		tenantID = idty.Tenant
	}

	// submit device accepted job
	if err := d.cOrch.SubmitProvisionDeviceJob(
		ctx,
		orchestrator.ProvisionDeviceReq{
			RequestId: reqId,
			DeviceID:  aset.DeviceId,
			TenantID:  tenantID,
			Device:    dev,
			Status:    dev.Status,
		}); err != nil {
		return errors.Wrap(err, "submit device provisioning job error")
	}

	return nil
}

func (d *DevAuth) setAuthSetStatus(
	ctx context.Context,
	deviceID string,
	authID string,
	status string,
) error {
	aset, err := d.db.GetAuthSetById(ctx, authID)
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			return err
		}
		return errors.Wrap(err, "db get auth set error")
	}

	if aset.DeviceId != deviceID {
		return ErrDevIdAuthIdMismatch
	}

	if aset.Status == status {
		return nil
	}

	currentStatus := aset.Status

	if aset.Status == model.DevStatusAccepted &&
		(status == model.DevStatusRejected || status == model.DevStatusPending) {
		deviceOID := oid.FromString(aset.DeviceId)
		// delete device token
		err := d.db.DeleteTokenByDevId(ctx, deviceOID)
		if err != nil && err != store.ErrTokenNotFound {
			return errors.Wrap(err, "db delete device token error")
		}
	}

	// if accepting an auth set
	if status == model.DevStatusAccepted {
		// reject all accepted auth sets for this device first
		err := d.db.RejectAuthSetsForDevice(ctx, deviceID)
		if err != nil && err != store.ErrAuthSetNotFound {
			return errors.Wrap(err, "failed to reject auth sets")
		}
	}

	if err := d.db.UpdateAuthSetById(ctx, aset.Id, model.AuthSetUpdate{
		Status: status,
	}); err != nil {
		return errors.Wrap(err, "db update device auth set error")
	}

	if status == model.DevStatusAccepted {
		return d.updateDeviceStatus(ctx, deviceID, status, currentStatus)
	}
	return d.updateDeviceStatus(ctx, deviceID, "", currentStatus)
}

func (d *DevAuth) RejectDeviceAuth(ctx context.Context, device_id string, auth_id string) error {
	aset, err := d.db.GetAuthSetById(ctx, auth_id)
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			return err
		}
		return errors.Wrap(err, "db get auth set error")
	} else if aset.Status != model.DevStatusPending && aset.Status != model.DevStatusAccepted {
		// device authentication set can be rejected only from 'accepted' or 'pending' statuses
		return ErrDevAuthBadRequest
	}

	err = d.cacheDeleteToken(ctx, device_id)
	if err != nil {
		return errors.Wrapf(err, "failed to delete token for %s from cache", device_id)
	}

	return d.setAuthSetStatus(ctx, device_id, auth_id, model.DevStatusRejected)
}

func (d *DevAuth) ResetDeviceAuth(ctx context.Context, device_id string, auth_id string) error {
	aset, err := d.db.GetAuthSetById(ctx, auth_id)
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			return err
		}
		return errors.Wrap(err, "db get auth set error")
	} else if aset.Status == model.DevStatusPreauth {
		// preauthorized auth set should not go into pending state
		return ErrDevAuthBadRequest
	}
	return d.setAuthSetStatus(ctx, device_id, auth_id, model.DevStatusPending)
}

func parseIdData(idData string) (map[string]interface{}, []byte, error) {
	var idDataStruct map[string]interface{}
	var idDataSha256 []byte

	err := json.Unmarshal([]byte(idData), &idDataStruct)
	if err != nil {
		return idDataStruct, idDataSha256, errors.Wrapf(
			err,
			"failed to parse identity data: %s",
			idData,
		)
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(idData))
	idDataSha256 = hash.Sum(nil)

	return idDataStruct, idDataSha256, nil
}

func (d *DevAuth) PreauthorizeDevice(
	ctx context.Context,
	req *model.PreAuthReq,
) (*model.Device, error) {
	// try add device, if a device with the given id_data exists -
	// the unique index on id_data will prevent it (conflict)
	// this is the only safeguard against id data conflict - we won't try to handle it
	// additionally on inserting the auth set (can't add an id data index on auth set - would
	// prevent key rotation)

	// FIXME: tenant_token is "" on purpose, will be removed

	l := log.FromContext(ctx)

	dev := model.NewDevice(req.DeviceId, req.IdData, req.PubKey)
	dev.Status = model.DevStatusPreauth

	idDataStruct, idDataSha256, err := parseIdData(req.IdData)
	if err != nil {
		return nil, MakeErrDevAuthBadRequest(err)
	}

	dev.IdDataStruct = idDataStruct
	dev.IdDataSha256 = idDataSha256

	err = d.db.AddDevice(ctx, *dev)
	switch err {
	case nil:
		break
	case store.ErrObjectExists:
		dev, err = d.db.GetDeviceByIdentityDataHash(ctx, idDataSha256)
		if err != nil {
			l.Error("failed to find device but could not preauthorize either")
			return nil, errors.New("failed to preauthorize device")
		}
		// if device exists we handle force: just add the authset, if force is in req
		if req.Force {
			// record authentication request
			authset := &model.AuthSet{
				Id:           req.AuthSetId,
				IdData:       req.IdData,
				IdDataStruct: idDataStruct,
				IdDataSha256: idDataSha256,
				PubKey:       req.PubKey,
				DeviceId:     dev.Id,
				Status:       model.DevStatusPreauth,
				Timestamp:    uto.TimePtr(time.Now()),
			}
			err = d.db.UpsertAuthSetStatus(ctx, authset)
			if err != nil {
				return nil, err
			}
			return dev, nil
		}
		return dev, ErrDeviceExists
	default:
		return nil, errors.Wrap(err, "failed to add device")
	}

	tenantId := ""
	idData := identity.FromContext(ctx)
	if idData != nil {
		tenantId = idData.Tenant
	}

	wfReq := orchestrator.UpdateDeviceStatusReq{
		RequestId: requestid.FromContext(ctx),
		Devices: []model.DeviceInventoryUpdate{{
			Id:       dev.Id,
			Revision: dev.Revision,
		}},
		TenantId: tenantId,
		Status:   dev.Status,
	}
	if err = d.cOrch.SubmitUpdateDeviceStatusJob(ctx, wfReq); err != nil {
		return nil, errors.Wrap(err, "update device status job error")
	}

	// record authentication request
	authset := model.AuthSet{
		Id:           req.AuthSetId,
		IdData:       req.IdData,
		IdDataStruct: idDataStruct,
		IdDataSha256: idDataSha256,
		PubKey:       req.PubKey,
		DeviceId:     req.DeviceId,
		Status:       model.DevStatusPreauth,
		Timestamp:    uto.TimePtr(time.Now()),
	}

	err = d.db.AddAuthSet(ctx, authset)
	switch err {
	case nil:
		if err := d.setDeviceIdentity(ctx, dev, tenantId); err != nil {
			return nil, err
		}
		return nil, nil
	case store.ErrObjectExists:
		dev, err = d.db.GetDeviceByIdentityDataHash(ctx, idDataSha256)
		if err != nil {
			l.Error("failed to find device but could not preauthorize either")
			return nil, errors.New("failed to preauthorize device")
		}
		return dev, ErrDeviceExists
	default:
		return nil, errors.Wrap(err, "failed to add auth set")
	}
}

func (d *DevAuth) RevokeToken(ctx context.Context, tokenID string) error {
	l := log.FromContext(ctx)
	tokenOID := oid.FromString(tokenID)

	var token *jwt.Token
	token, err := d.db.GetToken(ctx, tokenOID)
	if err != nil {
		return err
	}

	l.Warnf("Revoke token with jti: %s", tokenID)
	err = d.db.DeleteToken(ctx, tokenOID)

	if err == nil && d.cache != nil {
		err = d.cacheDeleteToken(ctx, token.Claims.Subject.String())
		err = errors.Wrapf(
			err,
			"failed to delete token for %s from cache",
			token.Claims.Subject.String(),
		)
	}
	return err
}

func verifyTenantClaim(ctx context.Context, verifyTenant bool, tenant string) error {
	l := log.FromContext(ctx)

	if verifyTenant {
		if tenant == "" {
			l.Errorf("No tenant claim in the token")
			return jwt.ErrTokenInvalid
		}
	} else if tenant != "" {
		l.Errorf("Unexpected tenant claim: %s in the token", tenant)
		return jwt.ErrTokenInvalid
	}

	return nil
}

func (d *DevAuth) validateJWTToken(ctx context.Context, jti oid.ObjectID, raw string) error {
	err := d.jwt.Validate(raw)
	if err != nil && d.jwtFallback != nil {
		err = d.jwtFallback.Validate(raw)
	}
	if err == jwt.ErrTokenExpired && jti.String() != "" {
		log.FromContext(ctx).Errorf("Token %s expired: %v", jti.String(), err)
		return d.handleExpiredToken(ctx, jti)
	} else if err != nil {
		log.FromContext(ctx).Errorf("Token %s invalid: %v", jti.String(), err)
		return jwt.ErrTokenInvalid
	}
	return nil
}

func (d *DevAuth) VerifyToken(ctx context.Context, raw string) error {
	l := log.FromContext(ctx)

	token := &jwt.Token{}
	err := token.UnmarshalJWT([]byte(raw), d.jwt.FromJWT)
	jti := token.Claims.ID
	if err != nil {
		l.Errorf("Token %s invalid: %v", jti.String(), err)
		return jwt.ErrTokenInvalid
	} else if !token.Claims.Device {
		l.Errorf("not a device token")
		return jwt.ErrTokenInvalid
	}

	err = verifyTenantClaim(ctx, d.verifyTenant, token.Claims.Tenant)
	if err == nil {
		err = d.checker.ValidateWithContext(ctx)
	}
	if err != nil {
		return err
	}

	origMethod := ctxhttpheader.FromContext(ctx, "X-Forwarded-Method")
	origUri := ctxhttpheader.FromContext(ctx, "X-Forwarded-Uri")
	origUri = purgeUriArgs(origUri)

	// throttle and try fetch token from cache - if cached, it was
	// already verified against the db checks below, we trust it
	cachedToken, err := d.cacheThrottleVerify(ctx, token, raw, origMethod, origUri)

	if err == cache.ErrTooManyRequests {
		return err
	} else if errRate := d.checkRateLimits(ctx); errRate != nil {
		return errRate
	}

	if cachedToken != "" && raw == cachedToken {
		// update device check-in time
		d.updateCheckInTime(
			ctx,
			token.Claims.Subject.String(),
			token.Claims.Tenant,
			nil,
		)
		return nil
	}

	// caching is best effort, don't fail
	if err != nil {
		l.Errorf("Failed to throttle for token %v: %s, continue.", token, err.Error())
	}

	// perform JWT signature and claims validation
	err = d.validateJWTToken(ctx, jti, raw)
	if err != nil {
		return err
	}

	// cache check was a MISS, hit the db for verification
	// check if token is in the system
	_, err = d.db.GetToken(ctx, jti)
	if err != nil {
		if err == store.ErrTokenNotFound {
			l.Errorf("Token %s not found", jti.String())
			return err
		}
		return errors.Wrapf(err, "Cannot get token with id: %s from database: %s", jti, err)
	}

	auth, err := d.db.GetAuthSetById(ctx, jti.String())
	if err != nil {
		if err == store.ErrAuthSetNotFound {
			l.Errorf("Auth set %s not found", jti.String())
			return err
		}
		return err
	}

	if auth.Status != model.DevStatusAccepted {
		return jwt.ErrTokenInvalid
	}

	// reject authentication for device that is in the process of
	// decommissioning
	dev, err := d.db.GetDeviceById(ctx, auth.DeviceId)
	if err != nil {
		return err
	}
	if dev.Decommissioning {
		l.Errorf(
			"Token %s rejected, device %s is being decommissioned",
			jti.String(),
			auth.DeviceId,
		)
		return jwt.ErrTokenInvalid
	}

	// update device check-in time
	d.updateCheckInTime(
		ctx,
		token.Claims.Subject.String(),
		token.Claims.Tenant,
		dev.CheckInTime,
	)

	// after successful token verification - cache it (best effort)
	_ = d.cacheSetToken(ctx, token, raw)

	return nil
}

func (d *DevAuth) handleExpiredToken(ctx context.Context, jti oid.ObjectID) error {
	err := d.db.DeleteToken(ctx, jti)
	if err == store.ErrTokenNotFound {
		l := log.FromContext(ctx)
		l.Errorf("Token %s not found", jti.String())
		return err
	}
	if err != nil {
		return errors.Wrapf(err, "Cannot delete token with jti: %s : %s", jti, err)
	}
	return jwt.ErrTokenExpired
}

// purgeUriArgs removes query string args from an uri string
// important for burst control (bursts are per uri without args)
func purgeUriArgs(uri string) string {
	return strings.Split(uri, "?")[0]
}

func (d *DevAuth) cacheThrottleVerify(
	ctx context.Context,
	token *jwt.Token,
	originalRaw,
	origMethod,
	origUri string,
) (string, error) {
	if d.cache == nil || d.cTenant == nil {
		return "", nil
	}

	// try get cached/precomputed limits
	limits, err := d.getApiLimits(ctx,
		token.Claims.Tenant,
		token.Claims.Subject.String())
	if err != nil {
		return "", err
	}

	// apply throttling and fetch cached token
	cached, err := d.cache.Throttle(ctx,
		originalRaw,
		*limits,
		token.Claims.Tenant,
		token.Claims.Subject.String(),
		cache.IdTypeDevice,
		origUri,
		origMethod)

	return cached, err
}

func (d *DevAuth) cacheSetToken(ctx context.Context, token *jwt.Token, raw string) error {
	if d.cache == nil {
		return nil
	}

	expireIn := time.Duration(token.Claims.ExpiresAt.Unix()-d.clock.Now().Unix()) * time.Second

	return d.cache.CacheToken(ctx,
		token.Claims.Tenant,
		token.Claims.Subject.String(),
		cache.IdTypeDevice,
		raw,
		expireIn)
}

func (d *DevAuth) getApiLimits(
	ctx context.Context,
	tid,
	did string,
) (*ratelimits.ApiLimits, error) {
	limits, err := d.cache.GetLimits(ctx, tid, did, cache.IdTypeDevice)
	if err != nil {
		return nil, err
	}

	if limits != nil {
		return limits, nil
	}

	dev, err := d.db.GetDeviceById(ctx, did)
	if err != nil {
		return nil, err
	}

	t, err := d.cTenant.GetTenant(ctx, tid)
	if err != nil {
		return nil, errors.Wrap(err, "request to get tenant failed")
	}
	if t == nil {
		return nil, errors.New("tenant not found")
	}

	finalLimits := apiLimitsOverride(t.ApiLimits.DeviceLimits, dev.ApiLimits)

	err = d.cache.CacheLimits(ctx, finalLimits, tid, did, cache.IdTypeDevice)

	return &finalLimits, err
}

func (d *DevAuth) cacheDeleteToken(ctx context.Context, did string) error {
	if d.cache == nil {
		return nil
	}

	idData := identity.FromContext(ctx)
	if idData == nil {
		return errors.New("can't unpack tenant identity data from context")
	}
	tid := idData.Tenant

	return d.cache.DeleteToken(ctx, tid, did, cache.IdTypeDevice)
}

// TODO move to 'ratelimits', as ApiLimits methods maybe?
func apiLimitsOverride(src, dest ratelimits.ApiLimits) ratelimits.ApiLimits {
	// override only if not default
	if dest.ApiQuota.MaxCalls != 0 && dest.ApiQuota.IntervalSec != 0 {
		src.ApiQuota.MaxCalls = dest.ApiQuota.MaxCalls
		src.ApiQuota.IntervalSec = dest.ApiQuota.IntervalSec
	}

	out := make([]ratelimits.ApiBurst, len(src.ApiBursts))
	copy(out, src.ApiBursts)

	for _, bdest := range dest.ApiBursts {
		found := false
		for i, bsrc := range src.ApiBursts {
			if bdest.Action == bsrc.Action &&
				bdest.Uri == bsrc.Uri {
				out[i].MinIntervalSec = bdest.MinIntervalSec
				found = true
			}
		}

		if !found {
			out = append(out,
				ratelimits.ApiBurst{
					Action:         bdest.Action,
					Uri:            bdest.Uri,
					MinIntervalSec: bdest.MinIntervalSec,
				},
			)
		}
	}

	src.ApiBursts = out
	return src
}

func (d *DevAuth) GetLimit(ctx context.Context, name string) (*model.Limit, error) {
	l := log.FromContext(ctx)
	var (
		limit *model.Limit
		err   error
	)
	if d.cache != nil {
		limit, err = d.cache.GetLimit(ctx, name)
		if err != nil {
			l.Warnf("error fetching limit from cache: %s", err.Error())
		}
	}
	if limit == nil {
		limit, err = d.db.GetLimit(ctx, name)
		if err != nil {
			if errors.Is(err, store.ErrLimitNotFound) {
				limit = &model.Limit{Name: name, Value: 0}
				err = nil
			} else {
				return nil, err
			}
		}
		if d.cache != nil {
			errCache := d.cache.SetLimit(ctx, limit)
			if errCache != nil {
				l.Warnf("failed to store limit %q in cache: %s", name, errCache.Error())
			}
		}
	}
	return limit, err
}

func (d *DevAuth) GetTenantLimit(
	ctx context.Context,
	name,
	tenant_id string,
) (*model.Limit, error) {
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant_id,
	})

	return d.GetLimit(ctx, name)
}

func (d *DevAuth) WithJWTFallbackHandler(handler jwt.Handler) *DevAuth {
	d.jwtFallback = handler
	return d
}

// WithTenantVerification will force verification of tenant token with tenant
// administrator when processing device authentication requests. Returns an
// updated devauth.
func (d *DevAuth) WithTenantVerification(c tenant.ClientRunner) *DevAuth {
	d.cTenant = c
	d.verifyTenant = true
	return d
}

func (d *DevAuth) WithCache(c cache.Cache) *DevAuth {
	d.cache = c
	return d
}

func (d *DevAuth) WithRatelimits(
	rl rate.Limiter,
	weights map[string]float64,
	defaultQuota float64,
) *DevAuth {
	if rl != nil {
		d.rateLimiter = rl
		d.rateLimiterWeights = weights
		d.rateLimiterWeightDefault = defaultQuota
	}
	return d
}

func (d *DevAuth) WithClock(c utils.Clock) *DevAuth {
	d.clock = c
	return d
}

func (d *DevAuth) SetTenantLimit(ctx context.Context, tenant_id string, limit model.Limit) error {
	l := log.FromContext(ctx)

	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant_id,
	})

	l.Infof("setting limit %v for tenant %v", limit, tenant_id)

	if err := d.db.PutLimit(ctx, limit); err != nil {
		l.Errorf("failed to save limit %v for tenant %v to database: %v",
			limit, tenant_id, err)
		return errors.Wrapf(err, "failed to save limit %v for tenant %v to database",
			limit, tenant_id)
	}
	if d.cache != nil {
		errCache := d.cache.SetLimit(ctx, &limit)
		if errCache != nil {
			l.Warnf("failed to store limit %q in cache: %s", limit.Name, errCache.Error())
		}
	}
	return nil
}

func (d *DevAuth) DeleteTenantLimit(ctx context.Context, tenant_id string, limit string) error {
	l := log.FromContext(ctx)

	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenant_id,
	})

	l.Infof("removing limit %v for tenant %v", limit, tenant_id)

	if err := d.db.DeleteLimit(ctx, limit); err != nil {
		l.Errorf("failed to delete limit %v for tenant %v to database: %v",
			limit, tenant_id, err)
		return errors.Wrapf(err, "failed to delete limit %v for tenant %v to database",
			limit, tenant_id)
	}
	if d.cache != nil {
		errCache := d.cache.DeleteLimit(ctx, limit)
		if errCache != nil {
			l.Warnf("error removing limit %q from cache: %s", limit, errCache.Error())
		}
	}
	return nil
}

func (d *DevAuth) GetDevCountByStatus(ctx context.Context, status string) (int, error) {
	return d.db.GetDevCountByStatus(ctx, status)
}

// canAcceptDevice checks if model.LimitMaxDeviceCount will be exceeded
func (d *DevAuth) canAcceptDevice(ctx context.Context) (bool, error) {
	limit, err := d.GetLimit(ctx, model.LimitMaxDeviceCount)
	if err != nil {
		return false, errors.Wrap(err, "can't get current device limit")
	}

	if limit.Value == 0 {
		return true, nil
	}

	accepted, err := d.db.GetDevCountByStatus(ctx, model.DevStatusAccepted)
	if err != nil {
		return false, errors.Wrap(err, "can't get current device count")
	}

	if uint64(accepted+1) <= limit.Value {
		return true, nil
	}

	return false, nil
}

func (d *DevAuth) DeleteTokens(
	ctx context.Context,
	tenantID string,
	deviceID string,
) error {
	var err error
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})

	if deviceID != "" {
		deviceOID := oid.FromString(deviceID)
		if deviceOID.String() == "" {
			return ErrInvalidAuthSetID
		}
		err = d.db.DeleteTokenByDevId(ctx, deviceOID)
	} else {
		if err := d.cacheFlush(ctx, tenantID); err != nil {
			return errors.Wrapf(
				err,
				"failed to flush cache when cleaning tokens for tenant %v",
				tenantID,
			)
		}

		err = d.db.DeleteTokens(ctx)
	}

	if err != nil && err != store.ErrTokenNotFound {
		return errors.Wrapf(
			err,
			"failed to delete tokens for tenant: %v, device id: %v",
			tenantID,
			deviceID,
		)
	}

	return nil
}

func (d *DevAuth) cacheFlush(ctx context.Context, tenantID string) error {
	if d.cache == nil {
		return nil
	}

	return d.cache.SuspendTenant(ctx, tenantID)
}

func (d *DevAuth) GetTenantDeviceStatus(
	ctx context.Context,
	tenantId,
	deviceId string,
) (*model.Status, error) {
	if tenantId != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantId,
		})
	}
	dev, err := d.db.GetDeviceById(ctx, deviceId)
	switch err {
	case nil:
		return &model.Status{Status: dev.Status}, nil
	case store.ErrDevNotFound:
		return nil, ErrDeviceNotFound
	default:
		return nil, errors.Wrapf(err, "get device %s failed", deviceId)

	}
}

func (d *DevAuth) updateCheckInTime(
	ctx context.Context,
	deviceId string,
	tenantId string,
	previous *time.Time,
) {
	var err error
	defer func() {
		if err != nil {
			log.FromContext(ctx).Errorf(
				"failed to update device check-in time for device %s: %s",
				deviceId, err.Error(),
			)
		}
	}()
	checkInTime := uto.TimePtr(time.Now().UTC())
	// in case cache is disabled, use mongo
	if d.cache == nil {
		if err = d.db.UpdateDevice(ctx,
			deviceId,
			model.DeviceUpdate{
				CheckInTime: checkInTime,
			}); err != nil {
			return
		}
	} else {
		// get check-in time from cache
		previous, err = d.cache.GetCheckInTime(ctx, tenantId, deviceId)
		if err != nil {
			return
		}
		// update check-in time in cache
		err = d.cache.CacheCheckInTime(ctx, checkInTime, tenantId, deviceId)
		if err != nil {
			return
		}
	}
	// compare data without a time of current and previous check-in time
	// and if it's different trigger reindexing (if enabled)
	// and save check-in time in the database
	if previous == nil ||
		(previous != nil &&
			!previous.Truncate(24*time.Hour).Equal(checkInTime.Truncate(24*time.Hour))) {
		// trigger reindexing
		if d.config.EnableReporting {
			if err = d.cOrch.SubmitReindexReporting(ctx, deviceId); err != nil {
				err = errors.Wrap(err, "reindex reporting job error")
				return
			}
		} else {
			// update check-in time in inventory
			if err := d.syncCheckInTime(ctx, checkInTime, deviceId, tenantId); err != nil {
				log.FromContext(ctx).Errorf(
					"failed to synchronize device check-in time with inventory: device %s: %s",
					deviceId, err.Error(),
				)
			}
		}
		// dump cached value to database
		if d.cache != nil {
			if err = d.db.UpdateDevice(ctx,
				deviceId,
				model.DeviceUpdate{
					CheckInTime: checkInTime,
				}); err != nil {
				return
			}
		}
	}
}

func (d *DevAuth) syncCheckInTime(
	ctx context.Context,
	checkInTime *time.Time,
	deviceId string,
	tenantId string,
) error {
	attributes := []model.DeviceAttribute{
		{
			Name:        "check_in_time",
			Description: nil,
			Value:       checkInTime,
			Scope:       InventoryScopeSystem,
		},
	}
	attrJson, err := json.Marshal(attributes)
	if err != nil {
		return errors.New("internal error: cannot marshal attributes into json")
	}
	if err := d.cOrch.SubmitUpdateDeviceInventoryJob(
		ctx,
		orchestrator.UpdateDeviceInventoryReq{
			RequestId:  requestid.FromContext(ctx),
			TenantId:   tenantId,
			DeviceId:   deviceId,
			Scope:      InventoryScopeSystem,
			Attributes: string(attrJson),
		}); err != nil {
		return errors.Wrap(err, "failed to start device inventory update job")
	}
	return nil
}
