# Copyright 2026 Northern.tech AS
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
import pytest

from testutils.infra.mongo import MongoClient
from testutils.common import wait_until_healthy

wait_until_healthy("backend-tests")


@pytest.fixture(scope="session")
def mongo():
    return MongoClient("mender-mongo:27017")


@pytest.fixture(scope="function")
def clean_mongo(mongo):
    """Fixture setting up a clean (i.e. empty database). Yields
    pymongo.MongoClient connected to the DB."""
    mongo.cleanup()
    yield mongo.client
