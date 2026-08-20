// Copyright 2026 Northern.tech AS
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

package inv

import (
	"context"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

const IdentityNameAttribute = "name"

func (i inventory) setIdentities(
	ctx context.Context,
	deviceId model.DeviceID,
	newAttributes model.DeviceAttributes,
) error {
	if len(newAttributes) == 0 {
		return nil
	}

	identities := make([]any, 0, len(newAttributes))
	for i, v := range newAttributes {
		if v.Scope == model.AttrScopeIdentity {
			identities = append(identities, newAttributes[i].Value)
		}
		if v.Scope == model.AttrScopeTags && v.Name == IdentityNameAttribute {
			identities = append(identities, newAttributes[i].Value)
		}
	}
	if len(identities) == 0 {
		return nil
	}

	return i.db.SetIdentity(ctx, deviceId, identities)
}
