package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/requestid"
)

func NewDefaultClientConfiguration() *client.Configuration {
	cfg := client.NewConfiguration()
	cfg.Middleware = defaultClientMiddleware
	return cfg
}

func NewDefaultClientConfigurationFromURL(u string) (*client.Configuration, error) {
	if len(u) < 1 {
		return nil, fmt.Errorf("url is empty")
	}

	u = strings.TrimSuffix(u, "/")

	cfg := NewDefaultClientConfiguration()
	if !strings.HasPrefix(u, "http") {
		// If the url doesn't include a scheme, default to http
		u = fmt.Sprintf("http://%s", u)
	}

	url, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	cfg.Host = url.Host
	cfg.Scheme = url.Scheme
	cfg.Servers = client.ServerConfigurations{{
		URL: url.String(),
	}}
	return cfg, nil
}

func defaultClientMiddleware(r *http.Request) {
	if id := requestid.FromContext(r.Context()); id != "" {
		requestid.SetRequestIDHeader(r, id)
	}
}
