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
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/mendersoftware/mender-server/services/deployments/storage"
)

var (
	ErrConnStrNoName = errors.New("connection string does not contain an account name")
	ErrConnStrNoKey  = errors.New("connection string does not contain an account key")
)

func (c *client) signParamsFromContext(
	ctx context.Context,
) (creds *azblob.SharedKeyCredential, proxyURL *url.URL, err error) {
	creds = c.credentials
	proxyURL = c.proxyURL
	if settings, _ := storage.SettingsFromContext(ctx); settings != nil {
		if settings.ConnectionString != nil {
			creds, err = keyFromConnString(*settings.ConnectionString)
		} else {
			creds, err = azblob.NewSharedKeyCredential(settings.Key, settings.Secret)
		}
		if err == nil && settings.ProxyURI != nil {
			proxyURL, err = url.Parse(*settings.ProxyURI)
		}
	}
	return creds, proxyURL, err
}

func connStringAttr(cs, key string) (string, bool) {
	var start, end int
	for {
		i := strings.Index(cs[start:], key)
		if i < 0 {
			return "", false
		}
		start += i
		if i == 0 || cs[start-1] == ';' {
			break
		}
		start += len(key)
	}
	start += len(key)
	i := strings.IndexRune(cs[start:], ';')
	if i < 0 {
		// cs ends with value
		end = len(cs)
	} else {
		end = start + i
	}
	return cs[start:end], true
}

func keyFromConnString(cs string) (*azblob.SharedKeyCredential, error) {
	const (
		attrName = "AccountName="
		attrKey  = "AccountKey="
	)
	var (
		accountName, accountKey string
		ok                      bool
	)
	if accountName, ok = connStringAttr(cs, attrName); !ok {
		return nil, ErrConnStrNoName
	}
	if accountKey, ok = connStringAttr(cs, attrKey); !ok {
		return nil, ErrConnStrNoKey
	}
	return azblob.NewSharedKeyCredential(accountName, accountKey)
}
