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
package log

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	l := New(Ctx{"foo": "bar"})
	assert.NotNil(t, l)
}

func TestNewFromLogger(t *testing.T) {
	baselog := logrus.New()
	baselog.Level = logrus.PanicLevel
	baselog.Out = ioutil.Discard

	l := NewFromLogger(baselog, Ctx{})
	assert.NotNil(t, l)
	assert.Equal(t, l.Logger.Level, logrus.PanicLevel)
	assert.Equal(t, l.Logger.Out, ioutil.Discard)
}

func TestSetup(t *testing.T) {
	// setup with debug on
	Setup(false)

	l := New(Ctx{"foo": "bar"})

	if l.Level() != logrus.InfoLevel {
		t.Fatalf("expected info level")
	}

	Setup(true)

	l = New(Ctx{"foo": "bar"})

	if l.Level() != logrus.DebugLevel {
		t.Fatalf("expected debug level")
	}
}

func TestWithFields(t *testing.T) {

	Setup(false)

	l := New(Ctx{})

	exp := map[string]interface{}{
		"bar":    1,
		"baz":    "cafe",
		"module": "foo",
	}
	l = l.F(Ctx{
		"bar": exp["bar"],
		"baz": exp["baz"],
	})

	if len(l.Data) != len(exp)-1 {
		t.Fatalf("log fields number mismatch: expected %v got %v",
			len(exp), len(l.Data))
	}

	for k, v := range l.Data {
		ev, ok := exp[k]
		if ok != true {
			t.Fatalf("unexpected key: %s", k)
		}
		if ev != v {
			t.Fatalf("value mismatch: got %+v expected %+v",
				v, ev)
		}
	}
}

func TestFromWithContext(t *testing.T) {
	ctx := context.Background()

	l := New(Ctx{"foo": "bar"})

	// we should get back the same logger
	ln := FromContext(WithContext(ctx, l))
	assert.Equal(t, l, ln)

	// since we're using a background context, a new, empty logger should be
	// returned
	ln2 := FromContext(context.Background())
	assert.NotEqual(t, ln2, ln)
	assert.Len(t, ln2.Data, 0)

	ctx = WithContext(context.Background(), l)
	assert.NotNil(t, ctx.Value(loggerContextKey))

	assert.Nil(t, ctx.Value(0))
}
