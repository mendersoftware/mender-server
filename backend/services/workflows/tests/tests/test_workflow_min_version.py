# Copyright 2024 Northern.tech AS
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
import uuid

import requests


def new_workflow(workflows_url, name, version):
    """create a new workflow"""
    req = {
        "name": name,
        "description": "some",
        "version": version,
        "schemaVersion": 1,
        "tasks": [
            {"name": "t1", "type": "http", "retries": 1, "retryDelaySeconds": 4}
        ],
        "inputParameters": ["newid"],
        "optionalParameters": ["someid"],
    }
    return requests.post(workflows_url + "/api/v1/metadata/workflows", json=req)


def start_workflow(workflows_url, name, version=""):
    req = {"newid": "1", "someid": "2"}
    return requests.post(
        workflows_url + "/api/v1/workflow/" + name,
        json=req,
        headers={"X-Workflows-Min-Version": version},
    )


def test_workflow_min_version(workflows_url):
    """
    Check that we can invoke a workflow with minimal required version
    """
    workflow_name = "wf1-" + str(uuid.uuid4())
    workflow_version = 4

    # first let's create a workflow
    res = new_workflow(workflows_url, workflow_name, workflow_version)
    assert res.status_code == 201

    # starting with the exact version should succeed
    res = start_workflow(workflows_url, workflow_name, str(workflow_version))
    assert res.status_code == 201

    # starting with a lower minimum version requirement should succeed
    res = start_workflow(workflows_url, workflow_name, str(workflow_version - 1))
    assert res.status_code == 201

    # starting with a higher minimum version requirement should fail
    res = start_workflow(workflows_url, workflow_name, str(workflow_version + 1))
    assert res.status_code == 404
