package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/tests/runner/client"
)

func init() {
	AddTestCase("test_login", mainTestRemove)
}

func mainTestRemove(t *testing.T, settings *TestSettings) error {
	t.Logf("login test starting\n")
	ctx := context.Background()
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	c, err := client.NewClientWithResponses(settings.ServerURL, client.WithHTTPClient(httpClient), client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+settings.jwt)
		return nil
	}))
	r, err := c.RemoveUser(ctx, "id")
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 204, r.StatusCode)

	return nil
}
