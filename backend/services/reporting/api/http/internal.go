// Copyright 2022 Northern.tech AS
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
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/reporting/app/reporting"
)

// InternalController contains internal end-points
type InternalController struct {
	reporting reporting.App
}

// NewInternalController returns a new InternalController
func NewInternalController(r reporting.App) *InternalController {
	return &InternalController{
		reporting: r,
	}
}

// Alive responds to GET /health/alive
func (h InternalController) Alive(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h InternalController) Health(c *gin.Context) {
	err := h.reporting.HealthCheck(c.Request.Context())
	if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}
