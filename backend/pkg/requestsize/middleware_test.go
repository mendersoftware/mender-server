package requestsize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	"github.com/stretchr/testify/assert"
)

func TestByteSizeToUnit(t *testing.T) {
	testCases := map[string]struct {
		Size     int
		Expected string
	}{
		"0 size": {
			Size:     0,
			Expected: "0 B",
		},
		"10 KiB": {
			Size:     10 * KiB,
			Expected: "10 KiB",
		},
		"1 MiB": {
			Size:     1 * MiB,
			Expected: "1 MiB",
		},
		"large number": {
			Size:     GiB * MiB, // 1 Pebibyte (PiB)
			Expected: "1024 TiB",
		},
		"negative number": {
			Size:     -1 * KiB,
			Expected: "-1024 B",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, byteSizeToUnit(tc.Size))
		})
	}

}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

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
