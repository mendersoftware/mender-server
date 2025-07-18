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
package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
)

const (
	EnvProd = "prod"
	EnvDev  = "dev"

	SettingHttps            = "https"
	SettingHttpsCertificate = SettingHttps + ".certificate"
	SettingHttpsKey         = SettingHttps + ".key"

	SettingListen        = "listen"
	SettingListenDefault = ":8080"

	SettingStorage = "storage"

	SettingDefaultStorage                = SettingStorage + ".default"
	SettingDefaultStorageDefault         = "aws"
	SettingStorageBucket                 = SettingStorage + ".bucket"
	SettingStorageBucketDefault          = "mender-artifact-storage"
	SettingStorageMaxImageSize           = SettingStorage + ".max_image_size"
	SettingStorageMaxImageSizeDefault    = 10 * 1024 * 1024 * 1024 // 10 GiB
	SettingStorageMaxGenerateSize        = SettingStorage + ".max_generate_data_size"
	SettingStorageMaxGenerateSizeDefault = 512 * 1024 * 1024 // 512 MiB

	SettingStorageProxyURI = SettingStorage + ".proxy_uri"

	SettingStorageEnableDirectUpload        = SettingStorage + ".enable_direct_upload"
	SettingStorageEnableDirectUploadDefault = false

	SettingStorageDirectUploadSkipVerify        = SettingStorage + ".direct_upload_skip_verify"
	SettingStorageDirectUploadSkipVerifyDefault = false

	SettingsStorageDownloadExpireSeconds        = SettingStorage + ".download_expire_seconds"
	SettingsStorageDownloadExpireSecondsDefault = 900
	SettingsStorageUploadExpireSeconds          = SettingStorage + ".upload_expire_seconds"
	SettingsStorageUploadExpireSecondsDefault   = 3600

	SettingsAws                       = "aws"
	SettingAwsS3Region                = SettingsAws + ".region"
	SettingAwsS3RegionDefault         = "us-east-1"
	SettingAwsS3ForcePathStyle        = SettingsAws + ".force_path_style"
	SettingAwsS3ForcePathStyleDefault = true
	SettingAwsS3UseAccelerate         = SettingsAws + ".use_accelerate"
	SettingAwsS3UseAccelerateDefault  = false
	SettingAwsURI                     = SettingsAws + ".uri"
	SettingAwsExternalURI             = SettingsAws + ".external_uri"
	SettingAwsUnsignedHeaders         = SettingsAws + ".unsigned_headers"
	SettingAwsUnsignedHeadersDefault  = "Accept-Encoding"

	SettingsAwsTagArtifact        = SettingsAws + ".tag_artifact"
	SettingsAwsTagArtifactDefault = false

	SettingsAwsAuth      = SettingsAws + ".auth"
	SettingAwsAuthKeyId  = SettingsAwsAuth + ".key"
	SettingAwsAuthSecret = SettingsAwsAuth + ".secret"
	SettingAwsAuthToken  = SettingsAwsAuth + ".token"

	SettingAzure                    = "azure"
	SettingAzureAuth                = SettingAzure + ".auth"
	SettingAzureConnectionString    = SettingAzureAuth + ".connection_string"
	SettingAzureSharedKey           = SettingAzureAuth + ".shared_key"
	SettingAzureSharedKeyAccount    = SettingAzureSharedKey + ".account_name"
	SettingAzureSharedKeyAccountKey = SettingAzureSharedKey + ".account_key"
	SettingAzureSharedKeyURI        = SettingAzureSharedKey + ".uri"

	SettingMongo        = "mongo-url"
	SettingMongoDefault = "mongodb://mongo-deployments:27017"

	SettingDbSSL        = "mongo_ssl"
	SettingDbSSLDefault = false

	SettingDbSSLSkipVerify        = "mongo_ssl_skipverify"
	SettingDbSSLSkipVerifyDefault = false

	SettingDbUsername = "mongo_username"
	SettingDbPassword = "mongo_password"

	SettingWorkflows        = "mender-workflows"
	SettingWorkflowsDefault = "http://mender-workflows-server:8080"

	SettingMiddleware        = "middleware"
	SettingMiddlewareDefault = EnvProd

	SettingInventoryAddr        = "inventory_addr"
	SettingInventoryAddrDefault = "http://mender-inventory:8080"

	SettingReportingAddr        = "reporting_addr"
	SettingReportingAddrDefault = ""

	SettingInventoryTimeout        = "inventory_timeout"
	SettingInventoryTimeoutDefault = 10

	// SettingPresignAlgorithm sets the algorithm used for signing
	// downloadable URLs. This option is currently ignored.
	SettingPresignAlgorithm        = "presign.algorithm"
	SettingPresignAlgorithmDefault = "HMAC256"

	// SettingPresignSecret sets the secret for generating signed url.
	// For HMAC type of algorithms the value must be a base64 encoded
	// secret. For public key signatures, the value must be a path to
	// the private key (not yet supported).
	SettingPresignSecret        = "presign.secret"
	SettingPresignSecretDefault = ""

	// SettingPresignExpireSeconds sets the amount of seconds it takes for
	// the signed URL to expire.
	SettingPresignExpireSeconds        = "presign.expire_seconds"
	SettingPresignExpireSecondsDefault = 900

	// SettingPresignHost sets the URL hostname (pointing to the gateway)
	// for the generated URL. If the configuration option is left blank
	// (default), it will try to use the X-Forwarded-Host header forwarded
	// by the proxy.
	SettingPresignHost        = "presign.url_hostname"
	SettingPresignHostDefault = ""

	// SettingPresignURLScheme sets the URL scheme used for generating the
	// pre-signed url.
	SettingPresignScheme        = "presign.url_scheme"
	SettingPresignSchemeDefault = "https"

	// SettingDisableNewReleasesFeature is a flag that turns off the new API end-points
	// related to releases; helpful in performing long-running maintenance and data
	// migrations on the artifacts and releases collections.
	SettingDisableNewReleasesFeature        = "disable_new_releases_feature"
	SettingDisableNewReleasesFeatureDefault = false

	// Max Request body size
	SettingMaxRequestSize        = "request_size_limit"
	SettingMaxRequestSizeDefault = 1024 * 1024 // 1 MiB
)

const (
	StorageTypeAWS   = "aws"
	StorageTypeAzure = "azure"
)

const (
	deprecatedSettingAwsS3Bucket               = SettingsAws + ".bucket"
	deprecatedSettingAwsS3MaxImageSize         = SettingsAws + ".max_image_size"
	deprecatedSettingsAwsDownloadExpireSeconds = SettingsAws + ".download_expire_seconds"
	deprecatedSettingsAwsUploadExpireSeconds   = SettingsAws + ".upload_expire_seconds"
)

// ValidateAwsAuth validates configuration of SettingsAwsAuth section if provided.
func ValidateAwsAuth(c config.Reader) error {

	if c.IsSet(SettingsAwsAuth) {
		required := []string{SettingAwsAuthKeyId, SettingAwsAuthSecret}
		for _, key := range required {
			if !c.IsSet(key) {
				return MissingOptionError(key)
			}

			if c.GetString(key) == "" {
				return MissingOptionError(key)
			}
		}
	}

	return nil
}

// ValidateHttps validates configuration of SettingHttps section if provided.
func ValidateHttps(c config.Reader) error {

	if c.IsSet(SettingHttps) {
		required := []string{SettingHttpsCertificate, SettingHttpsKey}
		for _, key := range required {
			if !c.IsSet(key) {
				return MissingOptionError(key)
			}

			value := c.GetString(key)
			if value == "" {
				return MissingOptionError(key)
			}

			if _, err := os.Stat(value); err != nil {
				return err
			}
		}
	}

	return nil
}

func ValidateStorage(c config.Reader) error {
	svc := c.GetString(SettingDefaultStorage)
	if svc != StorageTypeAWS && svc != StorageTypeAzure {
		return fmt.Errorf(
			`setting "%s" (%s) must be one of "aws" or "azure"`,
			SettingDefaultStorage, svc,
		)
	}
	return nil
}

// Generate error with missing required option message.
func MissingOptionError(option string) error {
	return fmt.Errorf("Required option: '%s'", option)
}

func applyAliases() {
	for _, alias := range Aliases {
		if config.Config.IsSet(alias.Alias) {
			config.Config.Set(alias.Key, config.Config.Get(alias.Alias))
		}
	}
}

func Setup(configPath string) error {
	err := config.FromConfigFile(configPath, Defaults)
	if err != nil {
		return fmt.Errorf("error loading configuration: %s", err)
	}

	// Enable setting config values by environment variables
	config.Config.SetEnvPrefix("DEPLOYMENTS")
	config.Config.AutomaticEnv()
	config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	if err := config.ValidateConfig(config.Config, Validators...); err != nil {
		return errors.WithMessage(err, "config: error validating configuration")
	}
	if config.Config.Get(SettingPresignSecret) == "" {
		log.NewEmpty().Warnf("'%s' not configured. Generating a random secret.",
			SettingPresignSecret,
		)
		var buf [32]byte
		n, err := io.ReadFull(rand.Reader, buf[:])
		if err != nil {
			return errors.WithMessagef(err,
				"failed to generate '%s'",
				SettingPresignSecret,
			)
		} else if n == 0 {
			return errors.Errorf(
				"failed to generate '%s'",
				SettingPresignSecret,
			)
		}
		secret := base64.StdEncoding.EncodeToString(buf[:n])
		config.Config.Set(SettingPresignSecret, secret)
	}
	applyAliases()
	return nil
}

var (
	Validators = []config.Validator{ValidateAwsAuth, ValidateHttps, ValidateStorage}
	// Aliases for deprecated configuration names to preserve backward compatibility.
	Aliases = []struct {
		Key   string
		Alias string
	}{
		{Key: SettingStorageBucket, Alias: deprecatedSettingAwsS3Bucket},
		{Key: SettingsStorageDownloadExpireSeconds,
			Alias: deprecatedSettingsAwsDownloadExpireSeconds},
		{Key: SettingsStorageUploadExpireSeconds, Alias: deprecatedSettingsAwsUploadExpireSeconds},
		{Key: SettingStorageMaxImageSize, Alias: deprecatedSettingAwsS3MaxImageSize},
	}

	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingDefaultStorage, Value: SettingDefaultStorageDefault},
		{Key: SettingAwsS3Region, Value: SettingAwsS3RegionDefault},
		{Key: SettingStorageBucket, Value: SettingStorageBucketDefault},
		{Key: SettingStorageDirectUploadSkipVerify,
			Value: SettingStorageDirectUploadSkipVerifyDefault},
		{Key: SettingStorageEnableDirectUpload, Value: SettingStorageEnableDirectUploadDefault},
		{Key: SettingAwsS3ForcePathStyle, Value: SettingAwsS3ForcePathStyleDefault},
		{Key: SettingAwsS3UseAccelerate, Value: SettingAwsS3UseAccelerateDefault},
		{Key: SettingAwsUnsignedHeaders, Value: SettingAwsUnsignedHeadersDefault},
		{Key: SettingStorageMaxImageSize, Value: SettingStorageMaxImageSizeDefault},
		{Key: SettingStorageMaxGenerateSize, Value: SettingStorageMaxGenerateSizeDefault},
		{Key: SettingsStorageDownloadExpireSeconds,
			Value: SettingsStorageDownloadExpireSecondsDefault},
		{Key: SettingsStorageUploadExpireSeconds, Value: SettingsStorageUploadExpireSecondsDefault},
		{Key: SettingMongo, Value: SettingMongoDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingWorkflows, Value: SettingWorkflowsDefault},
		{Key: SettingsAwsTagArtifact, Value: SettingsAwsTagArtifactDefault},
		{Key: SettingInventoryAddr, Value: SettingInventoryAddrDefault},
		{Key: SettingReportingAddr, Value: SettingReportingAddrDefault},
		{Key: SettingInventoryTimeout, Value: SettingInventoryTimeoutDefault},
		{Key: SettingPresignAlgorithm, Value: SettingPresignAlgorithmDefault},
		{Key: SettingPresignSecret, Value: SettingPresignSecretDefault},
		{Key: SettingPresignExpireSeconds, Value: SettingPresignExpireSecondsDefault},
		{Key: SettingPresignHost, Value: SettingPresignHostDefault},
		{Key: SettingPresignScheme, Value: SettingPresignSchemeDefault},
		{Key: SettingDisableNewReleasesFeature, Value: SettingDisableNewReleasesFeatureDefault},
		{Key: SettingMaxRequestSize, Value: SettingMaxRequestSizeDefault},
	}
)
