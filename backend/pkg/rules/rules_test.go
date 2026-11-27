package rules

import (
	"fmt"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestRuleIdentifier(t *testing.T) {
	okCases := []string{
		"",
		"Abc",
		"._-",
		strings.Repeat("a", 256),
	}
	invalidCases := []string{
		"id space",
		"id\nnewline",
		"id\ttab",
		"id@",
		"id/",
		"id$",
		strings.Repeat("a", 257),
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if err := Identifier(tc); err != nil {
				t.Errorf("expected '%s' to be valid, got err: %v", tc, err)
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if err := Identifier(tc); err == nil {
				t.Errorf("expected '%s' to be invalid", tc)
			}
		})
	}
}

func TestGoRuleName(t *testing.T) {
	okCases := []string{
		"",
		"A",
		"Name 123!@#-_.",
		"عليك تجربة الفالدي 😉",
		strings.Repeat("a", 256),
	}
	invalidCases := []string{
		"name\nnewline",
		"name\ttab",
		"name\x00null",
		strings.Repeat("a", 257),
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if err := Name(tc); err != nil {
				t.Errorf("expected '%s' to be valid, got err: %v", tc, err)
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if err := Name(tc); err == nil {
				t.Errorf("expected '%s' to be invalid", tc)
			}
		})
	}
}

func TestRuleText(t *testing.T) {
	okCases := []string{
		"",
		"A",
		"text 123!@#-_.",
		"我想要一隻小狗，汪汪，汪汪",
		"text\nnewline",
		"text\ttab",
		"text\rcarriage",
		"هل تعلم ان لدينا ميزة سجلات المراجعة؟ جرب الانتربرايز 👀",
		strings.Repeat("a", 4096),
	}
	invalidCases := []string{
		"text\x00null",
		"text\x0bverticalTab",
		"text\x0eshiftOut",
		strings.Repeat("a", 4097),
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if err := Text(tc); err != nil {
				t.Errorf("expected '%s' to be valid, got err: %v", tc, err)
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if err := Text(tc); err == nil {
				t.Errorf("expected '%s' to be invalid", tc)
			}
		})
	}
}

// assert.ErrorAssertionFunc helpers
func ErrorIs(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...any) bool {
		if tt, ok := t.(testing.TB); ok {
			tt.Helper()
		}
		return assert.ErrorIs(t, err, target, msgAndArgs...)
	}
}
func ErrorContains(contains string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...any) bool {
		if tt, ok := t.(testing.TB); ok {
			tt.Helper()
		}
		return assert.ErrorContains(t, err, contains, msgAndArgs...)
	}
}

func TestRulez(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name string

		Rules []validation.Rule
		Value any

		Error assert.ErrorAssertionFunc
	}{
		{
			Name:  "TenantID/ok",
			Rules: []validation.Rule{TenantID},
			Value: "123456789012345678901234",
			Error: assert.NoError,
		},
		{
			Name:  "TenantID/ok pointer",
			Rules: []validation.Rule{TenantID},
			Value: new("123456789012345678901234"),
			Error: assert.NoError,
		},
		{
			Name:  "TenantID/empty",
			Rules: []validation.Rule{TenantID},
			Value: "",
			Error: assert.NoError,
		},
		{
			Name:  "TenantID/skip nil",
			Rules: []validation.Rule{TenantID},
			Value: (*string)(nil),
			Error: assert.NoError,
		},
		{
			Name:  "TenantID/invalid charset",
			Rules: []validation.Rule{TenantID},
			Value: "thisisdefinitelynothex123",
			Error: ErrorIs(ErrInvalidTenantID),
		},
		{
			Name:  "TenantID/invalid length",
			Rules: []validation.Rule{TenantID},
			Value: "67",
			Error: ErrorIs(ErrInvalidTenantID),
		},
		{
			Name:  "TenantID/invalid type",
			Rules: []validation.Rule{TenantID},
			Value: true,
			Error: ErrorContains("invalid type bool for string"),
		},
		{
			Name:  "TenantID/invalid type <nil>",
			Rules: []validation.Rule{TenantID},
			Value: nil,
			Error: ErrorContains("invalid type <nil> for string"),
		},
		{
			Name:  "DeviceID/ok/bsonObjectID",
			Rules: []validation.Rule{DeviceID},
			Value: "123456789012345678901234",
			Error: assert.NoError,
		},
		{
			Name:  "DeviceID/ok/UUID",
			Rules: []validation.Rule{DeviceID},
			Value: "5c68e04e-682d-4130-a31c-68ace9841d89",
			Error: assert.NoError,
		},
		{
			Name:  "DeviceID/ok/pointer",
			Rules: []validation.Rule{DeviceID},
			Value: new("5c68e04e-682d-4130-a31c-68ace9841d89"),
			Error: assert.NoError,
		},
		{
			Name:  "DeviceID/skip nil",
			Rules: []validation.Rule{DeviceID},
			Value: (*string)(nil),
			Error: assert.NoError,
		},
		{
			Name:  "DeviceID/empty",
			Rules: []validation.Rule{DeviceID},
			Value: "",
			Error: assert.NoError,
		},
		{
			Name:  "DeviceID/invalid charset",
			Rules: []validation.Rule{DeviceID},
			Value: "thisisdefinitelynothex123",
			Error: ErrorIs(ErrInvalidDeviceID),
		},
		{
			Name:  "DeviceID/invalid length",
			Rules: []validation.Rule{DeviceID},
			Value: "5c68e04e682d4130a31c68ace9841d89",
			Error: ErrorIs(ErrInvalidDeviceID),
		},
		{
			Name:  "DeviceID/invalid type",
			Rules: []validation.Rule{DeviceID},
			Value: true,
			Error: ErrorContains("invalid type bool for string"),
		},
		{
			Name:  "Email/ok",
			Rules: []validation.Rule{Func(Email)},
			Value: "mender-test-bot@mender.io",
			Error: assert.NoError,
		},
		{
			Name:  "Email/ok",
			Rules: []validation.Rule{Func(Email)},
			Value: "mender-test-bot@mender.io",
			Error: assert.NoError,
		},
		{
			Name:  "Email/invalid type",
			Rules: []validation.Rule{Func(Email)},
			Value: true,
			Error: ErrorContains("invalid type bool for email"),
		},
		{
			Name:  "Email/invalid charset",
			Rules: []validation.Rule{Func(Email)},
			Value: "føø@bar.com",
			Error: ErrorContains("must contain ASCII characters only"),
		},
		{
			Name:  "Email/invalid format",
			Rules: []validation.Rule{Func(Email)},
			Value: "foobar",
			Error: ErrorContains("must be a valid email address"),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			err := validation.Validate(tc.Value, tc.Rules...)
			tc.Error(t, err)
		})
	}
}
