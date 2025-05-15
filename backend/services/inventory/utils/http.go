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

package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

// pagination constants
const (
	PageName       = "page"
	PerPageName    = "per_page"
	PageMin        = 1
	PageDefault    = 1
	PerPageMin     = 1
	PerPageMax     = 500
	PerPageDefault = 20
	LinkHdr        = "Link"
	LinkTmpl       = "<%s?%s>; rel=\"%s\""
	LinkPrev       = "prev"
	LinkNext       = "next"
	LinkFirst      = "first"
	DefaultScheme  = "http"
)

// error msgs
func MsgQueryParmMissing(name string) string {
	return fmt.Sprintf("Missing required param %s", name)
}

func MsgQueryParmOneOf(name string, allowed []string) string {
	return fmt.Sprintf("Param %s must be one of %v", name, allowed)
}

// query param parsing/validation
func ParseQueryParmUInt(
	r *http.Request,
	name string,
	required bool,
	min,
	max,
	def uint64,
) (uint64, error) {
	strVal := r.URL.Query().Get(name)

	if strVal == "" {
		if required {
			return 0, errors.New(MsgQueryParmMissing(name))
		} else {
			return def, nil
		}
	}

	uintVal, err := strconv.ParseUint(strVal, 10, 32)
	if err != nil {
		return 0, rest.ErrQueryParmInvalid(name, strVal)
	}

	if uintVal < min || uintVal > max {
		return 0, rest.ErrQueryParmLimit(name)
	}

	return uintVal, nil
}

func ParseQueryParmBool(r *http.Request, name string, required bool, def *bool) (*bool, error) {
	strVal := r.URL.Query().Get(name)

	if strVal == "" {
		if required {
			return nil, errors.New(MsgQueryParmMissing(name))
		} else {
			return def, nil
		}
	}

	boolVal, err := strconv.ParseBool(strVal)
	if err != nil {
		return nil, rest.ErrQueryParmInvalid(name, strVal)
	}

	return &boolVal, nil
}

func ParseQueryParmStr(
	r *http.Request,
	name string,
	required bool,
	allowed []string,
) (string, error) {
	val := r.URL.Query().Get(name)

	if val == "" {
		if required {
			return "", errors.New(MsgQueryParmMissing(name))
		}
	} else {
		if allowed != nil && !ContainsString(val, allowed) {
			return "", errors.New(MsgQueryParmOneOf(name, allowed))
		}
	}

	val, err := url.QueryUnescape(val)
	if err != nil {
		return "", rest.ErrQueryParmInvalid(name, val)
	}

	return val, nil
}
