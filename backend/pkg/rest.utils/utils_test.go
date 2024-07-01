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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRenderError(t *testing.T) {

	engine := gin.New()
	engine.GET("/test", func(c *gin.Context) {
		err := errors.New("test error")
		RenderError(c, http.StatusInternalServerError, err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/test", nil)
	engine.ServeHTTP(w, req)

	apiErr := Error{}
	_ = json.Unmarshal(w.Body.Bytes(), &apiErr)
	assert.EqualError(t, apiErr, "test error")
}
