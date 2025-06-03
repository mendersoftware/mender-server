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

import json
import os

import pytest

import openapi_client as oas

from client import make_authenticated_client


@pytest.mark.usefixtures("clean_db")
class TestInventorySearching:
    def test_inventory_searching(self, inventory_attributes):
        api_client = make_authenticated_client(is_device=False)
        internal_client = oas.InventoryInternalApi()
        extra_inventory_items = {
            "users_logged_in": 100,
            "open_connections": 1231,
            "open_ports": 523,
        }

        for key in extra_inventory_items.keys():
            it = list(inventory_attributes)
            it.append(
                oas.Attribute(
                    name=key, value=oas.AttributeValue(extra_inventory_items[key])
                )
            )

            did = "".join([format(i, "02x") for i in os.urandom(128)])
            internal_client.initialize_device(
                tenant_id="", device_new=oas.DeviceNew(id=did, attributes=it)
            )

        # HACK: Because of the API parameters for searching is not proper,
        # we have to interact with the api_client interface directly.
        for q in [{"users_logged_in": 100}, {"open_connections": 1231}]:
            req = api_client.param_serialize(
                "GET",
                "/api/management/v1/inventory/devices",
                query_params=q,
                header_params={
                    "Authorization": f"Bearer {api_client.configuration.access_token}"
                },
            )
            rsp = api_client.call_api(*req)
            rsp.read()
            rsp_json = json.loads(rsp.data.decode("utf-8"))
            assert len(rsp_json) == 1
