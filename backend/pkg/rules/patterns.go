package rules

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	identifierNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	nameRegex           = regexp.MustCompile(`^[\p{L}\p{M}\p{N}\p{P}\p{S}\p{Zs}]+$`)
	textRegex           = regexp.MustCompile(`^[\p{L}\p{M}\p{N}\p{P}\p{S}\p{Zs}\t\n\r]+$`)

	ErrIdentifierRegex = "identifier can only contain: upper/lowercase " +
		"alphanum, -(dash), _(underscore), .(period)"
	identifierPattern = validation.Match(identifierNameRegex).Error(ErrIdentifierRegex)

	ErrNameRegex = "name can only contain valid UTF-8 graphic Unicode characters"
	namePattern  = validation.Match(nameRegex).Error(ErrNameRegex)

	ErrTextRegex = "text can only contain valid UTF-8 graphic Unicode characters, " +
		"tabs, and line breaks"
	textPattern = validation.Match(textRegex).Error(ErrTextRegex)

	deviceGroupPattern = validation.Match(identifierNameRegex).Error(
		"group name can only contain: upper/lowercase " +
			"alphanum, -(dash), _(underscore), .(period)")
)
