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

import uuid
import pytest
import requests

from common import management_api_with_params
from internal_api import InternalAPIClient


@pytest.fixture
def device_id():
    client = InternalAPIClient()
    device_id = str(uuid.uuid4())
    new_device = {"device_id": device_id}
    r = client.provision_device_with_http_info(
        tenant_id="tenant-id", new_device=new_device, _preload_content=False
    )
    assert r.status == 201
    yield device_id
    r = client.decommission_device_with_http_info(
        tenant_id="tenant-id", device_id=device_id, _preload_content=False
    )
    assert r.status == 204


class TestAuditlogs:
    def test_config_device_set_auditlog(self, device_id, mmock_url):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        #
        # set the initial configuration
        configuration = {
            "key": "value",
            "another-key": "another-value",
            "dollar-key": "$",
        }
        r = client.set_device_configuration(
            device_id, request_body=configuration, _preload_content=False
        )
        assert r.status == 204
        #
        # verify the auditlogs
        res = requests.get(mmock_url + "/api/request/all")
        assert res.status_code == 200
        response = res.json()
        assert len(response) == 1
        expected = {
            "request": {
                "scheme": "http",
                "host": "mender-workflows-server",
                "port": "8080",
                "method": "POST",
                "path": "/api/v1/workflow/emit_auditlog",
                "queryStringParameters": {},
                "fragment": "",
                "headers": {
                    "Content-Type": ["application/json"],
                    "Accept-Encoding": ["gzip"],
                    "User-Agent": ["Go-http-client/1.1"],
                },
                "cookies": {},
            },
        }
        body = response[0]["request"].pop("body")
        del response[0]["request"]["headers"]["Content-Length"]
        assert expected["request"] == response[0]["request"]
        assert "set_configuration" in body

    def test_config_device_deploy_auditlog(self, device_id, mmock_url):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        #
        # set the initial configuration
        configuration = {
            "key": "value",
            "another-key": "another-value",
            "dollar-key": "$",
        }
        r = client.set_device_configuration(
            device_id, request_body=configuration, _preload_content=False
        )
        assert r.status == 204
        #
        # deploy the configuration
        request = {
            "retries": 1,
        }
        r = client.deploy_device_configuration(
            device_id, new_configuration_deployment=request, _preload_content=False
        )
        assert r.status == 200
        assert "deployment_id" in str(r.data)
        #
        # verify the auditlogs
        res = requests.get(mmock_url + "/api/request/all")
        assert res.status_code == 200
        responses = res.json()
        assert len(responses) == 3
        expected = [
            {
                "request": {
                    "scheme": "http",
                    "host": "mender-workflows-server",
                    "port": "8080",
                    "method": "POST",
                    "path": "/api/v1/workflow/emit_auditlog",
                    "queryStringParameters": {},
                    "fragment": "",
                    "headers": {
                        "Content-Type": ["application/json"],
                        "Accept-Encoding": ["gzip"],
                        "User-Agent": ["Go-http-client/1.1"],
                    },
                    "cookies": {},
                },
                "contains": "set_configuration",
            },
            {
                "request": {
                    "scheme": "http",
                    "host": "mender-workflows-server",
                    "port": "8080",
                    "method": "POST",
                    "path": "/api/v1/workflow/deploy_device_configuration",
                    "queryStringParameters": {},
                    "fragment": "",
                    "headers": {
                        "Content-Type": ["application/json"],
                        "Accept-Encoding": ["gzip"],
                        "User-Agent": ["Go-http-client/1.1"],
                    },
                    "cookies": {},
                },
            },
            {
                "request": {
                    "scheme": "http",
                    "host": "mender-workflows-server",
                    "port": "8080",
                    "method": "POST",
                    "path": "/api/v1/workflow/emit_auditlog",
                    "queryStringParameters": {},
                    "fragment": "",
                    "headers": {
                        "Content-Type": ["application/json"],
                        "Accept-Encoding": ["gzip"],
                        "User-Agent": ["Go-http-client/1.1"],
                    },
                    "cookies": {},
                },
                "contains": "deploy_configuration",
            },
        ]
        for expected, response in zip(expected, responses):
            body = response["request"].pop("body")
            del response["request"]["headers"]["Content-Length"]
            assert expected["request"] == response["request"]
            if expected.get("contains"):
                assert expected["contains"] in body
