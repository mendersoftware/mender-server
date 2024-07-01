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
package rbac

import (
	"context"
	"net/http"
	"strings"
)

type scopeContextKeyType int

const (
	scopeContextKey        scopeContextKeyType = 0
	ScopeHeader                                = "X-MEN-RBAC-Inventory-Groups"
	ScopeReleaseTagsHeader                     = "X-MEN-RBAC-Releases-Tags"
)

type Scope struct {
	DeviceGroups []string
	ReleaseTags  []string
}

// FromContext extracts current scope from context.Context
func FromContext(ctx context.Context) *Scope {
	val := ctx.Value(scopeContextKey)
	if v, ok := val.(*Scope); ok {
		return v
	}
	return nil
}

// WithContext adds scope to context `ctx` and returns the resulting context.
func WithContext(ctx context.Context, scope *Scope) context.Context {
	return context.WithValue(ctx, scopeContextKey, scope)
}

func ExtractScopeFromHeader(r *http.Request) *Scope {
	groupStr := r.Header.Get(ScopeHeader)
	tagsStr := r.Header.Get(ScopeReleaseTagsHeader)
	if len(groupStr) > 0 || len(tagsStr) > 0 {
		scope := Scope{}
		if len(groupStr) > 0 {
			scope.DeviceGroups = strings.Split(groupStr, ",")
		}
		if len(tagsStr) > 0 {
			scope.ReleaseTags = strings.Split(tagsStr, ",")
		}
		return &scope
	}
	return nil
}
