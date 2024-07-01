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
import pytest
import requests

from redo import retrier

MMOCK_URI = "http://mmock:8082"
WORKFLOWS_URI = "http://workflows-server:8080"


@pytest.fixture(scope="session")
def workflows_url():
    timeout = 10
    for _ in retrier(attempts=timeout, sleeptime=1):
        try:
            rsp = requests.request("GET", WORKFLOWS_URI + "/api/v1/health")
        except requests.exceptions.ConnectionError:
            continue
        if rsp.status_code < 300:
            break
        else:
            raise TimeoutError(
                f"Timed out waiting for service 'workflows' to become healthy"
            )
    yield WORKFLOWS_URI


@pytest.fixture(scope="function")
def mmock_url():
    res = requests.get(MMOCK_URI + "/api/request/reset")
    assert res.status_code == 200
    yield MMOCK_URI
