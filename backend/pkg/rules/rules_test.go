package rules

import (
	"fmt"
	"strings"
	"testing"
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
