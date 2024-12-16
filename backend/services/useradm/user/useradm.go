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

package useradm

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/mendersoftware/mender-server/pkg/apiclient"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"

	"github.com/mendersoftware/mender-server/services/useradm/client/tenant"
	"github.com/mendersoftware/mender-server/services/useradm/common"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/scope"
	"github.com/mendersoftware/mender-server/services/useradm/store"
	"github.com/mendersoftware/mender-server/services/useradm/utils"
)

var (
	ErrUserNotFound           = store.ErrUserNotFound
	ErrDuplicateEmail         = store.ErrDuplicateEmail
	ErrETagMismatch           = errors.New("entity tag did not match any records")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrAuthExpired            = errors.New("token expired")
	ErrAuthInvalid            = errors.New("token is invalid")
	ErrTenantAccountSuspended = errors.New("tenant account suspended")
	ErrInvalidTenantID        = errors.New("invalid tenant id")
	ErrTooManyTokens          = errors.New(
		"maximum number of personal acess tokens reached for this user")
	ErrDuplicateTokenName = errors.New(
		"Personal Access Token with a given name already exists")
	// password mismatch
	ErrCurrentPasswordMismatch = errors.New("current password mismatch")
	// modification of other user's password is not allowed
	ErrCannotModifyPassword = errors.New("password cannot be modified")
	// password too similar to the email
	ErrPassAndMailTooSimilar = errors.New("password is too similar to the email")
)

const (
	TenantStatusSuspended = "suspended"
	userIdMe              = "me"
)

//go:generate ../../../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	// Login accepts email/password, returns JWT
	Login(ctx context.Context, email model.Email, pass string,
		options *LoginOptions) (*jwt.Token, error)
	Logout(ctx context.Context, token *jwt.Token) error
	CreateUser(ctx context.Context, u *model.User) error
	CreateUserInternal(ctx context.Context, u *model.UserInternal) error
	UpdateUser(ctx context.Context, id string, u *model.UserUpdate) error
	Verify(ctx context.Context, token *jwt.Token) error
	GetUsers(ctx context.Context, fltr model.UserFilter) ([]model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	SetPassword(ctx context.Context, u model.UserUpdate) error

	// SignToken generates a signed
	// token using configuration & method set up in UserAdmApp
	SignToken(ctx context.Context, t *jwt.Token) (string, error)
	DeleteToken(ctx context.Context, id string) error

	// IssuePersonalAccessToken issues Personal Access Token
	IssuePersonalAccessToken(ctx context.Context, tr *model.TokenRequest) (string, error)
	// GetPersonalAccessTokens returns list of Personal Access Tokens
	GetPersonalAccessTokens(ctx context.Context, userID string) ([]model.PersonalAccessToken, error)

	DeleteTokens(ctx context.Context, tenantId, userId string) error

	CreateTenant(ctx context.Context, tenant model.NewTenant) error
	GetPlans(ctx context.Context, skip, limit int) []model.Plan
	GetPlanBinding(ctx context.Context) (*model.PlanBindingDetails, error)
}

type Config struct {
	// token issuer
	Issuer string
	// token expiration time
	ExpirationTimeSeconds int64
	// maximum number of log in tokens per user
	// zero means no limit
	LimitSessionsPerUser int
	// maximum number of personal access tokens per user
	// zero means no limit
	LimitTokensPerUser int
	// how often we should update personal access token
	// with last used timestamp
	TokenLastUsedUpdateFreqMinutes int
	// path to the private key, used to generate kid header in JWT and to get all the keys
	PrivateKeyPath string
	// PrivateKeyFileNamePattern holds the regular expression used
	// to get the key id from a filename
	PrivateKeyFileNamePattern string
}

type ApiClientGetter func() apiclient.HttpRunner

func simpleApiClientGetter() apiclient.HttpRunner {
	return &apiclient.HttpApi{}
}

type UserAdm struct {
	// JWT serialized/deserializer
	jwtHandlers  map[int]jwt.Handler
	db           store.DataStore
	config       Config
	verifyTenant bool
	cTenant      tenant.ClientRunner
	clientGetter ApiClientGetter
}

func NewUserAdm(jwtHandlers map[int]jwt.Handler, db store.DataStore, config Config) *UserAdm {

	return &UserAdm{
		jwtHandlers:  jwtHandlers,
		db:           db,
		config:       config,
		clientGetter: simpleApiClientGetter,
	}
}

func (u *UserAdm) HealthCheck(ctx context.Context) error {
	err := u.db.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "error reaching MongoDB")
	}

	if u.verifyTenant {
		err = u.cTenant.CheckHealth(ctx)
		if err != nil {
			return errors.Wrap(err, "Tenantadm service unhealthy")
		}
	}

	return nil
}

func (u *UserAdm) Login(ctx context.Context, email model.Email, pass string,
	options *LoginOptions) (*jwt.Token, error) {
	var ident identity.Identity
	l := log.FromContext(ctx)

	if email == "" {
		return nil, ErrUnauthorized
	}

	if u.verifyTenant {
		// check the user's tenant
		tenant, err := u.cTenant.GetTenant(ctx, string(email), u.clientGetter())

		if err != nil {
			return nil, errors.Wrap(err, "failed to check user's tenant")
		}

		if tenant == nil {
			return nil, ErrUnauthorized
		}

		if tenant.Status == TenantStatusSuspended {
			return nil, ErrTenantAccountSuspended
		}

		ident.Tenant = tenant.ID
		ctx = identity.WithContext(ctx, &ident)
	}

	//get user
	user, err := u.db.GetUserByEmail(ctx, email)

	if user == nil && err == nil {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to get user")
	}

	//verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
	if err != nil {
		return nil, ErrUnauthorized
	}

	//generate and save token
	t, err := u.generateToken(
		user.ID,
		scope.All,
		ident.Tenant,
		options.NoExpiry,
		common.KeyIdFromPath(u.config.PrivateKeyPath, u.config.PrivateKeyFileNamePattern),
	)
	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to generate token")
	}

	err = u.db.SaveToken(ctx, t)
	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to save token")
	}
	if u.config.LimitSessionsPerUser > 0 {
		err = u.db.EnsureSessionTokensLimit(ctx, t.Subject, u.config.LimitSessionsPerUser)
		if err != nil {
			return nil, errors.Wrap(err, "useradm: failed to ensure session tokens limit")
		}
	}

	if err = u.db.UpdateLoginTs(ctx, user.ID); err != nil {
		l.Warnf("failed to update login timestamp: %s", err.Error())
	}

	return t, nil
}

func (u *UserAdm) generateToken(subject, scope, tenant string,
	noExpiry bool, keyId int) (*jwt.Token, error) {
	id := oid.NewUUIDv4()
	subjectID := oid.FromString(subject)
	now := jwt.Time{Time: time.Now()}
	ret := &jwt.Token{
		KeyId: keyId,
		Claims: jwt.Claims{
			ID:        id,
			Subject:   subjectID,
			Issuer:    u.config.Issuer,
			IssuedAt:  now,
			NotBefore: now,
			Tenant:    tenant,
			Scope:     scope,
			User:      true,
		}}
	if !noExpiry {
		ret.Claims.ExpiresAt = &jwt.Time{
			Time: now.Add(time.Second * time.Duration(u.config.ExpirationTimeSeconds)),
		}
	}
	return ret, ret.Claims.Valid()
}

func (u *UserAdm) SignToken(ctx context.Context, t *jwt.Token) (string, error) {
	if _, ok := u.jwtHandlers[t.KeyId]; !ok {
		return "", common.ErrKeyIdNotFound
	}
	return u.jwtHandlers[t.KeyId].ToJWT(t)
}

func (u *UserAdm) Logout(ctx context.Context, token *jwt.Token) error {
	return u.db.DeleteToken(ctx, token.Subject, token.ID)
}

func (ua *UserAdm) CreateUser(ctx context.Context, u *model.User) error {
	if utils.CheckIfPassSimilarToEmailRaw(string(u.Email), string(u.Password)) {
		return ErrPassAndMailTooSimilar
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "failed to generate password hash")
	}
	u.Password = string(hash)

	return ua.doCreateUser(ctx, u, true)
}

func (ua *UserAdm) CreateUserInternal(ctx context.Context, u *model.UserInternal) error {
	if u.PasswordHash != "" {
		u.Password = u.PasswordHash
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "failed to generate password hash")
		}
		u.Password = string(hash)
	}

	return ua.doCreateUser(ctx, &u.User, u.ShouldPropagate())
}

func (ua *UserAdm) doCreateUser(ctx context.Context, u *model.User, propagate bool) error {
	var tenantErr error

	if u.ID == "" {
		//technically one of the functions down the oid.NewUUIDv4() stack can panic,
		//so what we could do is: wrap around it and recover and fall back
		//to non-rand based generation (NewUUIDv5(u.Email)).
		//on the other hand if rand failed maybe it is a good idea t not continue
		//after all.
		//the previous call to NewUUIDv5 produced exact the same ids for same
		//emails, which leads to problems once user changes the email address,
		//and then the old email is used in the new user creation.
		id := oid.NewUUIDv4()
		u.ID = id.String()
	}

	id := identity.FromContext(ctx)
	if ua.verifyTenant && propagate {
		tenantErr = ua.cTenant.CreateUser(ctx,
			&tenant.User{
				ID:       u.ID,
				Name:     string(u.Email),
				TenantID: id.Tenant,
			},
			ua.clientGetter())

		if tenantErr != nil && tenantErr != tenant.ErrDuplicateUser {
			return errors.Wrap(tenantErr, "useradm: failed to create user in tenantadm")
		}
	}

	if tenantErr == tenant.ErrDuplicateUser {
		// check if the user exists in useradm
		// if the user does not exists then we should try to remove the user from tenantadm
		user, err := ua.db.GetUserByEmail(ctx, u.Email)
		if err != nil {
			return errors.Wrap(err, "tenant data out of sync: failed to get user from db")
		}
		if user == nil {
			if compensateErr := ua.compensateTenantUser(
				ctx,
				u.ID,
				id.Tenant,
			); compensateErr != nil {
				tenantErr = compensateErr
			}
			return errors.Wrap(tenantErr, "tenant data out of sync")
		}
		return store.ErrDuplicateEmail
	}

	if err := ua.db.CreateUser(ctx, u); err != nil {
		if err == store.ErrDuplicateEmail {
			return err
		}
		if ua.verifyTenant && propagate {
			// if the user could not be created in the useradm database
			// try to remove the user from tenantadm
			if compensateErr := ua.compensateTenantUser(
				ctx,
				u.ID,
				id.Tenant,
			); compensateErr != nil {
				err = errors.Wrap(err, compensateErr.Error())
			}
		}

		return errors.Wrap(err, "useradm: failed to create user in the db")
	}

	return nil
}

func (ua *UserAdm) compensateTenantUser(ctx context.Context, userId, tenantId string) error {
	err := ua.cTenant.DeleteUser(ctx, tenantId, userId, ua.clientGetter())

	if err != nil {
		return errors.Wrap(err, "faield to delete tenant user")
	}
	return nil
}

func (ua *UserAdm) validateUserUpdate(
	ctx context.Context,
	user *model.User,
	u *model.UserUpdate,
	me bool,
) error {
	// user can change own password only
	if !me {
		if len(u.Password) > 0 {
			return ErrCannotModifyPassword
		}
	} else {
		// when changing own password or email address
		// user has to provide current password
		if len(u.Password) > 0 || (u.Email != "" && u.Email != user.Email) {
			if err := bcrypt.CompareHashAndPassword(
				[]byte(user.Password),
				[]byte(u.CurrentPassword),
			); err != nil {
				return ErrCurrentPasswordMismatch
			}
		}
	}
	return nil
}

func (ua *UserAdm) deleteAndInvalidateUserTokens(
	ctx context.Context,
	userID string,
	token *jwt.Token,
) error {
	var err error
	if token != nil {
		err = ua.db.DeleteTokensByUserIdExceptCurrentOne(ctx, userID, token.ID)
	} else {
		err = ua.db.DeleteTokensByUserId(ctx, userID)
	}
	return err
}

func (ua *UserAdm) UpdateUser(ctx context.Context, id string, userUpdate *model.UserUpdate) error {
	idty := identity.FromContext(ctx)
	me := idty.Subject == id
	user, err := ua.db.GetUserAndPasswordById(ctx, id)
	if err != nil {
		return errors.Wrap(err, "useradm: failed to get user")
	} else if user == nil {
		return store.ErrUserNotFound
	}

	if err := ua.validateUserUpdate(ctx, user, userUpdate, me); err != nil {
		return err
	}

	if userUpdate.ETag == nil {
		// Update without the support for etags.
		next := user.NextETag()
		userUpdate.ETagUpdate = &next
		userUpdate.ETag = user.ETag
		if user.ETag == nil {
			// If the ETag field is not set, assign the etag to nil
			user.ETag = &model.ETagNil
		}
	} else if user.ETag == nil || *userUpdate.ETag != *user.ETag {
		return ErrETagMismatch
	}

	if len(userUpdate.Email) > 0 && userUpdate.Email != user.Email {
		if ua.verifyTenant {
			err := ua.cTenant.UpdateUser(ctx,
				idty.Tenant,
				id,
				&tenant.UserUpdate{
					Name: string(userUpdate.Email),
				},
				ua.clientGetter())

			switch err {
			case nil:
				break
			case tenant.ErrDuplicateUser:
				return store.ErrDuplicateEmail
			case tenant.ErrUserNotFound:
				return store.ErrUserNotFound
			default:
				return errors.Wrap(err, "useradm: failed to update user in tenantadm")
			}
		}
	}

	if utils.CheckIfPassSimilarToEmail(user, userUpdate) {
		return ErrPassAndMailTooSimilar
	}

	_, err = ua.db.UpdateUser(ctx, id, userUpdate)
	switch err {
	case nil:
		// invalidate the JWT tokens but the one used to update the user
		err = ua.deleteAndInvalidateUserTokens(ctx, id, userUpdate.Token)
		err = errors.Wrap(err, "useradm: failed to invalidate tokens")

	case store.ErrUserNotFound:
		// We matched the user earlier, the ETag must have changed in
		// the meantime
		err = ErrETagMismatch
	case store.ErrDuplicateEmail:
		break

	default:
		err = errors.Wrap(err, "useradm: failed to update user information")
	}

	return err
}

func (ua *UserAdm) Verify(ctx context.Context, token *jwt.Token) error {

	if token == nil {
		return ErrUnauthorized
	}

	l := log.FromContext(ctx)

	if !token.Claims.User {
		l.Errorf("not a user token")
		return ErrUnauthorized
	}

	if ua.verifyTenant {
		if token.Claims.Tenant == "" {
			l.Errorf("Token has no tenant claim")
			return jwt.ErrTokenInvalid
		}
	} else if token.Claims.Tenant != "" {
		l.Errorf("Unexpected tenant claim: %s in the token", token.Claims.Tenant)
		return jwt.ErrTokenInvalid
	}

	user, err := ua.db.GetUserById(ctx, token.Claims.Subject.String())
	if user == nil && err == nil {
		return ErrUnauthorized
	}
	if err != nil {
		return errors.Wrap(err, "useradm: failed to get user")
	}

	dbToken, err := ua.db.GetTokenById(ctx, token.ID)
	if dbToken == nil && err == nil {
		return ErrUnauthorized
	}
	if err != nil {
		return errors.Wrap(err, "useradm: failed to get token")
	}

	// in case the token is a personal access token, update last used timestam
	// to not overload the database with writes to tokens collection, we do not
	// update the timestamp every time, but instead we wait some configurable
	// amount of time between updates
	if dbToken.TokenName != nil && ua.config.TokenLastUsedUpdateFreqMinutes > 0 {
		t := time.Now().Add(
			(-time.Minute * time.Duration(ua.config.TokenLastUsedUpdateFreqMinutes)))
		if dbToken.LastUsed == nil || dbToken.LastUsed.Before(t) {
			if err := ua.db.UpdateTokenLastUsed(ctx, token.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ua *UserAdm) GetUsers(ctx context.Context, fltr model.UserFilter) ([]model.User, error) {
	users, err := ua.db.GetUsers(ctx, fltr)
	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to get users")
	}

	return users, nil
}

func (ua *UserAdm) GetUser(ctx context.Context, id string) (*model.User, error) {
	if id == userIdMe {
		id = identity.FromContext(ctx).Subject
	}
	user, err := ua.db.GetUserById(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to get user")
	}

	return user, nil
}

func (ua *UserAdm) DeleteUser(ctx context.Context, id string) error {
	if ua.verifyTenant {
		identity := identity.FromContext(ctx)
		err := ua.cTenant.DeleteUser(ctx, identity.Tenant, id, ua.clientGetter())

		if err != nil {
			return errors.Wrap(err, "useradm: failed to delete user in tenantadm")
		}
	}

	err := ua.db.DeleteUser(ctx, id)
	if err != nil {
		return errors.Wrap(err, "useradm: failed to delete user")
	}

	// remove user tokens
	err = ua.db.DeleteTokensByUserId(ctx, id)
	if err != nil {
		return errors.Wrap(err, "useradm: failed to delete user tokens")
	}

	return nil
}

// WithTenantVerification produces a UserAdm instance which enforces
// tenant verification vs the tenantadm service upon /login.
func (u *UserAdm) WithTenantVerification(c tenant.ClientRunner) *UserAdm {
	u.verifyTenant = true
	u.cTenant = c
	return u
}

func (u *UserAdm) CreateTenant(ctx context.Context, tenant model.NewTenant) error {
	return nil
}

func (ua *UserAdm) SetPassword(ctx context.Context, uu model.UserUpdate) error {
	u, err := ua.db.GetUserByEmail(ctx, uu.Email)
	if err != nil {
		return errors.Wrap(err, "useradm: failed to get user by email")

	}
	if u == nil {
		return ErrUserNotFound
	}
	if utils.CheckIfPassSimilarToEmail(u, &uu) {
		return ErrPassAndMailTooSimilar
	}

	_, err = ua.db.UpdateUser(ctx, u.ID, &uu)

	// if we changed the password, invalidate the JWT tokens but the one used to update the user
	if err == nil && uu.Password != "" {
		if uu.Token != nil {
			err = ua.db.DeleteTokensByUserIdExceptCurrentOne(ctx, u.ID, uu.Token.ID)
		} else {
			err = ua.db.DeleteTokensByUserId(ctx, u.ID)
		}
	}
	if err != nil {
		return errors.Wrap(err, "useradm: failed to update user information")
	}

	return nil
}

func (ua *UserAdm) DeleteTokens(ctx context.Context, tenantId, userId string) error {
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantId,
	})

	var err error

	if userId != "" {
		err = ua.db.DeleteTokensByUserId(ctx, userId)
	} else {
		err = ua.db.DeleteTokens(ctx)
	}

	if err != nil && err != store.ErrTokenNotFound {
		return errors.Wrapf(
			err,
			"failed to delete tokens for tenant: %v, user id: %v",
			tenantId,
			userId,
		)
	}

	return nil
}

func (u *UserAdm) IssuePersonalAccessToken(
	ctx context.Context,
	tr *model.TokenRequest,
) (string, error) {
	id := identity.FromContext(ctx)
	if id == nil {
		return "", errors.New("identity not present in the context")
	}
	if u.config.LimitTokensPerUser > 0 {
		count, err := u.db.CountPersonalAccessTokens(ctx, id.Subject)
		if err != nil {
			return "", errors.Wrap(err, "useradm: failed to count personal access tokens")
		}
		if count >= int64(u.config.LimitTokensPerUser) {
			return "", ErrTooManyTokens
		}
	}
	//generate and save token
	keyId := common.KeyIdFromPath(u.config.PrivateKeyPath, u.config.PrivateKeyFileNamePattern)
	t, err := u.generateToken(
		id.Subject,
		scope.All,
		id.Tenant,
		tr.ExpiresIn == 0,
		keyId,
	)
	if err != nil {
		return "", errors.Wrap(err, "useradm: failed to generate token")
	}
	// update claims
	t.TokenName = tr.Name
	if tr.ExpiresIn > 0 {
		expires := jwt.Time{
			Time: time.Now().Add(time.Second * time.Duration(tr.ExpiresIn)),
		}
		t.ExpiresAt = &expires
	}

	err = u.db.SaveToken(ctx, t)
	if err == store.ErrDuplicateTokenName {
		return "", ErrDuplicateTokenName
	} else if err != nil {
		return "", errors.Wrap(err, "useradm: failed to save token")
	}

	if _, ok := u.jwtHandlers[keyId]; !ok {
		return "", common.ErrKeyIdNotFound
	}
	// sign token
	return u.jwtHandlers[keyId].ToJWT(t)
}

func (ua *UserAdm) GetPersonalAccessTokens(
	ctx context.Context,
	userID string,
) ([]model.PersonalAccessToken, error) {
	tokens, err := ua.db.GetPersonalAccessTokens(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "useradm: failed to get tokens")
	}

	return tokens, nil
}

func (ua *UserAdm) DeleteToken(ctx context.Context, id string) error {
	identity := identity.FromContext(ctx)
	if identity == nil {
		return errors.New("identity not present in the context")
	}
	err := ua.db.DeleteToken(ctx, oid.FromString(identity.Subject), oid.FromString(id))
	if err != nil {
		return errors.Wrap(err, "useradm: failed to delete token")
	}

	return nil
}

func (ua *UserAdm) GetPlans(ctx context.Context, skip, limit int) []model.Plan {
	if (skip + limit) <= len(model.PlanList) {
		return model.PlanList[skip:(skip + limit)]
	} else if skip <= len(model.PlanList) {
		return model.PlanList[skip:]
	}
	return []model.Plan{}
}

func (ua *UserAdm) GetPlanBinding(ctx context.Context) (*model.PlanBindingDetails, error) {
	if len(model.PlanList) == 1 {
		return &model.PlanBindingDetails{
			Plan: model.PlanList[0],
		}, nil
	}
	return &model.PlanBindingDetails{}, nil
}
