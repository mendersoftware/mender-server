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
package authz

import (
	"context"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"
)

var (
	ErrAuthzUnauthorized = errors.New("unauthorized")
	ErrAuthzNoAuth       = errors.New("authorization not present in header")
	ErrInvalidAuthHeader = errors.New("malformed Authorization header")
	ErrAuthzTokenInvalid = errors.New("invalid jwt")
)

// Authorizer defines the interface for checking the permissions of a given user(token) vs an action
// on a resource.
//
//go:generate ../../../utils/mockgen.sh
type Authorizer interface {
	// Authorize checks if the given user (identified by token) has permissions to an action on a
	// resource.
	// returns:
	// nil if authorization is granted
	// ErrAuthzUnauthorized otherwise
	// ErrAuthzTokenInvalid if can't parse token
	Authorize(ctx context.Context, token *jwt.Token, resource, action string) error
}
