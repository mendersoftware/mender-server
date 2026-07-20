//nolint:all // This is all test code
package common

import (
	"bytes"
	"context"
	"net/http"

	"github.com/mendersoftware/mender-server/pkg/api/client"
)

// RawRequest performs a request the generated client can't express (e.g.
// malformed JSON bodies for negative tests, or query parameters the client
// doesn't expose), authenticated with whatever bearer token is embedded in
// ctx (see JWTAuthContext).
func RawRequest(
	ctx context.Context, api *client.APIClient, method, path string, body []byte,
) (*http.Response, error) {
	cfg := api.GetConfig()

	// The generated client resolves the request URL from cfg.Host/cfg.Scheme
	// (see prepareRequest in pkg/api/client/client.go), overriding whatever
	// placeholder cfg.Servers/ServerURLWithContext would otherwise return.
	// Replicate that here so raw requests hit the same host.
	scheme := cfg.Scheme
	if scheme == "" {
		scheme = "https"
	}

	// bytes.NewReader(nil) is a valid, empty io.Reader (not a typed-nil
	// interface value), so this works whether or not body is nil.
	req, err := http.NewRequestWithContext(ctx, method, scheme+"://"+cfg.Host+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if jwt, ok := ctx.Value(client.ContextAccessToken).(string); ok {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return httpClient.Do(req)
}
