package model

import (
	"encoding/json"

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

type Software struct {
	ReleaseBase `bson:"inline"`
	Artifacts   []Image `json:"-" bson:"artifacts"`
}

// custom marshal to include the kind and compatible_types in the response
func (s Software) MarshalJSON() ([]byte, error) {
	var compatible_types []string
	for _, a := range s.Artifacts {
		compatible_types = append(
			compatible_types,
			a.DeviceTypesCompatible...,
		)
	}

	type Alias Software
	return json.Marshal(&struct {
		Alias
		Kind ReleaseKind `json:"kind"` // expose Kind
		// CompatibleTypes is the external name adopted by `mender-artifact`
		// for what is still `device_types_compatible` in the actual artifact
		// format. The reason for the name change is that `device_types_compatible``
		// can also hold `system_types_compatible` in the case where the Software
		// (aka Artifact) is a Manifest
		CompatibleTypes []string `json:"compatible_types"`
	}{
		Alias:           (Alias)(s),
		Kind:            s.ReleaseBase.Kind,
		CompatibleTypes: compatible_types,
	})
}

type SoftwareFilter struct {
	Names      []string    `json:"name"`
	NamePrefix string      `json:"name_prefix"`
	Kind       ReleaseKind `json:"kind"`
	UpdateType string      `json:"update_type"`
	Tags       []string    `json:"tags"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Sort       string      `json:"sort"`
}

func (s SoftwareFilter) Validate() error {
	return validation.ValidateStruct(&s,
		// Names and NamePrefix are mutually exclusive
		validation.Field(&s.Names,
			validation.When(s.NamePrefix != "",
				validation.Empty.Error("cannot be used with name_prefix")),
		),
		validation.Field(&s.NamePrefix,
			validation.When(len(s.Names) > 0,
				validation.Empty.Error("cannot be used with name")),
		),
		validation.Field(&s.Kind, validation.In(
			ReleaseKind(ReleaseKindRelease),
			ReleaseKind(ReleaseKindManifest),
		)),
		validation.Field(&s.Sort, validation.In(
			"modified:asc",
			"modified:desc",
			"name:asc",
			"name:desc",
			"tags:asc",
			"tags:desc",
		)),
	)
}
