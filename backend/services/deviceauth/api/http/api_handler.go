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
	"net/http"

	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
)

// thin API handler interface
type ApiHandler interface {
	// produce a rest.App with routing setup or an error
	Build() (http.Handler, error)
}

type InternalAPI struct {
	App devauth.App
	Db  store.DataStore
}

// NewInternalHandler returns a new InternalAPI
func NewInternalHandler(DevAuth devauth.App, db store.DataStore) *InternalAPI {
	return &InternalAPI{
		App: DevAuth,
		Db:  db,
	}
}

type ManagementAPI struct {
	App devauth.App
	Db  store.DataStore
}

// NewManagementHandler returns a new ManagementAPI
func NewManagementHandler(DevAuth devauth.App, db store.DataStore) *ManagementAPI {
	return &ManagementAPI{
		App: DevAuth,
		Db:  db,
	}
}
