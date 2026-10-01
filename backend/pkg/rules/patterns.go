package rules

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	identifierNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

	deviceGroupPattern = validation.Match(identifierNameRegex).Error(
		"group name can only contain: upper/lowercase " +
			"alphanum, -(dash), _(underscore), .(period)")
)
