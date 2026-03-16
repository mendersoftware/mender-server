package rules

import (
	"fmt"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func HasUnderlyingType(input interface{}, kind reflect.Kind) bool {
	v := reflect.ValueOf(input)

	if !v.IsValid() {
		return false
	}

	return v.Kind() == kind
}

func DeviceGroupName(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for device group", value)
	}

	return validation.Validate(value,
		validation.Required,
		deviceGroupNameSize,
		deviceGroupPattern,
	)
}

func LegacyDeviceGroupName(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for device group", value)
	}

	return validation.Validate(value,
		validation.Required,
		legacyDeviceGroupNameSize,
		deviceGroupPattern,
	)
}

func DeploymentName(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for deployment name", value)
	}

	return validation.Validate(value,
		validation.Required,
		deploymentNameSize,
	)
}

func Email(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for email", value)
	}
	return validation.Validate(value,
		emailSize,
		is.ASCII, is.EmailFormat,
	)
}
