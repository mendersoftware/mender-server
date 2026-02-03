package rules

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func DeviceGroupeName(value interface{}) error {
	s := fmt.Sprintf("%v", value)
	return validation.Validate(s,
		validation.Required,
		deviceGroupNameSize,
		deviceGroupPattern,
	)
}

func LegacyDeviceGroupeName(value interface{}) error {
	s := fmt.Sprintf("%v", value)
	return validation.Validate(s,
		validation.Required,
		legacyDeviceGroupNameSize,
		deviceGroupPattern,
	)
}
