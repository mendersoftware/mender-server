//nolint:all // This is all test code so we don't care
package common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/cmd/formatter"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	"github.com/google/uuid"
	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

const (
	localhost = "127.0.0.1"
)

type Config struct {
	ProjectName                 string
	SkipCleanup                 bool
	SkipCleanupOnFailure        bool
	PrintServiceLogsOnFailure   string
	PrintServiceStatusOnFailure string
}

type ComposeEnvironment struct {
	config Config

	serverURL string
	apiClient *oapiclient.APIClient
	project   *types.Project
}

func NewComposeEnvironment(config Config) *ComposeEnvironment {
	return &ComposeEnvironment{
		config: config,
	}
}

func (c *ComposeEnvironment) Setup(t *testing.T) {
	var (
		ctx     = t.Context()
		require = require.New(t)
	)

	t.Logf("Creating docker compose environment")

	compose, err := createComposeService(io.Discard)
	require.NoError(err, "failed to create compose service")

	project, cleanup, err := loadComposeProject(ctx, compose, c.config.ProjectName)
	require.NoError(err, "failed to load compose project")

	c.project = project
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	stacks, err := compose.List(ctx, api.ListOptions{All: true})
	require.NoError(err, "")

	projectAlreadyRunning := slices.ContainsFunc(stacks, func(s api.Stack) bool {
		return s.Name == c.config.ProjectName
	})
	if !projectAlreadyRunning {
		err = compose.Up(
			ctx,
			c.project,
			api.UpOptions{
				Create: api.CreateOptions{RemoveOrphans: true, Build: &api.BuildOptions{}},
				Start:  api.StartOptions{Wait: true},
			},
		)
		require.NoError(err, "failed to bring up docker compose environment")
	}

	err = healthCheckComposeEnvironment(ctx, compose, c.project)
	require.NoError(err, "failed to health check docker compose environment")

	t.Logf("Successfully created docker compose environment")

	c.serverURL = "traefik"
	if hostname := os.Getenv("MENDER_HOSTNAME"); hostname != "" {
		c.serverURL = hostname
	}

	// Simple workaround to "resolve" DNS without an actual DNS server
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// The server URL can't be resolved through DNS so we
		// resolve it manually here. This is essentially the same
		// as adding `[localhost] [host]` to `/etc/hosts`.
		if host == c.serverURL {
			host = localhost
		}

		return dialer.DialContext(
			ctx,
			network,
			net.JoinHostPort(host, port),
		)
	}
	c.apiClient = createMenderAPIClient(c.serverURL, dialContext)
}

func (c *ComposeEnvironment) APIClient() *oapiclient.APIClient {
	return c.apiClient
}

func (c *ComposeEnvironment) Teardown(t *testing.T) {
	var (
		config = c.config
		failed = t.Failed()
	)

	if config.SkipCleanup || (config.SkipCleanupOnFailure && failed) {
		return
	}

	if c.project == nil {
		// Nothing below here can be done if we're not managing our own compose environment
		return
	}

	var (
		ctx     = t.Context()
		output  = t.Output()
		require = require.New(t)
		name    = c.project.Name
	)

	compose, err := createComposeService(io.Discard)
	require.NoError(err)

	if config.PrintServiceStatusOnFailure != "" && failed {
		t.Logf("Printing service status for services %s in project: %s", config.PrintServiceStatusOnFailure, name)

		containers, err := compose.Ps(ctx, name, api.PsOptions{All: true})
		require.NoError(err, "failed to list docker compose containers")

		if config.PrintServiceStatusOnFailure != "*" {
			var (
				idx      = 0
				services = strings.Split(config.PrintServiceStatusOnFailure, ",")
			)

			for _, c := range containers {
				if !slices.Contains(services, c.Service) {
					continue
				}
				containers[idx] = c
				idx++
			}
			containers = containers[:idx]
		}

		apiClient, err := client.New(client.FromEnv)
		require.NoError(err, "failed to create docker api client")
		defer apiClient.Close()

		var inspect []container.InspectResponse
		for _, container := range containers {
			c, err := apiClient.ContainerInspect(ctx, container.ID, client.ContainerInspectOptions{})
			require.NoError(err)
			inspect = append(inspect, c.Container)
		}

		b, err := json.MarshalIndent(inspect, "", "  ")
		output.Write(b)
		require.NoError(err)
	}

	if config.PrintServiceLogsOnFailure != "" && failed {
		t.Logf("Printing service logs for services %s in project: %s", config.PrintServiceLogsOnFailure, name)
		var services []string
		if config.PrintServiceLogsOnFailure != "*" {
			services = strings.Split(config.PrintServiceLogsOnFailure, ",")
		}

		err := compose.Logs(
			ctx,
			name,
			formatter.NewLogConsumer(ctx, output, output, false, true, true),
			api.LogOptions{Services: services},
		)

		if err != nil {
			t.Logf("failed to gather container logs: %s", err)
		}
	}

	t.Logf("Tearing down project: %s", name)
	err = compose.Down(ctx, name, api.DownOptions{
		// Remove volume(s) so the tests can run from a clean
		// state every time
		Volumes: true,
	})
	require.NoError(err, "failed to tear down compose environment")
}

type User struct {
	Username string
	Password string
}

func (i *ComposeEnvironment) User(ctx context.Context) (User, error) {
	compose, err := createComposeService(io.Discard)
	if err != nil {
		return User{}, err
	}

	username := fmt.Sprintf("user-%s@docker.mender.io", uuid.New().String())
	password := uuid.New().String()
	exitCode, err := compose.Exec(ctx, i.config.ProjectName, api.RunOptions{
		Service: "useradm",
		Command: []string{"useradm", "create-user", "--username", username, "--password", password},
	})

	if exitCode != 0 {
		var statusErr cli.StatusError
		if errors.As(err, &statusErr) {
			return User{}, fmt.Errorf("unexpected exit code '%d' when creating initial user: %w", statusErr.StatusCode, statusErr.Cause)
		}
		return User{}, fmt.Errorf("unexpected exit code '%d' when creating initial user", exitCode)
	}

	return User{
		Username: username,
		Password: password,
	}, nil
}

type Tenant struct {
	ID          string
	TenantToken *string
}

func (i *ComposeEnvironment) TenantOfUser(ctx context.Context, u User) (Tenant, error) {
	return Tenant{}, nil
}

func BasicAuthContext(ctx context.Context, user User) context.Context {
	return context.WithValue(
		ctx,
		oapiclient.ContextBasicAuth,
		oapiclient.BasicAuth{UserName: user.Username, Password: user.Password},
	)
}

func JWTAuthContext(ctx context.Context, jwt string) context.Context {
	return context.WithValue(
		ctx,
		oapiclient.ContextAccessToken,
		jwt,
	)
}

type TestSettings struct {
	ServerURL string
	APIClient *oapiclient.APIClient
}

func createComposeService(output io.Writer) (api.Compose, error) {
	dockerCLI, err := command.NewDockerCli(
		command.WithCombinedStreams(output),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker CLI: %w", err)
	}

	err = dockerCLI.Initialize(
		&flags.ClientOptions{},
		command.WithAPIClientOptions(client.FromEnv),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker CLI: %w", err)
	}

	compose, err := compose.NewComposeService(dockerCLI)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker Compose service: %w", err)
	}

	return compose, nil
}

func loadComposeProject(ctx context.Context, compose api.Compose, projectName string) (*types.Project, func(), error) {
	configPaths := []string{
		"../../../../../docker-compose.yml",
		"../../../docker/docker-compose.backend-tests.yml",
	}

	project, err := compose.LoadProject(
		ctx,
		api.ProjectLoadOptions{
			ProjectName: projectName,
			ConfigPaths: configPaths,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project: %w", err)
	}

	s3, ok := project.Services["s3"]
	if !ok {
		return project, nil, nil
	}

	// Replicate `docker compose up` workaround for config mapping in `s3` service.
	//
	// Mounting a `Config` object (e.g `s3-conf`) as a file in a container is not actually
	// supported by the Docker Engine API unless you run docker compose in swarm mode (which we don't).
	//
	// The `docker compose up` CLI command works around this by mounting a temp file with
	// the `Config` content as a BindMount volume instead behind the scenes - effectively "faking it".
	//
	// As the Compose SDK interacts directly with the Docker Engine API, we need to do the same thing.
	config, ok := project.Configs["s3-conf"]
	if !ok {
		return nil, nil, fmt.Errorf("couldn't find s3-conf configuration object in project")
	}

	// We (unfortunately) can't use the tempfiles feature in `go test` because the docker-in-docker
	// used in CI doesn't allow mounting files from `/tmp` as volumes in containers.
	dir, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	source := path.Join(dir, "s3.conf")
	err = os.WriteFile(source, []byte(config.Content), 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write s3-conf to file: %w", err)
	}

	var target string
	for _, c := range s3.Command {
		if strings.HasPrefix(c, "-config=") {
			target = strings.Split(c, "=")[1]
		}
	}

	if target == "" {
		return nil, nil, fmt.Errorf("failed to resolve target destination for s3.conf from command: `%v`", s3.Command)
	}

	project, err = project.WithServicesTransform(func(name string, s types.ServiceConfig) (types.ServiceConfig, error) {
		if name == s3.Name {
			s.Volumes = append(s3.Volumes, types.ServiceVolumeConfig{
				Type:     string(mount.TypeBind),
				Source:   source,
				Target:   target,
				ReadOnly: true,
			})
			s.Configs = nil
		}
		return s, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to add volume mount to s3 service: %w", err)
	}

	return project, func() { os.Remove(source) }, nil
}

func healthCheckComposeEnvironment(ctx context.Context, compose api.Compose, project *types.Project) error {
	type healthCheck struct {
		name string
		url  string
		code int
	}
	// This is hopefully temporary because we should get this into the actual health checks
	// of the docker compose service itself so we can trust the "healthy/unhealthy" status
	// provided by the Docker Engine.
	healthChecks := []healthCheck{
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
		{name: "deviceauth-ingress", url: "https://traefik:443/api/devices/v1/authentication/auth_requests", code: http.StatusMethodNotAllowed},
	}

	for _, h := range healthChecks {
		var healthy bool
		for i := 0; i < 5; i++ {
			exit, err := compose.Exec(
				ctx,
				project.Name,
				api.RunOptions{
					Service: "traefik",
					Command: []string{
						"sh",
						"-c",
						fmt.Sprintf(
							"wget --no-check-certificate --quiet --tries=1 --timeout=1 --server-response %s 2>&1 | grep %d > /dev/null",
							h.url,
							h.code,
						),
					},
				},
			)

			if exit == 0 && err == nil {
				healthy = true
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		if !healthy {
			return fmt.Errorf("service '%s' did not become healthy in time", h.name)
		}
	}

	return nil
}

func createMenderAPIClient(
	serverURL string,
	dialContext func(ctx context.Context, network, addr string) (net.Conn, error),
) *oapiclient.APIClient {
	config := oapiclient.NewConfiguration()
	config.Host = serverURL
	config.Scheme = "https"
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext:     dialContext,
		},
	}
	return oapiclient.NewAPIClient(config)
}
