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

import os

import pymongo
import pytest
import requests

from client import CliIoTManager, MMockAPIClient, ManagementAPIClient

MMOCK_URL = os.environ.get("MMOCK_URL", "http://mmock:8081")
MONGO_URL = os.environ.get("MONGO_URL", "mongodb://mender-mongo")


@pytest.fixture(scope="session")
def mongo():
    return pymongo.MongoClient(MONGO_URL, uuidRepresentation="standard")


@pytest.fixture(scope="function")
def clean_mmock(request):
    mmock = MMockAPIClient(MMOCK_URL)
    mmock.reset()
    mmock.set_scenario(scenario="test", state=request.function.__name__)
    if request.cls:
        mmock.set_scenario(scenario="class", state=request.cls.__name__)
    return mmock


def mongo_cleanup(mgo: pymongo.MongoClient):
    dirty_dbs = [
        db["name"]
        for db in mgo.list_databases(
            filter={"name": {"$nin": ["admin", "config", "local"]}, "empty": False}
        )
    ]
    for db in dirty_dbs:
        for coll in mgo[db].list_collections(
            filter={
                "name": {"$ne": "migration_info"},
                "$or": [
                    {"options.capped": {"$exists": False}},
                    {"options.capped": False},
                ],
            }
        ):
            mgo[db][coll["name"]].delete_many({})


@pytest.fixture(scope="function")
def clean_mongo(mongo):
    mongo_cleanup(mongo)
    yield mongo


@pytest.fixture(scope="session")
def cli_iot_manager(pytestconfig):
    return CliIoTManager(pytestconfig.getoption("host").split(":")[0])
