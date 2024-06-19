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

package accesslog

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type AccessLogger struct {
	DisableLog func(c *gin.Context) bool
}

func (a AccessLogger) LogFunc(
	ctx context.Context,
	c *gin.Context,
	startTime time.Time,
) {
	logCtx := logrus.Fields{
		"clientip": c.ClientIP(),
		"method":   c.Request.Method,
		"path":     c.Request.URL.Path,
		"qs":       c.Request.URL.RawQuery,
		"ts": startTime.
			Truncate(time.Millisecond).
			Format(time.RFC3339Nano),
		"type":      c.Request.Proto,
		"useragent": c.Request.UserAgent(),
	}
	lc := fromContext(ctx)
	if lc != nil {
		lc.addFields(logCtx)
	}
	if r := recover(); r != nil {
		trace := collectTrace()
		logCtx["trace"] = trace
		logCtx["panic"] = r

		func() {
			// Try to respond with an internal server error.
			// If the connection is broken it might panic again.
			defer func() { recover() }() // nolint:errcheck
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New("internal error"),
			)
		}()
	} else if a.DisableLog != nil && a.DisableLog(c) {
		return
	}
	latency := time.Since(startTime)
	// We do not need more than 3 digit fraction
	if latency > time.Second {
		latency = latency.Round(time.Millisecond)
	} else if latency > time.Millisecond {
		latency = latency.Round(time.Microsecond)
	}
	code := c.Writer.Status()
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			code = StatusClientClosedConnection
		}
	default:
	}
	logCtx["responsetime"] = latency.String()
	logCtx["status"] = c.Writer.Status()
	logCtx["byteswritten"] = c.Writer.Size()

	var logLevel logrus.Level = logrus.InfoLevel
	if code >= 500 {
		logLevel = logrus.ErrorLevel
	} else if code >= 400 {
		logLevel = logrus.WarnLevel
	}
	if len(c.Errors) > 0 {
		errs := c.Errors.Errors()
		var errMsg string
		if len(errs) == 1 {
			errMsg = errs[0]
		} else {
			for i, err := range errs {
				errMsg = errMsg + fmt.Sprintf(
					"#%02d: %s\n", i+1, err,
				)
			}
		}
		logCtx["error"] = errMsg
	}
	log.FromContext(c.Request.Context()).
		WithFields(logCtx).
		Log(logLevel)
}

func (a AccessLogger) Middleware(c *gin.Context) {
	ctx := c.Request.Context()
	startTime := time.Now()
	ctx = withContext(ctx, &logContext{maxErrors: DefaultMaxErrors})
	c.Request = c.Request.WithContext(ctx)
	defer a.LogFunc(ctx, c, startTime)
	c.Next()
}

// Middleware provides accesslog middleware for the gin-gonic framework.
// This middleware will recover any panic from occurring in the API
// handler and log it to error level with panic and trace showing the panic
// message and traceback respectively.
// If an error status is returned in the response, the middleware tries
// to pop the topmost error from the gin.Context (c.Error) and puts it in
// the "error" context to the final log entry.
func Middleware() gin.HandlerFunc {
	return AccessLogger{}.Middleware
}
