# Copyright 2023 Northern.tech AS
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
    management_client_v2,
    internal_client,
    clean_db,
    mongo,
)

import bravado
import pytest


@pytest.mark.usefixtures("clean_db")
class TestGetAttributes:
    def test_get_attributes(
        self,
        management_client,
        management_client_v2,
        internal_client,
        inventory_attributes,
    ):
        # NOTE: bravado request and response checks are enabled for management API v2 client
        management_client_v2.getFiltersAttributes()
        attributeList = []
        attr = management_client.inventoryAttribute(
            name="foo", value="bar", scope="inventory", description="baz"
        )
        attributeList.append(attr)

        did = "some-device-id"
        internal_client.create_device(did, attributeList)
        res = management_client_v2.getFiltersAttributes()
        for attr in res:
            if attr.scope == "inventory":
                assert res[0].name == "foo"
                assert res[0].count == 1
