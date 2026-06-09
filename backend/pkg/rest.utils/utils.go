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

package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
)

func RenderError(c *gin.Context, code int, err error) {
	RenderErrorWithMessage(c, code, err, err.Error())
}

func RenderErrorWithMessage(c *gin.Context, code int, err error, apiMessage string) {
	// Wrap the private function to give equal call depth.
	renderErrorWithMessage(c, code, err, apiMessage)
}

func renderErrorWithMessage(c *gin.Context, code int, err error, apiMessage string) {
	ctx := c.Request.Context()
	_ = c.Error(err)
	err = &Error{
		Err:       apiMessage,
		RequestID: requestid.FromContext(ctx),
	}
	if c.Writer.Written() {
		// Skip 4
		// = rest.Render*
		// + rest.renderErrorWithMessage
		// + log.CollectTrace
		// + runtime.Callers
		log.FromContext(ctx).
			WithField("trace", log.CollectTrace(4)).
			Error("response already written")
	} else {
		c.JSON(code, err)
	}
}

func RenderInternalError(c *gin.Context, err error) {
	const msg = "internal error"
	if err == nil {
		err = errors.New(msg)
	}
	RenderErrorWithMessage(c,
		http.StatusInternalServerError,
		err, msg,
	)
}

func RenderUnavailable(c *gin.Context, err error) {
	const msg = "service unavailable"
	if err == nil {
		err = errors.New(msg)
	}
	RenderErrorWithMessage(c,
		http.StatusServiceUnavailable,
		err, msg,
	)
}
