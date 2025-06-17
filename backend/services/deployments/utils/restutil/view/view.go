// Copyright 2021 Northern.tech AS
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

package view

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// Headers
const (
	HttpHeaderLocation = "Location"
)

// Errors
var (
	ErrNotFound = errors.New("Resource not found")
)

type RESTView struct {
}

func (p *RESTView) RenderSuccessPost(c *gin.Context, id string) {
	c.Writer.Header().Add(
		HttpHeaderLocation,
		fmt.Sprintf("%s/%s", c.Request.URL.Path, id),
	)
	c.Status(http.StatusCreated)
}

func (p *RESTView) RenderSuccessGet(c *gin.Context, object interface{}) {
	c.JSON(http.StatusOK, object)
}

func (p *RESTView) RenderError(
	c *gin.Context,
	err error,
	status int,
) {
	rest.RenderError(c, status, err)
}

func (p *RESTView) RenderInternalError(
	c *gin.Context,
	err error,
) {
	rest.RenderInternalError(c, err)
}

func (p *RESTView) RenderErrorNotFound(c *gin.Context) {
	rest.RenderError(c, http.StatusNotFound, ErrNotFound)
}

func (p *RESTView) RenderSuccessDelete(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (p *RESTView) RenderSuccessPut(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (p *RESTView) RenderNoUpdateForDevice(c *gin.Context) {
	p.RenderEmptySuccessResponse(c)
}

// Success response with no data aka. 204 No Content
func (p *RESTView) RenderEmptySuccessResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (p *RESTView) RenderDeploymentLog(c *gin.Context, dlog model.DeploymentLog) {
	h := c.Writer

	h.Header().Set("Content-Type", "text/plain")
	h.WriteHeader(http.StatusOK)

	for _, m := range dlog.Messages {
		as := m.String()
		_, _ = h.Write([]byte(as))
		if !strings.HasSuffix(as, "\n") {
			_, _ = h.Write([]byte("\n"))
		}
	}
}
