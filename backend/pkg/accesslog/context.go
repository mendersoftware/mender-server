// Copyright 2024 Northern.tech AS
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

package accesslog

import (
	"context"
	"strings"
	"sync"
)

const (
	DefaultMaxErrors = 5
)

type AccessLogFormat string

type LogContext interface {
	PushError(err error) bool
	SetField(key string, value interface{})
}

type logContext struct {
	errors    []error
	mu        sync.Mutex
	maxErrors int
	fields    map[string]interface{}
}

func (c *logContext) SetField(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fields == nil {
		c.fields = make(map[string]interface{})
	}
	c.fields[key] = value
}

func (c *logContext) PushError(err error) bool {
	if c.maxErrors > 0 {
		c.mu.Lock()
		defer c.mu.Unlock()
		if len(c.errors) > c.maxErrors {
			return false
		}
	}
	c.errors = append(c.errors, err)
	return true
}

func (c *logContext) addFields(fields map[string]interface{}) {
	if c == nil {
		return
	}
	switch len(c.errors) {
	case 0:
	case 1:
		if c.errors[0] != nil {
			fields["error"] = c.errors[0].Error()
		}
	default:
		var s strings.Builder
		for i, err := range c.errors {
			if err != nil {
				s.WriteString(err.Error())
				if i < len(c.errors)-1 {
					s.WriteString("; ")
				}
			}
		}
		fields["error"] = s.String()
	}
	for key, value := range c.fields {
		fields[key] = value
	}
}

type logContextKey struct{}

func withContext(ctx context.Context, c *logContext) context.Context {
	return context.WithValue(ctx, logContextKey{}, c)
}

func fromContext(ctx context.Context) *logContext {
	if c, ok := ctx.Value(logContextKey{}).(*logContext); ok && c != nil {
		return c
	}
	return nil
}

func GetContext(ctx context.Context) LogContext {
	if lc := fromContext(ctx); lc != nil {
		return lc
	}
	return nil
}
