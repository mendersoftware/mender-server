//nolint:all // This is all test code
package opensource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/utils/types"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

// InventoryManagementV1Suite ports (a subset of) the OS inventory tests from
// mender-server/backend/tests/integration/test_inventory.py. Enterprise-only
// branches (tenant fixtures/useExistingTenant) and
// TestDeviceFiltering::test_search_v2_internal (the internal API is a
// tenant-scoped variant of the same search endpoint and behaves identically
// in open-source when no tenant_id is given -- see the docstring on the
// Python test itself) are intentionally not ported.
type InventoryManagementV1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string

	// createdDevices tracks every device this suite creates, so
	// TearDownSuite can decommission them again -- see the same field in
	// DevauthManagementV2Suite for why this matters on a shared environment.
	createdDevices   []string
	createdDevicesMu sync.Mutex
}

func (i *BackendIntegrationSuite) TestInventoryManagementV1() {
	suite.Run(i.T(), &InventoryManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *InventoryManagementV1Suite) SetupSuite() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, r, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.Equal(http.StatusOK, r.StatusCode)
	require.NotEmpty(token)
	s.JWT = token
}

// TearDownSuite decommissions every device this suite created, so this
// suite doesn't leak devices into other suites that assert exact (not
// delta) device/inventory counts against the shared environment.
func (s *InventoryManagementV1Suite) TearDownSuite() {
	if s.JWT == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = common.JWTAuthContext(ctx, s.JWT)
	for _, id := range s.createdDevices {
		_, _ = s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementDecommissionDevice(ctx, id).Execute()
	}
}

func (s *InventoryManagementV1Suite) trackDevice(id string) {
	if id == "" {
		return
	}
	s.createdDevicesMu.Lock()
	s.createdDevices = append(s.createdDevices, id)
	s.createdDevicesMu.Unlock()
}

// createAcceptedDevices creates n accepted devices concurrently (bounded),
// tracking their IDs for TearDownSuite cleanup.
func (s *InventoryManagementV1Suite) createAcceptedDevices(ctx context.Context, n int) ([]*common.Device, error) {
	devices := make([]*common.Device, n)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	for i := range n {
		eg.Go(func() error {
			d, err := common.NewAcceptedDevice(egCtx, s.APIClient, s.Tenant.TenantToken)
			if err != nil {
				return err
			}
			// Track immediately (trackDevice is mutex-guarded): if a
			// sibling goroutine fails, already-created devices must
			// still be cleaned up in TearDownSuite or they leak into
			// other suites' assertions.
			s.trackDevice(d.ID)
			devices[i] = d
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return devices, nil
}

// ---------------------------------------------------------------------
// TestGetDevices (test_inventory.py::TestGetDevices)
// ---------------------------------------------------------------------

func (s *InventoryManagementV1Suite) TestGetDevicesOk() {
	// ported from test_inventory.py::TestGetDevices::test_get_devices_ok
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	count, err := s.totalDeviceCount(ctx)
	require.NoError(err)

	_, err = s.createAcceptedDevices(ctx, 40)
	require.NoError(err)

	newCount, err := s.totalDeviceCount(ctx)
	require.NoError(err)
	assert.Equal(count+40, newCount)
}

func (s *InventoryManagementV1Suite) TestFilterDevicesOk() {
	// ported from test_inventory.py::TestGetDevices::test_filter_devices_ok
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	count, err := s.totalDeviceCount(ctx)
	require.NoError(err)

	devs, err := s.createAcceptedDevices(ctx, 40)
	require.NoError(err)

	newCount, err := s.totalDeviceCount(ctx)
	require.NoError(err)
	assert.Equal(count+40, newCount)

	// upload inventory attributes
	macs := make([]string, len(devs))
	for i, d := range devs {
		mac := "de:ad:be:ef:06:" + fmt.Sprint(i)
		macs[i] = mac
		payload := []client.DeviceAttributeRequest{
			{Name: "mac", Value: client.StringAsAttributeValueRequest(types.Pointer(mac))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
	}

	// get device with exact mac value
	apiDevs, err := s.listDevicesRaw(ctx, url.Values{
		"per_page": {"100"},
		"mac":      {"de:ad:be:ef:06:7"},
	})
	require.NoError(err)
	assert.Len(apiDevs, 1)

	// G8: numeric attribute filter (python test_inventory_searching.py::
	// test_inventory_searching): give one device an int-typed attribute and
	// filter on its exact value.
	numAttr := float32(31337)
	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(common.JWTAuthContext(ctx, devs[7].Token)).
		DeviceAttributeRequest([]client.DeviceAttributeRequest{
			{Name: "test-num-attr", Value: client.Float32AsAttributeValueRequest(&numAttr)},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

	apiDevs, err = s.listDevicesRaw(ctx, url.Values{
		"per_page":      {"100"},
		"test-num-attr": {"31337"},
	})
	require.NoError(err)
	assert.Len(apiDevs, 1)
}

// totalDeviceCount reads the X-Total-Count header from a minimal listing
// request (mirrors the Python original, which uses per_page=1 purely to
// read the header cheaply).
func (s *InventoryManagementV1Suite) totalDeviceCount(ctx context.Context) (int, error) {
	_, r, err := s.APIClient.DeviceInventoryManagementAPIAPI.
		ListDeviceInventories(ctx).PerPage(1).Execute()
	if err != nil {
		return 0, err
	}
	var count int
	_, err = fmt.Sscanf(r.Header.Get("X-Total-Count"), "%d", &count)
	return count, err
}

// ---------------------------------------------------------------------
// TestDevicePatchAttributes (test_inventory.py::TestDevicePatchAttributes)
// ---------------------------------------------------------------------

func (s *InventoryManagementV1Suite) TestDevicePatchAttributesOk() {
	// ported from test_inventory.py::TestDevicePatchAttributes::test_ok
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs := make([]*common.Device, 3)
	for i := range devs {
		d, err := s.newAcceptedDeviceWithSN(ctx)
		require.NoError(err)
		s.trackDevice(d.ID)
		devs[i] = d
	}

	for _, d := range devs {
		payload := []client.DeviceAttributeRequest{
			{Name: "mac", Value: client.StringAsAttributeValueRequest(types.Pointer("mac-new-" + d.ID))},
			// empty value for existing
			{Name: "sn", Value: client.StringAsAttributeValueRequest(types.Pointer(""))},
			// empty value for new
			{Name: "new-empty", Value: client.StringAsAttributeValueRequest(types.Pointer(""))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
	}

	for _, d := range devs {
		apiDev, _, err := s.APIClient.DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, d.ID).Execute()
		require.NoError(err)

		// Expected inventory count per scope:
		// {"inventory": 3, "identity": 1(status)+2(mac,sn from id_data), "system": 3}
		require.Len(apiDev.Attributes, 9)

		for _, a := range apiDev.Attributes {
			value := a.GetValue()
			str := ""
			if value.String != nil {
				str = *value.String
			}
			switch {
			case a.GetName() == "mac" && a.GetScope() == client.INVENTORY:
				assert.Equal("mac-new-"+apiDev.GetId(), str)
			case a.GetName() == "sn" && a.GetScope() == client.INVENTORY:
				assert.Equal("", str)
			case a.GetName() == "new-empty" && a.GetScope() == client.INVENTORY:
				assert.Equal("", str)
			case a.GetName() == "status" && a.GetScope() == client.IDENTITY:
				assert.Contains([]string{"accepted", "pending"}, str)
			case a.GetScope() != client.INVENTORY:
				assert.NotEqual("", str, "unexpected empty value for %s", a.GetName())
			default:
				assert.Fail("unexpected attribute " + a.GetName())
			}
		}
	}
}

func (s *InventoryManagementV1Suite) TestDevicePatchAttributesFailNoAttrValue() {
	// ported from test_inventory.py::TestDevicePatchAttributes::test_fail_no_attr_value
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	s.trackDevice(d.ID)

	// The generated client requires a Value on DeviceAttributeRequest, so
	// a payload that omits the "value" key entirely (as the Python
	// original sends) has to go over a raw request.
	body, err := json.Marshal([]map[string]string{{"name": "mac"}})
	require.NoError(err)
	resp, err := common.RawRequest(common.JWTAuthContext(ctx, d.Token), s.APIClient, http.MethodPatch,
		"/api/devices/v1/inventory/device/attributes", body)
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

// newAcceptedDeviceWithSN creates and accepts a device whose identity data
// carries both "mac" and "sn" fields, matching the Python fixtures
// (make_accepted_device(s) build id_data from {mac, sn}) so that the
// identity-scope attribute count matches the ported assertions exactly.
func (s *InventoryManagementV1Suite) newAcceptedDeviceWithSN(ctx context.Context) (*common.Device, error) {
	kp, err := common.NewKeyPair(common.KeyKindRSA)
	if err != nil {
		return nil, err
	}
	mac := uuid.NewString()
	sn := uuid.NewString()
	idData, err := json.Marshal(map[string]string{"mac": mac, "sn": sn})
	if err != nil {
		return nil, err
	}

	device := common.NewDeviceFromKeyPair(kp, string(idData))
	device.MAC = mac
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return nil, err
	}
	if err := device.Accept(ctx, s.APIClient); err != nil {
		return nil, err
	}
	// The first SubmitAuthRequest above ran before the device was
	// accepted, so it got a 401 and no token -- submit again now.
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return nil, err
	}
	return device, nil
}

// ---------------------------------------------------------------------
// TestDeviceFiltering (test_inventory.py::TestDeviceFiltering)
// ---------------------------------------------------------------------

func (s *InventoryManagementV1Suite) TestDeviceFilteringSearchV2() {
	// ported from test_inventory.py::TestDeviceFiltering::test_search_v2
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	type fixtureDev struct {
		id       string
		idx      float32
		artifact []string
	}

	devs := make([]fixtureDev, 3)
	ids := make([]string, 3)
	for i := range devs {
		d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		s.trackDevice(d.ID)

		payload := []client.DeviceAttributeRequest{
			{Name: "artifact", Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"v1"})},
			{Name: "idx", Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(i)))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)

		devs[i] = fixtureDev{id: d.ID, idx: float32(i), artifact: []string{"v1"}}
		ids[i] = d.ID
	}

	type expectedAttr struct {
		name  string
		scope client.Scope
		value client.AttributeValueRequest
	}
	invAttrs := func(dev fixtureDev) []expectedAttr {
		return []expectedAttr{
			{name: "idx", scope: client.INVENTORY, value: client.Float32AsAttributeValueRequest(types.Pointer(dev.idx))},
			{name: "artifact", scope: client.INVENTORY, value: client.ArrayOfStringAsAttributeValueRequest(&dev.artifact)},
		}
	}

	type expectedDev struct {
		id    string
		attrs []expectedAttr
	}

	// devices sorted by idx descending (all have artifact=["v1"], so none
	// are excluded by the $nin below)
	byIdxDesc := make([]fixtureDev, len(devs))
	copy(byIdxDesc, devs)
	sort.SliceStable(byIdxDesc, func(i, j int) bool { return byIdxDesc[i].idx > byIdxDesc[j].idx })

	cases := []struct {
		name       string
		params     client.SearchParams
		wantStatus int
		expected   []expectedDev // only checked when wantStatus == 200
	}{
		{
			name: "Test $eq single match",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(1)))},
				},
			},
			wantStatus: http.StatusOK,
			expected:   []expectedDev{{id: devs[1].id, attrs: invAttrs(devs[1])}},
		},
		{
			name: "Test $eq no-match",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "id_data", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(types.Pointer("illegal_data"))},
				},
			},
			wantStatus: http.StatusOK,
			expected:   []expectedDev{},
		},
		{
			name: "Test $nin, sort by descending idx",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$nin", Attribute: "artifact", Scope: client.INVENTORY, Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"v3"})},
				},
				Sort: []client.SortCriteria{
					{Attribute: "idx", Scope: client.INVENTORY, Order: "desc"},
				},
			},
			wantStatus: http.StatusOK,
			expected: []expectedDev{
				{id: byIdxDesc[0].id, attrs: invAttrs(byIdxDesc[0])},
				{id: byIdxDesc[1].id, attrs: invAttrs(byIdxDesc[1])},
				{id: byIdxDesc[2].id, attrs: invAttrs(byIdxDesc[2])},
			},
		},
		{
			name: "Error - missing type parameter",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Attribute: "artifact", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(types.Pointer("v1"))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Error - invalid filter scope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("user_defined"), Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(1)))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Error - invalid sort scope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(1)))},
				},
				Sort: []client.SortCriteria{
					{Attribute: "idx", Scope: client.Scope("user_defined"), Order: "desc"},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Error - invalid attribute scope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(1)))},
				},
				Attributes: []client.SelectAttribute{
					{Attribute: "idx", Scope: client.Scope("user_defined")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Error - invalid filter scope with path",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("../../../../Windows/system.ini"), Value: client.Float32AsAttributeValueRequest(types.Pointer(float32(1)))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Error - valid mongo query unsupported operation",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$type", Attribute: "artifact", Scope: client.INVENTORY, Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"int", "string", "array"})},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			results, r, err := s.APIClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.
				InventoryV2SearchDeviceInventories(ctx).
				SearchParams(tc.params).
				Execute()
			if tc.wantStatus == http.StatusOK {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			s.Require().Equal(tc.wantStatus, r.StatusCode)

			if tc.wantStatus != http.StatusOK {
				return
			}

			s.Require().Len(results, len(tc.expected), "unexpected number of results")
			for i, exp := range tc.expected {
				s.Assert().Equal(exp.id, results[i].GetId(), "unexpected device in response")
				for _, wantAttr := range exp.attrs {
					s.Assert().True(
						hasAttribute(results[i].Attributes, wantAttr.name, wantAttr.scope, wantAttr.value),
						"missing inventory attribute %s on device %s", wantAttr.name, exp.id,
					)
				}
			}
		})
	}
}

func hasAttribute(attrs []client.AttributeResponse, name string, scope client.Scope, value client.AttributeValueRequest) bool {
	for _, a := range attrs {
		if a.GetName() != name || a.GetScope() != scope {
			continue
		}
		v := a.GetValue()
		switch {
		case value.String != nil:
			return v.String != nil && *v.String == *value.String
		case value.Float32 != nil:
			return v.Float32 != nil && *v.Float32 == *value.Float32
		case value.ArrayOfString != nil:
			if v.ArrayOfString == nil || len(*v.ArrayOfString) != len(*value.ArrayOfString) {
				return false
			}
			for i, s := range *value.ArrayOfString {
				if (*v.ArrayOfString)[i] != s {
					return false
				}
			}
			return true
		case value.ArrayOfFloat32 != nil:
			if v.ArrayOfFloat32 == nil || len(*v.ArrayOfFloat32) != len(*value.ArrayOfFloat32) {
				return false
			}
			for i, f := range *value.ArrayOfFloat32 {
				if (*v.ArrayOfFloat32)[i] != f {
					return false
				}
			}
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------

// listDevicesRaw performs a GET against the management v1 inventory
// listing with arbitrary query parameters (e.g. `mac=<value>` attribute
// filters), which the generated client can't express since it only
// exposes page/per_page/sort/group/has_group as typed parameters.
func (s *InventoryManagementV1Suite) listDevicesRaw(ctx context.Context, query url.Values) ([]client.DeviceInventoryResponse, error) {
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodGet, "/api/management/v1/inventory/devices?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var devices []client.DeviceInventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}
