package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailDNSCheck(t *testing.T) {
	// The .invalid TLD is reserved (RFC 2606) and never resolves.
	const unresolvable = "user@mender.invalid"

	t.Cleanup(func() { SetEmailDNSCheck(true) })

	t.Run("enabled by default", func(t *testing.T) {
		assert.Error(t, Email(unresolvable))
		// Domains special-cased by the validator pass without DNS.
		assert.NoError(t, Email("user@example.com"))
	})

	t.Run("disabled", func(t *testing.T) {
		SetEmailDNSCheck(false)
		t.Cleanup(func() { SetEmailDNSCheck(true) })

		assert.NoError(t, Email(unresolvable))
		assert.Error(t, Email("not-an-email"))
		assert.Error(t, Email("user@"))
		assert.Error(t, Email(123))
	})

	t.Run("re-enabled", func(t *testing.T) {
		SetEmailDNSCheck(false)
		SetEmailDNSCheck(true)
		assert.Error(t, Email(unresolvable))
	})
}
