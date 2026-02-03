package rules

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	legacyDeviceGroupNameSize = validation.Length(0, 1024)
	deviceGroupNameSize       = validation.Length(0, 256)
)
