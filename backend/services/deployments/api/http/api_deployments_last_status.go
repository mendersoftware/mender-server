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

	"github.com/pkg/errors"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// device deployments last status handler
func (d *DeploymentsApiHandlers) GetDeviceDeploymentLastStatus(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	l := requestlog.GetRequestLogger(r)

	l.Debugf("starting")

	tenantId := r.PathParam("tenant")
	var req model.DeviceDeploymentLastStatusReq
	if err := r.DecodeJsonPayload(&req); err != nil {
		l.Errorf("error during DecodeJsonPayload: %s.", err.Error())
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.Wrap(err, "cannot parse device ids array"),
			http.StatusBadRequest,
		)
		return
	} else if len(req.DeviceIds) == 0 {
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			errors.Wrap(err, "device ids array cannot be empty"),
			http.StatusBadRequest,
		)
	}

	l.Debugf("querying %d devices ids", len(req.DeviceIds))
	ctx := r.Context()
	if tenantId != "" {
		ctx = identity.WithContext(
			ctx,
			&identity.Identity{
				Tenant: tenantId,
			},
		)
	}
	lastDeployments, err := d.app.GetDeviceDeploymentLastStatus(ctx, req.DeviceIds)
	switch err {
	default:
		d.view.RenderInternalError(w, r, err, l)
	case nil:
		l.Infof("outputting: %+v", lastDeployments)
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(lastDeployments)
	}
}
