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

from client import make_authenticated_client


@pytest.mark.usefixtures("clean_db")
class TestInventorySorting:
    def test_inventory_sorting(self, clean_db, inventory_attributes):
        internal_client = oas.InventoryInternalApi()
        management_client = oas.InventoryManagementApi(
            make_authenticated_client(is_device=False)
        )
        assert len(management_client.list_groups()) == 0
        numbers = [100, 1000, 1, 999]

        for n in numbers:
            it = list(inventory_attributes)
            it.append(oas.Attribute(name="number", value=oas.AttributeValue(n)))

            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.initialize_device(
                tenant_id="", device_new=oas.DeviceNew(id=did, attributes=it)
            )

        t = []
        r = management_client.list_device_inventories(sort="number:asc")
        for deviceInventoryList in r:
            if deviceInventoryList is None or deviceInventoryList.attributes is None:
                continue
            for i in deviceInventoryList.attributes:
                if i.name == "number":
                    t.append(i.value.actual_instance)

        assert sorted(numbers) == t

        t = []
        r = management_client.list_device_inventories(sort="number:desc")
        for deviceInventoryList in r:
            if deviceInventoryList is None or deviceInventoryList.attributes is None:
                continue
            for i in deviceInventoryList.attributes:
                if i.name == "number":
                    t.append(i.value.actual_instance)

        assert sorted(numbers, reverse=True) == t
