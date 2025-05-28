package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const DEFAULT_AUTH = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0." +
	"s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4"

type TestRequest struct {
	Method      string
	Path        string
	Body        interface{}
	ContentType string
	Auth        bool
	Token       string
}

func MakeTestRequest(r *TestRequest) *http.Request {
	var body io.Reader
	if r.Body != nil {
		bodyJSON, _ := json.Marshal(r.Body)
		body = bytes.NewReader(bodyJSON)
	}

	req, _ := http.NewRequest(r.Method, r.Path, body)

	if r.Auth {
		if r.Token == "" {
			req.Header.Set("Authorization", DEFAULT_AUTH)
		} else {
			req.Header.Set("Authorization", r.Token)
		}
	}

	if r.Body == nil || body == nil {
		return req
	}

	if r.ContentType == "" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", r.ContentType)
	}

	return req
}
