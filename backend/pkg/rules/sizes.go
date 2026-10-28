package rules

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	size256  = validation.Length(0, 256)
	size4096 = validation.Length(0, 4096)

	identifierSize = size256
	nameSize       = size256
	textSize       = size4096

	legacyDeviceGroupNameSize = validation.Length(0, 1024)
	deviceGroupNameSize       = size256

	deploymentNameSize = size256

	emailSize    = size256
	passwordSize = size256
)
