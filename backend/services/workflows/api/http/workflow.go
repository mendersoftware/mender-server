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

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
	"github.com/mendersoftware/mender-server/services/workflows/utils"
)

const (
	defaultTimeout = time.Second * 5
)

var (
	HeaderWorkflowMinVersion = "X-Workflows-Min-Version"
)

// WorkflowController container for end-points
type WorkflowController struct {
	// dataStore provides an interface to the database
	dataStore store.DataStore
	// nats provides an interface to the message bus
	nats nats.Client
}

// NewWorkflowController returns a new StatusController
func NewWorkflowController(dataStore store.DataStore, nats nats.Client) *WorkflowController {
	return &WorkflowController{
		dataStore: dataStore,
		nats:      nats,
	}
}

func (h WorkflowController) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	err := h.dataStore.Ping(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "error reaching MongoDB: " + err.Error(),
		})
		return
	}
	if !h.nats.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "not connected to nats",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterWorkflow responds to POST /api/v1/metadata/workflows
func (h WorkflowController) RegisterWorkflow(c *gin.Context) {
	var workflow model.Workflow
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad request",
		})
		return
	}
	if err = json.Unmarshal(rawData, &workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Error parsing JSON form: %s",
				err.Error()),
		})
		return
	}
	if workflow.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workflow missing name",
		})
		return
	}
	_, err = h.dataStore.InsertWorkflows(c, workflow)
	if err != nil {
		httpStatus := http.StatusBadRequest
		if err == store.ErrWorkflowAlreadyExists {
			httpStatus = http.StatusConflict
		}
		c.JSON(httpStatus, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Status(http.StatusCreated)
}

// GetWorkflows responds to GET /api/v1/metadata/workflows
func (h WorkflowController) GetWorkflows(c *gin.Context) {
	c.JSON(http.StatusOK, h.dataStore.GetWorkflows(c))
}

func (h WorkflowController) startWorkflowGetJob(
	c *gin.Context,
	inputParameters map[string]interface{},
	name string,
) ([]byte, string, string, error) {
	workflowVersion := ""
	if values := c.Request.Header[HeaderWorkflowMinVersion]; len(values) > 0 {
		workflowVersion = values[0]
	}

	var jobInputParameters []model.InputParameter
	for key, value := range inputParameters {
		valueSlice, ok := value.([]interface{})
		if ok {
			values := make([]string, 0, 10)
			for _, value := range valueSlice {
				valueString, err := utils.ConvertAnythingToString(value)
				if err == nil {
					values = append(values, valueString)
				}
			}
			jobInputParameters = append(jobInputParameters, model.InputParameter{
				Name:  key,
				Value: strings.Join(values, ","),
				Raw:   value,
			})
		} else {
			valueString, err := utils.ConvertAnythingToString(value)
			if err == nil {
				jobInputParameters = append(jobInputParameters, model.InputParameter{
					Name:  key,
					Value: valueString,
					Raw:   value,
				})
			}
		}
	}

	jobID := primitive.NewObjectID().Hex()
	job := &model.Job{
		ID:              jobID,
		InsertTime:      time.Now(),
		WorkflowName:    name,
		WorkflowVersion: workflowVersion,
		InputParameters: jobInputParameters,
	}

	workflow, err := h.dataStore.GetWorkflowByName(c, job.WorkflowName, job.WorkflowVersion)
	if err != nil {
		return nil, "", "", store.ErrWorkflowNotFound
	}

	if err := job.Validate(workflow); err != nil {
		return nil, "", "", err
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "failed to marshal the job")
	}

	topic := workflow.Topic
	if topic == "" {
		topic = model.DefaultTopic
	}
	subject := h.nats.StreamName() + "." + topic

	return jobJSON, job.ID, subject, nil
}

// StartWorkflow responds to POST /api/workflow/:name
func (h WorkflowController) StartWorkflow(c *gin.Context) {
	l := log.FromContext(c.Request.Context())

	var name string = c.Param("name")
	l.Infof("StartWorkflow starting workflow %s", name)

	var inputParameters map[string]interface{}
	if err := c.BindJSON(&inputParameters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unable to parse the input parameters: %s",
				err.Error()),
		})
		return
	}

	jobJSON, jobID, subject, err := h.startWorkflowGetJob(c, inputParameters, name)
	if err != nil {
		l.Error(err)
		statusCode := http.StatusBadRequest
		switch err {
		case store.ErrWorkflowNotFound:
			statusCode = http.StatusNotFound
		default:
		}
		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.nats.JetStreamPublish(subject, jobJSON)
	if err != nil {
		l.Error(errors.Wrap(err, "JetStreamPublish failed"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   jobID,
		"name": name,
	})
	l.Infof("StartWorkflow starting workflow %s : StatusCreated", name)
}

// StartBatchWorkflows responds to POST /api/workflow/:name/batch
func (h WorkflowController) StartBatchWorkflows(c *gin.Context) {
	l := log.FromContext(c.Request.Context())

	var name string = c.Param("name")
	l.Infof("StartWorkflow starting workflow %s", name)

	var inputParametersBatch []map[string]interface{}
	if err := c.BindJSON(&inputParametersBatch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unable to parse the input parameters: %s",
				err.Error()),
		})
		return
	}

	result := make([]map[string]string, 0, len(inputParametersBatch))
	for _, inputParameters := range inputParametersBatch {
		jobJSON, jobID, subject, err := h.startWorkflowGetJob(c, inputParameters, name)
		if err != nil {
			l.Error(err)
			statusCode := http.StatusBadRequest
			switch err {
			case store.ErrWorkflowNotFound:
				statusCode = http.StatusNotFound
			default:
			}
			c.JSON(statusCode, gin.H{
				"error": err.Error(),
			})
			return
		}
		jobResult := map[string]string{
			"id":   jobID,
			"name": name,
		}
		err = h.nats.JetStreamPublish(subject, jobJSON)
		if err != nil {
			l.Error(errors.Wrap(err, "JetStreamPublish failed"))
			delete(jobResult, "id")
			delete(jobResult, "name")
			jobResult["error"] = err.Error()
		}
		result = append(result, jobResult)
	}

	c.JSON(http.StatusCreated, result)
	l.Infof("StartWorkflow starting workflow %s : StatusCreated", name)
}

// GetWorkflowByNameAndID responds to GET /api/workflow/:name/:id
func (h WorkflowController) GetWorkflowByNameAndID(c *gin.Context) {
	var name string = c.Param("name")
	var id string = c.Param("id")

	job, err := h.dataStore.GetJobByNameAndID(c, name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	} else if job == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	job.PrepareForJSONMarshalling()
	c.JSON(http.StatusOK, job)
}
