package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/mendersoftware/mender-server/pkg/api/client"
)

func getPatternFromStructTag(fieldName string) *regexp.Regexp {
	field, ok := reflect.TypeOf(client.DeploymentV1{}).FieldByName(fieldName)
	if !ok {
		panic("Field " + fieldName + " not found in DeploymentV1")
	}

	tag := field.Tag.Get("validate")
	pattern := strings.TrimPrefix(tag, "regexp=")

	// Replace \\x with \x so the Go regex compiler evaluates the hex properly.
	pattern = strings.ReplaceAll(pattern, "\\\\", "\\")

	return regexp.MustCompile(pattern)
}

func TestOASPatternIdentifier(t *testing.T) {
	re := getPatternFromStructTag("Id")

	okCases := []string{
		"",
		"Abc",
		"._-",
	}
	invalidCases := []string{
		"id space",
		"id\nnewline",
		"id\ttab",
		"id@",
		"id/",
		"id$",
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if !re.MatchString(tc) {
				t.Errorf("expected '%s' to match pattern %s", tc, re.String())
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if re.MatchString(tc) {
				t.Errorf("expected '%s' to fail pattern %s", tc, re.String())
			}
		})
	}
}

func TestOASPatternName(t *testing.T) {
	re := getPatternFromStructTag("Name")

	okCases := []string{
		"",
		"A",
		"Name 123!@#-_.",
		"عليك تجربة الفالدي 😉",
	}
	invalidCases := []string{
		"name\nnewline",
		"name\ttab",
		"name\x00null",
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if !re.MatchString(tc) {
				t.Errorf("expected '%s' to match pattern %s", tc, re.String())
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if re.MatchString(tc) {
				t.Errorf("expected '%s' to fail pattern %s", tc, re.String())
			}
		})
	}
}

func TestOASPatternText(t *testing.T) {
	re := getPatternFromStructTag("ArtifactName")

	okCases := []string{
		"",
		"A",
		"text 123!@#-_.",
		"我想要一隻小狗，汪汪，汪汪",
		"text\nnewline",
		"text\ttab",
		"text\rcarriage",
		"هل تعلم ان لدينا ميزة سجلات المراجعة؟ جرب الانتربرايز 👀",
	}
	invalidCases := []string{
		"text\x00null",
		"text\x0bverticalTab",
		"text\x0eshiftOut",
	}

	for i, tc := range okCases {
		t.Run(fmt.Sprintf("Ok/%d", i), func(t *testing.T) {
			if !re.MatchString(tc) {
				t.Errorf("expected '%s' to match pattern %s", tc, re.String())
			}
		})
	}

	for i, tc := range invalidCases {
		t.Run(fmt.Sprintf("Invalid/%d", i), func(t *testing.T) {
			if re.MatchString(tc) {
				t.Errorf("expected '%s' to fail pattern %s", tc, re.String())
			}
		})
	}
}
