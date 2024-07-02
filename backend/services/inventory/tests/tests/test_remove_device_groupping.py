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

import os
import pytest


@pytest.mark.usefixtures("clean_db")
class TestGroupRemoving:
    def test_delete_device(
        self, management_client, internal_client, inventory_attributes
    ):
        d1 = "".join([format(i, "02x") for i in os.urandom(128)])
        internal_client.create_device(d1, inventory_attributes)

        g1 = "group-test-3"

        management_client.addDeviceToGroup(
            device=d1, group=management_client.group(group=g1)
        )
        assert len(management_client.getGroupDevices(g1)) == 1

        management_client.deleteDeviceInGroup(device=d1, group=g1, expected_error=False)
        assert len(management_client.getGroupDevices(g1, expected_error=True)) == 0

    def test_delete_device_non_existent_1(self, management_client):
        """ Delete non-existent device from non-existent group """
        g1 = "group-test-3-non-existent"
        management_client.deleteDeviceInGroup(
            device="404 device", group=g1, expected_error=True
        )

    def test_delete_device_non_existent_2(
        self, management_client, internal_client, inventory_attributes
    ):
        """ Delete existent device from non-existent group """
        d1 = "".join([format(i, "02x") for i in os.urandom(128)])
        internal_client.create_device(d1, inventory_attributes)

        management_client.deleteDeviceInGroup(
            device=d1, group="404 group", expected_error=True
        )
