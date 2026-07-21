package opensource

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
)

type HealthSuite struct {
	suite.Suite

	Env common.TestEnvironment
}

func (i *BackendIntegrationSuite) TestHealth() {
	suite.Run(i.T(), &HealthSuite{Env: i.environment})
}

// ported from test_health.py::TestHealthCheck::test_health_check
func (h *HealthSuite) TestHealthEndpoints() {
	ce, ok := h.Env.(*common.ComposeEnvironment)
	if !ok {
		// External environments expose no internal health endpoints.
		h.T().Skip("health endpoints are only reachable in the compose environment")
	}
	require.NoError(h.T(), ce.HealthCheck(h.T().Context()))
}
