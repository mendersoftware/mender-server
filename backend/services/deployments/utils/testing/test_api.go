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

package testing

import (
	"io"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/sirupsen/logrus"
)

type RouterTypeHandler func(pathExp string, handlerFunc rest.HandlerFunc) *rest.Route

func SetUpTestApi(route string, routeType RouterTypeHandler,
	handler func(w rest.ResponseWriter, r *rest.Request)) http.Handler {

	router, _ := rest.MakeRouter(routeType(route, handler))
	api := rest.NewApi()
	api.Use(
		&requestlog.RequestLogMiddleware{
			BaseLogger: &logrus.Logger{Out: io.Discard},
		},
		&requestid.RequestIdMiddleware{},
	)
	api.SetApp(router)

	rest.ErrorFieldName = "error"
	return api.MakeHandler()

}
