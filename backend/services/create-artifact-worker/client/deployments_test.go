// Copyright 2020 Northern.tech AS
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
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentsUploadArtifactInternal(t *testing.T) {
	t.Parallel()

	tc := []struct {
		tenantId string
		artId    string
		desc     string
		fname    string

		uriInternalUpload string

		code   int
		errmsg string
		err    error
	}{
		{
			tenantId: "1",
			artId:    "2",
			desc:     "foo",
			fname:    "artifact1",

			uriInternalUpload: "/api/internal/v1/deployments/tenants/1/artifacts",
			code:              201,
		},
		{
			tenantId: "2",
			artId:    "3",
			desc:     "foo",
			fname:    "artifact2",

			uriInternalUpload: "/api/internal/v1/deployments/tenants/2/artifacts",
			code:              201,
		},
		{
			tenantId: "2",
			artId:    "3",
			desc:     "foo",
			fname:    "artifact3",

			uriInternalUpload: "/api/internal/v1/deployments/tenants/2/artifacts",
			code:              500,
			errmsg:            "general service error",
			err:               errors.New("failed to upload artifact 3: http 500, reqid: 1234, msg: general service error"),
		},
	}

	for idx := range tc {
		tc := tc[idx]
		t.Run(fmt.Sprintf("tc: %d", idx), func(t *testing.T) {
			t.Parallel()

			ctx := context.TODO()

			tmp, err := ioutil.TempFile(".", tc.fname)
			assert.NoError(t, err)

			content := []byte("foobar")
			tmp.Write(content)

			defer tmp.Close()

			defer os.Remove(tmp.Name())

			name := filepath.Base(tmp.Name())

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				//verify url
				assert.Equal(t, tc.uriInternalUpload, req.URL.Path)

				//verify contents
				err := req.ParseMultipartForm(1024)
				assert.NoError(t, err)

				tid := req.FormValue("id")
				aid := req.FormValue("artifact_id")
				d := req.FormValue("description")

				assert.Equal(t, tc.artId, aid)
				assert.Equal(t, tc.tenantId, tid)
				assert.Equal(t, tc.desc, d)

				// file
				f, fh, err := req.FormFile("artifact")
				assert.NoError(t, err)
				assert.True(t, strings.HasPrefix(fh.Filename, tc.fname))
				b, err := ioutil.ReadAll(f)
				assert.NoError(t, err)

				rw.WriteHeader(tc.code)
				assert.Equal(t, content, b)

				// emit error if any
				if tc.err != nil {
					_, _ = rw.Write(restErr(t, tc.errmsg))
				}
			}))

			defer server.Close()

			c, err := NewDeployments(server.URL, true)
			assert.NoError(t, err)

			err = c.UploadArtifactInternal(ctx, name, tc.artId, tc.tenantId, tc.desc)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func restErr(t *testing.T, msg string) []byte {
	r := struct {
		Id  string `json:"request_id"`
		Err string `json:"error"`
	}{
		Id:  "1234",
		Err: msg,
	}

	b, err := json.Marshal(r)
	assert.NoError(t, err)

	return b
}
