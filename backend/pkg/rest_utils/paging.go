// Copyright 2024 Northern.tech AS
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

package rest_utils

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	micro_strings "github.com/mendersoftware/mender-server/pkg/strings"
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
	LinkTmpl       = "<%s>; rel=\"%s\""
	LinkPrev       = "prev"
	LinkNext       = "next"
	LinkFirst      = "first"
	DefaultScheme  = "http"
)

// error msgs
func MsgQueryParmInvalid(name string) string {
	return fmt.Sprintf("Can't parse param %s", name)
}

func MsgQueryParmMissing(name string) string {
	return fmt.Sprintf("Missing required param %s", name)
}

func MsgQueryParmLimit(name string) string {
	return fmt.Sprintf("Param %s is out of bounds", name)
}

func MsgQueryParmOneOf(name string, allowed []string) string {
	return fmt.Sprintf("Param %s must be one of %v", name, allowed)
}

// query param parsing/validation
func ParseQueryParmUInt(
	r *rest.Request,
	name string,
	required bool,
	min, max, def uint64,
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
		return 0, errors.New(MsgQueryParmInvalid(name))
	}

	if uintVal < min || uintVal > max {
		return 0, errors.New(MsgQueryParmLimit(name))
	}

	return uintVal, nil
}

func ParseQueryParmBool(r *rest.Request, name string, required bool, def *bool) (*bool, error) {
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
		return nil, errors.New(MsgQueryParmInvalid(name))
	}

	return &boolVal, nil
}

func ParseQueryParmStr(
	r *rest.Request,
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
		if allowed != nil && !micro_strings.ContainsString(val, allowed) {
			return "", errors.New(MsgQueryParmOneOf(name, allowed))
		}
	}

	return val, nil
}

// pagination helpers
func ParsePagination(r *rest.Request) (uint64, uint64, error) {
	page, err := ParseQueryParmUInt(r, PageName, false, PageMin, math.MaxUint64, PageDefault)
	if err != nil {
		return 0, 0, err
	}

	per_page, err := ParseQueryParmUInt(
		r,
		PerPageName,
		false,
		PerPageMin,
		PerPageMax,
		PerPageDefault,
	)
	if err != nil {
		return 0, 0, err
	}

	return page, per_page, nil
}

func MakePageLinkHdrs(r *rest.Request, page, per_page uint64, has_next bool) []string {
	var links []string
	if page > 1 {
		links = append(links, MakeLink(LinkPrev, r, page-1, per_page))
	}

	if has_next {
		links = append(links, MakeLink(LinkNext, r, page+1, per_page))
	}

	links = append(links, MakeLink(LinkFirst, r, 1, per_page))
	return links
}

// MakeLink creates a relative URL for insertion in the link header URL field.
func MakeLink(link_type string, r *rest.Request, page, per_page uint64) string {
	q := r.URL.Query()
	q.Set(PageName, strconv.Itoa(int(page)))
	q.Set(PerPageName, strconv.Itoa(int(per_page)))
	url := url.URL{
		Path:     r.URL.Path,
		RawPath:  r.URL.RawPath,
		RawQuery: q.Encode(),
		Fragment: r.URL.Fragment,
	}

	return fmt.Sprintf(LinkTmpl, url.String(), link_type)
}

// build URL using request 'r' and template, replace path params with
// elements from 'params' using lexical match as in strings.Replace()
func BuildURL(r *rest.Request, template string, params map[string]string) *url.URL {
	url := r.BaseUrl()

	path := template
	for k, v := range params {
		path = strings.Replace(path, k, v, -1)
	}
	url.Path = path

	return url
}
