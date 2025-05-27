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

import os

import pytest

import openapi_client as oas
import openapi_client.exceptions as api_exceptions

from client import make_authenticated_client


@pytest.mark.usefixtures("clean_db")
class TestGroupCreation:
    def test_get_groups_is_empty(self):
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        assert len(management_client.list_groups()) == 0

    def test_moving_device_group_1(self, inventory_attributes):
        """
        Create 1 device and move it amung 2 different groups
        """
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        management_client.add_devices_to_group(name="groupA", request_body=[did])
        group_a_devs = management_client.get_devices_in_group(name="groupA")
        assert len(group_a_devs) == 1

        management_client.add_devices_to_group(name="groupB", request_body=[did])

        with pytest.raises(api_exceptions.NotFoundException):
            management_client.get_devices_in_group(name="groupA")
        assert len(management_client.get_devices_in_group(name="groupB")) == 1

    def test_moving_devices_1(self, inventory_attributes):
        """
        Create 2 devices and move them amoung 2 different groups
        """
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()
        did1 = "device-id-1"
        did2 = "device-id-2"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did1, attributes=inventory_attributes),
        )
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did2, attributes=inventory_attributes),
        )

        management_client.add_devices_to_group(
            name="group-test-1", request_body=[did1, did2]
        )
        assert len(management_client.get_devices_in_group("group-test-1")) == 2

        management_client.add_devices_to_group(name="group-test-2", request_body=[did2])
        assert len(management_client.get_devices_in_group("group-test-1")) == 1
        assert len(management_client.get_devices_in_group("group-test-2")) == 1

        management_client.add_devices_to_group(name="group-test-2", request_body=[did1])
        with pytest.raises(api_exceptions.NotFoundException):
            management_client.get_devices_in_group("group-test-1")
        assert len(management_client.get_devices_in_group("group-test-2")) == 2

        management_client.add_devices_to_group(
            name="group-test-1", request_body=[did1, did2]
        )
        assert len(management_client.get_devices_in_group("group-test-1")) == 2
        with pytest.raises(api_exceptions.NotFoundException):
            management_client.get_devices_in_group("group-test-2")

    def test_get_groups(self, inventory_attributes):
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()
        for i in range(10):
            group = f"group{i}"
            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.initialize_device(
                tenant_id="",
                device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
            )
            management_client.add_devices_to_group(name=group, request_body=[did])

        assert len(management_client.list_groups()) == 10

    def test_get_groups_3(self, inventory_attributes):
        """
        Create 1 device, and move through 10 different groups
        """
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()

        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        for i in range(10):
            group = f"group{i}"
            management_client.add_devices_to_group(name=group, request_body=[did])
        assert len(management_client.list_groups()) == 1

    def test_has_group(self, inventory_attributes):
        """
        Verify has_group functionality
        """
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        assert len(management_client.list_groups()) == 0
        assert len(management_client.list_device_inventories(has_group=True)) == 0

        management_client.add_devices_to_group(
            name="has_group_test_1", request_body=[did]
        )
        assert len(management_client.list_device_inventories(has_group=True)) == 1

        management_client.remove_devices_from_group(
            name="has_group_test_1", request_body=[did]
        )
        assert len(management_client.list_device_inventories(has_group=True)) == 0

    def test_generic_groups_1(self, inventory_attributes):
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        internal_client = oas.InventoryInternalApi()
        total_groups = 10
        items_per_group = 2
        devices_in_groups = {}

        for i in range(total_groups):
            group = f"group{i}"
            for _ in range(items_per_group):
                device = "".join([format(i, "02x") for i in os.urandom(128)])
                internal_client.initialize_device(
                    tenant_id="",
                    device_new=oas.DeviceNew(
                        id=device, attributes=inventory_attributes
                    ),
                )
                devices_in_groups.setdefault(str(i), []).append(device)
                management_client.add_devices_to_group(
                    name=group, request_body=[device]
                )

        all_groups = management_client.list_groups()
        assert len(all_groups) == 10

        for idx, g in enumerate(all_groups):
            assert sorted(management_client.get_devices_in_group(name=g)) == sorted(
                devices_in_groups[str(idx)]
            )
