// Copyright 2022 Northern.tech AS
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

package mapping

import (
	"context"
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	"github.com/mendersoftware/mender-server/services/reporting/model"
	"github.com/mendersoftware/mender-server/services/reporting/store/mocks"
)

func TestNewMapper(t *testing.T) {
	m := NewMapper(nil)
	assert.NotNil(t, m)
}

func TestMapInventoryAttributes(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs       inventory.DeviceAttributes
		update      bool
		passthrough bool
		mapping     *model.Mapping
		out         inventory.DeviceAttributes
		err         error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
			},
			update: true,
			mapping: &model.Mapping{
				TenantID: tenantID,
				Inventory: []string{
					path.Join(model.ScopeInventory, "a1"),
					path.Join(model.ScopeInventory, "a2"),
				},
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
			},
		},
		"ok, no update": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeInventory},
			},
			update: false,
			mapping: &model.Mapping{
				TenantID: tenantID,
				Inventory: []string{
					path.Join(model.ScopeInventory, "a1"),
					path.Join(model.ScopeInventory, "a2"),
				},
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
			},
		},
		"ok, no update, passthrough": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeInventory},
			},
			update:      false,
			passthrough: true,
			mapping: &model.Mapping{
				TenantID: tenantID,
				Inventory: []string{
					path.Join(model.ScopeInventory, "a1"),
					path.Join(model.ScopeInventory, "a2"),
				},
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeInventory},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
			},
			update: true,
			err:    errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			if tc.update {
				ds.On("UpdateAndGetMapping",
					ctx,
					tenantID,
					mock.AnythingOfType("[]string"),
				).Return(tc.mapping, tc.err)
			} else {
				ds.On("GetMapping",
					ctx,
					tenantID,
				).Return(tc.mapping, tc.err)
			}

			mapper := NewMapper(ds)
			attrs, err := mapper.MapInventoryAttributes(ctx, tenantID, tc.attrs, tc.update, tc.passthrough)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, attrs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, attrs)
			}
		})
	}
}

func TestReverseInventoryAttributes(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs   inventory.DeviceAttributes
		mapping *model.Mapping
		out     inventory.DeviceAttributes
		err     error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
			},
			mapping: &model.Mapping{
				TenantID: tenantID,
				Inventory: []string{
					path.Join(model.ScopeInventory, "a1"),
					path.Join(model.ScopeInventory, "a2"),
				},
			},
			out: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v2", Scope: model.ScopeInventory},
			},
			err: errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			ds.On("GetMapping",
				ctx,
				tenantID,
			).Return(tc.mapping, tc.err)

			mapper := NewMapper(ds)
			attrs, err := mapper.ReverseInventoryAttributes(ctx, tenantID, tc.attrs)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, attrs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, attrs)
			}
		})
	}
}

func TestGetMapping(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs            inventory.DeviceAttributes
		inventoryMapping []string
		mapping          *model.Mapping
		err              error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
			},
			inventoryMapping: []string{
				path.Join(model.ScopeInventory, "a1"),
				path.Join(model.ScopeInventory, "a2"),
			},
			mapping: &model.Mapping{
				TenantID:  tenantID,
				Inventory: []string{"a1", "a2"},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a3", Value: "v2", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
			},
			inventoryMapping: []string{
				path.Join(model.ScopeInventory, "a1"),
				path.Join(model.ScopeInventory, "a2"),
				path.Join(model.ScopeInventory, "a3"),
				path.Join(model.ScopeInventory, "a2"),
			},
			err: errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			ds.On("UpdateAndGetMapping",
				ctx,
				tenantID,
				tc.inventoryMapping,
			).Return(tc.mapping, tc.err)

			mapper := newMapper(ds)
			mapping, err := mapper.updateAndGetMapping(ctx, tenantID, tc.attrs)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, mapping)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mapping, mapping)
			}
		})
	}
}

func TestMapAttributes(t *testing.T) {
	testCases := map[string]struct {
		attrs   inventory.DeviceAttributes
		mapping map[string]string
		out     inventory.DeviceAttributes
	}{
		"case 1": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
				{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
			},
			mapping: map[string]string{
				"a1": fmt.Sprintf(inventoryAttributeTemplate, 1),
				"a2": fmt.Sprintf(inventoryAttributeTemplate, 2),
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := mapAttributes(tc.attrs, tc.mapping, true, false)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestAttributesToFields(t *testing.T) {
	testCases := map[string]struct {
		in  []string
		out map[string]string
	}{
		"case 1": {
			in: []string{"a1", "a2"},
			out: map[string]string{
				"a1": fmt.Sprintf(inventoryAttributeTemplate, 1),
				"a2": fmt.Sprintf(inventoryAttributeTemplate, 2),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := attributesToFields(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestFieldsToAttributes(t *testing.T) {
	testCases := map[string]struct {
		in  []string
		out map[string]string
	}{
		"case 1": {
			in: []string{"a1", "a2"},
			out: map[string]string{
				fmt.Sprintf(inventoryAttributeTemplate, 1): "a1",
				fmt.Sprintf(inventoryAttributeTemplate, 2): "a2",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := fieldsToAttributes(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	const tenantID = "tenantID"

	ds := &mocks.DataStore{}
	ds.On("GetMapping",
		ctx,
		tenantID,
	).Return(&model.Mapping{
		TenantID: tenantID,
		Inventory: []string{
			path.Join(model.ScopeInventory, "a1"),
			path.Join(model.ScopeInventory, "a2"),
		},
	}, nil).Once()

	mapper := NewMapper(ds)

	// first map call will cache the mapping
	res, err := mapper.MapInventoryAttributes(ctx, tenantID, inventory.DeviceAttributes{
		{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
		{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, false, false)
	assert.NoError(t, err)
	assert.Equal(t, inventory.DeviceAttributes{
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, res)

	// second map call will resuse the cached mapping
	res, err = mapper.MapInventoryAttributes(ctx, tenantID, inventory.DeviceAttributes{
		{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
		{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, false, false)
	assert.NoError(t, err)
	assert.Equal(t, inventory.DeviceAttributes{
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, res)

	// first reverse call will also resuse the cached mapping
	res, err = mapper.ReverseInventoryAttributes(ctx, tenantID, inventory.DeviceAttributes{
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	})
	assert.NoError(t, err)
	assert.Equal(t, inventory.DeviceAttributes{
		{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
		{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, res)

	// call with an unseen attribute will trigger a new query to the data storage
	ds.On("GetMapping",
		ctx,
		tenantID,
	).Return(&model.Mapping{
		TenantID:  tenantID,
		Inventory: []string{"a1", "a2", "a3"},
	}, nil).Once()

	res, err = mapper.ReverseInventoryAttributes(ctx, tenantID, inventory.DeviceAttributes{
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: model.ScopeInventory},
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: model.ScopeInventory},
		{Name: fmt.Sprintf(inventoryAttributeTemplate, 3), Value: "v3", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	})
	assert.NoError(t, err)
	assert.Equal(t, inventory.DeviceAttributes{
		{Name: "a1", Value: "v1", Scope: model.ScopeInventory},
		{Name: "a2", Value: "v2", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeInventory},
		{Name: "a3", Value: "v3", Scope: model.ScopeSystem},
	}, res)
}
