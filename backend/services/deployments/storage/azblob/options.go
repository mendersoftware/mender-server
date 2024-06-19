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

package azblob

import (
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

const (
	BufferSizeMin     = 4 * 1024          // 4KiB
	BufferSizeDefault = 8 * BufferSizeMin // 32KiB - same default as used in io.Copy
)

type SharedKeyCredentials struct {
	AccountName string
	AccountKey  string

	URI *string // Optional
}

func (creds SharedKeyCredentials) azParams(
	containerName string,
) (containerURL string, azCreds *azblob.SharedKeyCredential, err error) {
	azCreds, err = azblob.NewSharedKeyCredential(creds.AccountName, creds.AccountKey)
	if err == nil {
		if creds.URI != nil {
			containerURL = *creds.URI
		} else {
			containerURL = fmt.Sprintf(
				"https://%s.blob.core.windows.net/%s",
				azCreds.AccountName(),
				containerName,
			)
		}
	}
	return containerURL, azCreds, err
}

type Options struct {
	ConnectionString *string
	SharedKey        *SharedKeyCredentials

	ProxyURI *url.URL

	BufferSize int64

	ContentType *string
}

func NewOptions(opts ...*Options) *Options {
	opt := &Options{
		BufferSize: BufferSizeDefault,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		if o.ConnectionString != nil {
			opt.ConnectionString = o.ConnectionString
		}
		if o.SharedKey != nil {
			opt.SharedKey = o.SharedKey
		}
		if o.ProxyURI != nil {
			opt.ProxyURI = o.ProxyURI
		}
		if o.ContentType != nil {
			opt.ContentType = o.ContentType
		}
		if o.BufferSize >= BufferSizeMin {
			opt.BufferSize = o.BufferSize
		}
	}
	return opt
}

func (opts *Options) SetConnectionString(connStr string) *Options {
	opts.ConnectionString = &connStr
	return opts
}

func (opts *Options) SetSharedKey(sk SharedKeyCredentials) *Options {
	opts.SharedKey = &sk
	return opts
}

func (opts *Options) SetProxyURI(proxyURI *url.URL) *Options {
	opts.ProxyURI = proxyURI
	return opts
}

func (opts *Options) SetContentType(typ string) *Options {
	opts.ContentType = &typ
	return opts
}

func (opts *Options) SetBufferSize(size int64) *Options {
	opts.BufferSize = size
	return opts
}
