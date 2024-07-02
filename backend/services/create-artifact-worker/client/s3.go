// Copyright 2022 Northern.tech AS
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
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type Storage interface {
	Download(ctx context.Context, url, path string) error
	Delete(ctx context.Context, url string) error
}

type storage struct {
	c *http.Client
}

func NewStorage(skipSsl bool) Storage {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSsl,
		},
	}

	c := &http.Client{
		Transport: tr,
	}

	return &storage{
		c: c,
	}
}

func (s *storage) Download(ctx context.Context, url, path string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	res, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var body string

		bbody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			body = "<failed to read body>"
		} else {
			body = string(bbody)
		}

		return errors.New(fmt.Sprintf(
			"failed to download artifact at url %s, http %d, response: \n %s",
			url,
			res.StatusCode,
			body,
		))
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}

func (s *storage) Delete(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutSec)
	defer cancel()

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	res, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusAccepted {
		var body string

		bbody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			body = "<failed to read body>"
		} else {
			body = string(bbody)
		}

		return errors.New(fmt.Sprintf(
			"failed to delete artifact at url %s, http %d, response: \n %s",
			url,
			res.StatusCode,
			body,
		))
	}

	return nil
}
