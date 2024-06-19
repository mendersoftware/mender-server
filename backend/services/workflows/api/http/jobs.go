// Copyright 2023 Northern.tech AS
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
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetJobByID responds to GET /api/jobs/:id
func (h WorkflowController) GetJobByID(c *gin.Context) {
	var id = c.Param("id")

	job, err := h.dataStore.GetJobByID(c, id)
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
