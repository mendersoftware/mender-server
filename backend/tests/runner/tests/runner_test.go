// this is the place where all the tests are called
// it is enough to create a your-test_test.go file with init containing:
// TestCases["your-test"] = your-test
// run with:
// go test -parallel 1 -count 1 -v github.com/mendersoftware/mender-server/tests/runner/tests -args -server-url=https://staging.hosted.mender.io -username=youruser -password=yourpass
package tests

import (
	"flag"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestIntegrationRun(t *testing.T) {
	if len(PasswordPattern) > 0 && len(UsernamePattern) > 0 {
		Username = strings.ReplaceAll(UsernamePattern, "%d", "1")
		Password = strings.ReplaceAll(PasswordPattern, "%d", "1")
	}
	settings := &TestSettings{
		ServerURL: ServerURL,
		Username:  Username,
		Password:  Password,
	}

	keys := slices.Collect(maps.Keys(TestCases))
	slices.Sort(keys)
	for _, name := range keys {
		test := TestCases[name]
		t.Logf("Running test %s\n", name)
		t.Run(name, func(t *testing.T) {
			assert.NoError(t,
				test(
					t,
					settings,
				),
			)
		})
	}
}
