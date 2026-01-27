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

import pytest
import urllib3

from mender_client import ApiClient, Configuration

# See https://docs.pytest.org/en/latest/writing_plugins.html#assertion-rewriting
pytest.register_assert_rewrite("testutils")


urllib3.disable_warnings()


def pytest_addoption(parser):
    parser.addoption(
        "--mender-url",
        action="store",
        default=(
            os.environ["TEST_MENDER_URL"]
            if "TEST_MENDER_URL" in os.environ
            else "https://traefik"
        ),
        help="Address for host hosting deviceconnect API (env: TEST_MENDER_URL)",
    )


def pytest_configure(config: pytest.Config):
    mender_url = config.getoption("mender_url")
    client_config = Configuration(host=mender_url)
    client_config.verify_ssl = False
    Configuration.set_default(client_config)
    ApiClient.set_default(ApiClient(client_config))
