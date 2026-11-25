package rules

import (
	"encoding/hex"
	"fmt"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"

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

type stringValidator[T any] func(T) error

func (f stringValidator[T]) Validate(value any) error {
	var valueT T
	switch t := value.(type) {
	case T:
		valueT = t
	case *T:
		if t == nil {
			break
		}
		valueT = *t
	default:
		return fmt.Errorf("invalid type %T for %T", value, valueT)
	}
	return f(valueT)
}

var TenantID tenantID

type tenantID struct{}

var ErrInvalidTenantID = validation.NewError("invalid_tenant_id", "must be a valid TenantID")

func (rule tenantID) Validate(value any) error {
	return stringValidator[string](rule.ValidateString).Validate(value)
}

func (tenantID) ValidateString(value string) error {
	var valid bool
	switch len(value) {
	case 0:
		valid = true
	case 24:
		_, err := hex.DecodeString(value)
		valid = err == nil
	}
	if !valid {
		return ErrInvalidTenantID
	}
	return nil
}

var DeviceID deviceID

type deviceID struct{}

var ErrInvalidDeviceID = validation.NewError("invalid_device_id", "must be a valid DeviceID")

func (rule deviceID) Validate(value any) error {
	return stringValidator[string](rule.ValidateString).Validate(value)
}

func (rule deviceID) ValidateString(value string) error {
	var valid bool
	switch len(value) {
	case 0:
		return nil
	case 24:
		_, err := hex.DecodeString(value)
		valid = err == nil
	case 36:
		_, err := uuid.Parse(value)
		valid = err == nil
	}
	if !valid {
		return ErrInvalidDeviceID
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
