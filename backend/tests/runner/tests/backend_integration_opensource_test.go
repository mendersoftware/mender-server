package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/google/uuid"
)

func getDockerComposeFilesOpenSource() []string {
	return []string{"../../../../docker-compose.yml"}
}

func seedComposeEnvironmentOpenSource(ctx context.Context, compose api.Compose, project *types.Project) ([]Tenant, error) {
	username := fmt.Sprintf("user-%s@docker.mender.io", uuid.New().String())
	password := uuid.New().String()
	exitCode, err := compose.Exec(ctx, project.Name, api.RunOptions{
		Service: "useradm",
		Command: []string{"useradm", "create-user", "--username", username, "--password", password},
	})

	if exitCode != 0 {
		var statusErr cli.StatusError
		if errors.As(err, &statusErr) {
			return nil, fmt.Errorf("unexpected exit code '%d' when creating initial user: %w", statusErr.StatusCode, statusErr.Cause)
		}
		return nil, fmt.Errorf("unexpected exit code '%d' when creating initial user", exitCode)
	}

	return []Tenant{
		{
			ID:              "",    // Irrelevant for opensource
			ServiceProvider: false, // Irrelevant for opensource
			Users:           []User{{Username: username, Password: password}},
		},
	}, nil

}

func getHealthChecksOpenSource() []healthCheck {
	return []healthCheck{
		{name: "deployments", url: "http://deployments:8080/api/internal/v1/deployments/health", code: http.StatusNoContent},
		{name: "deviceauth", url: "http://deviceauth:8080/api/internal/v1/devauth/health", code: http.StatusNoContent},
		{name: "deviceconfig", url: "http://deviceconfig:8080/api/internal/v1/deviceconfig/health", code: http.StatusNoContent},
		{name: "deviceconnect", url: "http://deviceconnect:8080/api/internal/v1/deviceconnect/health", code: http.StatusNoContent},
		{name: "inventory", url: "http://inventory:8080/api/internal/v1/inventory/health", code: http.StatusNoContent},
		{name: "iot-manager", url: "http://iot-manager:8080/api/internal/v1/iot-manager/health", code: http.StatusNoContent},
		{name: "useradm", url: "http://useradm:8080/api/internal/v1/useradm/health", code: http.StatusNoContent},
		{name: "workflows", url: "http://workflows:8080/api/v1/health", code: http.StatusNoContent},

		// Sometimes, the ingress routing rules in traefik takes a bit longer to register, so we add another
		// health check from outside the ingress to avoid spurious test failures.
		{name: "useradm-ingress", url: "https://traefik:443/api/management/v1/useradm/users", code: http.StatusUnauthorized},
	}
}

func isOpenSource() bool {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}

	return info.Main.Path == "github.com/mendersoftware/mender-server"
}
