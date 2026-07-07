//nolint:all // This is all test code so we don't care

package common

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
)

type TestEnvironment interface {
	Setup(t *testing.T)
	Teardown(t *testing.T)

	User(ctx context.Context) (User, error)
	TenantOfUser(ctx context.Context, u User) (Tenant, error)

	APIClient() *oapiclient.APIClient
}

func ResolveTestEnvironment() (TestEnvironment, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	environment := strings.ToLower(os.Getenv("TEST_ENVIRONMENT"))
	if environment == "external" {
		config, err := readExternalConfigFromEnv()
		if err != nil {
			return nil, err
		}
		return NewExternalEnvironment(config), nil
	} else {
		return NewComposeEnvironment(readComposeConfigFromEnv()), nil
	}
}

func readExternalConfigFromEnv() (ExternalEnvironmentConfig, error) {
	userFromEnv := func(name string) (User, error) {
		name = strings.ToUpper(name)
		value := os.Getenv(name)
		if value == "" {
			return User{}, fmt.Errorf("Missing '%s' in environment", value)
		}

		parts := strings.Split(value, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return User{},
				fmt.Errorf(
					"Invalid '%s' value in environment, should be '<username>:<password>', got %s",
					name,
					value,
				)
		}
		return User{Username: parts[0], Password: parts[1]}, nil
	}

	cfg := ExternalEnvironmentConfig{
		URL: os.Getenv("SERVER_URL"),
	}
	if cfg.URL == "" {
		return ExternalEnvironmentConfig{}, fmt.Errorf("Missing 'SERVER_URL' in environment")
	}

	var err error
	cfg.User, err = userFromEnv("USER")
	if err != nil {
		return ExternalEnvironmentConfig{}, err
	}

	return cfg, nil
}

func readComposeConfigFromEnv() ComposeEnvironmentConfig {
	cfg := ComposeEnvironmentConfig{
		ProjectName: "backend-tests",
	}

	if s := os.Getenv("PROJECT_NAME"); s != "" {
		cfg.ProjectName = s
	}

	if os.Getenv("SKIP_CLEANUP") != "" {
		cfg.SkipCleanup = true
	}

	if os.Getenv("SKIP_CLEANUP_ON_FAILURE") != "" {
		cfg.SkipCleanupOnFailure = true
	}

	cfg.PrintServiceLogsOnFailure = os.Getenv("PRINT_SERVICE_LOGS_ON_FAILURE")
	cfg.PrintServiceStatusOnFailure = os.Getenv("PRINT_SERVICE_STATUS_ON_FAILURE")

	return cfg
}
