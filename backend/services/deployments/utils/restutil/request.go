package restutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mt "github.com/mendersoftware/mender-server/pkg/testing"
)

func RunRequest(t *testing.T,
	handler http.Handler,
	request *http.Request) *mt.Recorded {

	request.Header.Set("X-MEN-RequestID", "test")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	return &mt.Recorded{T: t, Recorder: w}
}
