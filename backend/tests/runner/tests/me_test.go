package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	AddTestCase("test_z_me", mainTestMe)
}

func mainTestMe(t *testing.T, settings *TestSettings) error {
	t.Logf("login test starting\n")
	ctx := context.Background()
	r, err := settings.client.ShowMyUserSettingsWithResponse(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, r)

	assert.Equal(t, 200, r.StatusCode())
	t.Logf("test passed with data: %+v\n", r.JSON200)
	return nil
}
