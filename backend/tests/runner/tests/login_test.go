package tests

import (
	"context"
	"crypto/tls"
	b64 "encoding/base64"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/tests/runner/client"
)

func init() {
	AddTestCase("0_test_login", mainTestLogin)
}

func mainTestLogin(t *testing.T, settings *TestSettings) error {
	t.Logf("login test starting\n")
	ctx := context.Background()
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	c, err := client.NewClientWithResponses(settings.ServerURL, client.WithHTTPClient(httpClient))
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	basicAuth := b64.StdEncoding.EncodeToString([]byte(settings.Username + ":" + settings.Password))
	l, err := c.LoginWithResponse(
		ctx,
		client.LoginJSONRequestBody{},
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Basic "+basicAuth)
			return nil
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.NotNil(t, l.HTTPResponse)
	assert.Equal(t, 200, l.HTTPResponse.StatusCode)

	jwt := string(l.Body)
	assert.True(t, len(jwt) > 0)
	settings.jwt = jwt

	m, err := c.ShowOwnUserDataWithResponse(
		ctx,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+jwt)
			return nil
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 200, m.HTTPResponse.StatusCode)
	assert.Equal(t, settings.Username, m.JSON200.Email)

	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	settings.client, err = client.NewClientWithResponses(
		settings.ServerURL,
		client.WithHTTPClient(httpClient),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+settings.jwt)
			return nil
		}),
	)

	t.Logf("test passed with user data: %+v\n", m.JSON200)
	return nil
}
