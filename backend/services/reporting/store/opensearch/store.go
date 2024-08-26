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

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	_ "github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/model"
	"github.com/mendersoftware/mender-server/services/reporting/store"
)

type StoreOption func(*opensearchStore)

type opensearchStore struct {
	addresses                []string
	devicesIndexName         string
	devicesIndexShards       int
	devicesIndexReplicas     int
	deploymentsIndexName     string
	deploymentsIndexShards   int //nolint:unused //FIXME: this is field is never used
	deploymentsIndexReplicas int //nolint:unused //FIXME: this is field is never used
	client                   *opensearch.Client
}

func NewStore(opts ...StoreOption) (store.Store, error) {
	store := &opensearchStore{}
	for _, opt := range opts {
		opt(store)
	}

	cfg := opensearch.Config{
		Addresses: store.addresses,
	}
	osClient, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OpenSearch configuration")
	}

	store.client = osClient
	return store, nil
}

func WithServerAddresses(addresses []string) StoreOption {
	return func(s *opensearchStore) {
		s.addresses = addresses
	}
}

func WithDevicesIndexName(indexName string) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexName = indexName
	}
}

func WithDevicesIndexShards(indexShards int) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexShards = indexShards
	}
}

func WithDevicesIndexReplicas(indexReplicas int) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexReplicas = indexReplicas
	}
}

func WithDeploymentsIndexName(indexName string) StoreOption {
	return func(s *opensearchStore) {
		s.deploymentsIndexName = indexName
	}
}

func WithDeploymentsIndexShards(indexShards int) StoreOption {
	return func(s *opensearchStore) {
		s.deploymentsIndexShards = indexShards
	}
}

func WithDeploymentsIndexReplicas(indexReplicas int) StoreOption {
	return func(s *opensearchStore) {
		s.deploymentsIndexReplicas = indexReplicas
	}
}

type BulkAction struct {
	Type string
	Desc *BulkActionDesc
}

type BulkActionDesc struct {
	ID            string `json:"_id"`
	Index         string `json:"_index"`
	IfSeqNo       int64  `json:"_if_seq_no"`
	IfPrimaryTerm int64  `json:"_if_primary_term"`
	Routing       string `json:"routing"`
	Tenant        string
}

type BulkItem struct {
	Action *BulkAction
	Doc    interface{}
}

func (bad BulkActionDesc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID      string `json:"_id"`
		Index   string `json:"_index"`
		Routing string `json:"routing"`
	}{
		ID:      bad.ID,
		Index:   bad.Index,
		Routing: bad.Routing,
	})
}

func (ba BulkAction) MarshalJSON() ([]byte, error) {
	a := map[string]*BulkActionDesc{
		ba.Type: ba.Desc,
	}
	return json.Marshal(a)
}

func (bi BulkItem) Marshal() ([]byte, error) {
	action, err := json.Marshal(bi.Action)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(action)
	buf.WriteString("\n")

	if bi.Doc == nil {
		return buf.Bytes(), nil
	}

	if bi.Doc != nil {
		doc, err := json.Marshal(bi.Doc)
		if err != nil {
			return nil, err
		}
		buf.Write(doc)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

func (s *opensearchStore) BulkIndexDeployments(ctx context.Context,
	deployments []*model.Deployment) error {
	var data strings.Builder

	for _, deployment := range deployments {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "index",
			Desc: &BulkActionDesc{
				ID:      deployment.ID,
				Index:   s.GetDeploymentsIndex(deployment.TenantID),
				Routing: s.GetDeploymentsRoutingKey(deployment.TenantID),
			},
		})
		if err != nil {
			return err
		}
		deploymentJSON, err := json.Marshal(deployment)
		if err != nil {
			return err
		}
		data.WriteString(string(actionJSON) + "\n" + string(deploymentJSON) + "\n")
	}
	dataString := data.String()

	l := log.FromContext(ctx)
	l.Debugf("opensearch request: %s", dataString)

	req := opensearchapi.BulkRequest{
		Body: strings.NewReader(dataString),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	return nil
}

func (s *opensearchStore) BulkIndexDevices(ctx context.Context, devices []*model.Device,
	removedDevices []*model.Device) error {
	var data strings.Builder

	for _, device := range devices {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "index",
			Desc: &BulkActionDesc{
				ID:      device.GetID(),
				Index:   s.GetDevicesIndex(device.GetTenantID()),
				Routing: s.GetDevicesRoutingKey(device.GetTenantID()),
			},
		})
		if err != nil {
			return err
		}
		deviceJSON, err := json.Marshal(device)
		if err != nil {
			return err
		}
		data.WriteString(string(actionJSON) + "\n" + string(deviceJSON) + "\n")
	}
	for _, device := range removedDevices {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "delete",
			Desc: &BulkActionDesc{
				ID:      device.GetID(),
				Index:   s.GetDevicesIndex(device.GetTenantID()),
				Routing: s.GetDevicesRoutingKey(device.GetTenantID()),
			},
		})
		if err != nil {
			return err
		}
		data.WriteString(string(actionJSON) + "\n")
	}

	dataString := data.String()

	l := log.FromContext(ctx)
	l.Debugf("opensearch request: %s", dataString)

	req := opensearchapi.BulkRequest{
		Body: strings.NewReader(dataString),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	return nil
}

func (s *opensearchStore) Migrate(ctx context.Context) error {
	indexName := s.GetDevicesIndex("")
	template := fmt.Sprintf(indexDevicesTemplate,
		indexName,
		s.devicesIndexShards,
		s.devicesIndexReplicas,
	)
	err := s.migratePutIndexTemplate(ctx, indexName, template)
	if err == nil {
		err = s.migrateCreateIndex(ctx, indexName)
	}
	if err == nil {
		indexName = s.GetDeploymentsIndex("")
		template = fmt.Sprintf(indexDeploymentsTemplate,
			indexName,
			s.devicesIndexShards,
			s.devicesIndexReplicas,
		)
		err = s.migratePutIndexTemplate(ctx, indexName, template)
	}
	if err == nil {
		err = s.migrateCreateIndex(ctx, indexName)
	}
	return err
}

func (s *opensearchStore) migratePutIndexTemplate(ctx context.Context,
	indexName, template string) error {
	l := log.FromContext(ctx)
	l.Infof("put the index template for %s", indexName)

	req := opensearchapi.IndicesPutIndexTemplateRequest{
		Name: indexName,
		Body: strings.NewReader(template),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to put the index template")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return errors.Errorf("failed to set up the index template: %s", string(body))
	}
	return nil
}

func (s *opensearchStore) migrateCreateIndex(ctx context.Context, indexName string) error {
	l := log.FromContext(ctx)
	l.Infof("verify if the index %s exists", indexName)

	req := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to verify the index")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		l.Infof("create the index %s", indexName)

		req := opensearchapi.IndicesCreateRequest{
			Index: indexName,
		}
		res, err := req.Do(ctx, s.client)
		if err != nil {
			return errors.Wrap(err, "failed to create the index")
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return errors.New("failed to create the index")
		}
	} else if res.StatusCode != http.StatusOK {
		return errors.New("failed to verify the index")
	}

	return nil
}

func (s *opensearchStore) Ping(ctx context.Context) error {
	pingRequest := s.client.Ping.WithContext(ctx)
	_, err := s.client.Ping(pingRequest)
	return errors.Wrap(err, "failed to ping opensearch")
}

func (s *opensearchStore) AggregateDevices(ctx context.Context,
	query model.Query) (model.M, error) {
	id := identity.FromContext(ctx)
	indexName := s.GetDevicesIndex(id.Tenant)
	routingKey := s.GetDevicesRoutingKey(id.Tenant)
	return s.aggregate(ctx, indexName, routingKey, query)
}

func (s *opensearchStore) AggregateDeployments(ctx context.Context,
	query model.Query) (model.M, error) {
	id := identity.FromContext(ctx)
	indexName := s.GetDeploymentsIndex(id.Tenant)
	routingKey := s.GetDeploymentsRoutingKey(id.Tenant)
	return s.aggregate(ctx, indexName, routingKey, query)
}

func (s *opensearchStore) aggregate(ctx context.Context, indexName, routingKey string,
	query model.Query) (model.M, error) {
	l := log.FromContext(ctx)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	l.Debugf("es query: %v", buf.String())

	searchRequests := []func(*opensearchapi.SearchRequest){
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(false),
	}
	if routingKey != "" {
		searchRequests = append(searchRequests, s.client.Search.WithRouting(routingKey))
	}
	resp, err := s.client.Search(searchRequests...)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	var ret map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	l.Debugf("opensearch response: %v", ret)
	return ret, nil
}

func (s *opensearchStore) SearchDevices(ctx context.Context, query model.Query) (model.M, error) {
	id := identity.FromContext(ctx)
	indexName := s.GetDevicesIndex(id.Tenant)
	routingKey := s.GetDevicesRoutingKey(id.Tenant)
	return s.search(ctx, indexName, routingKey, query)
}

func (s *opensearchStore) SearchDeployments(ctx context.Context,
	query model.Query) (model.M, error) {
	id := identity.FromContext(ctx)
	indexName := s.GetDeploymentsIndex(id.Tenant)
	routingKey := s.GetDeploymentsRoutingKey(id.Tenant)
	return s.search(ctx, indexName, routingKey, query)
}

func (s *opensearchStore) search(ctx context.Context, indexName, routingKey string,
	query model.Query) (model.M, error) {
	l := log.FromContext(ctx)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	l.Debugf("es query: %v", buf.String())

	searchRequests := []func(*opensearchapi.SearchRequest){
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	}
	if routingKey != "" {
		searchRequests = append(searchRequests, s.client.Search.WithRouting(routingKey))
	}
	resp, err := s.client.Search(searchRequests...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	var ret map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	l.Debugf("opensearch response: %v", ret)
	return ret, nil
}

// GetDevicesIndexMapping retrieves the "devices*" index definition for tenant 'tid'
// existing fields, incl. inventory attributes, are found under 'properties'
// see: https://opensearch.org/docs/latest/api-reference/index-apis/get-index/
func (s *opensearchStore) GetDevicesIndexMapping(ctx context.Context,
	tid string) (map[string]interface{}, error) {
	idx := s.GetDevicesIndex(tid)
	return s.getIndexMapping(ctx, tid, idx)
}

// GetDeploymentsIndexMapping retrieves the "deployments*" index definition for tenant 'tid'
// existing fields, incl. inventory attributes, are found under 'properties'
// see: https://opensearch.org/docs/latest/api-reference/index-apis/get-index/
func (s *opensearchStore) GetDeploymentsIndexMapping(ctx context.Context,
	tid string) (map[string]interface{}, error) {
	idx := s.GetDeploymentsIndex(tid)
	return s.getIndexMapping(ctx, tid, idx)
}

func (s *opensearchStore) getIndexMapping(ctx context.Context,
	tid, idx string) (map[string]interface{}, error) {
	l := log.FromContext(ctx)
	req := opensearchapi.IndicesGetRequest{
		Index: []string{idx},
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get devices index from store, tid %s", tid)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.Errorf(
			"failed to get devices index from store, tid %s, code %d",
			tid, res.StatusCode,
		)
	}

	var indexRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indexRes); err != nil {
		return nil, err
	}

	index, ok := indexRes[idx]
	if !ok {
		return nil, errors.New("can't parse index defintion response")
	}

	indexM, ok := index.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index defintion response")
	}

	l.Debugf("index for tid %s\n%s\n", tid, indexM)
	return indexM, nil
}

// GetDevicesIndex returns the index name for the tenant tid
func (s *opensearchStore) GetDevicesIndex(tid string) string {
	return s.devicesIndexName
}

// GetDeploymentsIndex returns the index name for the tenant tid
func (s *opensearchStore) GetDeploymentsIndex(tid string) string {
	return s.deploymentsIndexName
}

// GetDevicesRoutingKey returns the routing key for the tenant tid
func (s *opensearchStore) GetDevicesRoutingKey(tid string) string {
	return tid
}

// GetDeploymentsRoutingKey returns the routing key for the tenant tid
func (s *opensearchStore) GetDeploymentsRoutingKey(tid string) string {
	return tid
}
