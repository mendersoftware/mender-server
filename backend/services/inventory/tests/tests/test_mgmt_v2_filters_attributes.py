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

import pytest

import openapi_client as oas

from client import make_authenticated_client


@pytest.mark.usefixtures("clean_db")
class TestGetAttributes:
    def test_get_attributes(
        self,
        inventory_attributes,
    ):
        auth_client = make_authenticated_client(is_device=False)
        internal_client = oas.InventoryInternalV1Api()
        management_client_v2 = oas.InventoryManagementV2Api(auth_client)
        assert len(management_client_v2.get_filterable_attributes()) == 0

        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        res = management_client_v2.get_filterable_attributes()

        # Expected set of name/scope/count
        expected = {(attr.name, attr.scope, 1) for attr in inventory_attributes}
        # Add server side generated attribute
        expected.add(("created_ts", "system", 1))

        actual = {(attr.name, attr.scope, attr.count) for attr in res}
        assert actual == expected
