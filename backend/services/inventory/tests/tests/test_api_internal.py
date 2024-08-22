#!/usr/bin/python
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
from common import (
    internal_client,
    management_client,
    mongo,
    clean_db,
    inventory_attributes,
)
import bravado
import pytest


class TestInternalApiTenantCreate:
    def test_create_ok(self, internal_client, clean_db):

        _, r = internal_client.create_tenant("foobar")
        assert r.status_code == 201

        assert "inventory-foobar" in clean_db.list_database_names()
        assert "migration_info" in clean_db["inventory-foobar"].list_collection_names()

    def test_create_twice(self, internal_client, clean_db):

        _, r = internal_client.create_tenant("foobar")
        assert r.status_code == 201

        # creating once more should not fail
        _, r = internal_client.create_tenant("foobar")
        assert r.status_code == 201

    def test_create_empty(self, internal_client):
        try:
            _, r = internal_client.create_tenant("")
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400


class TestInternalApiDeviceCreate:
    def test_create_ok(
        self, internal_client, management_client, clean_db, inventory_attributes,
    ):
        devid = "".join([format(i, "02x") for i in os.urandom(128)])
        _, r = internal_client.create_device(
            device_id=devid, attributes=inventory_attributes
        )
        assert r.status_code == 201

        dev = management_client.getDevice(device_id=devid)

        self._verify_inventory(inventory_attributes, dev.attributes)

    def test_create_twice_ok(
        self, internal_client, management_client, clean_db, inventory_attributes,
    ):
        # insert first device
        devid = "".join([format(i, "02x") for i in os.urandom(128)])
        _, r = internal_client.create_device(
            device_id=devid, attributes=inventory_attributes
        )
        assert r.status_code == 201

        # add extra attribute, modify existing
        new_attr = management_client.inventoryAttribute(
            name="new attr", value="new value", scope="inventory", description="desc",
        )

        existing = inventory_attributes[0]
        existing.value = "newval"
        existing.description = "newdesc"

        new_attrs = [new_attr, existing]

        # inventory_attributes will now act as 'expected' output attrs
        inventory_attributes.append(new_attr)

        # insert 'the same' device
        _, r = internal_client.create_device(device_id=devid, attributes=new_attrs)
        assert r.status_code == 201

        # verify update
        dev = management_client.getDevice(devid)

        self._verify_inventory(inventory_attributes, dev.attributes)

    def _verify_inventory(self, expected, inventory):
        # Filter only attributes within the inventory scope
        expected_inventory = list(filter(lambda a: a.scope == "inventory", expected))
        inventory = list(filter(lambda a: a.scope == "inventory", inventory))
        assert len(inventory) == len(
            expected_inventory
        ), "expected: %s / actual: %s" % (inventory, expected_inventory)
        for e in expected_inventory:
            found = [
                f
                for f in inventory
                if (
                    f.name == e.name
                    and f.value == e.value
                    and f.description == e.description
                )
            ]
            assert len(found) == 1, "Inventory data is incorrect"
