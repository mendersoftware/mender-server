package tests

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	openapi "github.com/mendersoftware/mender-server/tests/runner/client"
)

type TestSettings struct {
	ServerURL string
	Username  string
	Password  string
	//nolint:unused
	jwt string
	//nolint:unused
	client *openapi.APIClient
	//nolint:unused
	configuration *openapi.Configuration
}

type BackendIntegrationSuite struct {
	suite.Suite
	settings *TestSettings
}

var ServerURL string
var Username string
var Password string
var UsernamePattern string
var PasswordPattern string

func TestMain(m *testing.M) {
	flag.StringVar(&ServerURL, "server-url", "https://localhost", "Mender Server URL")
	flag.StringVar(&Username, "username", "demo@mender.io", "Mender Server login username")
	flag.StringVar(&Password, "password", "password123", "Mender Server user password")
	// the patterns are used to inform the test about what users and passwords were created during the setup phase
	// that is: what usernames and password had been created before the go test started
	flag.StringVar(&UsernamePattern, "username-pattern", "", "Mender Server login username pattern e.g: test-user-%d@mender.io")
	flag.StringVar(&PasswordPattern, "password-pattern", "", "Mender Server user password pattern e.g.: test-user-password-%d")
	flag.Parse()
	os.Exit(m.Run())
}

func (i *BackendIntegrationSuite) SetupSuite() {
	if len(PasswordPattern) > 0 && len(UsernamePattern) > 0 {
		Username = strings.ReplaceAll(UsernamePattern, "%d", "1")
		Password = strings.ReplaceAll(PasswordPattern, "%d", "1")
	}
	config := openapi.NewConfiguration()
	config.Host = ServerURL
	config.Scheme = "https"
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	i.settings = &TestSettings{
		ServerURL:     ServerURL,
		Username:      Username,
		Password:      Password,
		configuration: config,
		client:        openapi.NewAPIClient(config),
	}
}

func TestBackendIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BackendIntegrationSuite))
}

func (i *BackendIntegrationSuite) TestUseradmManagementV1() {
	suite.Run(i.T(), &UseradmManagementV1Suite{settings: i.settings})
}
