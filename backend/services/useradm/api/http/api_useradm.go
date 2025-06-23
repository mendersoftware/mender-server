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
package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/useradm/authz"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/store"
	useradm "github.com/mendersoftware/mender-server/services/useradm/user"
)

const (
	defaultTimeout = time.Second * 5
	hdrETag        = "ETag"
	hdrIfMatch     = "If-Match"
)

var (
	ErrAuthHeader   = errors.New("invalid or missing auth header")
	ErrUserNotFound = errors.New("user not found")

	ErrAuthzUnauthorized = errors.New("unauthorized")
	ErrAuthzNoAuth       = errors.New("authorization not present in header")
	ErrInvalidAuthHeader = errors.New("malformed Authorization header")
	ErrAuthzTokenInvalid = errors.New("invalid jwt")
)

type UserAdmApiHandlers struct {
	userAdm    useradm.App
	db         store.DataStore
	jwth       map[int]jwt.Handler
	config     Config
	authorizer authz.Authorizer
}

type Config struct {
	// maximum expiration time for Personal Access Token
	TokenMaxExpSeconds int

	JWTFallback jwt.Handler
}

// return an ApiHandler for user administration and authentiacation app
func NewUserAdmApiHandlers(
	userAdm useradm.App,
	db store.DataStore,
	jwth map[int]jwt.Handler,
	config Config,
	authorizer authz.Authorizer,
) *UserAdmApiHandlers {
	return &UserAdmApiHandlers{
		userAdm:    userAdm,
		db:         db,
		jwth:       jwth,
		config:     config,
		authorizer: authorizer,
	}
}

// Getting nil means that a rest error has allready been renderd
func (i *UserAdmApiHandlers) authTokenExtractor(c *gin.Context) *jwt.Token {
	tokstr, err := ExtractToken(c.Request)
	if err != nil {
		rest.RenderError(c, http.StatusUnauthorized, err)
		return nil
	}

	keyId := jwt.GetKeyId(tokstr)

	if _, ok := i.jwth[keyId]; !ok {
		// we have not found the corresponding handler for the key by id
		// on purpose we do not return common.ErrKeyIdNotFound -- to not allow
		// the enumeration attack
		rest.RenderError(c, http.StatusUnauthorized, ErrAuthzTokenInvalid)
		return nil
	}

	// parse token
	token, err := i.jwth[keyId].FromJWT(tokstr)
	if err != nil && i.config.JWTFallback != nil {
		token, err = i.config.JWTFallback.FromJWT(tokstr)
	}
	if err != nil {
		rest.RenderError(c, http.StatusUnauthorized, ErrAuthzTokenInvalid)
		return nil
	}

	// extract resource action
	action, err := ExtractResourceAction(c.Request)
	if err != nil {
		rest.RenderInternalError(c, err)
		return nil
	}

	ctx := c.Request.Context()

	//authorize, no authz = http 403
	err = i.authorizer.Authorize(ctx, token, action.Resource, action.Method)
	if err != nil {
		if err == ErrAuthzUnauthorized {
			rest.RenderError(c, http.StatusForbidden, ErrAuthzUnauthorized)
		} else if err == ErrAuthzTokenInvalid {
			rest.RenderError(c, http.StatusUnauthorized, ErrAuthzTokenInvalid)
		} else {
			rest.RenderInternalError(c, err)
		}
		return nil
	}

	return token
}

func (u *UserAdmApiHandlers) AliveHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (u *UserAdmApiHandlers) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := u.userAdm.HealthCheck(ctx)
	if err != nil {
		rest.RenderError(c, http.StatusServiceUnavailable, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (u *UserAdmApiHandlers) AuthLoginHandler(c *gin.Context) {
	ctx := c.Request.Context()

	//parse auth header
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		rest.RenderError(c, http.StatusUnauthorized, ErrAuthHeader)
		return
	}
	email := model.Email(strings.ToLower(user))

	options := &useradm.LoginOptions{}
	err := c.ShouldBindJSON(options)
	if err != nil && c.Request.ContentLength != 0 {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	token, err := u.userAdm.Login(ctx, email, pass, options)
	if err != nil {
		switch {
		case err == useradm.ErrUnauthorized || err == useradm.ErrTenantAccountSuspended:
			rest.RenderError(c, http.StatusUnauthorized, err)
		default:
			rest.RenderInternalError(c, err)
		}
		return
	}

	raw, err := u.userAdm.SignToken(ctx, token)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	writer := c.Writer
	writer.Header().Set("Content-Type", "application/jwt")
	_, _ = writer.Write([]byte(raw))
}

func (u *UserAdmApiHandlers) AuthLogoutHandler(c *gin.Context) {
	ctx := c.Request.Context()

	if tokenStr, err := ExtractToken(c.Request); err == nil {
		keyId := jwt.GetKeyId(tokenStr)
		if _, ok := u.jwth[keyId]; !ok {
			rest.RenderInternalError(c, err)
			return
		}

		token, err := u.jwth[keyId].FromJWT(tokenStr)
		if err != nil {
			rest.RenderInternalError(c, err)
			return
		}
		if err := u.userAdm.Logout(ctx, token); err != nil {
			rest.RenderInternalError(c, err)
			return
		}
	}

	c.Status(http.StatusAccepted)
}

func (u *UserAdmApiHandlers) AuthVerifyHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// note that the request has passed through authz - the token is valid
	token := u.authTokenExtractor(c)
	if token == nil {
		return
	}

	err := u.userAdm.Verify(ctx, token)
	if err != nil {
		if err == useradm.ErrUnauthorized {
			rest.RenderError(c, http.StatusUnauthorized, useradm.ErrUnauthorized)
		} else {
			rest.RenderInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)
}

func (u *UserAdmApiHandlers) CreateTenantUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := parseUserInternal(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	tenantId := c.Param("id")
	if tenantId == "" {
		rest.RenderError(c, http.StatusNotFound, errors.New("Entity not found"))
		return
	}
	ctx = getTenantContext(ctx, tenantId)
	err = u.userAdm.CreateUserInternal(ctx, user)
	if err != nil {
		if err == store.ErrDuplicateEmail {
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		} else {
			rest.RenderInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusCreated)

}

func (u *UserAdmApiHandlers) AddUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := parseUser(c)
	if err != nil {
		if err == model.ErrPasswordTooShort {
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		} else {
			rest.RenderError(c, http.StatusBadRequest, err)
		}
		return
	}

	err = u.userAdm.CreateUser(ctx, user)
	if err != nil {
		if err == store.ErrDuplicateEmail || err == useradm.ErrPassAndMailTooSimilar {
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		} else {
			rest.RenderInternalError(c, err)
		}
		return
	}

	c.Writer.Header().Add("Location", "users/"+string(user.ID))
	c.Status(http.StatusCreated)

}

func (u *UserAdmApiHandlers) GetUsersHandler(c *gin.Context) {
	ctx := c.Request.Context()

	if err := c.Request.ParseForm(); err != nil {
		err = errors.Wrap(err, "api: bad form parameters")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	fltr := model.UserFilter{}
	if err := fltr.ParseForm(c.Request.Form); err != nil {
		err = errors.Wrap(err, "api: invalid form values")
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	users, err := u.userAdm.GetUsers(ctx, fltr)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (u *UserAdmApiHandlers) GetTenantUsersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = identity.WithContext(ctx, &identity.Identity{
		Tenant: c.Param("id"),
	})
	c.Request = c.Request.WithContext(ctx)
	u.GetUsersHandler(c)
}

func (u *UserAdmApiHandlers) GetUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	user, err := u.userAdm.GetUser(ctx, id)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	if user == nil {
		rest.RenderError(c, http.StatusNotFound, ErrUserNotFound)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u *UserAdmApiHandlers) UpdateUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	idty := identity.FromContext(ctx)

	userUpdate, err := parseUserUpdate(c)
	if err != nil {
		if err == model.ErrPasswordTooShort {
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		} else {
			rest.RenderError(c, http.StatusBadRequest, err)
		}
		return
	}

	// extract the token used to update the user
	if tokenStr, err := ExtractToken(c.Request); err == nil {
		keyId := jwt.GetKeyId(tokenStr)
		if _, ok := u.jwth[keyId]; !ok {
			rest.RenderInternalError(c, err)
			return
		}

		token, err := u.jwth[keyId].FromJWT(tokenStr)
		if err != nil {
			rest.RenderInternalError(c, err)
			return
		}
		userUpdate.Token = token
	}

	id := c.Param("id")
	if strings.EqualFold(id, "me") {
		id = idty.Subject
	}
	err = u.userAdm.UpdateUser(ctx, id, userUpdate)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail,
			useradm.ErrCurrentPasswordMismatch,
			useradm.ErrPassAndMailTooSimilar,
			useradm.ErrCannotModifyPassword:
			rest.RenderError(c, http.StatusUnprocessableEntity, err)
		case store.ErrUserNotFound:
			rest.RenderError(c, http.StatusNotFound, err)
		case useradm.ErrETagMismatch:
			rest.RenderError(c, http.StatusConflict, err)
		default:
			rest.RenderInternalError(c, err)
		}
	}

	c.Status(http.StatusNoContent)
}

func (u *UserAdmApiHandlers) DeleteTenantUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tenantId := c.Param("id")
	if tenantId != "" {
		ctx = getTenantContext(ctx, tenantId)
	}

	err := u.userAdm.DeleteUser(ctx, c.Param("userid"))
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (u *UserAdmApiHandlers) DeleteUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	err := u.userAdm.DeleteUser(ctx, c.Param("id"))
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUser(c *gin.Context) (*model.User, error) {
	user := model.User{}

	//decode body
	err := c.ShouldBindJSON(&user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return &user, nil
}

func parseUserInternal(c *gin.Context) (*model.UserInternal, error) {
	user := model.UserInternal{}

	//decode body
	err := c.ShouldBindJSON(&user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return &user, nil
}

func parseUserUpdate(c *gin.Context) (*model.UserUpdate, error) {
	userUpdate := model.UserUpdate{}

	//decode body
	err := c.ShouldBindJSON(&userUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if err := userUpdate.Validate(); err != nil {
		return nil, err
	}

	return &userUpdate, nil
}

func (u *UserAdmApiHandlers) CreateTenantHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var newTenant model.NewTenant

	if err := c.ShouldBindJSON(&newTenant); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	if err := newTenant.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	err := u.userAdm.CreateTenant(ctx, model.NewTenant{
		ID: newTenant.ID,
	})
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func getTenantContext(ctx context.Context, tenantId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if tenantId != "" {
		id := &identity.Identity{
			Tenant: tenantId,
		}

		ctx = identity.WithContext(ctx, id)
	}

	return ctx
}

func (u *UserAdmApiHandlers) DeleteTokensHandler(c *gin.Context) {

	ctx := c.Request.Context()

	tenantId := c.Request.URL.Query().Get("tenant_id")
	if tenantId == "" {
		rest.RenderError(c, http.StatusBadRequest, errors.New("tenant_id must be provided"))
		return
	}
	userId := c.Request.URL.Query().Get("user_id")

	err := u.userAdm.DeleteTokens(ctx, tenantId, userId)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (u *UserAdmApiHandlers) SaveSettingsHandler(c *gin.Context) {
	u.saveSettingsHandler(c, false)
}

func (u *UserAdmApiHandlers) SaveSettingsMeHandler(c *gin.Context) {
	u.saveSettingsHandler(c, true)
}

func (u *UserAdmApiHandlers) saveSettingsHandler(c *gin.Context, me bool) {
	ctx := c.Request.Context()

	settings := &model.Settings{}
	err := c.ShouldBindJSON(settings)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, errors.New("cannot parse request body as json"))
		return
	}

	if err := settings.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	ifMatchHeader := c.Request.Header.Get(hdrIfMatch)

	settings.ETag = uuid.NewString()
	if me {
		id := identity.FromContext(ctx)
		if id == nil {
			rest.RenderInternalError(c, errors.New("identity not present"))
			return
		}
		err = u.db.SaveUserSettings(ctx, id.Subject, settings, ifMatchHeader)
	} else {
		err = u.db.SaveSettings(ctx, settings, ifMatchHeader)
	}
	if err == store.ErrETagMismatch {
		rest.RenderError(c, http.StatusPreconditionFailed, err)
		return
	} else if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (u *UserAdmApiHandlers) GetSettingsHandler(c *gin.Context) {
	u.getSettingsHandler(c, false)
}

func (u *UserAdmApiHandlers) GetSettingsMeHandler(c *gin.Context) {
	u.getSettingsHandler(c, true)
}

func (u *UserAdmApiHandlers) getSettingsHandler(c *gin.Context, me bool) {
	ctx := c.Request.Context()

	var settings *model.Settings
	var err error
	if me {
		id := identity.FromContext(ctx)
		if id == nil {
			rest.RenderInternalError(c, errors.New("identity not present"))
			return
		}
		settings, err = u.db.GetUserSettings(ctx, id.Subject)
	} else {
		settings, err = u.db.GetSettings(ctx)
	}

	if err != nil {
		rest.RenderInternalError(c, err)
		return
	} else if settings == nil {
		settings = &model.Settings{
			Values: model.SettingsValues{},
		}
	}

	if settings.ETag != "" {
		c.Writer.Header().Set(hdrETag, settings.ETag)
	}
	c.JSON(http.StatusOK, settings)
}

func (u *UserAdmApiHandlers) IssueTokenHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var tokenRequest model.TokenRequest

	if err := c.ShouldBindJSON(&tokenRequest); err != nil {
		rest.RenderError(c, http.StatusBadRequest, errors.New("cannot parse request body as json"))
		return
	}
	if err := tokenRequest.Validate(u.config.TokenMaxExpSeconds); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	token, err := u.userAdm.IssuePersonalAccessToken(ctx, &tokenRequest)
	switch err {
	case nil:
		writer := c.Writer
		writer.Header().Set("Content-Type", "application/jwt")
		_, _ = writer.Write([]byte(token))
	case useradm.ErrTooManyTokens:
		rest.RenderError(c, http.StatusUnprocessableEntity, err)
	case useradm.ErrDuplicateTokenName:
		rest.RenderError(c, http.StatusConflict, err)
	default:
		rest.RenderInternalError(c, err)
	}
}

func (u *UserAdmApiHandlers) GetTokensHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id := identity.FromContext(ctx)
	if id == nil {
		rest.RenderInternalError(c, errors.New("identity not present"))
		return
	}

	tokens, err := u.userAdm.GetPersonalAccessTokens(ctx, id.Subject)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (u *UserAdmApiHandlers) DeleteTokenHandler(c *gin.Context) {
	ctx := c.Request.Context()

	err := u.userAdm.DeleteToken(ctx, c.Param("id"))
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// plans and plan binding

func (u *UserAdmApiHandlers) GetPlansHandler(c *gin.Context) {
	ctx := c.Request.Context()
	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	plans := u.userAdm.GetPlans(ctx, int((page-1)*perPage), int(perPage))
	if plans == nil {
		plans = []model.Plan{}
	}

	c.JSON(http.StatusOK, plans)
}

func (u *UserAdmApiHandlers) GetPlanBindingHandler(c *gin.Context) {
	ctx := c.Request.Context()

	planBinding, err := u.userAdm.GetPlanBinding(ctx)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, planBinding)
}
