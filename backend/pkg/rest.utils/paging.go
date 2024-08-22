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

package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

const (
	PerPageDefault = 20
	PerPageMax     = 500

	pageQueryParam    = "page"
	perPageQueryParam = "per_page"
)

var (
	ErrPerPageLimit = errors.Errorf(
		`parameter "per_page" above limit (max: %d)`, PerPageMax,
	)
)

// ParsePagingParameters parses the paging parameters from the URL query
// string and returns the parsed page, per_page or a parsing error respectively.
func ParsePagingParameters(r *http.Request) (int64, int64, error) {
	q := r.URL.Query()
	var (
		err     error
		page    int64
		perPage int64
	)
	qPage := q.Get(pageQueryParam)
	if qPage == "" {
		page = 1
	} else {
		page, err = strconv.ParseInt(qPage, 10, 64)
		if err != nil {
			return -1, -1, errors.Errorf(
				"invalid page query: \"%s\"",
				qPage,
			)
		} else if page < 1 {
			return -1, -1, errors.New("invalid page query: " +
				"value must be a non-zero positive integer",
			)
		}
	}

	qPerPage := q.Get(perPageQueryParam)
	if qPerPage == "" {
		perPage = PerPageDefault
	} else {
		perPage, err = strconv.ParseInt(qPerPage, 10, 64)
		if err != nil {
			return -1, -1, errors.Errorf(
				"invalid per_page query: \"%s\"",
				qPerPage,
			)
		} else if perPage < 1 {
			return -1, -1, errors.New("invalid per_page query: " +
				"value must be a non-zero positive integer",
			)
		} else if perPage > PerPageMax {
			return page, perPage, ErrPerPageLimit
		}
	}
	return page, perPage, nil
}

type PagingHints struct {
	// TotalCount provides the total count of elements available,
	// if provided adds another link to the last page available.
	TotalCount *int64

	// HasNext instructs adding the "next" link header. This option
	// has no effect if TotalCount is given.
	HasNext *bool

	// Pagination parameters
	Page, PerPage *int64
}

func NewPagingHints() *PagingHints {
	return new(PagingHints)
}

func (h *PagingHints) SetTotalCount(totalCount int64) *PagingHints {
	h.TotalCount = &totalCount
	return h
}

func (h *PagingHints) SetHasNext(hasNext bool) *PagingHints {
	h.HasNext = &hasNext
	return h
}

func (h *PagingHints) SetPage(page int64) *PagingHints {
	h.Page = &page
	return h
}

func (h *PagingHints) SetPerPage(perPage int64) *PagingHints {
	h.PerPage = &perPage
	return h
}

func MakePagingHeaders(r *http.Request, hints ...*PagingHints) ([]string, error) {
	// Parse hints
	hint := new(PagingHints)
	for _, h := range hints {
		if h == nil {
			continue
		}
		if h.HasNext != nil {
			hint.HasNext = h.HasNext
		}
		if h.TotalCount != nil {
			hint.TotalCount = h.TotalCount
		}
		if h.Page != nil {
			hint.Page = h.Page
		}
		if h.PerPage != nil {
			hint.PerPage = h.PerPage
		}
	}
	if hint.Page == nil || hint.PerPage == nil {
		page, perPage, err := ParsePagingParameters(r)
		if err != nil {
			return nil, err
		}
		hint.Page, hint.PerPage = &page, &perPage
	}
	locationURL := url.URL{
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
		Fragment: r.URL.Fragment,
	}
	q := locationURL.Query()
	// Ensure per_page is set
	q.Set(perPageQueryParam, strconv.FormatInt(*hint.PerPage, 10))
	links := make([]string, 0, 4)
	q.Set(pageQueryParam, "1")
	locationURL.RawQuery = q.Encode()
	links = append(links, fmt.Sprintf(
		"<%s>; rel=\"first\"", locationURL.String(),
	))
	if (*hint.Page) > 1 {
		q.Set(pageQueryParam, strconv.FormatInt(*hint.Page-1, 10))
		locationURL.RawQuery = q.Encode()
		links = append(links, fmt.Sprintf(
			"<%s>; rel=\"prev\"", locationURL.String(),
		))
	}

	// TotalCount takes precedence over HasNext
	if hint.TotalCount != nil && *hint.TotalCount > 0 {
		lastPage := (*hint.TotalCount-1) / *hint.PerPage + 1
		if *hint.Page < lastPage {
			// Add "next" link
			q.Set(pageQueryParam, strconv.FormatUint(uint64(*hint.Page)+1, 10))
			locationURL.RawQuery = q.Encode()
			links = append(links, fmt.Sprintf(
				"<%s>; rel=\"next\"", locationURL.String(),
			))
		}
		// Add "last" link
		q.Set(pageQueryParam, strconv.FormatInt(lastPage, 10))
		locationURL.RawQuery = q.Encode()
		links = append(links, fmt.Sprintf(
			"<%s>; rel=\"last\"", locationURL.String(),
		))
	} else if hint.HasNext != nil && *hint.HasNext {
		q.Set(pageQueryParam, strconv.FormatUint(uint64(*hint.Page)+1, 10))
		locationURL.RawQuery = q.Encode()
		links = append(links, fmt.Sprintf(
			"<%s>; rel=\"next\"", locationURL.String(),
		))
	}

	return links, nil
}
