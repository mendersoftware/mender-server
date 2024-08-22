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


def test_update_device_inventory(mmock_url, workflows_url):
    # start the provision_device workflow
    request_id = "1234567890"
    device_status = "accepted"
    attributes = '{"foo":"bar"}'
    tenant_id = "1"
    device_id = "2"
    res = requests.post(
        workflows_url + "/api/v1/workflow/update_device_inventory",
        json={
            "request_id": request_id,
            "attributes": attributes,
            "tenant_id": tenant_id,
            "device_id": device_id,
            "scope": "identity",
        },
    )
    assert res.status_code == 201
    # verify the response
    response = res.json()
    assert response is not None
    assert type(response) is dict
    assert response["name"] == "update_device_inventory"
    assert response["id"] is not None
    # get the job details, every second until done
    for i in range(10):
        time.sleep(1)
        res = requests.get(
            workflows_url + "/api/v1/workflow/update_device_inventory/" + response["id"]
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
    assert {"name": "attributes", "value": attributes} in response["inputParameters"]
    assert {"name": "device_id", "value": device_id} in response["inputParameters"]
    assert {"name": "tenant_id", "value": tenant_id} in response["inputParameters"]
    assert response["status"] == "done"
    assert len(response["results"]) == 1
    assert response["results"][0]["success"] == True
    assert response["results"][0]["httpResponse"]["statusCode"] == 200
    #  verify the mock server has been correctly called
    res = requests.get(mmock_url + "/api/request/all")
    assert res.status_code == 200
    response = res.json()
    assert len(response) == 1
    expected = {
        "request": {
            "scheme": "http",
            "host": "mender-inventory",
            "port": "8080",
            "method": "PATCH",
            "path": "/api/internal/v1/inventory/tenants/"
            + tenant_id
            + "/device/"
            + device_id
            + "/attribute/scope/identity",
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Content-Type": ["application/json"],
                "Accept-Encoding": ["gzip"],
                "Content-Length": ["13"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": attributes,
        },
    }
    assert expected["request"] == response[0]["request"]
