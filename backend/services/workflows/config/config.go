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
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"

	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

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
	SettingNatsSubscriberTopicDefault = "default"

	// SettingNatsSubscriberDurable is the config key for the nats subscriber durable name
	SettingNatsSubscriberDurable = "nats_subscriber_durable"
	// SettingNatsSubscriberDurableDefault is the default value for the nats subscriber durable name
	SettingNatsSubscriberDurableDefault = "workflows-worker"

	SettingsNats = "nats"

	SettingNatsConsumer                  = SettingsNats + ".consumer"
	SettingNatsConsumerAckWait           = SettingNatsConsumer + ".ack_wait"
	SettingNatsConsumerAckWaitDefault    = "30s"
	SettingNatsConsumerMaxDeliver        = SettingNatsConsumer + ".max_deliver"
	SettingNatsConsumerMaxDeliverDefault = 3
	SettingNatsConsumerMaxPending        = SettingNatsConsumer + ".max_pending"
	SettingNatsConsumerMaxPendingDefault = 1000

	SettingNatsJetStream                = SettingsNats + ".stream"
	SettingNatsJetStreamReplicas        = SettingNatsJetStream + ".replicas"
	SettingNatsJetStreamReplicasDefault = 3

	// SettingMongo is the config key for the mongo URLz
	SettingMongo = "mongo-url"
	// SettingMongoDefault is the default value for the mongo URL
	SettingMongoDefault = "mongodb://mender-mongo:27017"

	// SettingDbName is the config key for the mongo database name
	SettingDbName = "mongo-dbname"
	// SettingDbNameDefault is the default value for the mongo database name
	SettingDbNameDefault = "workflows"

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

	// SettingSMTPHost is the config key for the SMTP host
	SettingSMTPHost = "smtp_host"
	// SettingSMTPHostDefault is the default value for the SMTP host
	SettingSMTPHostDefault = "localhost:25"

	// SettingSMTPAuthMechanism is the config key for the SMTP auth mechanism
	SettingSMTPAuthMechanism = "smtp_auth_mechanism"
	// SettingSMTPAuthMechanismDefault is the default value for the SMTP auth mechanism
	SettingSMTPAuthMechanismDefault = "PLAIN"

	// SettingSMTPUsername is the config key for the SMTP username
	SettingSMTPUsername = "smtp_username"

	// SettingSMTPPassword is the config key for the SMTP password
	SettingSMTPPassword = "smtp_password"

	// SettingWorkflowsPath is the config key for the workflows path
	SettingWorkflowsPath = "workflows_path"

	// SettingConcurrency is the config key for the concurrency limit
	SettingConcurrency = "concurrency"
	// SettingConcurrencyDefault is the default value for the concurrency limit
	SettingConcurrencyDefault = 10

	// SettingDebugLog is the config key for the truning on the debug log
	SettingDebugLog = "debug_log"
	// SettingDebugLogDefault is the default value for the debug log enabling
	SettingDebugLogDefault = false
)

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingNatsURI, Value: SettingNatsURIDefault},
		{Key: SettingNatsStreamName, Value: SettingNatsStreamNameDefault},
		{Key: SettingNatsSubscriberTopic, Value: SettingNatsSubscriberTopicDefault},
		{Key: SettingNatsSubscriberDurable, Value: SettingNatsSubscriberDurableDefault},
		{Key: SettingNatsJetStreamReplicas, Value: SettingNatsJetStreamReplicasDefault},
		{Key: SettingMongo, Value: SettingMongoDefault},
		{Key: SettingDbName, Value: SettingDbNameDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingSMTPHost, Value: SettingSMTPHostDefault},
		{Key: SettingSMTPAuthMechanism, Value: SettingSMTPAuthMechanismDefault},
		{Key: SettingConcurrency, Value: SettingConcurrencyDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingNatsConsumerAckWait, Value: SettingNatsConsumerAckWaitDefault},
		{Key: SettingNatsConsumerMaxDeliver, Value: SettingNatsConsumerMaxDeliverDefault},
		{Key: SettingNatsConsumerMaxPending, Value: SettingNatsConsumerMaxPendingDefault},
	}
)

func GetNatsConsumerConfig(c config.Reader) (consumer nats.ConsumerConfig, err error) {
	streamName := c.GetString(SettingNatsStreamName)
	consumer.Filter = streamName + "." + c.GetString(SettingNatsSubscriberTopic)
	consumer.AckWait = c.GetDuration(SettingNatsConsumerAckWait)
	consumer.MaxDeliver = c.GetInt(SettingNatsConsumerMaxDeliver)
	consumer.MaxPending = c.GetInt(SettingNatsConsumerMaxPending)
	return consumer, errors.WithMessage(
		consumer.Validate(),
		`invalid settings "nats.consumer"`,
	)
}
