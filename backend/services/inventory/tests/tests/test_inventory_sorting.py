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
from common import (
    inventory_attributes,
    management_client,
    internal_client,
    clean_db,
    mongo,
)

import pytest
import os


@pytest.mark.usefixtures("clean_db")
class TestInventorySorting:
    def test_inventory_sorting(
        self, management_client, internal_client, inventory_attributes
    ):
        numbers = [100, 1000, 1, 999]

        for n in range(20):
            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.create_device(did, inventory_attributes)

        for n in numbers:
            it = list(inventory_attributes)
            it.append(internal_client.Attribute(name="number", value=n))

            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.create_device(did, it)

        t = []
        r = management_client.getAllDevices(sort="number:asc")
        for deviceInventoryList in r:
            for i in deviceInventoryList.attributes:
                if i.name == "number":
                    t.append(i.value)

        assert sorted(numbers) == t

        t = []
        r = management_client.getAllDevices(sort="number:desc")
        for deviceInventoryList in r:
            for i in deviceInventoryList.attributes:
                if i.name == "number":
                    t.append(i.value)

        assert sorted(numbers, reverse=True) == t
