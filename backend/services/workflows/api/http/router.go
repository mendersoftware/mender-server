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

package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	"github.com/mendersoftware/mender-server/services/workflows/store"
)

// API URL used by the HTTP router
const (
	APIURLStatus = "/status"

	APIURLHealth        = "/api/v1/health"
	APIURLWorkflow      = "/api/v1/workflow/:name"
	APIURLWorkflowBatch = "/api/v1/workflow/:name/batch"
	APIURLWorkflowID    = "/api/v1/workflow/:name/:id"
	APIURLJobsID        = "/api/v1/jobs/:id"

	APIURLWorkflows = "/api/v1/metadata/workflows"
)

// NewRouter returns the gin router
func NewRouter(dataStore store.DataStore, nats nats.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	ctx := context.Background()
	l := log.FromContext(ctx)

	router.Use(routerLogger(l))
	router.Use(gin.Recovery())

	status := NewStatusController()
	router.GET(APIURLStatus, status.Status)

	workflow := NewWorkflowController(dataStore, nats)
	router.GET(APIURLHealth, workflow.HealthCheck)

	router.POST(APIURLWorkflow, workflow.StartWorkflow)
	router.POST(APIURLWorkflowBatch, workflow.StartBatchWorkflows)
	router.GET(APIURLWorkflowID, workflow.GetWorkflowByNameAndID)

	router.POST(APIURLWorkflows, workflow.RegisterWorkflow)
	router.GET(APIURLWorkflows, workflow.GetWorkflows)
	router.GET(APIURLJobsID, workflow.GetJobByID)

	return router
}
