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

package config

import (
	"github.com/mendersoftware/mender-server/pkg/config"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

	// SettingOpenSearchAddresses is the config key for the opensearch addresses
	SettingOpenSearchAddresses = "opensearch_addresses"
	// SettingOpenSearchAddressesDefault is the default value for the opensearch addresses
	SettingOpenSearchAddressesDefault = "http://localhost:9200"

	// SettingOpenSearchDevicesIndexName is the config key for the opensearch devices
	// index name
	SettingOpenSearchDevicesIndexName = "opensearch_devices_index_name"
	// SettingOpenSearchDevicesIndexNameDefault is the default value for the opensearch
	// devices index name
	SettingOpenSearchDevicesIndexNameDefault = "devices"

	// SettingOpenSearchDevicesIndexShards is the config key for the opensearch devices
	// index shards
	SettingOpenSearchDevicesIndexShards = "opensearch_devices_index_shards"
	// SettingOpenSearchDevicesIndexShardsDefault is the default value for the opensearch
	// devices index shards
	SettingOpenSearchDevicesIndexShardsDefault = 1

	// SettingOpenSearchDevicesIndexReplicas is the config key for the opensearch devices
	// index replicas
	SettingOpenSearchDevicesIndexReplicas = "opensearch_devices_index_replicas"
	// SettingOpenSearchDevicesIndexReplicasDefault is the default value for the
	// opensearch devices index replicas
	SettingOpenSearchDevicesIndexReplicasDefault = 0

	// SettingOpenSearchDeploymentsIndexName is the config key for the opensearch deployments
	// index name
	SettingOpenSearchDeploymentsIndexName = "opensearch_deployments_index_name"
	// SettingOpenSearchDeploymentsIndexNameDefault is the default value for the opensearch
	// deployments index name
	SettingOpenSearchDeploymentsIndexNameDefault = "deployments"

	// SettingOpenSearchDeploymentsIndexShards is the config key for the opensearch deployments
	// index shards
	SettingOpenSearchDeploymentsIndexShards = "opensearch_deployments_index_shards"
	// SettingOpenSearchDeploymentsIndexShardsDefault is the default value for the opensearch
	// deployments index shards
	SettingOpenSearchDeploymentsIndexShardsDefault = 1

	// SettingOpenSearchDeploymentsIndexReplicas is the config key for the opensearch deployments
	// index replicas
	SettingOpenSearchDeploymentsIndexReplicas = "opensearch_deployments_index_replicas"
	// SettingOpenSearchDeploymentsIndexReplicasDefault is the default value for the
	// opensearch deployments index replicas
	SettingOpenSearchDeploymentsIndexReplicasDefault = 0

	// SettingDeploymentsAddr is the config key for the deviceauth service address
	SettingDeploymentsAddr = "deployments_addr"
	// SettingDeploymentsAddrDefault is the default value for the deployments service address
	SettingDeploymentsAddrDefault = "http://mender-deployments:8080/"

	// SettingDeviceAuthAddr is the config key for the deviceauth service address
	SettingDeviceAuthAddr = "deviceauth_addr"
	// SettingDeviceAuthAddrDefault is the default value for the deviceauth service address
	SettingDeviceAuthAddrDefault = "http://mender-device-auth:8080/"

	// SettingInventoryAddr is the config key for the inventory service address
	SettingInventoryAddr = "inventory_addr"
	// SettingInventoryAddrDefault is the default value for the inventory service address
	SettingInventoryAddrDefault = "http://mender-inventory:8080/"

	// SettingMongo is the config key for the mongo URL
	SettingMongo = "mongo_url"
	// SettingMongoDefault is the default value for the mongo URL
	SettingMongoDefault = "mongodb://mender-mongo:27017"

	// SettingDbName is the config key for the mongo database name
	SettingDbName = "mongo_dbname"
	// SettingDbNameDefault is the default value for the mongo database name
	SettingDbNameDefault = "reporting"

	// SettingDbSSL is the config key for the mongo SSL setting
	SettingDbSSL = "mongo_ssl"
	// SettingDbSSLDefault is the default value for the mongo SSL setting
	SettingDbSSLDefault = false

	// SettingDbSSLSkipVerify is the config key for the mongo SSL skip verify setting
	SettingDbSSLSkipVerify = "mongo_ssl_skipverify"
	// SettingDbSSLSkipVerifyDefault is the default value for the mongo SSL skip verify setting
	SettingDbSSLSkipVerifyDefault = false

	// SettingDbUsername is the config key for the mongo username
	SettingDbUsername = "mongo_username"

	// SettingDbPassword is the config key for the mongo password
	SettingDbPassword = "mongo_password"

	// SettingNatsURI is the config key for the nats uri
	SettingNatsURI = "nats_uri"
	// SettingNatsURIDefault is the default value for the nats uri
	SettingNatsURIDefault = "nats://mender-nats:4222"

	// SettingNatsStreamName is the config key for the nats streaem name
	SettingNatsStreamName = "nats_stream_name"
	// SettingNatsStreamNameDefault is the default value for the nats stream name
	SettingNatsStreamNameDefault = "WORKFLOWS"

	// SettingNatsSubscriberTopic is the config key for the nats subscriber topic name
	SettingNatsSubscriberTopic = "nats_subscriber_topic"
	// SettingNatsSubscriberTopicDefault is the default value for the nats subscriber topic name
	SettingNatsSubscriberTopicDefault = "reporting"

	// SettingNatsSubscriberDurable is the config key for the nats subscriber durable name
	SettingNatsSubscriberDurable = "nats_subscriber_durable"
	// SettingNatsSubscriberDurableDefault is the default value for the nats subscriber durable
	// name
	SettingNatsSubscriberDurableDefault = "reporting"

	// SettingReindexBatchSize is the num of buffered requests processed together
	SettingReindexBatchSize        = "reindex_batch_size"
	SettingReindexBatchSizeDefault = 100

	// SettingWorkerConcurrency defines the number of concurrent worker
	// threads that exist at the same time (defaults to 10)
	SettingWorkerConcurrency        = "worker_concurrency"
	SettingWorkerConcurrencyDefault = 10

	// SettingReindexTimeMsec is the max time after which reindexing is triggered
	// (even if buffered requests didn't reach reindex_batch_size yet)
	SettingReindexMaxTimeMsec        = "reindex_max_time_msec"
	SettingReindexMaxTimeMsecDefault = 1000

	// SettingDebugLog is the config key for the truning on the debug log
	SettingDebugLog = "debug_log"
	// SettingDebugLogDefault is the default value for the debug log enabling
	SettingDebugLogDefault = false
)

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingOpenSearchAddresses, Value: SettingOpenSearchAddressesDefault},
		{Key: SettingOpenSearchDevicesIndexName,
			Value: SettingOpenSearchDevicesIndexNameDefault},
		{Key: SettingOpenSearchDevicesIndexShards,
			Value: SettingOpenSearchDevicesIndexShardsDefault},
		{Key: SettingOpenSearchDevicesIndexReplicas,
			Value: SettingOpenSearchDevicesIndexReplicasDefault},
		{Key: SettingOpenSearchDeploymentsIndexName,
			Value: SettingOpenSearchDeploymentsIndexNameDefault},
		{Key: SettingOpenSearchDeploymentsIndexShards,
			Value: SettingOpenSearchDeploymentsIndexShardsDefault},
		{Key: SettingOpenSearchDeploymentsIndexReplicas,
			Value: SettingOpenSearchDeploymentsIndexReplicasDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingDeploymentsAddr, Value: SettingDeploymentsAddrDefault},
		{Key: SettingDeviceAuthAddr, Value: SettingDeviceAuthAddrDefault},
		{Key: SettingInventoryAddr, Value: SettingInventoryAddrDefault},
		{Key: SettingMongo, Value: SettingMongoDefault},
		{Key: SettingDbName, Value: SettingDbNameDefault},
		{Key: SettingNatsURI, Value: SettingNatsURIDefault},
		{Key: SettingNatsStreamName, Value: SettingNatsStreamNameDefault},
		{Key: SettingNatsSubscriberTopic, Value: SettingNatsSubscriberTopicDefault},
		{Key: SettingNatsSubscriberDurable, Value: SettingNatsSubscriberDurableDefault},
		{Key: SettingReindexMaxTimeMsec, Value: SettingReindexMaxTimeMsecDefault},
		{Key: SettingReindexBatchSize, Value: SettingReindexBatchSizeDefault},
		{Key: SettingWorkerConcurrency, Value: SettingWorkerConcurrencyDefault},
	}
)
