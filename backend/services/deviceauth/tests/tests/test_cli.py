#!/usr/bin/python
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
import asyncio
import json
import tornado.web
from contextlib import contextmanager

import pytest

import mockserver
from common import (
    DevAuthorizer,
    Device,
    clean_db,
    cli,
    device_api,
    device_auth_req,
    get_fake_tenantadm_addr,
    get_fake_workflows_addr,
    internal_api,
    make_devices,
    make_fake_tenant_token,
    management_api,
    mongo,
)

DB_NAME = "deviceauth"
DB_MIGRATION_COLLECTION = "migration_info"
DB_DEVICES_COLLECTION = "devices"
DB_VERSION = "1.10.0"
TENANT_IDS = ["tid1", "tid2", "tid3"]


MIGRATED_TENANT_DBS = {
    "tenant-stale-1": "0.0.1",
    "tenant-stale-2": "0.2.0",
    "tenant-stale-3": "1.0.0",
    "tenant-current": "1.10.0",
    "tenant-future": "2.0.0",
}


@pytest.fixture(scope="function")
def fake_migrated_db(clean_db, mongo, request):
    """Init a default db to version passed in 'request'. Does not run the actual
    migrations, just records DB version in proper collection."""
    version = request.param
    mongo_set_version(mongo, DB_NAME, version)


def mongo_set_version(mongo, dbname, version):
    major, minor, patch = [int(x) for x in version.split(".")]

    version = {
        "major": major,
        "minor": minor,
        "patch": patch,
    }

    mongo[dbname][DB_MIGRATION_COLLECTION].insert_one({"version": version})


def make_tenant_db(tenant_id):
    return "{}-{}".format(DB_NAME, tenant_id)


def mongo_insert_tenant_data(mongo, dbname, tid):
    mongo[dbname][DB_DEVICES_COLLECTION].insert_one({"tenant_id": tid})


class TestMigration:
    @staticmethod
    def verify_db_and_collections(client, dbname):
        dbs = client.list_database_names()
        assert dbname in dbs

        colls = client[dbname].list_collection_names()
        assert DB_MIGRATION_COLLECTION in colls

    @staticmethod
    def verify_migration(db, expected_version):
        major, minor, patch = [int(x) for x in expected_version.split(".")]
        version = {
            "version.major": major,
            "version.minor": minor,
            "version.patch": patch,
        }

        mi = db[DB_MIGRATION_COLLECTION].find_one(version)
        print("found migration:", mi)

        assert mi

    @staticmethod
    def verify(cli, mongo, dbname, version):
        TestMigration.verify_db_and_collections(mongo, dbname)
        TestMigration.verify_migration(mongo[dbname], version)


class TestListTenants:
    def test_no_tenants(self, cli):
        assert not cli.list_tenants()[1], "expected no tenants (empty output)"


# runs 'last' since it drops/reinits the default db, which breaks deviceauth under test:
# - the indexes are destroyed by 'clean_db/fake_migrated_db'
# - even though we ensure them on every write - mgo caches their names
#   and thinks they're still there
# - writes by deviceauth are then essentially incorrect (duplicate data) and other tests fail
@pytest.mark.last
class TestCliMigrate:
    def test_ok_no_db(self, cli, clean_db, mongo):
        code, _, _ = cli.migrate()
        assert code == 0
        TestMigration.verify(cli, mongo, DB_NAME, DB_VERSION)

    @pytest.mark.parametrize("fake_migrated_db", ["0.0.1"], indirect=True)
    def test_ok_stale_db(self, cli, fake_migrated_db, mongo):
        code, _, _ = cli.migrate()
        assert code == 0
        TestMigration.verify(cli, mongo, DB_NAME, DB_VERSION)

    @pytest.mark.parametrize("fake_migrated_db", ["1.10.0"], indirect=True)
    def test_ok_current_db(self, cli, fake_migrated_db, mongo):
        code, _, _ = cli.migrate()
        assert code == 0
        TestMigration.verify(cli, mongo, DB_NAME, DB_VERSION)

    @pytest.mark.parametrize("fake_migrated_db", ["2.0.0"], indirect=True)
    def test_ok_future_db(self, cli, fake_migrated_db, mongo):
        code, _, _ = cli.migrate()
        assert code == 0
        TestMigration.verify(cli, mongo, DB_NAME, "2.0.0")
