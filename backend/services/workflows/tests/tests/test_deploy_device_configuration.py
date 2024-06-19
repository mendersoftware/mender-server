# Copyright 2021 Northern.tech AS
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
import requests
import time


def test_deploy_device_configuration(mmock_url, workflows_url):
    # start the deploy_device_configuration workflow
    request_id = "1234567890"
    tenant_id = "tenant"
    deployment_id = "deployment_id"
    device_id = "device_id"
    configuration = '{"key":"value"}'
    res = requests.post(
        workflows_url + "/api/v1/workflow/deploy_device_configuration",
        json={
            "request_id": request_id,
            "tenant_id": tenant_id,
            "deployment_id": deployment_id,
            "device_id": device_id,
            "configuration": configuration,
            "retries": 0,
        },
    )
    assert res.status_code == 201
    # verify the response
    response = res.json()
    assert response is not None
    assert type(response) is dict
    assert response["name"] == "deploy_device_configuration"
    assert response["id"] is not None
    # get the job details, every second until done
    for i in range(10):
        time.sleep(1)
        res = requests.get(
            workflows_url
            + "/api/v1/workflow/deploy_device_configuration/"
            + response["id"]
        )
        assert res.status_code == 200
        # if status is done, break
        response = res.json()
        assert response is not None
        assert type(response) is dict
        if response["status"] == "done":
            break
    # verify the status
    assert {"name": "request_id", "value": request_id} in response["inputParameters"]
    assert {"name": "tenant_id", "value": tenant_id} in response["inputParameters"]
    assert {"name": "deployment_id", "value": deployment_id} in response[
        "inputParameters"
    ]
    assert {"name": "device_id", "value": device_id} in response["inputParameters"]
    assert {"name": "configuration", "value": configuration} in response[
        "inputParameters"
    ]
    assert {"name": "retries", "value": "0"} in response["inputParameters"]
    assert response["status"] == "done"
    assert len(response["results"]) == 2
    assert response["results"][0]["success"] == True
    assert response["results"][0]["httpResponse"]["statusCode"] == 201
    assert response["results"][1]["success"] == True
    assert response["results"][1]["httpResponse"]["statusCode"] == 202
    #  verify the mock server has been correctly called
    res = requests.get(mmock_url + "/api/request/all")
    assert res.status_code == 200
    response = res.json()
    assert len(response) == 2
    expected = [
        {
            "request": {
                "scheme": "http",
                "host": "mender-deployments",
                "port": "8080",
                "method": "POST",
                "path": "/api/internal/v1/deployments/tenants/"
                + tenant_id
                + "/configuration/deployments/"
                + deployment_id
                + "/devices/"
                + device_id,
                "queryStringParameters": {},
                "fragment": "",
                "headers": {
                    "Content-Type": ["application/json"],
                    "Accept-Encoding": ["gzip"],
                    "Content-Length": ["88"],
                    "User-Agent": ["Go-http-client/1.1"],
                    "X-Men-Requestid": [request_id],
                },
                "cookies": {},
                "body": '{"configuration":"{\\"key\\":\\"value\\"}","name":"configuration-'
                + deployment_id
                + '","retries":0}',
            }
        },
        {
            "request": {
                "scheme": "http",
                "host": "mender-deviceconnect",
                "port": "8080",
                "method": "POST",
                "path": "/api/internal/v1/deviceconnect/tenants/"
                + tenant_id
                + "/devices/"
                + device_id
                + "/check-update",
                "queryStringParameters": {},
                "fragment": "",
                "headers": {
                    "Accept-Encoding": ["gzip"],
                    "Content-Length": ["0"],
                    "User-Agent": ["Go-http-client/1.1"],
                    "X-Men-Requestid": [request_id],
                },
                "cookies": {},
                "body": "",
            }
        },
    ]
    assert expected[0]["request"] == response[0]["request"]
    assert expected[1]["request"] == response[1]["request"]
