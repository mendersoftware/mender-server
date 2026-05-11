package model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SoftwareTagsFilter struct {
	Kind ReleaseKind
}

func (s SoftwareTagsFilter) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(
			&s.Kind,
			validation.Required,
			validation.In(
				ReleaseKind(ReleaseKindRelease),
				ReleaseKind(ReleaseKindManifest),
			),
		),
	)
}
