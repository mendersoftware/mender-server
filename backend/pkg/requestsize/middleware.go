package requestsize

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

const (
	KiB = 1024       // 1 KiB
	MiB = KiB * 1024 // 1 MiB
	GiB = MiB * 1024 // 1 GiB
)

func Middleware(size int64) gin.HandlerFunc {
	if size < 0 {
		panic("requestsize Middleware initialized with negativ size")
	}

	return func(c *gin.Context) {
		// if the Content-Length in header and is greater than size limit
		if c.Request.ContentLength > int64(size) {
			rest.RenderError(c, http.StatusRequestEntityTooLarge, rest.ErrTooLarge)
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, size)

		c.Next()
	}
}

func CheckBodySize(c *gin.Context) error {
	if err := c.Request.ParseForm(); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, maxBytesError) {
			return rest.ErrTooLarge
		}
		return err
	}
	return nil
}

// Size = 1 MiB
func MiB1() gin.HandlerFunc {
	return Middleware(MiB)
}
