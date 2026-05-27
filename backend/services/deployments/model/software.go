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
	Kind        ReleaseKind `json:"kind" bson:"kind,omitempty"`
}

// custom marshal to include the kind in the response
func (s Software) MarshalJSON() ([]byte, error) {
	type Alias Software
	return json.Marshal(&struct {
		Alias
		Kind ReleaseKind `json:"kind"` // expose Kind
	}{
		Alias: (Alias)(s),
		Kind:  s.ReleaseBase.Kind,
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
