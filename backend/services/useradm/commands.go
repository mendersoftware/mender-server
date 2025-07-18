// Copyright 2022 Northern.tech AS
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
	"fmt"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/term"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/useradm/client/tenant"
	. "github.com/mendersoftware/mender-server/services/useradm/config"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/store/mongo"
	useradm "github.com/mendersoftware/mender-server/services/useradm/user"
)

// safeReadPassword reads a user password from a terminal in a safe way (without
// echoing the characters input by the user)
func safeReadPassword() (string, error) {
	stdinfd := int(os.Stdin.Fd())

	if !term.IsTerminal(stdinfd) {
		return "", errors.New("stdin is not a terminal")
	}

	fmt.Fprintf(os.Stderr, "Enter password: ")
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", errors.Wrap(err, "failed to read password")
	}
	fmt.Fprintf(os.Stderr, "\n")

	return string(raw), nil
}

func commandCreateUser(
	c config.Reader,
	username model.Email,
	password, userId, tenantId string,
) error {
	ctx := context.Background()
	l := log.NewEmpty()

	l.Debugf("create user '%s'", username)

	if password == "" {
		var err error
		if password, err = safeReadPassword(); err != nil {
			return err
		}
	}

	u := model.UserInternal{
		User: model.User{
			Email:    username,
			Password: password,
		},
	}

	if userId != "" {
		u.ID = userId
	}

	if err := u.Validate(); err != nil {
		return errors.Wrap(err, "user validation failed")
	}

	db, err := mongo.GetDataStoreMongo(dataStoreMongoConfigFromAppConfig(c))
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	ua := useradm.NewUserAdm(nil, db, useradm.Config{})
	if tadmAddr := c.GetString(SettingTenantAdmAddr); tadmAddr != "" {
		l.Infof("setting up tenant verification")

		tc := tenant.NewClient(tenant.Config{
			TenantAdmAddr: tadmAddr,
		})

		ua = ua.WithTenantVerification(tc)
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantId,
		})
	}

	if err := ua.CreateUserInternal(ctx, &u); err != nil {
		return errors.Wrap(err, "creating user failed")
	}

	fmt.Println(u.ID)

	return nil
}

func getTenantContext(tenantId string) context.Context {
	ctx := context.Background()
	if tenantId != "" {
		id := &identity.Identity{
			Tenant: tenantId,
		}

		ctx = identity.WithContext(ctx, id)
	}

	return ctx
}

func commandMigrate(c config.Reader, tenantId string) error {
	l := log.New(log.Ctx{})

	l.Printf("User Administration Service starting up")

	if tenantId != "" {
		l.Printf("migrating tenant %v", tenantId)
	} else {
		l.Printf("migrating all the tenants")
	}

	db, err := mongo.NewDataStoreMongo(dataStoreMongoConfigFromAppConfig(c))

	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	// we want to apply migrations
	db = db.WithAutomigrate()

	ctx := context.Background()

	if tenantId != "" {
		err = db.MigrateTenant(ctx, mongo.DbVersion, tenantId)
	} else {
		err = db.Migrate(ctx, mongo.DbVersion)
	}
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil

}

func commandSetPassword(
	c config.Reader,
	username model.Email,
	password, tenantId string,
) error {
	l := log.NewEmpty()

	l.Debugf("set password for '%s'", username)

	if password == "" {
		var err error
		if password, err = safeReadPassword(); err != nil {
			return err
		}
	}

	db, err := mongo.GetDataStoreMongo(dataStoreMongoConfigFromAppConfig(c))
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	ua := useradm.NewUserAdm(nil, db, useradm.Config{})

	u := model.User{
		Email:    username,
		Password: password,
	}

	if err := u.Validate(); err != nil {
		return errors.Wrap(err, "user validation failed")
	}

	ctx := getTenantContext(tenantId)

	uu := model.UserUpdate{
		Email:    username,
		Password: password,
	}

	if err := ua.SetPassword(ctx, uu); err != nil {
		return errors.Wrap(err, "setting password failed")
	}

	return nil
}
