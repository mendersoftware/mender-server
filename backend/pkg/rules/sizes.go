package rules

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	size256 = validation.Length(0, 256)

	legacyDeviceGroupNameSize = validation.Length(0, 1024)
	deviceGroupNameSize       = size256

	deploymentNameSize = size256

	emailSize    = size256
	passwordSize = size256
)
