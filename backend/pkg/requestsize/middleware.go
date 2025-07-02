package requestsize

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func Middleware(size int64) gin.HandlerFunc {
	if size < 0 {
		panic("requestsize Middleware initialized with negativ size")
	}
	tooLargeErr := http.MaxBytesError{
		Limit: size,
	}
	return func(c *gin.Context) {
		// if the Content-Length in header and is greater than size limit
		if c.Request.ContentLength > int64(size) {
			rest.RenderError(c, http.StatusRequestEntityTooLarge, &tooLargeErr)
			c.Abort()
			return
		}
		// wrapping a nil body with MaxBytesReader causes a nil pointer dereference panic
		// when calling c.ShouldBindJSON()
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, size)
		}

		c.Next()
	}
}
