// Copyright 2026 Northern.tech AS
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

package middlewares

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func renderErrorWithMessage(c *gin.Context, code int, err error, apiMessage string) {
	ctx := c.Request.Context()
	c.JSON(code, &rest.Error{
		Err:       apiMessage,
		RequestID: requestid.FromContext(ctx),
	})
}

func renderError(c *gin.Context, code int, err error) {
	renderErrorWithMessage(c, code, err, err.Error())
}

// ErrorHandler handles common errors if not already handled by the API handler.
func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if ctx.Writer.Written() {
			return
		}
		err := ctx.Errors.Last()
		if err == nil {
			return // Fallback to default (200 OK)
		}
		var (
			netErr        net.Error
			validationErr validation.Errors
			maxBytesErr   *http.MaxBytesError
		)
		switch {
		case errors.As(err, &netErr) && netErr.Timeout():
			// NOTE: context.DeadlineExceeded is also caught here
			renderErrorWithMessage(ctx, http.StatusGatewayTimeout, err, "gateway timeout")
		case errors.As(err, &maxBytesErr):
			renderError(ctx, http.StatusRequestEntityTooLarge, err)
		case err.IsType(gin.ErrorTypeBind), errors.As(err, &validationErr):
			renderError(ctx, http.StatusBadRequest, err)
		case errors.Is(err, context.Canceled):
			ctx.Writer.WriteHeader(499)
		default:
			ctx.JSON(http.StatusInternalServerError, rest.Error{
				Err:       "internal error",
				RequestID: requestid.FromContext(ctx.Request.Context()),
			})
		}
	}
}
