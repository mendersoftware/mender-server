package tests

import (
	"strconv"
	"testing"

	"github.com/mendersoftware/mender-server/tests/runner/client"
)

// TestSettings contains settings for running a test.
// all the private fields are the ones that tests can fill
// (thanks to getting a pointer to objects of this type)
// for other tests to use.
type TestSettings struct {
	ServerURL string
	Username  string
	Password  string
	jwt       string
	client    *client.ClientWithResponses
}

type MainTestFunc func(t *testing.T, settings *TestSettings) error

var TestCases = make(map[string]MainTestFunc)

func AddTestCase(name string, main MainTestFunc) {
	i := 0
	for {
		if _, defined := TestCases[name]; defined {
			i++
			name = name + "_" + strconv.Itoa(i)
		} else {
			break
		}
	}
	TestCases[name] = main
}
