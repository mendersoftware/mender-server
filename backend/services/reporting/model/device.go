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

package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Device struct {
	ID                  *string             `json:"id"`
	TenantID            *string             `json:"tenant_id,omitempty"`
	Location            *string             `json:"location,omitempty"`
	IdentityAttributes  InventoryAttributes `json:"identity_attributes,omitempty"`
	InventoryAttributes InventoryAttributes `json:"inventory_attributes,omitempty"`
	MonitorAttributes   InventoryAttributes `json:"monitor_attributes,omitempty"`
	SystemAttributes    InventoryAttributes `json:"system_attributes,omitempty"`
	TagsAttributes      InventoryAttributes `json:"tags_attributes,omitempty"`
	UpdatedAt           *time.Time          `json:"updated_at,omitempty"`
	LastCheckInDate     *time.Time          `json:"check_in_time,omitempty"`
}

func NewDevice(tenantID, id string) *Device {
	return &Device{
		ID:       &id,
		TenantID: &tenantID,
	}
}

func (a *Device) AppendAttr(attr *InventoryAttribute) error {
	switch attr.Scope {
	case ScopeIdentity:
		a.IdentityAttributes = append(a.IdentityAttributes, attr)
		return nil
	case ScopeInventory:
		a.InventoryAttributes = append(a.InventoryAttributes, attr)
		return nil
	case ScopeMonitor:
		a.MonitorAttributes = append(a.MonitorAttributes, attr)
		return nil
	case ScopeSystem:
		a.SystemAttributes = append(a.SystemAttributes, attr)
		return nil
	case ScopeTags:
		a.TagsAttributes = append(a.TagsAttributes, attr)
		return nil
	default:
		return errors.New("unknown attribute scope " + attr.Scope)
	}
}

func (a *Device) GetID() string {
	if a.ID != nil {
		return *a.ID
	}
	return ""
}

func (a *Device) SetID(val string) *Device {
	a.ID = &val
	return a
}

func (a *Device) GetTenantID() string {
	if a.TenantID != nil {
		return *a.TenantID
	}
	return ""
}

func (a *Device) SetTenantID(val string) *Device {
	a.TenantID = &val
	return a
}

func (a *Device) GetUpdatedAt() time.Time {
	if a.UpdatedAt != nil {
		return *a.UpdatedAt
	}
	return time.Time{}
}

func (a *Device) SetUpdatedAt(val time.Time) *Device {
	if !val.IsZero() {
		a.UpdatedAt = &val
	}
	return a
}

func (a *Device) SetLastCheckIn(val time.Time) *Device {
	if !val.IsZero() {
		a.LastCheckInDate = &val
	}
	return a
}

type InventoryAttributes []*InventoryAttribute

type InventoryAttribute struct {
	Scope   string
	Name    string
	String  []string
	Numeric []float64
	Boolean []bool
}

func NewInventoryAttribute(s string) *InventoryAttribute {
	return &InventoryAttribute{
		Scope: s,
	}
}

func (a *InventoryAttribute) IsStr() bool {
	return a.String != nil
}

func (a *InventoryAttribute) IsNum() bool {
	return a.Numeric != nil
}

func (a *InventoryAttribute) IsBool() bool {
	return a.Boolean != nil
}

func (a *InventoryAttribute) SetName(val string) *InventoryAttribute {
	a.Name = val
	return a
}

func (a *InventoryAttribute) SetString(val string) *InventoryAttribute {
	a.String = []string{val}
	a.Boolean = nil
	a.Numeric = nil
	return a
}

func (a *InventoryAttribute) SetStrings(val []string) *InventoryAttribute {
	a.String = val
	a.Boolean = nil
	a.Numeric = nil
	return a
}

func (a *InventoryAttribute) SetNumeric(val float64) *InventoryAttribute {
	a.Numeric = []float64{val}
	a.Boolean = nil
	a.String = nil
	return a
}

func (a *InventoryAttribute) SetNumerics(val []float64) *InventoryAttribute {
	a.Numeric = val
	a.String = nil
	a.Boolean = nil
	return a
}

func (a *InventoryAttribute) SetBoolean(val bool) *InventoryAttribute {
	a.Boolean = []bool{val}
	a.Numeric = nil
	a.String = nil
	return a
}

func (a *InventoryAttribute) SetBooleans(val []bool) *InventoryAttribute {
	a.Boolean = val
	a.Numeric = nil
	a.String = nil
	return a
}

// SetVal inspects the 'val' type and sets the correct subtype field
// useful for translating from inventory attributes (interface{})
func (a *InventoryAttribute) SetVal(val interface{}) *InventoryAttribute {
	switch val := val.(type) {
	case bool:
		a.SetBoolean(val)
	case float64:
		a.SetNumeric(val)
	case string:
		a.SetString(val)
	case []interface{}:
		switch val[0].(type) {
		case bool:
			bools := make([]bool, len(val))
			for i, v := range val {
				bools[i] = v.(bool)
			}
			a.SetBooleans(bools)
		case float64:
			nums := make([]float64, len(val))
			for i, v := range val {
				nums[i] = v.(float64)
			}
			a.SetNumerics(nums)
		case string:
			strs := make([]string, len(val))
			for i, v := range val {
				strs[i] = v.(string)
			}
			a.SetStrings(strs)
		}
	}

	return a
}

func (d *Device) MarshalJSON() ([]byte, error) {
	// TODO: smarter encoding, without explicit rewrites?
	m := make(map[string]interface{})
	m[FieldNameID] = d.ID
	m[FieldNameTenantID] = d.TenantID
	m[FieldNameLocation] = d.Location
	if d.LastCheckInDate != nil {
		m[FieldNameCheckIn] = d.LastCheckInDate
	}

	attributes := append(d.IdentityAttributes, d.InventoryAttributes...)
	attributes = append(attributes, d.MonitorAttributes...)
	attributes = append(attributes, d.SystemAttributes...)
	attributes = append(attributes, d.TagsAttributes...)

	for _, a := range attributes {
		name, val := a.Map()
		m[name] = val
	}

	return json.Marshal(m)
}

func (a *InventoryAttribute) Map() (string, interface{}) {
	var val interface{}
	var typ Type

	if a.IsStr() {
		typ = TypeStr
		val = a.String
	} else if a.IsNum() {
		typ = TypeNum
		val = a.Numeric
	} else if a.IsBool() {
		typ = TypeBool
		val = a.Boolean
	}

	name := ToAttr(a.Scope, a.Name, typ)

	return name, val
}

// maybeParseAttr decides if a given field is an attribute and parses
// it's name + scope
func MaybeParseAttr(field string) (string, string, error) {
	scope := ""
	name := ""

	for _, s := range []string{ScopeIdentity, ScopeInventory, ScopeMonitor,
		ScopeSystem, ScopeTags} {
		if strings.HasPrefix(field, s+"_") {
			scope = s
			break
		}
	}

	if scope != "" {
		for _, s := range []string{typeStr, typeNum} {
			if strings.HasSuffix(field, "_"+s) {
				// strip the prefix/suffix
				start := strings.Index(field, "_")
				end := strings.LastIndex(field, "_")

				name = field[start+1 : end]
			}
		}
	}

	return scope, name, nil
}
