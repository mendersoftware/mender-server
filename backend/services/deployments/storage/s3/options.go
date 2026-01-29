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

package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mendersoftware/mender-server/pkg/utils/types"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
)

const (
	kib = 1024
	mib = kib * 1024

	DefaultBufferSize = 10 * mib
	DefaultExpire     = 15 * time.Minute
)

var (
	validAtLeast5MiB = validation.Min(MultipartMinSize).
		Error("must be at least 5MiB")
)

// Subset of model.StorageSettings applicable to s3.
type storageSettings struct {
	// BucketName defines the bucket name.
	BucketName *string

	// StaticCredentials that overrides AWS config.
	StaticCredentials *StaticCredentials `json:"auth"`

	// Region where the bucket lives
	Region *string
	// ExternalURI is the URI used for signing requests.
	ExternalURI *url.URL
	// URI is the URI for the s3 API.
	URI *url.URL

	LiteralBucketURI bool

	// ProxyURI is used for rewriting presigned requests, pointing the
	// requests to the proxy URL instead of the direct URL to s3.
	ProxyURI *url.URL

	// ForcePathStyle encodes bucket in the API path.
	ForcePathStyle bool
	// UseAccelerate enables s3 Accelerate
	UseAccelerate bool
}

func newFromParent(
	defaults *storageSettings,
	parent *model.StorageSettings,
) (*storageSettings, error) {
	var err error
	ret := new(storageSettings)
	ret.patch(defaults)
	if parent.Bucket != "" {
		ret.BucketName = &parent.Bucket
	}
	if parent.Region != "" {
		ret.Region = &parent.Region
	}
	if parent.Key != "" && parent.Secret != "" {
		ret.StaticCredentials = &StaticCredentials{
			Key:    parent.Key,
			Secret: parent.Secret,
			Token:  parent.Token,
		}
	}
	if parent.ExternalUri != "" {
		ret.ExternalURI, err = url.Parse(parent.ExternalUri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse aws.external_uri: %w", err)
		}
	}
	if parent.ProxyURI != nil {
		ret.ProxyURI, err = url.Parse(*parent.ProxyURI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse storage.proxy_uri: %w", err)
		}
	}
	if parent.Uri != "" {
		ret.URI, err = url.Parse(parent.Uri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse aws.uri: %w", err)
		}
	}
	if parent.LiteralBucketURI != nil {
		ret.LiteralBucketURI = *parent.LiteralBucketURI
	}
	if parent.ForcePathStyle != ret.ForcePathStyle {
		ret.ForcePathStyle = parent.ForcePathStyle
	}
	if parent.UseAccelerate != ret.UseAccelerate {
		ret.UseAccelerate = parent.UseAccelerate
	}
	return ret, nil
}

func (s storageSettings) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.StaticCredentials),
	)
}

// endpointResolver implements custom endpoint resolution for OSS CNAME mode.
// When CNAMEMode is enabled, the bucket name is NOT included in the URL.
// This is required for Alibaba OSS CNAME addressing.
type endpointResolver url.URL

func (r endpointResolver) ResolveEndpoint(
	ctx context.Context,
	params s3.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	// Return the CNAME endpoint directly without bucket in hostname
	// The bucket is still tracked internally for signing purposes
	return smithyendpoints.Endpoint{
		URI: url.URL(r),
	}, nil
}

func (s storageSettings) options(opts *s3.Options) {
	if s.StaticCredentials != nil {
		opts.Credentials = *s.StaticCredentials
	}
	if s.Region != nil {
		opts.Region = *s.Region
	}

	// OSS CNAME Mode: Use custom endpoint resolver to prevent bucket
	// from being added to the hostname
	//
	if s.URI != nil {
		if s.URI.Scheme == "" {
			s.URI.Scheme = "https"
		}
		if s.LiteralBucketURI {
			opts.EndpointResolverV2 = endpointResolver(*s.URI)
		} else {
			opts.BaseEndpoint = types.Pointer(s.URI.String())
		}
	}

	opts.UsePathStyle = s.ForcePathStyle
	opts.UseAccelerate = s.UseAccelerate
}

func (s storageSettings) presignOptions(opts *s3.PresignOptions) {
	opts.ClientOptions = append(opts.ClientOptions, s.options)
	if s.ExternalURI != nil {
		opts.ClientOptions = append(opts.ClientOptions, func(clientOpts *s3.Options) {
			if s.ExternalURI != nil {
				if s.ExternalURI.Scheme == "" {
					s.ExternalURI.Scheme = "https"
				}
				if s.LiteralBucketURI {
					clientOpts.EndpointResolverV2 = endpointResolver(
						*s.ExternalURI,
					)
				} else {
					clientOpts.BaseEndpoint = types.Pointer(
						s.ExternalURI.String(),
					)
				}
			}
		})
	}
}

func (s *storageSettings) patch(setting *storageSettings) *storageSettings {
	if setting == nil {
		return s
	}
	if setting.BucketName != nil {
		s.BucketName = setting.BucketName
	}
	if setting.StaticCredentials != nil {
		s.StaticCredentials = setting.StaticCredentials
	}
	if setting.Region != nil {
		s.Region = setting.Region
	}
	if setting.ExternalURI != nil {
		s.ExternalURI = setting.ExternalURI
	}
	if setting.URI != nil {
		s.URI = setting.URI
	}
	if setting.ProxyURI != nil {
		s.ProxyURI = setting.ProxyURI
	}
	if setting.ForcePathStyle != s.ForcePathStyle {
		s.ForcePathStyle = setting.ForcePathStyle
	}
	if setting.UseAccelerate != s.UseAccelerate {
		s.UseAccelerate = setting.UseAccelerate
	}
	return s
}

type Options struct {
	storageSettings

	// ContentType of the uploaded objects
	ContentType *string
	// FilenameSuffix adds the suffix to the content-disposition for object downloads
	FilenameSuffix *string
	// DefaultExpire is the fallback presign expire duration
	// (defaults to 15min).
	DefaultExpire *time.Duration
	// BufferSize sets the buffer size allocated for uploads.
	// This implicitly sets the upper limit for upload size:
	// BufferSize * 10000 (defaults to: 5MiB).
	BufferSize *int

	// UnsignedHeaders forces the driver to skip the named headers from the
	// being signed.
	UnsignedHeaders []string

	// Transport sets an alternative RoundTripper used by the Go HTTP
	// client.
	Transport http.RoundTripper
}

func NewOptions(opts ...*Options) *Options {
	defaultBufferSize := DefaultBufferSize
	ret := &Options{
		BufferSize: &defaultBufferSize,
	}
	for _, opt := range opts {
		ret.storageSettings.patch(&opt.storageSettings)
		if opt.LiteralBucketURI != ret.LiteralBucketURI {
			ret.LiteralBucketURI = opt.LiteralBucketURI
		}
		if opt.DefaultExpire != nil {
			ret.DefaultExpire = opt.DefaultExpire
		}
		if opt.ContentType != nil {
			ret.ContentType = opt.ContentType
		}
		if opt.BufferSize != nil {
			ret.BufferSize = opt.BufferSize
		}
		if opt.UnsignedHeaders != nil {
			ret.UnsignedHeaders = opt.UnsignedHeaders
		}
		if opt.Transport != nil {
			ret.Transport = opt.Transport
		}
	}
	return ret
}

func (opts Options) Validate() error {
	return validation.ValidateStruct(&opts,
		validation.Field(&opts.storageSettings),
		validation.Field(&opts.BufferSize, validAtLeast5MiB),
	)
}

func (opts *Options) SetBucketName(bucketName string) *Options {
	opts.storageSettings.BucketName = &bucketName
	return opts
}

func (opts *Options) SetStaticCredentials(key, secret, sessionToken string) *Options {
	opts.StaticCredentials = &StaticCredentials{
		Key:    key,
		Secret: secret,
		Token:  sessionToken,
	}
	return opts
}

func (opts *Options) SetRegion(region string) *Options {
	opts.Region = &region
	return opts
}

func (opts *Options) SetContentType(contentType string) *Options {
	opts.ContentType = &contentType
	return opts
}

func (opts *Options) SetFilenameSuffix(suffix string) *Options {
	opts.FilenameSuffix = &suffix
	return opts
}

func (opts *Options) SetExternalURI(externalURI *url.URL) *Options {
	opts.ExternalURI = externalURI
	return opts
}

func (opts *Options) SetURI(URI *url.URL) *Options {
	opts.URI = URI
	return opts
}

func (opts *Options) SetLiteralBucketURI(literal bool) *Options {
	opts.LiteralBucketURI = literal
	return opts
}

func (opts *Options) SetProxyURI(proxyURI *url.URL) *Options {
	opts.ProxyURI = proxyURI
	return opts
}

func (opts *Options) SetForcePathStyle(forcePathStyle bool) *Options {
	opts.ForcePathStyle = forcePathStyle
	return opts
}

func (opts *Options) SetUseAccelerate(useAccelerate bool) *Options {
	opts.UseAccelerate = useAccelerate
	return opts
}

func (opts *Options) SetDefaultExpire(defaultExpire time.Duration) *Options {
	opts.DefaultExpire = &defaultExpire
	return opts
}

func (opts *Options) SetBufferSize(bufferSize int) *Options {
	opts.BufferSize = &bufferSize
	return opts
}

func (opts *Options) SetUnsignedHeaders(unsignedHeaders []string) *Options {
	opts.UnsignedHeaders = unsignedHeaders
	return opts
}

func (opts *Options) SetTransport(transport http.RoundTripper) *Options {
	opts.Transport = transport
	return opts
}

type Signer struct {
	v4.Signer
	// Google Cloud Storage does not tolerate signing the Accept-Encoding header
	UnsignedHeaders []string
}

func NewSigner(
	unsignedHeaders []string,
	optFns ...func(signer *v4.SignerOptions),
) v4.HTTPSigner {
	for i := range unsignedHeaders {
		unsignedHeaders[i] = textproto.CanonicalMIMEHeaderKey(unsignedHeaders[i])
	}
	return &Signer{
		Signer:          *v4.NewSigner(optFns...),
		UnsignedHeaders: unsignedHeaders,
	}
}

func (s Signer) SignHTTP(
	ctx context.Context,
	credentials aws.Credentials,
	r *http.Request,
	payloadHash, service, region string,
	signingTime time.Time,
	optFns ...func(*v4.SignerOptions)) error {
	unsignedHeaders := make(http.Header)
	for _, hdr := range s.UnsignedHeaders {
		if value, ok := r.Header[hdr]; ok {
			unsignedHeaders[hdr] = value
			delete(r.Header, hdr)
		}
	}
	err := s.Signer.SignHTTP(
		ctx, credentials,
		r, payloadHash,
		service, region,
		signingTime, optFns...,
	)
	for key, value := range unsignedHeaders {
		r.Header[key] = value
	}
	return err
}

func (opts *Options) toS3Options() (
	clientOpts func(*s3.Options),
	presignOpts func(*s3.PresignOptions),
) {
	clientOpts = func(s3Opts *s3.Options) {
		opts.options(s3Opts)
		if len(opts.UnsignedHeaders) > 0 {
			s3Opts.HTTPSignerV4 = NewSigner(opts.UnsignedHeaders)
		}
		roundTripper := opts.Transport
		if roundTripper == nil {
			roundTripper = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: storage.GetRootCAs(),
				},
			}
		}
		s3Opts.HTTPClient = &http.Client{
			Transport: roundTripper,
		}
	}

	expires := DefaultExpire
	if opts.DefaultExpire != nil {
		expires = *opts.DefaultExpire
	}
	presignOpts = func(s3Opts *s3.PresignOptions) {
		opts.presignOptions(s3Opts)
		s3.WithPresignExpires(expires)(s3Opts)
	}
	return clientOpts, presignOpts
}
