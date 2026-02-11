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
import requests
import signal

import bson
import pymongo

import mender_client
from mender_client.api import DeviceConfigureInternalAPIApi


MMOCK_URI = "http://mmock:8082"


def pytest_addoption(parser):
    parser.addoption(
        "--host",
        action="store",
        default=os.environ["TESTING_HOST"]
        if "TESTING_HOST" in os.environ
        else "localhost",
        help="Address for host hosting deviceconfig API (env: TEST_HOST)",
    )


def pytest_configure(config):
    host = config.getoption("host")
    mender_client.Configuration.set_default(
        mender_client.Configuration(host="http://" + host)
    )


@pytest.fixture(scope="session")
def mongo():
    return pymongo.MongoClient("mongodb://mender-mongo")


def mongo_cleanup(mongo: pymongo.MongoClient):
    dbs = mongo.list_databases(
        filter={"name": {"$nin": ["admin", "config", "local", "workflows"]}},
        nameOnly=True,
    )
    for db_name in (db["name"] for db in dbs):
        db = mongo[db_name]
        for coll in db.list_collection_names(
            filter={
                "name": {"$ne": "migration_info"},
                "$or": [
                    {"options.capped": {"$exists": False}},
                    {"options.capped": False},
                ],
            }
        ):
            db[coll].delete_many({})


@pytest.fixture(scope="function")
def clean_mongo(mongo):
    mongo_cleanup(client=mongo)
    yield mongo
    mongo_cleanup(client=mongo)


@pytest.fixture(scope="function")
def tenant(tenant_id=None):
    """
    This fixture provisions a new tenant database.
    :param tenant_id: can be indirectly overridden with
                      @pytest.mark.fixture decorator.
    """
    if tenant_id is None:
        tenant_id = str(bson.objectid.ObjectId())
    client = DeviceConfigureInternalAPIApi()
    client.device_config_internal_provision_tenant(new_tenant=mender_client.NewTenant(tenant_id=tenant_id))
    yield tenant_id


@pytest.fixture(scope="function")
def timeout(request, timeout_sec=30):
    """"""
    alrm_handler = signal.getsignal(signal.SIGALRM)

    def timeout(signum, frame):
        raise TimeoutError("%s did not finish in time" % request.function.__name__)

    signal.signal(signal.SIGALRM, timeout)
    signal.alarm(timeout_sec)
    yield
    signal.signal(signal.SIGALRM, alrm_handler)


@pytest.fixture(scope="function")
def mmock_url():
    res = requests.get(MMOCK_URI + "/api/request/reset")
    assert res.status_code == 200
    yield MMOCK_URI
