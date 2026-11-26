package rules

import (
	"fmt"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func HasUnderlyingType(input interface{}, kind reflect.Kind) bool {
	v := reflect.ValueOf(input)

	if !v.IsValid() {
		return false
	}

	return v.Kind() == kind
}

func ObjectId(value interface{}) error {
	if validation.IsEmpty(value) {
		return nil
	}

	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for identifier", value)
	}

	if _, err := bson.ObjectIDFromHex(value.(string)); err != nil {
		return fmt.Errorf("invalid object id hex string: %w", err)
	}
	return nil
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

func NameRule() validation.Rule {
	return validation.By(func(value interface{}) error {
		if validation.IsEmpty(value) {
			return nil
		}

		v := reflect.ValueOf(value)
		for v.IsValid() && v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}

		if !v.IsValid() {
			return nil
		}

		return Name(v.Interface())
	})
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

// emailDNSCheckDisabled controls whether Email skips the DNS (MX/A) lookup
// of the e-mail domain. The zero value keeps the lookup enabled.
var emailDNSCheckDisabled bool

// SetEmailDNSCheck enables or disables the DNS (MX/A) lookup performed by
// Email on the e-mail domain. The check is enabled by default; disable it on
// air-gapped installations where public domains cannot be resolved.
func SetEmailDNSCheck(enabled bool) {
	emailDNSCheckDisabled = !enabled
}

func Email(value interface{}) error {
	if !HasUnderlyingType(value, reflect.String) {
		return fmt.Errorf("invalid type %T for email", value)
	}

	// is.Email enforces:
	//	1 - Length between 6 and 254 characters.
	//	2 - W3C HTML5 / RFC 5322 compliant email regex.
	//	3 - The e-mail domain has an MX or A/AAAA DNS record.
	// is.EmailFormat enforces only 1 and 2 (no DNS lookup).
	// Reference: https://github.com/asaskevich/govalidator
	emailRule := is.Email
	if emailDNSCheckDisabled {
		emailRule = is.EmailFormat
	}

	return validation.Validate(value,
		is.ASCII,
		emailRule,
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
