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

import requests
import pytest
import os


@pytest.mark.usefixtures("clean_db")
class TestInventorySearching:
    def test_inventory_searching(
        self, management_client, internal_client, inventory_attributes
    ):
        extra_inventory_items = {
            "users_logged_in": 100,
            "open_connections": 1231,
            "open_ports": 523,
        }

        for i in extra_inventory_items.keys():
            it = list(inventory_attributes)
            it.append(
                management_client.inventoryAttribute(
                    name=i, value=extra_inventory_items[i]
                )
            )

            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.create_device(did, it)

        r = requests.get(
            management_client.client.swagger_spec.api_url + "/devices",
            params=({"users_logged_in": 100}),
            verify=False,
        )
        assert len(r.json()) == 1

        r = requests.get(
            management_client.client.swagger_spec.api_url + "/devices",
            params=({"open_connections": 1231}),
            verify=False,
        )
        assert len(r.json()) == 1
