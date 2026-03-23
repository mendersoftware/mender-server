package tests

import (
	"flag"
	"log"
	"os"
	"strings"
	"testing"

	oapiclient "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/stretchr/testify/suite"
)

var (
	projectName                 string
	skipCleanup                 bool
	skipCleanupOnFailure        bool
	printServiceLogsOnFailure   string
	printServiceStatusOnFailure string

	externalServerURL string
	externalTenants   []Tenant
)

type TestSettings struct {
	ServerURL string
	APIClient *oapiclient.APIClient
	Tenants   []Tenant
}

type User struct {
	Username string
	Password string
}

type Tenant struct {
	ID              string
	ServiceProvider bool
	Users           []User
}

func TestMain(m *testing.M) {
	flag.StringVar(&projectName, "project-name", "backend-tests",
		"The name to use for the docker compose environment",
	)
	flag.BoolVar(&skipCleanup, "skip-cleanup", false,
		"Skip cleanup of the docker compose environment",
	)
	flag.BoolVar(&skipCleanupOnFailure, "skip-cleanup-on-failure", false,
		"Skip cleanup of the docker compose environment if at least one test failed",
	)
	flag.StringVar(&printServiceStatusOnFailure, "print-service-status-on-failure", "",
		"Comma separated list of docker compose service named to describe on  test failure. Asterisk (*) prints all.",
	)
	flag.StringVar(&printServiceLogsOnFailure, "print-service-logs-on-failure", "",
		"Comma separated list of docker compose service named to print logs from on  test failure. Asterisk (*) prints all.",
	)

	var (
		serverURL string
		usernames string
		passwords string
	)

	flag.StringVar(&serverURL, "server-url", "",
		"Run tests against an existing environment instead of creating a docker compose environment as part of the test.",
	)
	flag.StringVar(&usernames, "usernames", "",
		"Comma separated usernames of users in existing environment (length must match `passwords`). At least one user is required.",
	)
	flag.StringVar(&passwords, "passwords", "",
		"Comma separated usernames of users in existing environment (length must match `usernames`). At least one password is required.",
	)

	flag.Parse()

	if serverURL != "" {
		externalServerURL = serverURL
		names := strings.Split(usernames, ",")
		passwords := strings.Split(passwords, ",")

		if len(names) == 0 {
			log.Printf("At least one username and password is required")
			os.Exit(1)
		}

		if len(names) != len(passwords) {
			log.Printf("Number of usernames and passwords does not match")
			os.Exit(1)
		}

		var tenant Tenant
		for idx := range names {
			tenant.Users = append(tenant.Users, User{
				Username: names[idx],
				Password: passwords[idx],
			})
		}
		externalTenants = []Tenant{tenant}
	}

	os.Exit(m.Run())
}

func TestBackendIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BackendIntegrationSuite))
}
