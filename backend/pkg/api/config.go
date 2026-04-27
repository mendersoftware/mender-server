package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/requestid"
)

func NewDefaultClientConfiguration() *client.Configuration {
	cfg := client.NewConfiguration()
	cfg.Middleware = defaultClientMiddleware
	return cfg
}

func NewDefaultClientConfigurationFromURL(u string) (*client.Configuration, error) {
	cfg := NewDefaultClientConfiguration()
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
