# Copyright 2025 Northern.tech AS
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

import os

import pytest

import openapi_client as oas
import openapi_client.exceptions as api_exceptions

from client import make_authenticated_client


@pytest.mark.usefixtures("clean_db")
class TestGroupRemoving:
    def test_delete_device(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        d1 = "".join([format(i, "02x") for i in os.urandom(128)])
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=d1, attributes=inventory_attributes),
        )

        g1 = "group-test-3"

        management_client.add_devices_to_group(name=g1, request_body=[d1])
        assert len(management_client.get_devices_in_group(g1)) == 1

        management_client.remove_devices_from_group_with_http_info(
            name=g1, request_body=[d1]
        )

        with pytest.raises(api_exceptions.NotFoundException):
            management_client.get_devices_in_group(name=g1)

    def test_delete_device_non_existent_1(self):
        """Delete non-existent device from non-existent group"""
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        g1 = "group-test-3-non-existent"
        rsp = management_client.remove_devices_from_group(
            name=g1, request_body=["404-device"]
        )
        assert rsp.updated_count is None or rsp.updated_count == 0

    def test_delete_device_non_existent_2(self, inventory_attributes):
        """Delete existent device from non-existent group"""
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        d1 = "".join([format(i, "02x") for i in os.urandom(128)])
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=d1, attributes=inventory_attributes),
        )

        rsp = management_client.remove_devices_from_group(
            name="404_group", request_body=[d1]
        )
        assert rsp.updated_count is None or rsp.updated_count == 0
