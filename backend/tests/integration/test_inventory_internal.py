# Copyright 2026 Northern.tech AS
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
from mender_client import Configuration, ApiClient, api
from mender_client.models.attribute_v2 import AttributeV2
import uuid
import redo

from testutils.api.client import ApiClient as ApiClientOld
import testutils.api.deviceauth as deviceauth

from testutils.common import (
    create_user,
    make_accepted_devices,
)

from testutils.infra.cli import CliUseradm, CliDeviceauth


@pytest.fixture(scope="function")
def clean_migrated_mongo(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()

    deviceauth_cli.migrate()
    useradm_cli.migrate()

    yield clean_mongo


class TestPatchAttributes:
    def test_inventory_patch_attributes_internal(self, clean_migrated_mongo):
        login = "user-foo@acme.com"
        password = "correcthorse"
        user = create_user(login, password)

        configuration = Configuration.get_default()
        host = configuration.host
        configuration.username = login
        configuration.password = password
        api_client_management = ApiClient(configuration)
        user_token = api.UserAdministrationManagementAPIApi(
            api_client_management
        ).login()

        tenant_token = ""
        devauthd = ApiClientOld(deviceauth.URL_DEVICES)
        devauthm = ApiClientOld(deviceauth.URL_MGMT)
        make_accepted_devices(devauthd, devauthm, user_token, tenant_token, 2)

        configuration.access_token = user_token
        inventory_management = api.DeviceInventoryManagementAPIApi(
            api_client_management
        )
        for _ in redo.retrier(attempts=3, sleeptime=1):
            devices = inventory_management.list_device_inventories(
                _headers={"Authorization": f"Bearer {user_token}"}, per_page=32, page=1
            )
            if len(devices) > 0:
                break
        device_id_to_update = devices[0].id

        inventory_internal = api.DeviceInventoryInternalAPIApi(
            ApiClient(Configuration(host="http://inventory:8080"))
        )
        ts = [
            AttributeV2.from_dict(
                {"name": "mac", "scope": "identity", "value": "de:ad:be:ef:06:12"}
            )
        ]
        inventory_internal.update_inventory_for_a_device_scope_wise(
            tenant_id="", device_id=device_id_to_update, attribute_v2=ts
        )

        current_devices = inventory_management.list_device_inventories(
            _headers={"Authorization": f"Bearer {user_token}"}, per_page=32, page=1
        )

        found = False
        for device in current_devices:
            if device.id == device_id_to_update:
                found = any(
                    [
                        attribute.name == "mac"
                        and attribute.value.actual_instance == "de:ad:be:ef:06:12"
                        and attribute.scope == "identity"
                        for attribute in device.attributes
                    ]
                )
        assert found

        ts = [
            AttributeV2.from_dict(
                {"name": "mac", "scope": "identity", "value": "de:ad:be:ef:06:24"}
            )
        ]
        inventory_internal.update_inventory_for_a_device_scope_wise(
            tenant_id="", device_id=device_id_to_update, attribute_v2=ts
        )

        current_devices = inventory_management.list_device_inventories(
            _headers={"Authorization": f"Bearer {user_token}"}, per_page=32, page=1
        )

        found = False
        for device in current_devices:
            if device.id == device_id_to_update:
                found = any(
                    [
                        attribute.name == "mac"
                        and attribute.value.actual_instance == "de:ad:be:ef:06:24"
                        and attribute.scope == "identity"
                        for attribute in device.attributes
                    ]
                )
        assert found

        ts = [
            AttributeV2.from_dict(
                {"name": "newone", "scope": "system", "value": "brandnew"}
            )
        ]
        inventory_internal.update_inventory_for_a_device_scope_wise(
            tenant_id="", device_id=device_id_to_update, attribute_v2=ts
        )

        current_devices = inventory_management.list_device_inventories(
            _headers={"Authorization": f"Bearer {user_token}"}, per_page=32, page=1
        )

        found = False
        for device in current_devices:
            if device.id == device_id_to_update:
                found = any(
                    [
                        attribute.name == "newone"
                        and attribute.value.actual_instance == "brandnew"
                        and attribute.scope == "system"
                        for attribute in device.attributes
                    ]
                )
        assert found
        configuration.host = host
