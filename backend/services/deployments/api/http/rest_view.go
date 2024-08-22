// Copyright 2019 Northern.tech AS
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

package http

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type RESTView interface {
	RenderSuccessGet(w rest.ResponseWriter, object interface{})
	RenderError(w rest.ResponseWriter, r *rest.Request, err error, status int, l *log.Logger)
	RenderInternalError(w rest.ResponseWriter, r *rest.Request, err error, l *log.Logger)
	RenderNoUpdateForDevice(w rest.ResponseWriter)
	RenderSuccessPost(w rest.ResponseWriter, r *rest.Request, id string)
	RenderEmptySuccessResponse(w rest.ResponseWriter)
	RenderErrorNotFound(w rest.ResponseWriter, r *rest.Request, l *log.Logger)
	RenderDeploymentLog(w rest.ResponseWriter, dlog model.DeploymentLog)
	RenderSuccessDelete(w rest.ResponseWriter)
	RenderSuccessPut(w rest.ResponseWriter)
}
