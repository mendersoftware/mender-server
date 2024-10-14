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
	"errors"
	"sort"
	"time"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	"github.com/mendersoftware/mender-server/services/reporting/mapping"
	"github.com/mendersoftware/mender-server/services/reporting/model"
	"github.com/mendersoftware/mender-server/services/reporting/store"
)

//go:generate ../../../../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	GetMapping(ctx context.Context, tid string) (*model.Mapping, error)
	GetSearchableInvAttrs(ctx context.Context, tid string) ([]model.FilterAttribute, error)
	AggregateDevices(ctx context.Context, aggregateParams *model.AggregateParams) (
		[]model.DeviceAggregation, error)
	SearchDevices(ctx context.Context, searchParams *model.SearchParams) (
		[]inventory.Device, int, error)
	AggregateDeployments(ctx context.Context, aggregateParams *model.AggregateDeploymentsParams) (
		[]model.DeviceAggregation, error)
	SearchDeployments(ctx context.Context, searchParams *model.DeploymentsSearchParams) (
		[]model.Deployment, int, error)
}

type app struct {
	store  store.Store
	mapper mapping.Mapper
	ds     store.DataStore
}

func NewApp(store store.Store, ds store.DataStore) App {
	mapper := mapping.NewMapper(ds)
	return &app{
		store:  store,
		mapper: mapper,
		ds:     ds,
	}
}

// HealthCheck performs a health check and returns an error if it fails
func (a *app) HealthCheck(ctx context.Context) error {
	err := a.ds.Ping(ctx)
	if err == nil {
		err = a.store.Ping(ctx)
	}
	return err
}

// GetMapping returns the mapping for the specified tenant
func (app *app) GetMapping(ctx context.Context, tid string) (*model.Mapping, error) {
	return app.ds.GetMapping(ctx, tid)
}

// AggregateDevices aggregates device data
func (app *app) AggregateDevices(
	ctx context.Context,
	aggregateParams *model.AggregateParams,
) ([]model.DeviceAggregation, error) {
	searchParams := &model.SearchParams{
		Filters:  aggregateParams.Filters,
		Groups:   aggregateParams.Groups,
		TenantID: aggregateParams.TenantID,
	}
	if err := app.mapSearchParams(ctx, searchParams); err != nil {
		return nil, err
	}
	query, err := model.BuildQuery(*searchParams)
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

	if err := app.mapAggregations(ctx, searchParams.TenantID,
		aggregateParams.Aggregations); err != nil {
		return nil, err
	}
	aggregations, err := model.BuildAggregations(aggregateParams.Aggregations)
	if err != nil {
		return nil, err
	}

	query = query.WithSize(0).With(map[string]interface{}{
		"aggs": aggregations,
	})
	esRes, err := app.store.AggregateDevices(ctx, query)
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

// storeToDeviceAggregations translates ES results directly to device aggregations
func (a *app) storeToDeviceAggregations(
	ctx context.Context, tenantID string, aggregationsS map[string]interface{},
) ([]model.DeviceAggregation, error) {
	aggs := []model.DeviceAggregation{}
	for name, aggregationS := range aggregationsS {
		if _, ok := aggregationS.(map[string]interface{}); !ok {
			continue
		}
		bucketsS, ok := aggregationS.(map[string]interface{})["buckets"].([]interface{})
		if !ok {
			continue
		}
		items := make([]model.DeviceAggregationItem, 0, len(bucketsS))
		for _, bucket := range bucketsS {
			bucketMap, ok := bucket.(map[string]interface{})
			if !ok {
				return nil, errors.New("can't process store bucket item")
			}
			key, ok := bucketMap["key"].(string)
			if !ok {
				return nil, errors.New("can't process store key attribute")
			}
			count, ok := bucketMap["doc_count"].(float64)
			if !ok {
				return nil, errors.New("can't process store doc_count attribute")
			}
			item := model.DeviceAggregationItem{
				Key:   key,
				Count: int(count),
			}
			subaggs, err := a.storeToDeviceAggregations(ctx, tenantID, bucketMap)
			if err == nil && len(subaggs) > 0 {
				item.Aggregations = subaggs
			}
			items = append(items, item)
		}

		otherCount := 0
		if count, ok := aggregationS.(map[string]interface{})["sum_other_doc_count"].(float64); ok {
			otherCount = int(count)
		}

		aggs = append(aggs, model.DeviceAggregation{
			Name:       name,
			Items:      items,
			OtherCount: otherCount,
		})
	}
	return aggs, nil
}

// SearchDevices searches device data
func (app *app) SearchDevices(
	ctx context.Context,
	searchParams *model.SearchParams,
) ([]inventory.Device, int, error) {
	if err := app.mapSearchParams(ctx, searchParams); err != nil {
		return nil, 0, err
	}
	query, err := model.BuildQuery(*searchParams)
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
				model.FieldNameID: searchParams.DeviceIDs,
			},
		})
	}

	esRes, err := app.store.SearchDevices(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	res, total, err := app.storeToInventoryDevs(ctx, searchParams.TenantID, esRes)
	if err != nil {
		return nil, 0, err
	}

	return res, total, err
}

func (app *app) mapAggregations(ctx context.Context, tenantID string,
	aggregations []model.AggregationTerm) error {
	attributes := make(inventory.DeviceAttributes, 0, len(aggregations))
	for i := range aggregations {
		attributes = append(attributes, inventory.DeviceAttribute{
			Name:  aggregations[i].Attribute,
			Scope: aggregations[i].Scope,
		})
	}
	attributes, err := app.mapper.MapInventoryAttributes(ctx, tenantID,
		attributes, false, true)
	if err == nil {
		for i, attr := range attributes {
			aggregations[i].Attribute = attr.Name
			aggregations[i].Scope = attr.Scope
			if len(aggregations[i].Aggregations) > 0 {
				err = app.mapAggregations(ctx, tenantID, aggregations[i].Aggregations)
				if err != nil {
					break
				}
			}
		}
	}
	return err
}

func (app *app) mapSearchParams(ctx context.Context, searchParams *model.SearchParams) error {
	if len(searchParams.Filters) > 0 {
		attributes := make(inventory.DeviceAttributes, 0, len(searchParams.Attributes))
		for i := 0; i < len(searchParams.Filters); i++ {
			attributes = append(attributes, inventory.DeviceAttribute{
				Name:        searchParams.Filters[i].Attribute,
				Scope:       searchParams.Filters[i].Scope,
				Value:       searchParams.Filters[i].Value,
				Description: &searchParams.Filters[i].Type,
			})
		}
		attributes, err := app.mapper.MapInventoryAttributes(ctx, searchParams.TenantID,
			attributes, false, true)
		if err != nil {
			return err
		}
		searchParams.Filters = make([]model.FilterPredicate, 0, len(searchParams.Filters))
		for _, attribute := range attributes {
			searchParams.Filters = append(searchParams.Filters, model.FilterPredicate{
				Attribute: attribute.Name,
				Scope:     attribute.Scope,
				Value:     attribute.Value,
				Type:      *attribute.Description,
			})
		}
	}
	if len(searchParams.Attributes) > 0 {
		attributes := make(inventory.DeviceAttributes, 0, len(searchParams.Attributes))
		for i := 0; i < len(searchParams.Attributes); i++ {
			attributes = append(attributes, inventory.DeviceAttribute{
				Name:  searchParams.Attributes[i].Attribute,
				Scope: searchParams.Attributes[i].Scope,
			})
		}
		attributes, err := app.mapper.MapInventoryAttributes(ctx, searchParams.TenantID,
			attributes, false, false)
		if err != nil {
			return err
		}
		searchParams.Attributes = make([]model.SelectAttribute, 0, len(searchParams.Attributes))
		for _, attribute := range attributes {
			searchParams.Attributes = append(searchParams.Attributes, model.SelectAttribute{
				Attribute: attribute.Name,
				Scope:     attribute.Scope,
			})
		}
	}
	if len(searchParams.Sort) > 0 {
		attributes := make(inventory.DeviceAttributes, 0, len(searchParams.Sort))
		for i := 0; i < len(searchParams.Sort); i++ {
			attributes = append(attributes, inventory.DeviceAttribute{
				Name:        searchParams.Sort[i].Attribute,
				Scope:       searchParams.Sort[i].Scope,
				Description: &searchParams.Sort[i].Order,
			})
		}
		attributes, err := app.mapper.MapInventoryAttributes(ctx, searchParams.TenantID,
			attributes, false, false)
		if err != nil {
			return err
		}
		searchParams.Sort = make([]model.SortCriteria, 0, len(searchParams.Attributes))
		for _, attribute := range attributes {
			searchParams.Sort = append(searchParams.Sort, model.SortCriteria{
				Attribute: attribute.Name,
				Scope:     attribute.Scope,
				Order:     *attribute.Description,
			})
		}
	}

	return nil
}

// storeToInventoryDevs translates ES results directly to inventory devices
func (a *app) storeToInventoryDevs(
	ctx context.Context, tenantID string, storeRes map[string]interface{},
) ([]inventory.Device, int, error) {
	devs := []inventory.Device{}

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
		res, err := a.storeToInventoryDev(ctx, tenantID, v)
		if err != nil {
			return nil, 0, err
		}

		devs = append(devs, *res)
	}

	return devs, int(total), nil
}

func (a *app) storeToInventoryDev(ctx context.Context, tenantID string,
	storeRes interface{}) (*inventory.Device, error) {
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

	// if query has a 'fields' clause, all results will be arrays incl. device id, so extract it
	id, ok := sourceM["id"].(string)
	if !ok {
		idarr, ok := sourceM["id"].([]interface{})
		if !ok {
			return nil, errors.New(
				"can't parse device id as neither single value nor array",
			)
		}

		id, ok = idarr[0].(string)
		if !ok {
			return nil, errors.New(
				"can't parse device id as neither single value nor array",
			)
		}
	}

	ret := &inventory.Device{
		ID: inventory.DeviceID(id),
	}
	t := getTime(sourceM, model.FieldNameCheckIn)
	if t != nil && !t.IsZero() {
		ret.LastCheckinDate = t
	}
	attrs := []inventory.DeviceAttribute{}

	for k, v := range sourceM {
		s, n, err := model.MaybeParseAttr(k)
		if err != nil {
			return nil, err
		}

		if vArray, ok := v.([]interface{}); ok && len(vArray) == 1 {
			v = vArray[0]
		}

		if n != "" {
			a := inventory.DeviceAttribute{
				Name:  model.Redot(n),
				Scope: s,
				Value: v,
			}

			if a.Scope == model.ScopeSystem &&
				a.Name == model.AttrNameUpdatedAt {
				ret.UpdatedTs = parseTime(v)
			} else if a.Scope == model.ScopeSystem &&
				a.Name == model.AttrNameCreatedAt {
				ret.CreatedTs = parseTime(v)
			}

			attrs = append(attrs, a)
		}
	}

	attributes, err := a.mapper.ReverseInventoryAttributes(ctx, tenantID, attrs)
	if err != nil {
		return nil, err
	}
	ret.Attributes = attributes

	return ret, nil
}

func getTime(m map[string]interface{}, s string) *time.Time {
	if v, ok := m[s]; ok && v != nil {
		timeString := ""
		if vString, ok := v.(string); ok && len(vString) > 0 {
			timeString = v.(string)
		}
		if vArray, ok := v.([]interface{}); ok && len(vArray) > 0 {
			timeString = v.([]interface{})[0].(string)
		}
		if len(timeString) > 0 {
			t, e := time.Parse(time.RFC3339, timeString)
			if e != nil {
				return nil
			}
			if ok {
				return &t
			}
		}
	}
	return nil
}

func parseTime(v interface{}) time.Time {
	val, _ := v.(string)
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t
	}
	return time.Time{}
}

func (app *app) GetSearchableInvAttrs(
	ctx context.Context,
	tid string,
) ([]model.FilterAttribute, error) {
	l := log.FromContext(ctx)

	index, err := app.store.GetDevicesIndexMapping(ctx, tid)
	if err != nil {
		return nil, err
	}

	// inventory attributes are under 'mappings.properties'
	mappings, ok := index["mappings"]
	if !ok {
		return nil, errors.New("can't parse index mappings")
	}

	mappingsM, ok := mappings.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index mappings")
	}

	props, ok := mappingsM["properties"]
	if !ok {
		return nil, errors.New("can't parse index properties")
	}

	propsM, ok := props.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index properties")
	}

	attrs := []inventory.DeviceAttribute{}
	for k := range propsM {
		s, n, err := model.MaybeParseAttr(k)

		if err != nil {
			return nil, err
		}

		if n != "" {
			attrs = append(attrs, inventory.DeviceAttribute{Name: n, Scope: s})
		}
	}
	attributes, err := app.mapper.ReverseInventoryAttributes(ctx, tid, attrs)
	if err != nil {
		return nil, err
	}

	ret := []model.FilterAttribute{}
	for _, attr := range attributes {
		ret = append(ret, model.FilterAttribute{Name: attr.Name, Scope: attr.Scope, Count: 1})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[j].Scope > ret[i].Scope {
			return true
		}

		if ret[j].Scope < ret[i].Scope {
			return false
		}

		return ret[j].Name > ret[i].Name
	})

	l.Debugf("parsed searchable attributes %v\n", ret)

	return ret, nil
}
