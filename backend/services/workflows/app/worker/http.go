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

package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
	"github.com/mendersoftware/mender-server/services/workflows/model"
)

var makeHTTPRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
	var httpClient = &http.Client{
		Timeout: timeout,
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func processHTTPTask(
	httpTask *model.HTTPTask,
	ps *processor.JobStringProcessor,
	jp *processor.JobProcessor,
	l *log.Logger,
) (*model.TaskResult, error) {
	uri := ps.ProcessJobString(httpTask.URI)
	l.Infof("processHTTPTask: starting with: method=%s uri=%s",
		httpTask.Method,
		uri,
	)

	var payloadString string
	if len(httpTask.FormData) > 0 {
		form := url.Values{}
		for key, value := range httpTask.FormData {
			key = ps.ProcessJobString(key)
			key = ps.MaybeExecuteGoTemplate(key)
			value = ps.ProcessJobString(value)
			value = ps.MaybeExecuteGoTemplate(value)
			form.Add(key, value)
		}
		payloadString = form.Encode()
	} else if httpTask.JSON != nil {
		payloadJSON := jp.ProcessJSON(httpTask.JSON, ps)
		payloadBytes, err := json.Marshal(payloadJSON)
		if err != nil {
			return nil, err
		}
		payloadString = string(payloadBytes)
	} else {
		payloadString = ps.ProcessJobString(httpTask.Body)
		payloadString = ps.MaybeExecuteGoTemplate(payloadString)
	}
	payload := strings.NewReader(payloadString)

	req, err := http.NewRequest(httpTask.Method, uri, payload)
	if err != nil {
		return nil, err
	}

	if httpTask.ContentType != "" {
		req.Header.Add("Content-Type", httpTask.ContentType)
	}

	var headersToBeSent []string
	for name, value := range httpTask.Headers {
		headerValue := ps.ProcessJobString(value)
		req.Header.Add(name, headerValue)
		headersToBeSent = append(headersToBeSent,
			fmt.Sprintf("%s: %s", name, headerValue))
	}

	l.Debugf("processHTTPTask makeHTTPRequest '%v'", req)
	res, err := makeHTTPRequest(req, time.Duration(httpTask.ReadTimeOut)*time.Second)
	l.Debugf("processHTTPTask makeHTTPRequest returned '%v','%v'", res, err)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	resBody, _ := ioutil.ReadAll(res.Body)

	var success bool
	if len(httpTask.StatusCodes) == 0 {
		success = true
		if res.StatusCode >= 400 {
			success = false
		}
	} else {
		success = false
		for _, statusCode := range httpTask.StatusCodes {
			if statusCode == res.StatusCode {
				success = true
				break
			}
		}
	}

	result := &model.TaskResult{
		Success: success,
		HTTPRequest: &model.TaskResultHTTPRequest{
			URI:     uri,
			Method:  httpTask.Method,
			Body:    payloadString,
			Headers: headersToBeSent,
		},
		HTTPResponse: &model.TaskResultHTTPResponse{
			StatusCode: res.StatusCode,
			Body:       string(resBody),
		},
	}

	return result, nil
}
