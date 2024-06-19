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

package store

import (
	"context"
	"errors"

	"github.com/mendersoftware/mender-server/pkg/mongo/oid"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/model"
)

var (
	// user not found
	ErrUserNotFound = errors.New("user not found")
	// token not found
	ErrTokenNotFound = errors.New("token not found")
	// duplicated email address
	ErrDuplicateEmail = errors.New("user with a given email already exists")
	// duplicated Personal Access Token name
	ErrDuplicateTokenName = errors.New("Personal Access Token with a given name already exists")
	// etag doesn't match
	ErrETagMismatch = errors.New("ETag doesn't match")
)

//go:generate ../utils/mockgen.sh
type DataStore interface {
	// Ping the storage service - verifying network connection.
	Ping(ctx context.Context) error
	// CreateUser persists the user
	CreateUser(ctx context.Context, u *model.User) error
	// Update user information - password or/and email address and
	// returns the updated user
	UpdateUser(ctx context.Context, id string, u *model.UserUpdate) (*model.User, error)
	UpdateLoginTs(ctx context.Context, id string) error
	//GetUserByEmail returns nil,nil if not found
	GetUserByEmail(ctx context.Context, email model.Email) (*model.User, error)
	GetUserById(ctx context.Context, id string) (*model.User, error)
	GetUserAndPasswordById(ctx context.Context, id string) (*model.User, error)
	GetUsers(ctx context.Context, fltr model.UserFilter) ([]model.User, error)
	DeleteUser(ctx context.Context, id string) error
	SaveToken(ctx context.Context, token *jwt.Token) error
	GetTokenById(ctx context.Context, id oid.ObjectID) (*jwt.Token, error)
	DeleteToken(ctx context.Context, userID, tokenID oid.ObjectID) error
	EnsureSessionTokensLimit(ctx context.Context, userID oid.ObjectID, sessionsLimit int) error

	// deletes all tenant's tokens (identity in context)
	DeleteTokens(ctx context.Context) error

	GetPersonalAccessTokens(ctx context.Context, userID string) ([]model.PersonalAccessToken, error)
	UpdateTokenLastUsed(ctx context.Context, id oid.ObjectID) error
	CountPersonalAccessTokens(ctx context.Context, userID string) (int64, error)

	// deletes user tokens
	DeleteTokensByUserId(ctx context.Context, userId string) error
	DeleteTokensByUserIdExceptCurrentOne(
		ctx context.Context,
		userId string,
		tokenID oid.ObjectID,
	) error

	SaveSettings(ctx context.Context, s *model.Settings, etag string) error
	GetSettings(ctx context.Context) (*model.Settings, error)
	SaveUserSettings(ctx context.Context, userID string, s *model.Settings, etag string) error
	GetUserSettings(ctx context.Context, userID string) (*model.Settings, error)
}
