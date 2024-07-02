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

package http

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
)

const (
	defaultTimeout = time.Second * 10
)

// StatusController contains status-related end-points
type StatusController struct {
	app                     app.App
	gracefulShutdownTimeout time.Duration
}

// NewStatusController returns a new StatusController
func NewStatusController(app app.App, gracefulShutdownTimeout time.Duration) *StatusController {
	return &StatusController{app: app, gracefulShutdownTimeout: gracefulShutdownTimeout}
}

// Alive responds to GET /alive
func (h StatusController) Alive(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Health responds to GET /health
func (h StatusController) Health(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := h.app.HealthCheck(ctx)
	if err != nil {
		l.Error(errors.Wrap(err, "health check failed"))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

// Shutdown responds to GET /shutdown
func (h StatusController) Shutdown(c *gin.Context) {
	pid := os.Getpid()
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		ctx := c.Request.Context()
		log.FromContext(ctx).Error(errors.Wrap(err, "shutdown failed"))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}
	time.Sleep(h.gracefulShutdownTimeout)
	c.Writer.WriteHeader(http.StatusAccepted)
}
