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

package deployments

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/utils"
)

const (
	urlDeviceDeployments    = "/api/internal/v1/deployments/tenants/:tid/deployments/devices"
	urlLastDeviceDeployment = "/api/internal/v1/deployments/tenants/:tid/devices/deployments/last"
	defaultTimeout          = 10 * time.Second
)

//go:generate ../../x/mockgen.sh
type Client interface {
	// GetDeployments retrieves a list of deployments by ID
	GetDeployments(
		ctx context.Context,
		tenantID string,
		IDs []string,
	) ([]*DeviceDeployment, error)
	// GetLatestDeployment retrieves the latest deployment for a given devices
	GetLatestFinishedDeployment(
		ctx context.Context,
		tenantID string,
		deviceIDs []string,
	) ([]LastDeviceDeployment, error)
}

type client struct {
	client  *http.Client
	urlBase string
}

func NewClient(urlBase string) Client {
	return &client{
		client:  &http.Client{},
		urlBase: urlBase,
	}
}

func (c *client) GetDeployments(
	ctx context.Context,
	tenantID string,
	IDs []string,
) ([]*DeviceDeployment, error) {
	const maxDeploymentIDs = 20 // API constraint
	l := log.FromContext(ctx)

	url := utils.JoinURL(c.urlBase, urlDeviceDeployments)
	url = strings.Replace(url, ":tid", tenantID, 1)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}

	var (
		i, j    int
		body    io.ReadCloser
		devDevs []*DeviceDeployment
		rsp     *http.Response
	)
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	for i < len(IDs) && err == nil {
		j += maxDeploymentIDs
		if j > len(IDs) {
			j = len(IDs)
		}
		q := req.URL.Query()
		q.Set("page", "1")
		q.Set("per_page", strconv.Itoa(j-i))
		q.Del("id")
		for k := i; k < j; k++ {
			q.Add("id", IDs[k])
		}
		req.URL.RawQuery = q.Encode()
		rsp, err = c.client.Do(req) //nolint:bodyclose
		if err != nil {
			err = errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
			break
		}
		body = rsp.Body

		switch rsp.StatusCode {
		case http.StatusNotFound:
			// pass
		case http.StatusOK:
			dec := json.NewDecoder(rsp.Body)
			var batch []*DeviceDeployment
			if err = dec.Decode(&batch); err != nil {
				err = errors.Wrap(err, "failed to parse request body")
				break
			}
			if devDevs == nil {
				devDevs = batch
			} else {
				devDevs = append(devDevs, batch...)
			}
		default:
			err = errors.Errorf("%s %s request failed with status %v",
				req.Method, req.URL, rsp.Status)
			l.Errorf(err.Error())
		}
		body.Close()
		body = nil
		i = j
	}
	if len(devDevs) == 0 {
		return nil, err
	}

	return devDevs, err
}

func (c *client) GetLatestFinishedDeployment(
	ctx context.Context,
	tenantID string,
	deviceIDs []string,
) ([]LastDeviceDeployment, error) {
	l := log.FromContext(ctx)

	url := utils.JoinURL(c.urlBase, urlLastDeviceDeployment)
	url = strings.Replace(url, ":tid", tenantID, 1)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	getReq := &GetLastDeviceDeploymentReq{
		DeviceIDs: deviceIDs,
	}

	body, err := json.Marshal(getReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to serialize get last device deployment request")
	}

	rd := bytes.NewReader(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, rd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if rsp.StatusCode != http.StatusOK {
		err := errors.Errorf("%s %s request failed with status %v",
			req.Method, req.URL, rsp.Status)
		l.Errorf(err.Error())
		return nil, err
	}

	dec := json.NewDecoder(rsp.Body)
	var lastDeployments GetLastDeviceDeploymentRsp
	if err = dec.Decode(&lastDeployments); err != nil {
		return nil, errors.Wrap(err, "failed to parse request body")
	} else if len(lastDeployments.DeviceDeploymentLastStatuses) == 0 {
		return nil, nil
	}
	return lastDeployments.DeviceDeploymentLastStatuses, nil
}
