// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package model

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	validPlanName  = regexp.MustCompile(`^[a-zA-Z0-9_\\-]*$`)
	PlanList       = []Plan{}
	PlanNameToPlan = map[string]*Plan{}
)

type Features struct {
	// RBAC
	RBAC bool `json:"rbac" bson:"rbac"`

	// audit logs
	AuditLogs bool `json:"audit_logs" bson:"audit_logs"`

	// dynamic groups
	DynamicGroups bool `json:"dynamic_groups" bson:"dynamic_groups"`

	// remote terminal
	Terminal bool `json:"terminal" bson:"terminal"`

	// file transfer
	FileTransfer bool `json:"file_transfer" bson:"file_transfer"`

	// configuration
	Configuration bool `json:"configuration" bson:"configuration"`

	// monitoring
	Monitoring bool `json:"monitoring" bson:"monitoring"`

	// reporting
	Reporting bool `json:"reporting" bson:"reporting"`
}

type PlanLimits struct {
	// maximum number of devices
	Devices int `json:"devices" bson:"devices"`

	// maximum number of users
	Users int `json:"users" bson:"users"`

	// audit logs history in days
	AuditLogsDays int `json:"audit_logs_days" bson:"audit_logs_days"`
}

type Plan struct {
	// unique name used as identifier
	Name string `json:"name" bson:"_id"`

	// name to display
	DisplayName string `json:"display_name" bson:"display_name"`

	// feature set
	Features *Features `json:"features" bson:"features"`
}

type PlanBindingDetails struct {
	// tenant can only have single plan binding
	TenantID string `json:"-" bson:"_id"`

	// plan name
	Plan Plan `json:"plan" bson:"plan"`

	// limits
	Limits *PlanLimits `json:"limits,omitempty" bson:"limits,omitempty"`
}

type PlanBinding struct {
	// tenant can only have single plan binding
	TenantID string `json:"-" bson:"_id"`

	// plan name
	PlanName string `json:"plan_name" bson:"plan_name"`

	// limits
	Limits *PlanLimits `json:"limits,omitempty" bson:"limits,omitempty"`
}

func ValidatePlanName(name string) error {
	//plan of no name is not accepted
	if !(len(name) > 0) {
		return errors.New("plan name cannot be empty")
	}

	//plan name must be composed of letters or numbers with possible _ or -
	if !validPlanName.MatchString(name) {
		return errors.New("invalid plan name")
	}
	return nil
}

func (p *Plan) Validate() error {
	//plan of not valid name is not accepted
	if err := ValidatePlanName(p.Name); err != nil {
		return err
	}

	return nil
}

func jsonSerializerDeserializer(serialize interface{}, deserialized interface{}) error {
	b, err := json.Marshal(serialize)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, deserialized)
	return err
}

// LoadPlans loads Plan definitions from filePath.
// This function also checks for the validity of the definitions.
func LoadPlans(filepath string) error {
	var res struct {
		Plans map[string]interface{} `json:"plans" yaml:"plans"`
	}
	var decoder func(r io.Reader, v interface{}) error
	switch path.Ext(filepath) {
	case ".json":
		decoder = func(r io.Reader, v interface{}) error {
			return json.NewDecoder(r).
				Decode(v)
		}
	case ".yaml", ".yml":
		decoder = func(r io.Reader, v interface{}) error {
			return yaml.NewDecoder(r).
				Decode(v)
		}
	default:
		return fmt.Errorf(
			"file extension for %q not recognized",
			path.Base(filepath),
		)
	}
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = decoder(f, &res)
	if err != nil {
		return err
	}
	return loadPlans(res.Plans)
}

func loadPlans(
	rawPlans map[string]interface{},
) error {
	var (
		err error

		newPlanList []Plan
		newPlans    map[string]*Plan
	)

	if rawPlans != nil {
		if err := jsonSerializerDeserializer(rawPlans, &newPlans); err != nil {
			return fmt.Errorf(
				"failed to parse plans: %w",
				err,
			)
		}
		newPlanList, err = buildPlanList(newPlans)
		if err != nil {
			return err
		}
	}

	PlanList = newPlanList
	PlanNameToPlan = newPlans
	return nil
}

func buildPlanList(newPlans map[string]*Plan) ([]Plan, error) {
	newPlanList := make([]Plan, 0, len(newPlans))
	for name, plan := range newPlans {
		if plan == nil {
			return nil, fmt.Errorf("plan %q is empty", name)
		}
		plan.Name = name

		err := plan.Validate()
		if err != nil {
			return nil, fmt.Errorf(
				"invalid plan definition: %w",
				err,
			)
		}
		newPlanList = append(newPlanList, *plan)
	}
	sort.Slice(newPlanList, func(i, j int) bool {
		return newPlanList[i].Name < newPlanList[j].Name
	})
	return newPlanList, nil
}
