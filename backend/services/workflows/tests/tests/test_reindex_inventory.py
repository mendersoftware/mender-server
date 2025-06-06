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


def test_reindex_inventory(mmock_url, workflows_url):
    request_id = "1234567890"
    tenant_id = "1"
    device_id = "2"
    service = "monitoring"

    res = requests.post(
        workflows_url + "/api/v1/workflow/reindex_inventory",
        json={
            "request_id": request_id,
            "tenant_id": tenant_id,
            "device_id": device_id,
            "service": service,
        },
    )

    assert res.status_code == 201

    response = res.json()
    assert response is not None
    assert type(response) is dict
    assert response["name"] == "reindex_inventory"
    assert response["id"] is not None

    for i in range(10):
        time.sleep(1)
        res = requests.get(
            workflows_url + "/api/v1/workflow/reindex_inventory/" + response["id"]
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

    assert {"name": "request_id", "value": request_id} in response["inputParameters"]
    assert {"name": "device_id", "value": device_id} in response["inputParameters"]
    assert {"name": "tenant_id", "value": tenant_id} in response["inputParameters"]
    assert {"name": "service", "value": service} in response["inputParameters"]
    assert response["status"] == "done"
    assert len(response["results"]) == 1
    assert response["results"][0]["success"] == True
    assert response["results"][0]["httpResponse"]["statusCode"] == 202
