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
package routing

import (
	"net/http"
	"reflect"
	"runtime"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	rtest "github.com/ant0ine/go-json-rest/rest/test"
)

func TestSupportsMethod(t *testing.T) {

	var sets = []struct {
		exp       bool
		method    string
		supported []string
	}{
		{
			true,
			http.MethodOptions,
			[]string{
				http.MethodGet,
				http.MethodPut,
				http.MethodOptions,
			},
		},
		{
			false,
			http.MethodOptions,
			[]string{
				http.MethodGet,
				http.MethodPut,
			},
		},
	}

	for _, tv := range sets {
		if supportsMethod(tv.method, tv.supported) != tv.exp {
			t.Errorf("failed case: %+v", tv)
		}
	}
}

// We can't compare functions, so let's take the hard way and extract
// func name from runtime
func funcName(f interface{}) string {
	p := reflect.ValueOf(f).Pointer()
	rfunc := runtime.FuncForPC(p)
	return rfunc.Name()
}

func TestAutogenOptionRoutes(t *testing.T) {
	// make sure that dummy and options are different to prevent
	// the compiler making this a single symbol
	dummy := func(w rest.ResponseWriter, r *rest.Request) {
		// dummy
		w.WriteJson(struct {
			x int
		}{
			2,
		})
	}
	options := func(w rest.ResponseWriter, r *rest.Request) {
		// dummy
		w.WriteJson(struct {
			x int
		}{
			1,
		})
	}
	gen := func(methods []string) rest.HandlerFunc {
		return options
	}

	routes := []*rest.Route{
		// expecting rest.Options(..) to be added for /foo
		rest.Get("/foo", dummy),
		rest.Post("/foo", dummy),

		// no extra OPTIONS handler for /bar
		rest.Get("/bar", dummy),
		rest.Options("/bar", dummy),
	}

	augmented := AutogenOptionsRoutes(routes, gen)

	type expHandler map[string]rest.HandlerFunc
	exp := map[string]expHandler{
		"/foo": {
			http.MethodGet:     dummy,
			http.MethodPost:    dummy,
			http.MethodOptions: options,
		},
		"/bar": {
			http.MethodGet:     dummy,
			http.MethodOptions: dummy,
		},
	}

	// we're expecting 5 handlers in total
	expCount := 5
	if len(augmented) != expCount {
		t.Errorf("got %d handlers instead of %d", len(augmented), expCount)
	}

	for _, r := range augmented {
		v, ok := exp[r.PathExp]
		if ok != true {
			t.Errorf("failed with route %+v, route not present", r)
		}

		h, ok := v[r.HttpMethod]
		if ok != true {
			t.Errorf("failed with route %+v, method not present", r)
		}

		if funcName(r.Func) != funcName(h) {
			t.Errorf("failed with route %+v, different handler", r)
		}
	}
}

func TestAutogenOptionHeaders(t *testing.T) {

	suppmeth := []string{
		http.MethodGet,
		http.MethodPut,
	}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, _ := rest.MakeRouter(
		rest.Options("/test", AllowHeaderOptionsGenerator(suppmeth)),
	)

	api.SetApp(router)

	rec := rtest.RunRequest(t, api.MakeHandler(),
		rtest.MakeSimpleRequest(http.MethodOptions,
			"http://1.2.3.4/test", nil))

	allowmeth := rec.Recorder.HeaderMap[http.CanonicalHeaderKey("Allow")]

	// expecting only 2 allowed methods (should OPTIONS be
	// included in Allow too?)
	if len(allowmeth) != 2 {
		t.Errorf("too many allowed methods: %+v", allowmeth)
	}

	for _, sh := range suppmeth {
		var found bool
		for _, s := range allowmeth {
			if sh == s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("supported method %s not in allowed: %+v",
				sh, allowmeth)
		}
	}
}
