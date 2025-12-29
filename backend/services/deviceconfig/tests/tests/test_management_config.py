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

from common import management_api_with_params
from internal_v1 import InternalAPIClient, ProvisionDevice
from management_v1 import (
    ManagementAPIClient,
    NewConfigurationDeployment,
    ApiException as ManagementApiException,
)


@pytest.fixture
def device_id():
    client = InternalAPIClient()
    device_id = uuid.uuid4()
    r = client.device_config_internal_provision_device_with_http_info(
        tenant_id="tenant-id", provision_device=ProvisionDevice(device_id=device_id)
    )
    assert r.status_code == 201
    yield str(device_id)
    r = client.device_config_internal_decommission_device_with_http_info(
        tenant_id="tenant-id", device_id=str(device_id)
    )
    assert r.status_code == 204


class TestManagementConfig:
    def test_config_device_set_get_remove(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        #
        # get the configuration (empty)
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {"id": device_id, "reported": {}, "configured": {}} == {
            k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")
        }
        assert "updated_ts" in data.keys()
        #
        # set the initial configuration
        configuration = {
            "key": "value",
            "another-key": "another-value",
            "dollar-key": "$",
        }
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # get the configuration
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {
            "id": device_id,
            "reported": {},
            "configured": {
                "key": "value",
                "another-key": "another-value",
                "dollar-key": "$",
            },
        } == {k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")}
        assert "updated_ts" in data.keys()
        #
        # replace the configuration
        configuration = {
            "key": "update-value",
            "additional-key": "",
        }
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # get the configuration
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {
            "id": device_id,
            "reported": {},
            "configured": {"key": "update-value", "additional-key": ""},
        } == {k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")}
        assert "updated_ts" in data.keys()
        #
        # remove the configuration
        configuration = {}
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # get the configuration
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {"id": device_id, "reported": {}, "configured": {}} == {
            k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")
        }
        assert "updated_ts" in data.keys()

    def test_config_device_replace_key_with_empty_value(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        #
        # get the configuration (empty)
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {"id": device_id, "reported": {}, "configured": {}} == {
            k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")
        }
        assert "updated_ts" in data.keys()
        #
        # set the initial configuration
        configuration = {
            "key": "value",
            "another-key": "another-value",
        }
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # get the configuration
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {
            "id": device_id,
            "reported": {},
            "configured": {"key": "value", "another-key": "another-value"},
        } == {k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")}
        assert "updated_ts" in data.keys()
        #
        # replace the configuration
        configuration = {
            "key": "value",
            "another-key": "",
        }
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # get the configuration
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert {
            "id": device_id,
            "reported": {},
            "configured": {"key": "value", "another-key": ""},
        } == {k: (str(data[k]) if k == "id" else data[k]) for k in ("id", "reported", "configured")}
        assert "updated_ts" in data.keys()

    def test_config_device_value_number(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "key": "value",
            "another-key": 1234,
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_value_none(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "key": "value",
            "another-key": None,
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_value_boolean(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "key": "value",
            "another-key": False,
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_value_dict(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "key": "value",
            "another-key": {},
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_value_list(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "key": "value",
            "another-key": [],
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_key_too_long(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        configuration = {
            "k" * 4097: "value",
        }
        with pytest.raises(ManagementApiException) as excinfo:
            ManagementAPIClient.device_config_management_set_device_configuration.raw_function(
                client, device_id, request_body=configuration,
            )
        assert excinfo.value.status == 400

    def test_config_device_deploy(self, device_id):
        user_id = str(uuid.uuid4())
        client = management_api_with_params(user_id=user_id, tenant_id="tenant-id")
        #
        # set the initial configuration
        configuration = {
            "key": "value",
            "another-key": "another-value",
            "dollar-key": "$",
        }
        r = client.device_config_management_set_device_configuration_with_http_info(
            device_id, request_body=configuration
        )
        assert r.status_code == 204
        #
        # deploy the configuration
        r = client.device_config_management_deploy_device_configuration_with_http_info(
            device_id,
            new_configuration_deployment=NewConfigurationDeployment(retries=1)
        )
        assert r.status_code == 200
        assert r.data is not None and r.data.deployment_id is not None
        #
        # get the deployment ID
        r = client.device_config_management_get_device_configuration(device_id)
        data = r.to_dict()
        assert "deployment_id" in data.keys()
        assert "" != data["deployment_id"]
