package rules

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	groupNameRegex = regexp.MustCompile("^[A-Za-z0-9_-]*$")

	deviceGroupPattern = validation.Match(groupNameRegex).Error(
		"group name can only contain: upper/lowercase " +
			"alphanum, -(dash), _(underscore)")
)
