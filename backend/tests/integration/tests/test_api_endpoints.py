# Copyright 2024 Northern.tech AS
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


import glob
import json
import logging
import os

import pytest
import requests
import yaml

from testutils.api.client import GATEWAY_HOSTNAME
from testutils.infra.container_manager.kubernetes_manager import isK8S

logging.basicConfig(format="%(asctime)s %(message)s")
logger = logging.getLogger("test_api_endpoints")
logger.setLevel(logging.INFO)


def get_api_docs(repo):
    files = glob.glob(os.path.join(os.sep + "docs", repo, "*.yml"))
    for file in files:
        basename = os.path.basename(file)
        kind = (
            basename.startswith("management_")
            and "management"
            or basename.startswith("devices_")
            and "devices"
            or "internal"
        )
        with open(file) as f:
            data = yaml.load(f, Loader=yaml.FullLoader)
            yield kind, data


def get_api_endpoints(repo):
    for kind, data in get_api_docs(repo):
        if data.get("swagger"):
            scheme, host, base_path = (
                data["schemes"][0],
                data["host"],
                data.get("basePath", "/"),
            )
        elif data.get("openapi"):
            parts = data["servers"][0]["url"].split("/", 3)
            scheme = parts[0].rstrip(":")
            host, base_path = parts[2:4]
        else:
            logger.error(f"unknown specification file: {json.dumps(data)}")
            raise ValueError(
                "Unknown specification file, only swagger and openapi 3 are supported!"
            )
        for path, path_value in data["paths"].items():
            for method, definition in path_value.items():
                # ignore the shutdown endpoint
                if path.rstrip("/").endswith("/shutdown"):
                    continue
                returns_401 = (
                    len(definition.get("security") or ()) > 0
                    or len(data.get("security") or ()) > 0
                    or path.rstrip("/").endswith("/verify")  # JWT token verifications
                    or path.rstrip("/").endswith("/2faqr")  # 2FA QR code
                    or path.rstrip("/").endswith("/2faverify")  # 2FA code verification
                    or path.rstrip("/").endswith(
                        "/auth/magic/{id}"
                    )  # token authentication
                )
                if path.rstrip("/").endswith("/auth_requests"):
                    returns_401 = False  # device auth endpoint,
                    # the way we do these tests will return 400 before
                    # 401 so it cannot be tested here like that

                yield {
                    "kind": kind,
                    "returns_401": returns_401,
                    "method": method,
                    "scheme": scheme,
                    "host": host,
                    "path": base_path.rstrip("/") + path,
                }


def get_all_api_endpoints(repos):
    for repo in repos:
        for endpoint in get_api_endpoints(repo):
            yield (
                endpoint["kind"],
                endpoint["returns_401"],
                endpoint["method"],
                endpoint["scheme"],
                endpoint["host"],
                endpoint["path"],
            )


class BaseTestAPIEndpoints:
    def do_test_api_endpoints(
        self, kind, returns_401, method, scheme, host, path, get_endpoint_url
    ):
        assert method in ("get", "post", "put", "delete", "patch")
        requests_method = getattr(requests, method)
        if host == "hosted.mender.io" or kind in ("management", "devices"):
            base_url = f"{scheme}://{GATEWAY_HOSTNAME}"
        else:
            base_url = get_endpoint_url(f"{scheme}://{host}")
        r = requests_method(
            base_url + "/" + path.lstrip("/"), verify=False, timeout=2.0
        )
        if returns_401:
            assert 401 == int(r.status_code)
        else:
            assert 401 != int(r.status_code)
            assert (
                int(r.status_code) >= 200
                and int(r.status_code) < 500
                and int(r.status_code) != 405
            )


class TestAPIEndpoints(BaseTestAPIEndpoints):
    REPOS = (
        "deployments",
        "deviceauth",
        "deviceconfig",
        "deviceconnect",
        "inventory",
        "iot-manager",
        "useradm",
        "workflows",
    )

    @pytest.mark.parametrize(
        "kind,returns_401,method,scheme,host,path", get_all_api_endpoints(REPOS),
    )
    def test_api_endpoints(
        self, kind, returns_401, method, scheme, host, path, get_endpoint_url
    ):
        self.do_test_api_endpoints(
            kind, returns_401, method, scheme, host, path, get_endpoint_url
        )


class TestAPIEndpointsEnterprise(BaseTestAPIEndpoints):
    REPOS = (
        "auditlogs",
        "deployments",
        "deviceauth",
        "deviceconfig",
        "deviceconnect",
        "devicemonitor",
        "inventory",
        "iot-manager",
        "tenantadm",
        "useradm",
        "workflows",
    )

    @pytest.mark.parametrize(
        "kind,returns_401,method,scheme,host,path", get_all_api_endpoints(REPOS),
    )
    def test_api_endpoints(
        self, kind, returns_401, method, scheme, host, path, get_endpoint_url
    ):
        self.do_test_api_endpoints(
            kind, returns_401, method, scheme, host, path, get_endpoint_url
        )
