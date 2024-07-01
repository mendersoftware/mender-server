// Copyright 2021 Northern.tech AS
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

package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDevice(t *testing.T) {
	const tenant = "tenant"
	const id = "id"

	device := &Device{}
	assert.Equal(t, "", device.GetID())
	assert.Equal(t, "", device.GetTenantID())

	device = NewDevice(tenant, id)
	assert.Equal(t, id, device.GetID())
	assert.Equal(t, tenant, device.GetTenantID())
	assert.Equal(t, time.Time{}, device.GetUpdatedAt())

	const newId = "new_id"
	const newTenant = "new_tenant"
	now := time.Date(2010, 9, 22, 6, 5, 0, 0, time.UTC)

	device.SetID(newId)
	device.SetTenantID(newTenant)
	device.SetUpdatedAt(now)
	assert.Equal(t, newId, device.GetID())
	assert.Equal(t, newTenant, device.GetTenantID())
	assert.Equal(t, now, device.GetUpdatedAt())

	err := device.AppendAttr(NewInventoryAttribute("dummy"))
	assert.NotNil(t, err)
	assert.Equal(t, "unknown attribute scope dummy", err.Error())

	attr := NewInventoryAttribute(ScopeIdentity).SetName("a1").SetVal("a")
	assert.True(t, attr.IsStr())
	err = device.AppendAttr(attr)
	assert.Nil(t, err)

	attr = NewInventoryAttribute(ScopeInventory).SetName("a2").SetVal(float64(1))
	assert.True(t, attr.IsNum())
	err = device.AppendAttr(attr)
	assert.Nil(t, err)

	attr = NewInventoryAttribute(ScopeMonitor).SetName("a3").SetVal(true)
	assert.True(t, attr.IsBool())
	err = device.AppendAttr(attr)
	assert.Nil(t, err)

	err = device.AppendAttr(NewInventoryAttribute(ScopeTags).
		SetName("a4").SetVal([]interface{}{"a", "b"}))
	assert.Nil(t, err)

	err = device.AppendAttr(NewInventoryAttribute(ScopeSystem).
		SetName("a5").SetVal([]interface{}{float64(1.0), float64(2.0)}))
	assert.Nil(t, err)

	err = device.AppendAttr(NewInventoryAttribute(ScopeSystem).
		SetName("2").SetVal([]interface{}{true, true}))
	assert.Nil(t, err)

	_, err = json.Marshal(device)
	assert.Nil(t, err)
}

func TestMaybeParseAttr(t *testing.T) {
	scope, name, err := MaybeParseAttr("monitor_a1_str")
	assert.Nil(t, err)
	assert.Equal(t, "monitor", scope)
	assert.Equal(t, "a1", name)
}
