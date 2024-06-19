// Copyright 2021 Northern.tech AS
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

package testing

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type Part struct {
	ContentType string
	ImageData   []byte
	FieldName   string
	FieldValue  string
}

// MakeMultipartRequest returns a http.Request.
func MakeMultipartRequest(
	method string,
	urlStr string,
	contentType string,
	payload []Part,
) *http.Request {
	body_buf := new(bytes.Buffer)
	body_writer := multipart.NewWriter(body_buf)
	for _, part := range payload {
		mh := make(textproto.MIMEHeader)
		mh.Set("Content-Type", part.ContentType)
		if part.ContentType == "" && part.ImageData == nil {
			mh.Set("Content-Disposition", "form-data; name=\""+part.FieldName+"\"")
		} else {
			mh.Set(
				"Content-Disposition", "form-data; name=\""+
					part.FieldName+"\"; filename=\"artifact-213.tar.gz\"",
			)
		}
		part_writer, err := body_writer.CreatePart(mh)
		if nil != err {
			panic(err.Error())
		}
		if part.ContentType == "" && part.ImageData == nil {
			b := []byte(part.FieldValue)
			_, _ = io.Copy(part_writer, bytes.NewReader(b))
		} else {
			_, _ = io.Copy(part_writer, bytes.NewReader(part.ImageData))
		}
	}
	body_writer.Close()

	r, err := http.NewRequest(method, urlStr, bytes.NewReader(body_buf.Bytes()))
	if err != nil {
		panic(err)
	}
	r.Header.Set("Accept-Encoding", "gzip")
	if payload != nil {
		r.Header.Set("Content-Type", contentType+";boundary="+body_writer.Boundary())
	}

	return r
}
