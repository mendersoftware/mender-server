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

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	inventory "github.com/mendersoftware/mender-server/services/inventory/inv"
	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const (
	apiUrlLegacy       = "/api/0.1.0"
	apiUrlManagementV1 = "/api/management/v1/inventory"
	apiUrlDevicesV1    = "/api/devices/v1/inventory"

	uriDevices          = "/devices"
	uriDevice           = "/devices/:id"
	uriDeviceTags       = "/devices/:id/tags"
	uriDeviceGroups     = "/devices/:id/group"
	uriDeviceGroup      = "/devices/:id/group/:name"
	uriGroups           = "/groups"
	uriGroupsName       = "/groups/:name"
	uriGroupsDevices    = "/groups/:name/devices"
	uriAttributes       = "/attributes"
	uriDeviceAttributes = "/device/attributes"

	apiUrlInternalV1         = "/api/internal/v1/inventory"
	uriInternalAlive         = "/alive"
	uriInternalHealth        = "/health"
	uriInternalTenants       = "/tenants"
	uriInternalDevices       = "/tenants/:tenant_id/devices"
	urlInternalDevicesStatus = "/tenants/:tenant_id/devices/status/:status"
	uriInternalDeviceDetails = "/tenants/:tenant_id/devices/:device_id"
	uriInternalDeviceGroups  = "/tenants/:tenant_id/devices/:device_id/groups"
	urlInternalAttributes    = "/tenants/:tenant_id/device/:device_id/attribute/scope/:scope"
	urlInternalReindex       = "/tenants/:tenant_id/devices/:device_id/reindex"
	apiUrlManagementV2       = "/api/management/v2/inventory"
	urlFiltersAttributes     = "/filters/attributes"
	urlFiltersSearch         = "/filters/search"

	apiUrlInternalV2         = "/api/internal/v2/inventory"
	urlInternalFiltersSearch = "/tenants/:tenant_id/filters/search"

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

func (i *InternalAPI) LivelinessHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (i *InternalAPI) HealthCheckHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	err := i.App.HealthCheck(ctx)
	if err != nil {
		rest.RenderError(c, http.StatusServiceUnavailable, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// `sort` paramater value is an attribute name with optional direction (desc or asc)
// separated by colon (:)
//
// eg. `sort=attr_name1` or `sort=attr_name1:asc`
func parseSortParam(c *gin.Context) (*store.Sort, error) {
	sortStr, err := utils.ParseQueryParmStr(c.Request, queryParamSort, false, nil)
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
func parseFilterParams(c *gin.Context) ([]store.Filter, error) {
	knownParams := []string{
		utils.PageName,
		utils.PerPageName,
		queryParamSort,
		queryParamHasGroup,
		queryParamGroup,
	}
	filters := make([]store.Filter, 0)
	var filter store.Filter
	for name := range c.Request.URL.Query() {
		if utils.ContainsString(name, knownParams) {
			continue
		}
		valueStr, err := utils.ParseQueryParmStr(c.Request, name, false, nil)
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

func (i *ManagementAPI) GetDevicesHandler(c *gin.Context) {
	ctx := c.Request.Context()

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	hasGroup, err := utils.ParseQueryParmBool(c.Request, queryParamHasGroup, false, nil)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	groupName, err := utils.ParseQueryParmStr(c.Request, "group", false, nil)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	sort, err := parseSortParam(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	filters, err := parseFilterParams(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	ld := store.ListQuery{Skip: int((page - 1) * perPage),
		Limit:     int(perPage),
		Filters:   filters,
		Sort:      sort,
		HasGroup:  hasGroup,
		GroupName: groupName}

	devs, totalCount, err := i.App.ListDevices(ctx, ld)

	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	hasNext := totalCount > int(page*perPage)

	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
		SetHasNext(hasNext).
		SetTotalCount(int64(totalCount))

	links, err := rest.MakePagingHeaders(c.Request, hints)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	for _, l := range links {
		c.Writer.Header().Add("Link", l)
	}
	// the response writer will ensure the header name is in Kebab-Pascal-Case
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	c.JSON(http.StatusOK, devs)
}

func (i *ManagementAPI) GetDeviceHandler(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")

	dev, err := i.App.GetDevice(ctx, model.DeviceID(deviceID))
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	if dev == nil {
		rest.RenderError(c,
			http.StatusNotFound,
			store.ErrDevNotFound,
		)
		return
	}
	if dev.TagsEtag != "" {
		c.Header("ETag", dev.TagsEtag)
	}
	c.JSON(http.StatusOK, dev)
}

func (i *ManagementAPI) DeleteDeviceInventoryHandler(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")

	err := i.App.ReplaceAttributes(ctx, model.DeviceID(deviceID),
		model.DeviceAttributes{}, model.AttrScopeInventory, "")
	if err != nil && err != store.ErrDevNotFound {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *InternalAPI) DeleteDeviceHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tenantId := c.Param("tenant_id")
	if tenantId != "" {
		id := &identity.Identity{
			Tenant: tenantId,
		}
		ctx = identity.WithContext(ctx, id)
	}

	deviceID := c.Param("device_id")

	err := i.App.DeleteDevice(ctx, model.DeviceID(deviceID))
	if err != nil && err != store.ErrDevNotFound {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *InternalAPI) AddDeviceHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tenantId := c.Param("tenant_id")
	if tenantId != "" {
		id := &identity.Identity{
			Tenant: tenantId,
		}
		ctx = identity.WithContext(ctx, id)
	}

	dev, err := parseDevice(c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	err = dev.Attributes.Validate()
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	err = i.App.AddDevice(ctx, dev)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Writer.Header().Add("Location", "devices/"+dev.ID.String())
	c.Status(http.StatusCreated)
}

func (i *ManagementAPI) UpdateDeviceAttributesHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var idata *identity.Identity
	if idata = identity.FromContext(ctx); idata == nil || !idata.IsDevice {
		rest.RenderError(c,
			http.StatusUnauthorized,
			errors.New("unauthorized"),
		)
		return
	}
	deviceID := model.DeviceID(idata.Subject)
	//extract attributes from body
	attrs, err := parseAttributes(c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			err,
		)
		return
	}
	i.updateDeviceAttributes(c, attrs, deviceID, model.AttrScopeInventory, "")
}

func (i *ManagementAPI) UpdateDeviceTagsHandler(c *gin.Context) {
	// get device ID from uri
	deviceID := model.DeviceID(c.Param("id"))
	if len(deviceID) < 1 {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("device id cannot be empty"),
		)
		return
	}

	ifMatchHeader := c.Request.Header.Get("If-Match")

	// extract attributes from body
	attrs, err := parseAttributes(c)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			err,
		)
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

	i.updateDeviceAttributes(c, attrs, deviceID, model.AttrScopeTags, ifMatchHeader)
}

func (i *ManagementAPI) updateDeviceAttributes(
	c *gin.Context,
	attrs model.DeviceAttributes,
	deviceID model.DeviceID,
	scope string,
	etag string,
) {
	ctx := c.Request.Context()

	var err error

	// upsert or replace the attributes
	if c.Request.Method == http.MethodPatch {
		err = i.App.UpsertAttributesWithUpdated(ctx, deviceID, attrs, scope, etag)
	} else if c.Request.Method == http.MethodPut {
		err = i.App.ReplaceAttributes(ctx, deviceID, attrs, scope, etag)
	} else {
		rest.RenderError(c,
			http.StatusMethodNotAllowed,
			errors.New("method not alllowed"),
		)
		return
	}

	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
	case inventory.ErrTooManyAttributes:
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	case inventory.ErrETagDoesntMatch:
		rest.RenderError(c,
			http.StatusPreconditionFailed,
			cause,
		)
		return
	}
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (i *InternalAPI) PatchDeviceAttributesInternalHandler(
	c *gin.Context,
) {
	ctx := c.Request.Context()
	tenantId := c.Param("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	deviceId := c.Param("device_id")
	if len(deviceId) < 1 {
		rest.RenderError(c, http.StatusBadRequest, errors.New("device id cannot be empty"))
		return
	}
	//extract attributes from body
	attrs, err := parseAttributes(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}
	for i := range attrs {
		attrs[i].Scope = c.Param("scope")
		if attrs[i].Name == checkInTimeParamName && attrs[i].Scope == checkInTimeParamScope {
			t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", attrs[i].Value))
			if err != nil {
				rest.RenderError(c, http.StatusBadRequest, err)
				return
			}
			attrs[i].Value = t
		}
	}

	//upsert the attributes
	err = i.App.UpsertAttributes(ctx, model.DeviceID(deviceId), attrs)
	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
		rest.RenderError(c, http.StatusBadRequest, cause)
		return
	}
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (i *ManagementAPI) DeleteDeviceGroupHandler(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")
	groupName := c.Param("name")

	err := i.App.UnsetDeviceGroup(ctx, model.DeviceID(deviceID), model.GroupName(groupName))
	if err != nil {
		cause := errors.Cause(err)
		if cause != nil {
			if cause.Error() == store.ErrDevNotFound.Error() {
				rest.RenderError(c, http.StatusNotFound, err)
				return
			}
		}
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (i *ManagementAPI) AddDeviceToGroupHandler(c *gin.Context) {
	ctx := c.Request.Context()

	devId := c.Param("id")

	var group InventoryApiGroup
	err := c.ShouldBindJSON(&group)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "failed to decode device group data"))
		return
	}

	if err = group.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	err = i.App.UpdateDeviceGroup(ctx, model.DeviceID(devId), model.GroupName(group.Group))
	if err != nil {
		if cause := errors.Cause(err); cause != nil && cause == store.ErrDevNotFound {
			rest.RenderError(c, http.StatusNotFound, err)
			return
		}
		rest.RenderInternalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (i *ManagementAPI) GetDevicesByGroupHandler(c *gin.Context) {
	ctx := c.Request.Context()

	group := c.Param("name")

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	//get one extra device to see if there's a 'next' page
	ids, totalCount, err := i.App.ListDevicesByGroup(
		ctx,
		model.GroupName(group),
		int((page-1)*perPage),
		int(perPage),
	)
	if err != nil {
		if err == store.ErrGroupNotFound {
			rest.RenderError(c, http.StatusNotFound, err)

		} else {
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}
		return
	}

	hasNext := totalCount > int(page*perPage)

	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
		SetHasNext(hasNext).
		SetTotalCount(int64(totalCount))

	links, err := rest.MakePagingHeaders(c.Request, hints)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}
	for _, l := range links {
		c.Writer.Header().Add("Link", l)
	}
	// the response writer will ensure the header name is in Kebab-Pascal-Case
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	c.JSON(http.StatusOK, ids)
}

func (i *ManagementAPI) AppendDevicesToGroup(c *gin.Context) {
	var deviceIDs []model.DeviceID
	ctx := c.Request.Context()
	groupName := model.GroupName(c.Param("name"))
	if err := groupName.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&deviceIDs); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "invalid payload schema"),
		)
		return
	} else if len(deviceIDs) == 0 {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("no device IDs present in payload"),
		)
		return
	}
	updated, err := i.App.UpdateDevicesGroup(
		ctx, deviceIDs, groupName,
	)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (i *ManagementAPI) DeleteGroupHandler(c *gin.Context) {
	ctx := c.Request.Context()

	groupName := model.GroupName(c.Param("name"))
	if err := groupName.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	updated, err := i.App.DeleteGroup(ctx, groupName)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (i *ManagementAPI) ClearDevicesGroupHandler(c *gin.Context) {
	var deviceIDs []model.DeviceID
	ctx := c.Request.Context()

	groupName := model.GroupName(c.Param("name"))
	if err := groupName.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if err := c.ShouldBindJSON(&deviceIDs); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "invalid payload schema"),
		)
		return
	} else if len(deviceIDs) == 0 {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("no device IDs present in payload"),
		)
		return
	}

	updated, err := i.App.UnsetDevicesGroup(ctx, deviceIDs, groupName)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func parseDevice(c *gin.Context) (*model.Device, error) {
	dev := model.Device{}

	//decode body
	err := c.ShouldBindJSON(&dev)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if err := dev.Validate(); err != nil {
		return nil, err
	}

	return &dev, nil
}

func parseAttributes(c *gin.Context) (model.DeviceAttributes, error) {
	var attrs model.DeviceAttributes

	err := c.ShouldBindJSON(&attrs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	err = attrs.Validate()
	if err != nil {
		return nil, err
	}

	return attrs, nil
}

func (i *ManagementAPI) GetGroupsHandler(c *gin.Context) {
	var fltr []model.FilterPredicate
	ctx := c.Request.Context()

	query := c.Request.URL.Query()
	status := query.Get("status")
	if status != "" {
		fltr = []model.FilterPredicate{{
			Attribute: "status",
			Scope:     "identity",
			Type:      "$eq",
			Value:     status,
		}}
	}

	groups, err := i.App.ListGroups(ctx, fltr)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	if groups == nil {
		groups = []model.GroupName{}
	}

	c.JSON(http.StatusOK, groups)
}

func (i *ManagementAPI) GetDeviceGroupHandler(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")

	group, err := i.App.GetDeviceGroup(ctx, model.DeviceID(deviceID))
	if err != nil {
		if err == store.ErrDevNotFound {
			rest.RenderError(c,
				http.StatusNotFound,
				store.ErrDevNotFound,
			)
		} else {
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}
		return
	}

	ret := map[string]*model.GroupName{"group": nil}

	if group != "" {
		ret["group"] = &group
	}

	c.JSON(http.StatusOK, ret)
}

type newTenantRequest struct {
	TenantID string `json:"tenant_id" valid:"required"`
}

func (t newTenantRequest) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.TenantID, validation.Required),
	)
}
func (i *InternalAPI) CreateTenantHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var newTenant newTenantRequest

	if err := c.ShouldBindJSON(&newTenant); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if err := newTenant.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	err := i.App.CreateTenant(ctx, model.NewTenant{
		ID: newTenant.TenantID,
	})
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (i *ManagementAPI) FiltersAttributesHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// query the database
	attributes, err := i.App.GetFiltersAttributes(ctx)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	// in case of nil make sure we return empty list
	if attributes == nil {
		attributes = []model.FilterAttribute{}
	}

	c.JSON(http.StatusOK, attributes)
}

func (i *ManagementAPI) FiltersSearchHandler(c *gin.Context) {
	ctx := c.Request.Context()

	//extract attributes from body
	searchParams, err := parseSearchParams(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	// query the database
	devs, totalCount, err := i.App.SearchDevices(ctx, *searchParams)
	if err != nil {
		if strings.Contains(err.Error(), "BadValue") {
			rest.RenderError(c, http.StatusBadRequest, err)
		} else {
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}
		return
	}

	// the response writer will ensure the header name is in Kebab-Pascal-Case
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	c.JSON(http.StatusOK, devs)
}

func (i *InternalAPI) InternalFiltersSearchHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("tenant_id")
	if tenantId != "" {
		ctx = getTenantContext(ctx, tenantId)
	}

	//extract attributes from body
	searchParams, err := parseSearchParams(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	// query the database
	devs, totalCount, err := i.App.SearchDevices(ctx, *searchParams)
	if err != nil {
		if strings.Contains(err.Error(), "BadValue") {
			rest.RenderError(c, http.StatusBadRequest, err)
		} else {
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}
		return
	}

	// the response writer will ensure the header name is in Kebab-Pascal-Case
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	c.JSON(http.StatusOK, devs)
}

func getTenantContext(ctx context.Context, tenantId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if tenantId == "" {
		return ctx
	}
	id := &identity.Identity{
		Tenant: tenantId,
	}

	ctx = identity.WithContext(ctx, id)

	return ctx
}

func (i *InternalAPI) InternalDevicesStatusHandler(c *gin.Context) {
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

	ctx := c.Request.Context()

	tenantID := c.Param("tenant_id")
	ctx = getTenantContext(ctx, tenantID)

	status := c.Param("status")

	err := c.ShouldBindJSON(&devices)
	if err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "cant parse devices"),
		)
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
		result, err = i.App.UpsertDevicesStatuses(ctx, devices, attrs)
	case StatusDecommissioned:
		// Delete Inventory
		result, err = i.App.DeleteDevices(ctx, getIdsFromDevices(devices))
	default:
		// Unrecognized status
		rest.RenderError(c,
			http.StatusNotFound,
			errors.Errorf("unrecognized status: %s", status),
		)
		return
	}
	if err == store.ErrWriteConflict {
		rest.RenderError(c,
			http.StatusConflict,
			err,
		)
		return
	} else if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (i *InternalAPI) GetDeviceGroupsInternalHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	deviceID := c.Param("device_id")
	group, err := i.App.GetDeviceGroup(ctx, model.DeviceID(deviceID))
	if err != nil {
		if err == store.ErrDevNotFound {
			rest.RenderError(c,
				http.StatusNotFound,
				store.ErrDevNotFound,
			)
		} else {
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}
		return
	}

	res := model.DeviceGroups{}
	if group != "" {
		res.Groups = append(res.Groups, string(group))
	}

	c.JSON(http.StatusOK, res)
}

func (i *InternalAPI) ReindexDeviceDataHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tenantId := c.Param("tenant_id")
	ctx = getTenantContext(ctx, tenantId)

	deviceId := c.Param("device_id")
	if len(deviceId) < 1 {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("device id cannot be empty"),
		)
		return
	}

	serviceName, err := utils.ParseQueryParmStr(c.Request, "service", false, nil)
	// inventory service accepts only reindex requests from devicemonitor
	if err != nil || serviceName != "devicemonitor" {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("unsupported service"),
		)
		return
	}

	// check devicemonitor alerts
	alertsCount, err := i.App.CheckAlerts(ctx, deviceId)
	if err != nil {
		rest.RenderInternalError(c, err)
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
	err = i.App.UpsertAttributes(ctx, model.DeviceID(deviceId), attrs)
	cause := errors.Cause(err)
	switch cause {
	case store.ErrNoAttrName:
		rest.RenderError(c, http.StatusBadRequest, cause)
		return
	}
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func getIdsFromDevices(devices []model.DeviceUpdate) []model.DeviceID {
	ids := make([]model.DeviceID, len(devices))
	for i, dev := range devices {
		ids[i] = dev.Id
	}
	return ids
}

func parseSearchParams(c *gin.Context) (*model.SearchParams, error) {
	var searchParams model.SearchParams

	if err := c.ShouldBindJSON(&searchParams); err != nil {
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
