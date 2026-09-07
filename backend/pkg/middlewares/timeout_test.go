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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handler completes and writes a response", func(t *testing.T) {
		engine := gin.New()
		engine.Use(Timeout(time.Second))
		engine.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("handler exceeds the deadline and observes context cancellation", func(t *testing.T) {
		engine := gin.New()
		engine.Use(Timeout(10 * time.Millisecond))
		handlerReturned := make(chan struct{})
		engine.GET("/test", func(c *gin.Context) {
			defer close(handlerReturned)
			select {
			case <-c.Request.Context().Done():
				return
			case <-time.After(time.Second):
				c.String(http.StatusOK, "too slow")
			}
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		engine.ServeHTTP(w, req)

		select {
		case <-handlerReturned:
		case <-time.After(time.Second):
			t.Fatal("handler did not observe context cancellation in time")
		}

		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "request deadline exceeded", body["error"])
	})

	t.Run("request context is given a deadline", func(t *testing.T) {
		const duration = 50 * time.Millisecond
		engine := gin.New()
		engine.Use(Timeout(duration))

		var (
			deadline time.Time
			ok       bool
		)
		start := time.Now()
		engine.GET("/test", func(c *gin.Context) {
			deadline, ok = c.Request.Context().Deadline()
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		engine.ServeHTTP(w, req)

		require.True(t, ok, "expected request context to carry a deadline")
		assert.WithinDuration(t, start.Add(duration), deadline, 20*time.Millisecond)
	})
}
