package contenttype

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/testing/rest"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}
func TestGinMiddleware(t *testing.T) {
	json := "application/json"
	text := "text/html"

	testCases := []struct {
		Name string

		Request *http.Request
		Options []string

		ResponseCode int
	}{{
		Name: "ok, json",
		Request: rest.MakeTestRequest(&rest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
			Body:   "foo",
		}),
		ResponseCode: 200,
	}, {
		Name: "ok, no body",
		Request: rest.MakeTestRequest(&rest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
		}),
		ResponseCode: 200,
	}, {
		Name: "ok, text/html",
		Request: rest.MakeTestRequest(&rest.TestRequest{
			Method:      http.MethodPost,
			Path:        "http://localhost/test",
			ContentType: text,
		}),
		Options:      []string{text},
		ResponseCode: 200,
	}, {
		Name: "ok, text/html or json",
		Request: rest.MakeTestRequest(&rest.TestRequest{
			Method:      http.MethodPost,
			Path:        "http://localhost/test",
			ContentType: text,
		}),
		Options:      []string{text, json},
		ResponseCode: 200,
	}, {
		Name: "error, conflicting Content-Type",
		Request: rest.MakeTestRequest(&rest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
			Body:   "foo",
		}),
		Options:      []string{text},
		ResponseCode: 415,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			router := gin.New()
			if tc.Options != nil {
				router.Use(Middleware(tc.Options...))
			} else {
				router.Use(CheckJSON())
			}
			router.POST("/test", func(c *gin.Context) {
				c.Status(200)
			})

			w := httptest.NewRecorder()

			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.ResponseCode, w.Code)
		})
	}
}
