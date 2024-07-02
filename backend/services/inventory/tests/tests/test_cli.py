#!/usr/bin/python
# Copyright 2022 Northern.tech AS
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
from common import cli, mongo, clean_db


DB_NAME = "inventory"
MIGRATION_COLLECTION = "migration_info"
DB_VERSION = "0.2.0"


class TestMigration:
    @staticmethod
    def verify_db_and_collections(client, dbname):
        dbs = client.list_database_names()
        assert dbname in dbs

        colls = client[dbname].list_collection_names()
        assert MIGRATION_COLLECTION in colls

    @staticmethod
    def verify_migration(db, expected_version):
        major, minor, patch = [int(x) for x in expected_version.split(".")]
        version = {
            "version.major": major,
            "version.minor": minor,
            "version.patch": patch,
        }

        mi = db[MIGRATION_COLLECTION].find_one(version)
        print("found migration:", mi)
        assert mi


class TestCliMigration(TestMigration):
    def test_migrate(self, cli, clean_db, mongo):
        code, _, stderr = cli.migrate()
        assert code == 0, f"Unexpected exit code, stderr: {stderr}"

        TestMigration.verify_db_and_collections(mongo, DB_NAME)
        TestMigration.verify_migration(mongo[DB_NAME], DB_VERSION)


class TestCliMigrationMultitenant(TestMigration):
    def test_migrate(self, cli, clean_db, mongo):
        code, _, stderr = cli.migrate(tenant_id="foobar")
        assert code == 0, f"Unexpected exit code, stderr: {stderr}"

        tenant_db = DB_NAME + "-foobar"
        TestMigration.verify_db_and_collections(mongo, tenant_db)
        TestMigration.verify_migration(mongo[tenant_db], DB_VERSION)
