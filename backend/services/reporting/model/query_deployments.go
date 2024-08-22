// Copyright 2023 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

func BuildDeploymentsQuery(params DeploymentsSearchParams) (Query, error) {
	query := NewQuery()

	for _, f := range params.Filters {
		fpart, err := getFilterPart(FilterPredicate{
			Attribute: f.Attribute,
			Type:      f.Type,
			Value:     f.Value,
		})
		if err != nil {
			return nil, err
		}
		query = fpart.AddTo(query)
	}

	if len(params.DeploymentGroups) > 0 {
		fpart, err := getFilterPart(FilterPredicate{
			Attribute: FieldNameDeploymentGroups,
			Type:      "$in",
			Value:     params.DeploymentGroups,
		})
		if err != nil {
			return nil, err
		}
		query = fpart.AddTo(query)
	}

	for _, s := range params.Sort {
		sort := NewDeploymentsSort(s)
		query = sort.AddTo(query)
	}

	query = query.WithPage(params.Page, params.PerPage)

	if len(params.Attributes) > 0 {
		sel := NewDeploymentsSelect(params.Attributes)
		query = sel.AddTo(query)
	}

	return query, nil
}

type deploymentsSort struct {
	attrStr  string
	attrNum  string
	attrBool string
	order    string
}

func NewDeploymentsSort(sc DeploymentsSortCriteria) *deploymentsSort {
	order := sc.Order
	if order == "" {
		order = SortOrderAsc
	}
	return &deploymentsSort{
		attrStr:  sc.Attribute,
		attrNum:  sc.Attribute,
		attrBool: sc.Attribute,
		order:    order,
	}
}

func (s *deploymentsSort) AddTo(q Query) Query {
	q = q.
		WithSort(
			M{
				s.attrStr: M{
					"order":         s.order,
					"unmapped_type": "keyword",
				},
			},
		).WithSort(
		M{
			s.attrNum: M{
				"order":         s.order,
				"unmapped_type": "double",
			},
		},
	)

	return q
}

type deploymentsSel struct {
	attrs []DeploymentsSelectAttribute
}

func NewDeploymentsSelect(attrs []DeploymentsSelectAttribute) *deploymentsSel {
	return &deploymentsSel{
		attrs: attrs,
	}
}

func (s *deploymentsSel) AddTo(q Query) Query {
	fields := []string{}

	for _, a := range s.attrs {
		fields = append(fields, a.Attribute)
	}

	// always include the deployment id
	fields = append(fields, "id")

	return q.With(map[string]interface{}{
		"fields":  fields,
		"_source": false,
	})

}
