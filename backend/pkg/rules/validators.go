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

func Identifier(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for identifier", value)
	}

	return validation.Validate(value,
		identifierSize,
		identifierPattern,
	)
}

func Name(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for name", value)
	}

	return validation.Validate(value,
		nameSize,
		namePattern,
	)
}

func Text(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for text", value)
	}

	return validation.Validate(value,
		textSize,
		textPattern,
	)
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

	// is.Email enforces:
	//	1 - Length between 6 and 254 characters.
	//	2 - W3C HTML5 / RFC 5322 compliant email regex.
	// Reference: https://github.com/asaskevich/govalidator
	return validation.Validate(value,
		is.ASCII,
		is.Email,
	)
}

func Password(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for password", value)
	}

	return validation.Validate(value,
		passwordSize,
		is.ASCII,
	)
}
