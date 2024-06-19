// Copyright 2021 Northern.tech AS
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
package main

import (
	"context"

	"github.com/mendersoftware/mender-server/services/useradm/authz"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/scope"
)

// SimpleAuthz is a trivial authorizer, mostly ensuring
// proper permission check for the 'create initial user' case.
type SimpleAuthz struct {
}

// Authorize makes SimpleAuthz implement the Authorizer interface.
func (sa *SimpleAuthz) Authorize(
	_ context.Context,
	token *jwt.Token,
	resource,
	action string,
) error {
	if token == nil {
		return authz.ErrAuthzUnauthorized
	}

	tokenScope := token.Claims.Scope

	// allow all actions on all services for 'mender.*'
	if tokenScope == scope.All {
		return nil
	}

	return authz.ErrAuthzUnauthorized
}
