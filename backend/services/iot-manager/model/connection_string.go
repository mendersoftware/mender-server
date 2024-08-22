// Copyright 2022 Northern.tech AS
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

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
)

const (
	csDelimiter    = ";"
	csVarSeparator = "="
	omitted        = "...<omitted>"

	csKeyHostName              = "HostName"
	csKeySharedAccessKey       = "SharedAccessKey"
	csKeySharedAccessKeyName   = "SharedAccessKeyName"
	csKeySharedAccessSignature = "SharedAccessSignature"
	csKeyDeviceId              = "DeviceId"
	csKeyModuleId              = "ModuleId"
)

var (
	ErrConnectionStringTooLong = errors.New(
		"connection string can be no longer than 4096 characters",
	)
	ErrHostnameTrust = errors.New(
		"hostname does not refer to a trusted domain",
	)

	trustedHostnames hostnameValidator
)

type ResolveError string
type lookupHost func(hostname string) error

var lookupHostFunc lookupHost = func(hostname string) error {
	_, err := net.LookupHost(hostname)
	return err
}

func (err ResolveError) Error() string {
	return fmt.Sprintf("failed to lookup host with name '%s'", string(err))
}

func SetTrustedHostnames(hostnames []string) {
	trustedHostnames = newHostnameValidator(hostnames)
}

// ConnectionString implements the Azure connection string format and the
// SharedAccessSignature authz algorithm.
// The implementation is based on the official python SDK.
// https://github.com/Azure/azure-iot-sdk-python
type ConnectionString struct {
	HostName        string        `cs:"HostName" bson:"hostname"`
	GatewayHostName string        `cs:"GatewayHostName" bson:"gateway_hostname,omitempty"`
	Name            string        `cs:"SharedAccessKeyName" bson:"name,omitempty"`
	DeviceID        string        `cs:"DeviceId" bson:"device_id,omitempty"`
	ModuleID        string        `cs:"ModuleId" bson:"module_id,omitempty"`
	Key             crypto.String `cs:"SharedAccessKey" bson:"access_key"`
	Signature       string        `cs:"SharedAccessSignature" bson:"-"`
}

func ParseConnectionString(connection string) (*ConnectionString, error) {
	cs := new(ConnectionString)
	csArgs := strings.Split(connection, csDelimiter)
	for _, arg := range csArgs {
		kv := strings.SplitN(arg, csVarSeparator, 2)
		if len(kv) != 2 {
			return nil, errors.New("invalid connectionstring format")
		}
		switch kv[0] {
		case csKeyHostName:
			cs.HostName = kv[1]
		case csKeySharedAccessKey:
			key, err := base64.StdEncoding.DecodeString(kv[1])
			if err != nil {
				return nil, errors.Wrap(err, "shared access key format")
			}
			cs.Key = crypto.String(key)
		case csKeySharedAccessKeyName:
			cs.Name = kv[1]
		case csKeySharedAccessSignature:
			cs.Signature = kv[1]
		case csKeyDeviceId:
			cs.DeviceID = kv[1]
		case csKeyModuleId:
			cs.ModuleID = kv[1]
		default:
			return nil, fmt.Errorf("invalid connection string key: %s", kv[0])
		}
	}
	return cs, errors.Wrap(cs.Validate(), "connection string invalid")
}

func (cs ConnectionString) IsZero() bool {
	rVal := reflect.ValueOf(cs)
	n := rVal.NumField()
	for i := 0; i < n; i++ {
		if !rVal.Field(i).IsZero() {
			return false
		}
	}
	return true
}

func (cs ConnectionString) Validate() error {
	if cs.IsZero() {
		return nil
	}
	err := validation.ValidateStruct(&cs,
		validation.Field(&cs.HostName, validation.Required, trustedHostnames),
		validation.Field(&cs.Key, validation.Required),
		validation.Field(&cs.GatewayHostName, validation.When(
			cs.GatewayHostName != "", trustedHostnames,
		)),
	)
	if err != nil {
		return err
	}
	if cs.DeviceID == "" && cs.Name == "" {
		return errors.New("one of 'DeviceId' or 'SharedAccessKeyName' must be set")
	}
	if len(cs.String()) > 4096 {
		return ErrConnectionStringTooLong
	}
	return nil
}

func (cs ConnectionString) Authorization(expireAt time.Time) string {
	qURI := url.QueryEscape(cs.HostName)
	msg := fmt.Sprintf("%s\n%d", qURI, expireAt.Unix())
	signer := hmac.New(sha256.New, []byte(cs.Key))
	_, _ = signer.Write([]byte(msg))
	sign := signer.Sum(nil)
	sign64 := base64.StdEncoding.EncodeToString(sign)
	token := fmt.Sprintf("SharedAccessSignature sr=%s&sig=%s&se=%d",
		qURI,
		url.QueryEscape(sign64),
		expireAt.Unix(),
	)
	if cs.Name != "" {
		token += "&skn=" + cs.Name
	}
	return token
}

func (cs ConnectionString) string(omit bool) string {
	val := reflect.ValueOf(cs)
	typ := val.Type()
	n := typ.NumField()
	var res = make([]string, 0, n)
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("cs")
		if tag == "" {
			continue
		}
		fieldVal := val.Field(i)
		if fieldVal.Len() == 0 {
			continue
		}
		switch typ := fieldVal.Interface().(type) {
		case []byte:
			res = append(res, tag+"="+base64.StdEncoding.EncodeToString(typ))
		case crypto.String:
			var value string
			if omit {
				value = base64.StdEncoding.EncodeToString([]byte(typ[:3])) + omitted
			} else {
				value = base64.StdEncoding.EncodeToString([]byte(typ))
			}
			res = append(res, tag+"="+value)
		case string:
			res = append(res, tag+"="+typ)
		default:
			continue
		}
	}
	txt := strings.Join(res, csDelimiter)
	return txt
}

func (cs ConnectionString) String() string {
	return cs.string(false)
}

func (cs ConnectionString) MarshalText() ([]byte, error) {
	return []byte(cs.string(true)), nil
}

func (cs *ConnectionString) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	connStr, err := ParseConnectionString(string(b))
	if err != nil {
		return err
	}
	*cs = *connStr
	return nil
}

type hostnameValidator [][]string

func newHostnameValidator(hostnames []string) hostnameValidator {
	// Compile the list of hostnames into a set of hostnames split by separator '.'
	var ret = make(hostnameValidator, len(hostnames))
	var j int
	for i := range hostnames {
		if hostnames[i] == "" {
			continue
		}
		ret[i] = strings.Split(hostnames[i], ".")
		j++
	}
	ret = ret[:j]
	return ret
}

func (patterns hostnameValidator) matchHostname(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	if len(hostname) == 0 {
		return false
	}
	hostParts := strings.Split(hostname, ".")
	for _, patternParts := range patterns {
		if len(patternParts) != len(hostParts) {
			continue
		}
		allMatch := true
		for i, patternPart := range patternParts {
			if patternPart == "*" {
				continue
			}
			if patternPart != hostParts[i] {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}
	return false
}

func (patterns hostnameValidator) Validate(v interface{}) error {
	if len(patterns) == 0 {
		return errors.New("[PROG ERR(hostnameValidator)] no trusted hostnames configured")
	}
	hostname, ok := v.(string)
	if !ok {
		return errors.New("[PROG ERR(hostnameValidator)] validating non-string hostname")
	}
	hostname = strings.SplitN(hostname, ":", 2)[0]
	if !patterns.matchHostname(hostname) {
		return ErrHostnameTrust
	} else if err := lookupHostFunc(hostname); err != nil {
		return ResolveError(hostname)
	}
	return nil
}
