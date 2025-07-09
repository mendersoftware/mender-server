package requestsize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

const (
	KiB = 1024       // 1 KiB
	MiB = KiB * 1024 // 1 MiB
)

func createTestBody(size int) *[]byte {
	slice := make([]byte, size)
	for i := range size {
		slice[i] = byte(1)
	}
	return &slice
}

func TestRequestSizeMiddleware(t *testing.T) {

	testCases := []struct {
		Name string

		Size    int64
		Request *http.Request

		ResponseCode int
	}{{
		Name: "ok, 0 B limit ",
		Size: 0,
		Request: rtest.MakeTestRequest(&rtest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
			Body:   nil,
		}),
		ResponseCode: 200,
	}, {
		Name: "ok, size smaller than limit  ",
		Size: KiB,
		Request: rtest.MakeTestRequest(&rtest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
			Body:   createTestBody(10),
		}),
		ResponseCode: 200,
	}, {
		Name: "error, bigger than the limit",
		Size: KiB,
		Request: rtest.MakeTestRequest(&rtest.TestRequest{
			Method: http.MethodPost,
			Path:   "http://localhost/test",
			Body:   createTestBody(MiB),
		}),
		ResponseCode: 413,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			router := gin.New()

			router.Use(Middleware(tc.Size))

			router.POST("/test", func(c *gin.Context) {
				c.Status(200)
			})

			w := httptest.NewRecorder()

			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.ResponseCode, w.Code)
		})
	}
}
