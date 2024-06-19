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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/identity"
	midentity "github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"
	u "github.com/mendersoftware/mender-server/pkg/rest_utils"

	inventory "github.com/mendersoftware/mender-server/services/inventory/inv"
	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const (
	uriDevices       = "/api/0.1.0/devices"
	uriDevice        = "/api/0.1.0/devices/#id"
	uriDeviceTags    = "/api/0.1.0/devices/#id/tags"
	uriDeviceGroups  = "/api/0.1.0/devices/#id/group"
	uriDeviceGroup   = "/api/0.1.0/devices/#id/group/#name"
	uriAttributes    = "/api/0.1.0/attributes"
	uriGroups        = "/api/0.1.0/groups"
	uriGroupsName    = "/api/0.1.0/groups/#name"
	uriGroupsDevices = "/api/0.1.0/groups/#name/devices"

	apiUrlInternalV1         = "/api/internal/v1/inventory"
	uriInternalAlive         = apiUrlInternalV1 + "/alive"
	uriInternalHealth        = apiUrlInternalV1 + "/health"
	uriInternalTenants       = apiUrlInternalV1 + "/tenants"
	uriInternalDevices       = apiUrlInternalV1 + "/tenants/#tenant_id/devices"
	urlInternalDevicesStatus = apiUrlInternalV1 + "/tenants/#tenant_id/devices/status/#status"
	uriInternalDeviceDetails = apiUrlInternalV1 + "/tenants/#tenant_id/devices/#device_id"
	uriInternalDeviceGroups  = apiUrlInternalV1 + "/tenants/#tenant_id/devices/#device_id/groups"
	urlInternalAttributes    = apiUrlInternalV1 +
		"/tenants/#tenant_id/device/#device_id/attribute/scope/#scope"
	urlInternalReindex   = apiUrlInternalV1 + "/tenants/#tenant_id/devices/#device_id/reindex"
	apiUrlManagementV2   = "/api/management/v2/inventory"
	urlFiltersAttributes = apiUrlManagementV2 + "/filters/attributes"
	urlFiltersSearch     = apiUrlManagementV2 + "/filters/search"

	apiUrlInternalV2         = "/api/internal/v2/inventory"
	urlInternalFiltersSearch = apiUrlInternalV2 + "/tenants/#tenant_id/filters/search"

	hdrTotalCount = "X-Total-Count"
)

const (
	queryParamGroup          = "group"
	queryParamSort           = "sort"
	queryParamHasGroup       = "has_group"
	queryParamValueSeparator = ":"
	queryParamScopeSeparator = "/"
	sortOrderAsc             = "asc"
	sortOrderDesc            = "desc"
	sortAttributeNameIdx     = 0
	sortOrderIdx             = 1
)

const (
	DefaultTimeout = time.Second * 10
)

const (
	checkInTimeParamName  = "check_in_time"
	checkInTimeParamScope = "system"
)

// model of device's group name response at /devices/:id/group endpoint
type InventoryApiGroup struct {
	Group model.GroupName `json:"group"`
}

func (g InventoryApiGroup) Validate() error {
	return g.Group.Validate()
}

type inventoryHandlers struct {
	inventory inventory.InventoryApp
}

// return an ApiHandler for device admission app
func NewInventoryApiHandlers(i inventory.InventoryApp) ApiHandler {
	return &inventoryHandlers{
		inventory: i,
	}
}

func wrapRoutes(middleware rest.Middleware, routes ...*rest.Route) []*rest.Route {
	for _, route := range routes {
		route.Func = middleware.MiddlewareFunc(route.Func)
	}
	return routes
}

func (i *inventoryHandlers) Build() (http.Handler, error) {
	//this will override the framework's error resp to the desired one:
	// {"error": "msg"}
	// instead of:
	// {"Error": "msg"}
	rest.ErrorFieldName = "error"

	api := rest.NewApi()

	api.Use(
		&requestlog.RequestLogMiddleware{},
		&requestid.RequestIdMiddleware{},
		&accesslog.AccessLogMiddleware{
			Format: accesslog.SimpleLogFormat,
			DisableLog: func(statusCode int, r *rest.Request) bool {
				if r.URL.Path == uriInternalAlive ||
					r.URL.Path == uriInternalHealth &&
						statusCode < 300 {
					// Skips the health/liveliness probes
					return true
				}
				return false
			},
		},
		&rest.ContentTypeCheckerMiddleware{},
	)
	internalRoutes := []*rest.Route{
		rest.Get(uriInternalAlive, i.LivelinessHandler),
		rest.Get(uriInternalHealth, i.HealthCheckHandler),

		rest.Patch(urlInternalAttributes, i.PatchDeviceAttributesInternalHandler),
		rest.Post(urlInternalReindex, i.ReindexDeviceDataHandler),

		rest.Post(uriInternalTenants, i.CreateTenantHandler),
		rest.Post(uriInternalDevices, i.AddDeviceHandler),
		rest.Delete(uriInternalDeviceDetails, i.DeleteDeviceHandler),
		rest.Post(urlInternalDevicesStatus, i.InternalDevicesStatusHandler),
		rest.Get(uriInternalDeviceGroups, i.GetDeviceGroupsInternalHandler),
		rest.Post(urlInternalFiltersSearch, i.InternalFiltersSearchHandler),
	}

	publicRoutes := AutogenOptionsRoutes([]*rest.Route{
		rest.Get(uriDevices, i.GetDevicesHandler),
		rest.Get(uriDevice, i.GetDeviceHandler),
		rest.Delete(uriDevice, i.DeleteDeviceInventoryHandler),
		rest.Delete(uriDeviceGroup, i.DeleteDeviceGroupHandler),
		rest.Delete(uriGroupsName, i.DeleteGroupHandler),
		rest.Delete(uriGroupsDevices, i.ClearDevicesGroupHandler),
		rest.Patch(uriAttributes, i.UpdateDeviceAttributesHandler),
		rest.Put(uriAttributes, i.UpdateDeviceAttributesHandler),
		rest.Put(uriDeviceGroups, i.AddDeviceToGroupHandler),
		rest.Patch(uriGroupsDevices, i.AppendDevicesToGroup),
		rest.Put(uriDeviceTags, i.UpdateDeviceTagsHandler),
		rest.Patch(uriDeviceTags, i.UpdateDeviceTagsHandler),

		rest.Get(uriDeviceGroups, i.GetDeviceGroupHandler),
		rest.Get(uriGroups, i.GetGroupsHandler),
		rest.Get(uriGroupsDevices, i.GetDevicesByGroupHandler),

		rest.Get(urlFiltersAttributes, i.FiltersAttributesHandler),
		rest.Post(urlFiltersSearch, i.FiltersSearchHandler),
	}, AllowHeaderOptionsGenerator)
	publicRoutes = wrapRoutes(&identity.IdentityMiddleware{
		UpdateLogger: true,
	}, publicRoutes...)
	routes := append(internalRoutes, publicRoutes...)

	app, err := rest.MakeRouter(routes...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}
	api.SetApp(app)

	return api.MakeHandler(), nil

}

func (i *inventoryHandlers) LivelinessHandler(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (i *inventoryHandlers) HealthCheckHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	err := i.inventory.HealthCheck(ctx)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// `sort` paramater value is an attribute name with optional direction (desc or asc)
// separated by colon (:)
//
// eg. `sort=attr_name1` or `sort=attr_name1:asc`
func parseSortParam(r *rest.Request) (*store.Sort, error) {
	sortStr, err := utils.ParseQueryParmStr(r, queryParamSort, false, nil)
	if err != nil {
		return nil, err
	}
	if sortStr == "" {
		return nil, nil
	}
	sortValArray := strings.Split(sortStr, queryParamValueSeparator)
	attrNameWithScope := strings.SplitN(
		sortValArray[sortAttributeNameIdx],
		queryParamScopeSeparator,
		2,
	)
	var scope, attrName string
	if len(attrNameWithScope) == 1 {
		scope = model.AttrScopeInventory
		attrName = attrNameWithScope[0]
	} else {
		scope = attrNameWithScope[0]
		attrName = attrNameWithScope[1]
	}
	sort := store.Sort{AttrName: attrName, AttrScope: scope}
	if len(sortValArray) == 2 {
		sortOrder := sortValArray[sortOrderIdx]
		if sortOrder != sortOrderAsc && sortOrder != sortOrderDesc {
			return nil, errors.New("invalid sort order")
		}
		sort.Ascending = sortOrder == sortOrderAsc
	}
	return &sort, nil
}

// Filter paramaters name are attributes name. Value can be prefixed
// with equality operator code (`eq` for =), separated from value by colon (:).
// Equality operator default value is `eq`
//
// eg. `attr_name1=value1` or `attr_name1=eq:value1`
func parseFilterParams(r *rest.Request) ([]store.Filter, error) {
	knownParams := []string{
		utils.PageName,
		utils.PerPageName,
		queryParamSort,
		queryParamHasGroup,
		queryParamGroup,
	}
	filters := make([]store.Filter, 0)
	var filter store.Filter
	for name := range r.URL.Query() {
		if utils.ContainsString(name, knownParams) {
			continue
		}
		valueStr, err := utils.ParseQueryParmStr(r, name, false, nil)
		if err != nil {
			return nil, err
		}

		attrNameWithScope := strings.SplitN(name, queryParamScopeSeparator, 2)
		var scope, attrName string
		if len(attrNameWithScope) == 1 {
			scope = model.AttrScopeInventory
			attrName = attrNameWithScope[0]
		} else {
			scope = attrNameWithScope[0]
			attrName = attrNameWithScope[1]
		}
		filter = store.Filter{AttrName: attrName, AttrScope: scope}

		// make sure we parse ':'s in value, it's either:
		// not there
		// after a valid operator specifier
		// or/and inside the value itself(mac, etc), in which case leave it alone
		sepIdx := strings.Index(valueStr, ":")
		if sepIdx == -1 {
			filter.Value = valueStr
			filter.Operator = store.Eq
		} else {
			validOps := []string{"eq"}
			for _, o := range validOps {
				if valueStr[:sepIdx] == o {
					switch o {
					case "eq":
						filter.Operator = store.Eq
						filter.Value = valueStr[sepIdx+1:]
					}
					break
				}
			}

			if filter.Value == "" {
				filter.Value = valueStr
				filter.Operator = store.Eq
			}
		}

		floatValue, err := strconv.ParseFloat(filter.Value, 64)
		if err == nil {
			filter.ValueFloat = &floatValue
		}

		timeValue, err := time.Parse("2006-01-02T15:04:05Z", filter.Value)
		if err == nil {
			filter.ValueTime = &timeValue
		}

		filters = append(filters, filter)
	}
	return filters, nil
}

func (i *inventoryHandlers) GetDevicesHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	page, perPage, err := utils.ParsePagination(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	hasGroup, err := utils.ParseQueryParmBool(r, queryParamHasGroup, false, nil)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	groupName, err := utils.ParseQueryParmStr(r, "group", false, nil)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	sort, err := parseSortParam(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	filters, err := parseFilterParams(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	ld := store.ListQuery{Skip: int((page - 1) * perPage),
		Limit:     int(perPage),
		Filters:   filters,
		Sort:      sort,
		HasGroup:  hasGroup,
		GroupName: groupName}

	devs, totalCount, err := i.inventory.ListDevices(ctx, ld)

	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	hasNext := totalCount > int(page*perPage)
	links := utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add("Link", l)
	}
	// the response writer will ensure the header name is in Kebab-Pascal-Case
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	_ = w.WriteJson(devs)
}

func (i *inventoryHandlers) GetDeviceHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	deviceID := r.PathParam("id")

	dev, err := i.inventory.GetDevice(ctx, model.DeviceID(deviceID))
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}
	if dev == nil {
		u.RestErrWithLog(w, r, l, store.ErrDevNotFound, http.StatusNotFound)
		return
	}
	if dev.TagsEtag != "" {
		w.Header().Set("ETag", dev.TagsEtag)
	}

	_ = w.WriteJson(dev)
}

func (i *inventoryHandlers) DeleteDeviceInventoryHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	deviceID := r.PathParam("id")

	err := i.inventory.ReplaceAttributes(ctx, model.DeviceID(deviceID),
		model.DeviceAttributes{}, model.AttrScopeInventory, "")
	if err != nil && err != store.ErrDevNotFound {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (i *inventoryHandlers) DeleteDeviceHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	tenantId := r.PathParam("tenant_id")
	if tenantId != "" {
		id := &midentity.Identity{
			Tenant: tenantId,
		}
		ctx = midentity.WithContext(ctx, id)
	}

	l := log.FromContext(ctx)

	deviceID := r.PathParam("device_id")

	err := i.inventory.DeleteDevice(ctx, model.DeviceID(deviceID))
	if err != nil && err != store.ErrDevNotFound {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (i *inventoryHandlers) AddDeviceHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	tenantId := r.PathParam("tenant_id")
	if tenantId != "" {
		id := &midentity.Identity{
			Tenant: tenantId,
		}
		ctx = midentity.WithContext(ctx, id)
	}

	l := log.FromContext(ctx)

	dev, err := parseDevice(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = dev.Attributes.Validate()
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = i.inventory.AddDevice(ctx, dev)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.Header().Add("Location", "devices/"+dev.ID.String())
	w.WriteHeader(http.StatusCreated)
}

func (i *inventoryHandlers) UpdateDeviceAttributesHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)
	var idata *identity.Identity
	if idata = identity.FromContext(ctx); idata == nil || !idata.IsDevice {
		u.RestErrWithLog(w, r, l, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}
	deviceID := model.DeviceID(idata.Subject)
	//extract attributes from body
	attrs, err := parseAttributes(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}
	i.updateDeviceAttributes(w, r, ctx, attrs, deviceID, model.AttrScopeInventory, "")
}

func (i *inventoryHandlers) UpdateDeviceTagsHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	// get device ID from uri
	deviceID := model.DeviceID(r.PathParam("id"))
	if len(deviceID) < 1 {
		u.RestErrWithLog(w, r, l, errors.New("device id cannot be empty"), http.StatusBadRequest)
		return
	}

	ifMatchHeader := r.Header.Get("If-Match")

	// extract attributes from body
	attrs, err := parseAttributes(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	// set scope and timestamp for tags attributes
	now := time.Now()
	for i := range attrs {
		attrs[i].Scope = model.AttrScopeTags
		if attrs[i].Timestamp == nil {
			attrs[i].Timestamp = &now
		}
	}

	i.updateDeviceAttributes(w, r, ctx, attrs, deviceID, model.AttrScopeTags, ifMatchHeader)
}

func (i *inventoryHandlers) updateDeviceAttributes(
	w rest.ResponseWriter,
	r *rest.Request,
	ctx context.Context,
	attrs model.DeviceAttributes,
	deviceID model.DeviceID,
	scope string,
	etag string,
) {
	l := log.FromContext(ctx)
	var err error

	// upsert or replace the attributes
	if r.Method == http.MethodPatch {
		err = i.inventory.UpsertAttributesWithUpdated(ctx, deviceID, attrs, scope, etag)
	} else if r.Method == http.MethodPut {
		err = i.inventory.ReplaceAttributes(ctx, deviceID, attrs, scope, etag)
	} else {
		u.RestErrWithLog(w, r, l, errors.New("method not alllowed"), http.StatusMethodNotAllowed)
		return
	}

	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
	case inventory.ErrTooManyAttributes:
		u.RestErrWithLog(w, r, l, cause, http.StatusBadRequest)
		return
	case inventory.ErrETagDoesntMatch:
		u.RestErrWithInfoMsg(w, r, l, cause, http.StatusPreconditionFailed, cause.Error())
		return
	}
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (i *inventoryHandlers) PatchDeviceAttributesInternalHandler(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	tenantId := r.PathParam("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	l := log.FromContext(ctx)

	deviceId := r.PathParam("device_id")
	if len(deviceId) < 1 {
		u.RestErrWithLog(w, r, l, errors.New("device id cannot be empty"), http.StatusBadRequest)
		return
	}
	//extract attributes from body
	attrs, err := parseAttributes(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}
	for i := range attrs {
		attrs[i].Scope = r.PathParam("scope")
		if attrs[i].Name == checkInTimeParamName && attrs[i].Scope == checkInTimeParamScope {
			t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", attrs[i].Value))
			if err != nil {
				u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
				return
			}
			attrs[i].Value = t
		}
	}

	//upsert the attributes
	err = i.inventory.UpsertAttributes(ctx, model.DeviceID(deviceId), attrs)
	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
		u.RestErrWithLog(w, r, l, cause, http.StatusBadRequest)
		return
	}
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (i *inventoryHandlers) DeleteDeviceGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	deviceID := r.PathParam("id")
	groupName := r.PathParam("name")

	err := i.inventory.UnsetDeviceGroup(ctx, model.DeviceID(deviceID), model.GroupName(groupName))
	if err != nil {
		cause := errors.Cause(err)
		if cause != nil {
			if cause.Error() == store.ErrDevNotFound.Error() {
				u.RestErrWithLog(w, r, l, err, http.StatusNotFound)
				return
			}
		}
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (i *inventoryHandlers) AddDeviceToGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	devId := r.PathParam("id")

	var group InventoryApiGroup
	err := r.DecodeJsonPayload(&group)
	if err != nil {
		u.RestErrWithLog(
			w, r, l, errors.Wrap(err, "failed to decode device group data"),
			http.StatusBadRequest)
		return
	}

	if err = group.Validate(); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = i.inventory.UpdateDeviceGroup(ctx, model.DeviceID(devId), model.GroupName(group.Group))
	if err != nil {
		if cause := errors.Cause(err); cause != nil && cause == store.ErrDevNotFound {
			u.RestErrWithLog(w, r, l, err, http.StatusNotFound)
			return
		}
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (i *inventoryHandlers) GetDevicesByGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	group := r.PathParam("name")

	page, perPage, err := utils.ParsePagination(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	//get one extra device to see if there's a 'next' page
	ids, totalCount, err := i.inventory.ListDevicesByGroup(
		ctx,
		model.GroupName(group),
		int((page-1)*perPage),
		int(perPage),
	)
	if err != nil {
		if err == store.ErrGroupNotFound {
			u.RestErrWithLog(w, r, l, err, http.StatusNotFound)
		} else {
			u.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	hasNext := totalCount > int(page*perPage)

	links := utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add("Link", l)
	}
	// the response writer will ensure the header name is in Kebab-Pascal-Case
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	_ = w.WriteJson(ids)
}

func (i *inventoryHandlers) AppendDevicesToGroup(w rest.ResponseWriter, r *rest.Request) {
	var deviceIDs []model.DeviceID
	ctx := r.Context()
	l := log.FromContext(ctx)
	groupName := model.GroupName(r.PathParam("name"))
	if err := groupName.Validate(); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	if err := r.DecodeJsonPayload(&deviceIDs); err != nil {
		u.RestErrWithLog(w, r, l,
			errors.Wrap(err, "invalid payload schema"),
			http.StatusBadRequest,
		)
		return
	} else if len(deviceIDs) == 0 {
		u.RestErrWithLog(w, r, l,
			errors.New("no device IDs present in payload"),
			http.StatusBadRequest,
		)
		return
	}
	updated, err := i.inventory.UpdateDevicesGroup(
		ctx, deviceIDs, groupName,
	)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}
	_ = w.WriteJson(updated)
}

func (i *inventoryHandlers) DeleteGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	groupName := model.GroupName(r.PathParam("name"))
	if err := groupName.Validate(); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	updated, err := i.inventory.DeleteGroup(ctx, groupName)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(updated)
}

func (i *inventoryHandlers) ClearDevicesGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	var deviceIDs []model.DeviceID
	ctx := r.Context()
	l := log.FromContext(ctx)

	groupName := model.GroupName(r.PathParam("name"))
	if err := groupName.Validate(); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	if err := r.DecodeJsonPayload(&deviceIDs); err != nil {
		u.RestErrWithLog(w, r, l,
			errors.Wrap(err, "invalid payload schema"),
			http.StatusBadRequest,
		)
		return
	} else if len(deviceIDs) == 0 {
		u.RestErrWithLog(w, r, l,
			errors.New("no device IDs present in payload"),
			http.StatusBadRequest,
		)
		return
	}

	updated, err := i.inventory.UnsetDevicesGroup(ctx, deviceIDs, groupName)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(updated)
}

func parseDevice(r *rest.Request) (*model.Device, error) {
	dev := model.Device{}

	//decode body
	err := r.DecodeJsonPayload(&dev)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if err := dev.Validate(); err != nil {
		return nil, err
	}

	return &dev, nil
}

func parseAttributes(r *rest.Request) (model.DeviceAttributes, error) {
	var attrs model.DeviceAttributes

	err := r.DecodeJsonPayload(&attrs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	err = attrs.Validate()
	if err != nil {
		return nil, err
	}

	return attrs, nil
}

func (i *inventoryHandlers) GetGroupsHandler(w rest.ResponseWriter, r *rest.Request) {
	var fltr []model.FilterPredicate
	ctx := r.Context()

	l := log.FromContext(ctx)

	query := r.URL.Query()
	status := query.Get("status")
	if status != "" {
		fltr = []model.FilterPredicate{{
			Attribute: "status",
			Scope:     "identity",
			Type:      "$eq",
			Value:     status,
		}}
	}

	groups, err := i.inventory.ListGroups(ctx, fltr)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	if groups == nil {
		groups = []model.GroupName{}
	}

	_ = w.WriteJson(groups)
}

func (i *inventoryHandlers) GetDeviceGroupHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	deviceID := r.PathParam("id")

	group, err := i.inventory.GetDeviceGroup(ctx, model.DeviceID(deviceID))
	if err != nil {
		if err == store.ErrDevNotFound {
			u.RestErrWithLog(w, r, l, store.ErrDevNotFound, http.StatusNotFound)
		} else {
			u.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	ret := map[string]*model.GroupName{"group": nil}

	if group != "" {
		ret["group"] = &group
	}

	_ = w.WriteJson(ret)
}

type newTenantRequest struct {
	TenantID string `json:"tenant_id" valid:"required"`
}

func (t newTenantRequest) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.TenantID, validation.Required),
	)
}

func (i *inventoryHandlers) CreateTenantHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	var newTenant newTenantRequest

	if err := r.DecodeJsonPayload(&newTenant); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	if err := newTenant.Validate(); err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err := i.inventory.CreateTenant(ctx, model.NewTenant{
		ID: newTenant.TenantID,
	})
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (i *inventoryHandlers) FiltersAttributesHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	// query the database
	attributes, err := i.inventory.GetFiltersAttributes(ctx)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	// in case of nil make sure we return empty list
	if attributes == nil {
		attributes = []model.FilterAttribute{}
	}

	_ = w.WriteJson(attributes)
}

func (i *inventoryHandlers) FiltersSearchHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	//extract attributes from body
	searchParams, err := parseSearchParams(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	// query the database
	devs, totalCount, err := i.inventory.SearchDevices(ctx, *searchParams)
	if err != nil {
		if strings.Contains(err.Error(), "BadValue") {
			u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		} else {
			u.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	// the response writer will ensure the header name is in Kebab-Pascal-Case
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	_ = w.WriteJson(devs)
}

func (i *inventoryHandlers) InternalFiltersSearchHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.PathParam("tenant_id")
	if tenantId != "" {
		ctx = getTenantContext(ctx, tenantId)
	}

	//extract attributes from body
	searchParams, err := parseSearchParams(r)
	if err != nil {
		u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	// query the database
	devs, totalCount, err := i.inventory.SearchDevices(ctx, *searchParams)
	if err != nil {
		if strings.Contains(err.Error(), "BadValue") {
			u.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		} else {
			u.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	// the response writer will ensure the header name is in Kebab-Pascal-Case
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	_ = w.WriteJson(devs)
}

func getTenantContext(ctx context.Context, tenantId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if tenantId == "" {
		return ctx
	}
	id := &midentity.Identity{
		Tenant: tenantId,
	}

	ctx = midentity.WithContext(ctx, id)

	return ctx
}

func (i *inventoryHandlers) InternalDevicesStatusHandler(w rest.ResponseWriter, r *rest.Request) {
	const (
		StatusDecommissioned = "decommissioned"
		StatusAccepted       = "accepted"
		StatusRejected       = "rejected"
		StatusPreauthorized  = "preauthorized"
		StatusPending        = "pending"
		StatusNoAuth         = "noauth"
	)
	var (
		devices []model.DeviceUpdate
		result  *model.UpdateResult
	)

	ctx := r.Context()
	l := log.FromContext(ctx)

	tenantID := r.PathParam("tenant_id")
	ctx = getTenantContext(ctx, tenantID)

	status := r.PathParam("status")

	err := r.DecodeJsonPayload(&devices)
	if err != nil {
		u.RestErrWithLog(w, r, l, errors.Wrap(err, "cant parse devices"), http.StatusBadRequest)
		return
	}

	switch status {
	case StatusAccepted, StatusPreauthorized,
		StatusPending, StatusRejected,
		StatusNoAuth:
		// Update statuses
		attrs := model.DeviceAttributes{{
			Name:  "status",
			Scope: model.AttrScopeIdentity,
			Value: status,
		}}
		result, err = i.inventory.UpsertDevicesStatuses(ctx, devices, attrs)
	case StatusDecommissioned:
		// Delete Inventory
		result, err = i.inventory.DeleteDevices(ctx, getIdsFromDevices(devices))
	default:
		// Unrecognized status
		u.RestErrWithLog(w, r, l,
			errors.Errorf("unrecognized status: %s", status),
			http.StatusNotFound,
		)
		return
	}
	if err == store.ErrWriteConflict {
		u.RestErrWithLog(w, r, l, err, http.StatusConflict)
		return
	} else if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(result)
}

func (i *inventoryHandlers) GetDeviceGroupsInternalHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()

	l := log.FromContext(ctx)

	tenantId := r.PathParam("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	deviceID := r.PathParam("device_id")
	group, err := i.inventory.GetDeviceGroup(ctx, model.DeviceID(deviceID))
	if err != nil {
		if err == store.ErrDevNotFound {
			u.RestErrWithLog(w, r, l, store.ErrDevNotFound, http.StatusNotFound)
		} else {
			u.RestErrWithLogInternal(w, r, l, err)
		}
		return
	}

	res := model.DeviceGroups{}
	if group != "" {
		res.Groups = append(res.Groups, string(group))
	}

	_ = w.WriteJson(res)
}

func (i *inventoryHandlers) ReindexDeviceDataHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	tenantId := r.PathParam("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	l := log.FromContext(ctx)

	deviceId := r.PathParam("device_id")
	if len(deviceId) < 1 {
		u.RestErrWithLog(w, r, l, errors.New("device id cannot be empty"), http.StatusBadRequest)
		return
	}

	serviceName, err := utils.ParseQueryParmStr(r, "service", false, nil)
	// inventory service accepts only reindex requests from devicemonitor
	if err != nil || serviceName != "devicemonitor" {
		u.RestErrWithLog(w, r, l, errors.New("unsupported service"), http.StatusBadRequest)
		return
	}

	// check devicemonitor alerts
	alertsCount, err := i.inventory.CheckAlerts(ctx, deviceId)
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	alertsPresent := false
	if alertsCount > 0 {
		alertsPresent = true
	}
	attrs := model.DeviceAttributes{
		model.DeviceAttribute{
			Name:  model.AttrNameNumberOfAlerts,
			Scope: model.AttrScopeMonitor,
			Value: alertsCount,
		},
		model.DeviceAttribute{
			Name:  model.AttrNameAlerts,
			Scope: model.AttrScopeMonitor,
			Value: alertsPresent,
		},
	}

	// upsert monitor attributes
	err = i.inventory.UpsertAttributes(ctx, model.DeviceID(deviceId), attrs)
	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
		u.RestErrWithLog(w, r, l, cause, http.StatusBadRequest)
		return
	}
	if err != nil {
		u.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getIdsFromDevices(devices []model.DeviceUpdate) []model.DeviceID {
	ids := make([]model.DeviceID, len(devices))
	for i, dev := range devices {
		ids[i] = dev.Id
	}
	return ids
}

func parseSearchParams(r *rest.Request) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	if err := r.DecodeJsonPayload(&searchParams); err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if searchParams.Page < 1 {
		searchParams.Page = utils.PageDefault
	}
	if searchParams.PerPage < 1 {
		searchParams.PerPage = utils.PerPageDefault
	}

	if err := searchParams.Validate(); err != nil {
		return nil, err
	}

	return &searchParams, nil
}
