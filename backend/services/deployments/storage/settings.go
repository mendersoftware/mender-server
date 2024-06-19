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

package storage

import (
	"context"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type awsSettingsContextKey struct{}

func SettingsWithContext(ctx context.Context, set *model.StorageSettings) context.Context {
	return context.WithValue(ctx, awsSettingsContextKey{}, set)
}

func SettingsFromContext(ctx context.Context) (*model.StorageSettings, bool) {
	set, ok := ctx.Value(awsSettingsContextKey{}).(*model.StorageSettings)
	if ok {
		return set, true
	}
	return nil, false
}
