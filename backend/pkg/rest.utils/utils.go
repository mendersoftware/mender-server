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

	"github.com/mendersoftware/mender-server/pkg/requestid"
)

func RenderError(c *gin.Context, code int, err error) {
	ctx := c.Request.Context()
	_ = c.Error(err)
	err = &Error{
		Err:       err.Error(),
		RequestID: requestid.FromContext(ctx),
	}
	c.JSON(code, err)
}

func RenderErrorWithMessage(c *gin.Context, code int, err error, apiMessage string) {
	ctx := c.Request.Context()
	_ = c.Error(err)
	err = &Error{
		Err:       apiMessage,
		RequestID: requestid.FromContext(ctx),
	}
	c.JSON(code, err)
}

func RenderInternalError(c *gin.Context, err error) {
	msg := "internal error"
	if err != nil {
		RenderErrorWithMessage(c,
			http.StatusInternalServerError,
			err, msg,
		)
		return
	}
	RenderError(c,
		http.StatusInternalServerError,
		errors.New(msg),
	)
}
