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

import json
import requests
import time


def test_update_device_status(mmock_url, workflows_url):
    # start the provision_device workflow
    request_id = "1234567890"
    device_status = "accepted"
    devices = [
        {"id": "1", "revision": 1},
        {"id": "2", "revision": 1},
        {"id": "3", "revision": 1},
    ]
    devices_json = json.dumps(devices, separators=(",", ":"))
    tenant_id = "1"
    res = requests.post(
        workflows_url + "/api/v1/workflow/update_device_status",
        json={
            "request_id": request_id,
            "devices": devices,
            "device_status": device_status,
            "tenant_id": tenant_id,
        },
    )
    assert res.status_code == 201
    # verify the response
    response = res.json()
    assert response is not None
    assert type(response) is dict
    assert response["name"] == "update_device_status"
    assert response["id"] is not None
    # get the job details, every second until done
    for i in range(10):
        time.sleep(1)
        res = requests.get(
            workflows_url + "/api/v1/workflow/update_device_status/" + response["id"]
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
    assert {"name": "devices", "value": devices_json.strip("[]")} in response[
        "inputParameters"
    ]
    assert response["status"] == "done", response
    assert len(response["results"]) == 2
    assert response["results"][0]["success"] == True
    assert response["results"][0]["httpResponse"]["statusCode"] == 200
    #  verify the mock server has been correctly called
    res = requests.get(mmock_url + "/api/request/all")
    assert res.status_code == 200
    response = res.json()
    assert len(response) == 2
    expected_req = [
        {
            "scheme": "http",
            "host": "mender-inventory",
            "port": "8080",
            "method": "POST",
            "path": "/api/internal/v1/inventory/tenants/"
            + tenant_id
            + "/devices/status/"
            + device_status,
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Content-Type": ["application/json"],
                "Accept-Encoding": ["gzip"],
                "Content-Length": ["73"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": devices_json,
        },
        {
            "scheme": "http",
            "host": "mender-iot-manager",
            "port": "8080",
            "method": "PUT",
            "path": "/api/internal/v1/iot-manager/tenants/"
            + tenant_id
            + "/bulk/devices/status/"
            + device_status,
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Content-Type": ["application/json"],
                "Accept-Encoding": ["gzip"],
                "Content-Length": ["73"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": devices_json,
        },
    ]
    assert expected_req == [res["request"] for res in response]
