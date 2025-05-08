package contenttype

import (
	"mime"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func CheckJSON() gin.HandlerFunc {
	return Middleware("application/json")
}

func Middleware(contentTypes ...string) gin.HandlerFunc {
	if len(contentTypes) == 0 {
		panic("contenttype.Middleware initialized with no parameters")
	}

	errMsg := "Bad Content-Type, expected "
	for i, ctype := range contentTypes {
		if i > 0 {
			errMsg += ", "
		}
		errMsg += "'" + ctype + "'"
	}

	return func(c *gin.Context) {
		if c.Request.ContentLength > 0 {
			contentType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if err != nil {
				rest.RenderError(c,
					http.StatusBadRequest,
					errors.WithMessage(err, "Invalid value for Content-Type header"),
				)
				c.Abort()
				return
			}

			if !slices.Contains(contentTypes, contentType) {
				rest.RenderError(c,
					http.StatusUnsupportedMediaType,
					errors.New(errMsg),
				)
				c.Abort()
			}
		}
	}
}
