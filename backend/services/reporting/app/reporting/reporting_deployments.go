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

package reporting

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

// AggregateDeployments aggregates deployments data
func (app *app) AggregateDeployments(
	ctx context.Context,
	aggregateParams *model.AggregateDeploymentsParams,
) ([]model.DeviceAggregation, error) {
	searchParams := &model.DeploymentsSearchParams{
		Filters:          aggregateParams.Filters,
		DeploymentGroups: aggregateParams.DeploymentGroups,
		TenantID:         aggregateParams.TenantID,
	}
	query, err := model.BuildDeploymentsQuery(*searchParams)
	if err != nil {
		return nil, err
	}
	if searchParams.TenantID != "" {
		query = query.Must(model.M{
			"term": model.M{
				model.FieldNameTenantID: searchParams.TenantID,
			},
		})
	}

	aggregations, err := model.BuildDeploymentsAggregations(aggregateParams.Aggregations)
	if err != nil {
		return nil, err
	}

	query = query.WithSize(0).With(map[string]interface{}{
		"aggs": aggregations,
	})
	esRes, err := app.store.AggregateDeployments(ctx, query)
	if err != nil {
		return nil, err
	}

	aggregationsS, ok := esRes["aggregations"].(map[string]interface{})
	if !ok {
		return nil, errors.New("can't process store aggregations slice")
	}
	res, err := app.storeToDeviceAggregations(ctx, searchParams.TenantID, aggregationsS)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SearchDevices searches device data
func (app *app) SearchDeployments(
	ctx context.Context,
	searchParams *model.DeploymentsSearchParams,
) ([]model.Deployment, int, error) {
	query, err := model.BuildDeploymentsQuery(*searchParams)
	if err != nil {
		return nil, 0, err
	}

	if searchParams.TenantID != "" {
		query = query.Must(model.M{
			"term": model.M{
				model.FieldNameTenantID: searchParams.TenantID,
			},
		})
	}

	if len(searchParams.DeviceIDs) > 0 {
		query = query.Must(model.M{
			"terms": model.M{
				model.FieldNameDeviceID: searchParams.DeviceIDs,
			},
		})
	}

	if len(searchParams.DeploymentIDs) > 0 {
		query = query.Must(model.M{
			"terms": model.M{
				model.FieldNameDeploymentID: searchParams.DeploymentIDs,
			},
		})
	}

	esRes, err := app.store.SearchDeployments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	res, total, err := app.storeToDeployments(ctx, searchParams.TenantID, esRes)
	if err != nil {
		return nil, 0, err
	}

	return res, total, err
}

// storeToInventoryDevs translates ES results directly to inventory devices
func (a *app) storeToDeployments(
	ctx context.Context, tenantID string, storeRes map[string]interface{},
) ([]model.Deployment, int, error) {
	depls := []model.Deployment{}

	hitsM, ok := storeRes["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("can't process store hits map")
	}

	hitsTotalM, ok := hitsM["total"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("can't process total hits struct")
	}

	total, ok := hitsTotalM["value"].(float64)
	if !ok {
		return nil, 0, errors.New("can't process total hits value")
	}

	hitsS, ok := hitsM["hits"].([]interface{})
	if !ok {
		return nil, 0, errors.New("can't process store hits slice")
	}

	for _, v := range hitsS {
		res, err := a.storeToDeployment(ctx, tenantID, v)
		if err != nil {
			return nil, 0, err
		}

		depls = append(depls, *res)
	}

	return depls, int(total), nil
}

func (a *app) storeToDeployment(ctx context.Context, tenantID string,
	storeRes interface{}) (*model.Deployment, error) {
	resM, ok := storeRes.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't process individual hit")
	}

	// if query has a 'fields' clause, use 'fields' instead of '_source'
	sourceM, ok := resM["_source"].(map[string]interface{})
	if !ok {
		sourceM, ok = resM["fields"].(map[string]interface{})
		if !ok {
			return nil, errors.New("can't process hit's '_source' nor 'fields'")
		}
	}

	source, err := json.Marshal(sourceM)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal result into JSON")
	}

	ret := &model.Deployment{}
	err = json.Unmarshal(source, ret)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal result from JSON")
	}

	return ret, nil
}
