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


def test_decommission_device(mmock_url, workflows_url):
    do_decommission_device(mmock_url, workflows_url, "")


def test_decommission_device_with_tenant_id(mmock_url, workflows_url):
    do_decommission_device(mmock_url, workflows_url, "123456789012345678901234")


def do_decommission_device(mmock_url, workflows_url, tenant_id):
    # start the decommission device workflow
    device_id = "1"
    request_id = "1234567890"
    res = requests.post(
        workflows_url + "/api/v1/workflow/decommission_device",
        json={
            "request_id": request_id,
            "device_id": device_id,
            "tenant_id": tenant_id,
        },
    )
    assert res.status_code == 201
    # verify the response
    response = res.json()
    assert response is not None
    assert type(response) is dict
    assert response["name"] == "decommission_device"
    assert response["id"] is not None
    # get the job details, every second until done
    for i in range(10):
        time.sleep(1)
        res = requests.get(
            workflows_url + "/api/v1/workflow/decommission_device/" + response["id"]
        )
        if res.status_code == 404:
            continue
        assert res.status_code == 200
        # if status is done, break
        response = res.json()
        assert response is not None
        assert type(response) is dict
        if response["status"] == "done":
            break
    else:
        raise TimeoutError("timeout waiting for workflow to finish")
    # verify the status
    assert {"name": "request_id", "value": request_id} in response["inputParameters"]
    assert {"name": "device_id", "value": device_id} in response["inputParameters"]
    assert response["status"] == "done"
    # #  verify the mock server has been correctly called
    res = requests.get(mmock_url + "/api/request/all")
    assert res.status_code == 200
    response = res.json()
    # assert len(response) == 4
    expected = [
        {
            "scheme": "http",
            "host": "mender-deployments",
            "port": "8080",
            "method": "DELETE",
            "path": "/api/internal/v1/deployments/tenants/"
            + tenant_id
            + "/deployments/devices/"
            + device_id,
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": ["gzip"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": "",
        },
        {
            "scheme": "http",
            "host": "mender-deviceconnect",
            "port": "8080",
            "method": "DELETE",
            "path": "/api/internal/v1/deviceconnect/tenants/"
            + tenant_id
            + "/devices/"
            + device_id,
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": ["gzip"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": "",
        },
        {
            "scheme": "http",
            "host": "mender-deviceconfig",
            "port": "8080",
            "method": "DELETE",
            "path": "/api/internal/v1/deviceconfig/tenants/"
            + tenant_id
            + "/devices/"
            + device_id,
            "queryStringParameters": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": ["gzip"],
                "User-Agent": ["Go-http-client/1.1"],
                "X-Men-Requestid": [request_id],
            },
            "cookies": {},
            "body": "",
        },
        {
            "body": "",
            "cookies": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": [
                    "gzip",
                ],
                "User-Agent": [
                    "Go-http-client/1.1",
                ],
                "X-Men-Requestid": [
                    "1234567890",
                ],
            },
            "host": "mender-iot-manager",
            "method": "DELETE",
            "path": f"/api/internal/v1/iot-manager/tenants/{tenant_id}/devices/{device_id}",
            "port": "8080",
            "queryStringParameters": {},
            "scheme": "http",
        },
        {
            "body": "",
            "cookies": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": [
                    "gzip",
                ],
                "User-Agent": [
                    "Go-http-client/1.1",
                ],
                "X-Men-Requestid": [
                    "1234567890",
                ],
            },
            "host": "mender-inventory",
            "method": "DELETE",
            "path": f"/api/internal/v1/inventory/tenants/{tenant_id}/devices/{device_id}",
            "port": "8080",
            "queryStringParameters": {},
            "scheme": "http",
        },
        {
            "body": "",
            "cookies": {},
            "fragment": "",
            "headers": {
                "Accept-Encoding": [
                    "gzip",
                ],
                "User-Agent": [
                    "Go-http-client/1.1",
                ],
                "X-Men-Requestid": [
                    "1234567890",
                ],
            },
            "host": "mender-device-auth",
            "method": "DELETE",
            "path": f"/api/internal/v1/devauth/tenants/{tenant_id}/devices/{device_id}",
            "port": "8080",
            "queryStringParameters": {},
            "scheme": "http",
        },
    ]
    assert expected == [actual["request"] for actual in response]
