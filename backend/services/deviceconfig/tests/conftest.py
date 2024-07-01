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

import devices_api
import internal_api
import management_api


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
    devices_api.Configuration.set_default(
        devices_api.Configuration(
            host="http://" + host + "/api/devices/v1/deviceconfig"
        )
    )
    internal_api.Configuration.set_default(
        internal_api.Configuration(
            host="http://" + host + "/api/internal/v1/deviceconfig"
        )
    )
    management_api.Configuration.set_default(
        management_api.Configuration(
            host="http://" + host + "/api/management/v1/deviceconfig"
        )
    )


@pytest.fixture(scope="session")
def mongo():
    return pymongo.MongoClient("mongodb://mender-mongo")


def mongo_cleanup(client):
    dbs = client.list_database_names()
    for db in dbs:
        if db in ["local", "admin", "config"]:
            continue
        client.drop_database(db)


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
    client = internal_api.InternalAPIClient()
    client.provision_tenant(new_tenant=internal_api.NewTenant(tenant_id=tenant_id))
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
