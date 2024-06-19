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

package http

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

func pageLinkHdrs(c *gin.Context, page, perPage, total int) {
	url := &url.URL{
		Path:     c.Request.URL.Path,
		RawQuery: c.Request.URL.RawQuery,
		Fragment: c.Request.URL.Fragment,
	}

	query := url.Query()

	query.Set("page", "1")
	query.Set("per_page", fmt.Sprintf("%d", perPage))
	url.RawQuery = query.Encode()
	Link := fmt.Sprintf(`<%s>;rel="first"`, url.String())
	// Previous page
	if page > 1 {
		query.Set("page", fmt.Sprintf("%d", page-1))
		url.RawQuery = query.Encode()
		Link = fmt.Sprintf(`%s, <%s>;rel="previous"`, Link, url.String())
	}

	// Next page
	if total > (perPage*page - 1) {
		query.Set("page", fmt.Sprintf("%d", page+1))
		url.RawQuery = query.Encode()
		Link = fmt.Sprintf(`%s, <%s>;rel="next"`, Link, url.String())

	}
	c.Header("Link", Link)
}
