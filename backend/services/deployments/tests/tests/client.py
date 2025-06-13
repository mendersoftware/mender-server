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

import hashlib
import hmac
import io
import json
import logging
import os.path
import random
import socket

import jwt
from base64 import b64encode
from datetime import datetime
from collections import OrderedDict
from contextlib import contextmanager
from uuid import uuid4

import docker
import pytest  # noqa
import pytz
import requests

from config import pytest_config

from bson.objectid import ObjectId
from base64 import b64encode
from datetime import datetime, timedelta
import hmac
import uuid
import management_v2 as mv2
import management_v1 as mv1
import devices_v1 as dv1
import internal_v1 as iv1

DEPLOYMENTS_BASE_URL = "http://{}/api/{}/v1/deployments"


class BaseApiClient:
    def make_api_url(self, path=None):
        if path is not None:
            return os.path.join(
                self.api_url, path if not path.startswith("/") else path[1:]
            )
        return self.api_url


class RequestsApiClient(requests.Session):
    # TODO: convert to make_session() helper
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.verify = False


class ArtifactsClientError(Exception):
    def __init__(self, message="", response=None):
        self.response = response
        super().__init__(message)


def create_authz(sub, is_user=True, tenant_id=None):
    hdr = {"alg": "HS256", "typ": "JWT"}
    claims = {"sub": sub, "mender.user": is_user}
    if tenant_id is not None:
        claims["mender.tenant"] = tenant_id

    def _b64encode_url(obj):
        if isinstance(obj, dict):
            return (
                b64encode(json.dumps(obj).encode(), b"-_").decode("UTF-8").rstrip("=")
            )
        elif isinstance(obj, str):
            return b64encode(obj.encode(), b"-_").decode("UTF-8").rstrip("=")
        else:
            return b64encode(obj, b"-_").decode("UTF-8").rstrip("=")

    hdr64 = _b64encode_url(hdr)
    claims64 = _b64encode_url(claims)
    jwt = f"{hdr64}.{claims64}"
    sig = hmac.HMAC(b"supersecretsecret", jwt.encode(), digestmod=hashlib.sha256)
    jwt += f".{_b64encode_url(sig.digest())}"
    return jwt


class ArtifactsClient(BaseApiClient):
    log = logging.getLogger("client.Client")

    def __init__(self, sub=None, tenant_id=None):
        if sub is None:
            sub = str(uuid4())
        self._jwt = create_authz(sub, tenant_id=tenant_id)
        self.api_url = DEPLOYMENTS_BASE_URL.format(
            pytest_config.getoption("host"), "management"
        )
        super().__init__()

    @staticmethod
    def make_upload_meta(meta):
        order = ["description", "size", "artifact_id", "artifact"]

        upload_meta = OrderedDict()
        for entry in order:
            if entry in meta:
                upload_meta[entry] = meta[entry]
        return upload_meta

    def get_jwt(self):
        return self._jwt

    def add_artifact(self, description="", size=0, data=None):
        """Create new artifact with provided upload data. Data must be a file like
        object.

        Returns artifact ID or raises ArtifactsClientError if response checks
        failed
        """
        # prepare upload data for multipart/form-data
        files = ArtifactsClient.make_upload_meta(
            {
                "description": (None, description),
                "size": (None, str(size)),
                "artifact": ("firmware", data, "application/octet-stream", {}),
            }
        )
        rsp = requests.post(
            self.make_api_url("/artifacts"),
            files=files,
            verify=False,
            headers={"Authorization": f"Bearer {self._jwt}"},
        )
        # should have be created
        try:
            assert rsp.status_code == 201
            loc = rsp.headers.get("Location", None)
            assert loc
        except AssertionError:
            raise ArtifactsClientError("add failed", rsp)

        loc = rsp.headers.get("Location", None)
        artid = os.path.basename(loc)
        return artid

    @staticmethod
    def make_generate_meta(meta):
        order = [
            "name",
            "description",
            "device_types_compatible",
            "type",
            "args",
            "file",
        ]

        upload_meta = OrderedDict()
        for entry in order:
            if entry in meta:
                upload_meta[entry] = meta[entry]
        return upload_meta

    def generate_artifact(
        self,
        name="",
        description="",
        device_types_compatible="",
        type="",
        args="",
        data=None,
    ):
        """Generate a new artifact with provided upload data.
        Data must be a file like object.

        Returns artifact ID or raises ArtifactsClientError if response checks
        failed
        """
        # prepare upload data for multipart/form-data
        files = ArtifactsClient.make_generate_meta(
            {
                "name": (None, name),
                "description": (None, description),
                "device_types_compatible": (None, device_types_compatible),
                "type": (None, type),
                "args": (None, args),
                "file": ("firmware", data, "application/octet-stream", {}),
            }
        )
        rsp = requests.post(
            self.make_api_url("/artifacts/generate"),
            files=files,
            verify=False,
            headers={"Authorization": f"Bearer {self._jwt}"},
        )
        # should have be created
        try:
            assert rsp.status_code == 201
            loc = rsp.headers.get("Location", None)
            assert loc
        except AssertionError:
            raise ArtifactsClientError("add failed", rsp)

        loc = rsp.headers.get("Location", None)
        artid = os.path.basename(loc)
        return artid

    def delete_artifact(self, artid=""):
        try:
            management_v1_client(jwt=self._jwt).delete_artifact(id=artid)
        except mv1.rest.ApiException as e:
            raise ArtifactsClientError("delete failed", e.status)

    def list_artifacts(self):
        rsp = management_v1_client(jwt=self._jwt).list_artifacts_with_http_info()
        try:
            assert rsp[1] == 200
        except AssertionError:
            raise ArtifactsClientError("get failed", rsp)
        return rsp[0]

    def show_artifact(self, artid=""):
        rsp = management_v1_client(jwt=self._jwt).show_artifact_with_http_info(id=artid)
        try:
            assert rsp[1] == 200
        except AssertionError:
            raise ArtifactsClientError("get failed", rsp)
        return rsp[0]

    @contextmanager
    def with_added_artifact(self, description="", size=0, data=None):
        """Acts as a context manager, adds artifact and yields artifact ID and deletes
        it upon completion"""
        artid = None
        try:
            artid = self.add_artifact(description=description, size=size, data=data)
            yield artid
        finally:
            if artid is not None:
                self.delete_artifact(artid)

    class UploadURL:
        def __init__(self, identifier: str, uri: str, expire: str):
            self._id = identifier
            self._uri = uri
            self._expire = expire

        @property
        def id(self) -> str:
            return self._id

        @property
        def uri(self) -> str:
            return self._uri

        @property
        def expire(self) -> str:
            return self._expire

    def make_upload_url(self):
        rsp = requests.post(
            self.make_api_url("/artifacts/directupload"),
            "",
            headers={"Authorization": f"Bearer {self._jwt}"},
        )
        try:
            assert rsp.status_code == 200
        except AssertionError:
            raise ArtifactsClientError(
                f"unexpected HTTP status code: {rsp.status_code}", rsp
            )
        body = rsp.json()
        return ArtifactsClient.UploadURL(body["id"], body["uri"], body["expire"])

    def complete_upload(self, identifier, body=""):
        rsp = requests.post(
            self.make_api_url(f"/artifacts/directupload/{identifier}/complete"),
            data=body,
            headers={
                "Authorization": f"Bearer {self._jwt}",
                "Content-Type": "application/json",
            },
        )
        try:
            assert rsp.status_code == 202
        except AssertionError:
            raise ArtifactsClientError(
                f"unexpected HTTP status code: {rsp.status_code}", rsp
            )
        return rsp


class SimpleArtifactsClient(ArtifactsClient):
    """Simple swagger based client for artifacts. Cannot be used as Pytest base class"""

    def __init__(self):
        super().__init__()


class DeploymentsClient(BaseApiClient):
    log = logging.getLogger("client.Client")

    def __init__(self):
        self.api_url = DEPLOYMENTS_BASE_URL.format(
            pytest_config.getoption("host"), "management"
        )
        super().__init__()

    def make_new_deployment(self, *args, **kwargs):
        return mv1.NewDeployment(*args, **kwargs)

    def add_deployment(self, dep):
        """Posts new deployment `dep`"""
        r = management_v1_client(jwt="foo").create_deployment_with_http_info(dep)
        loc = r[2]["Location"]
        depid = os.path.basename(loc)

        self.log.debug("added new deployment with ID: %s", depid)
        return depid

    def abort_deployment(self, depid):
        """Abort deployment with `ID `depid`"""
        management_v1_client(jwt="foo").abort_deployment_with_http_info(
            deployment_id=depid,
            abort_deployment_request=mv1.AbortDeploymentRequest(status="aborted"),
        )

    @contextmanager
    def with_added_deployment(self, dep):
        """Acts as a context manager, adds artifact and yields artifact ID and deletes
        it upon completion"""
        depid = self.add_deployment(dep)
        yield depid
        try:
            self.abort_deployment(depid)
        except mv1.rest.ApiException:
            self.log.warning("deployment: %s already finished", depid)

    def verify_deployment_stats(self, depid, expected):
        stats = management_v1_client(jwt="foo").deployment_status_statistics(
            deployment_id=depid
        )
        stat_names = [
            "success",
            "pending",
            "failure",
            "downloading",
            "installing",
            "rebooting",
            "noartifact",
            "already_installed",
            "aborted",
            "pause_before_installing",
            "pause_before_committing",
            "pause_before_rebooting",
        ]
        for s in stat_names:
            exp = expected.get(s, 0)
            current = getattr(stats, s) or 0
            assert exp == current


class DeviceClient(BaseApiClient):
    """Swagger based device API client. Can be used a Pytest base class"""

    spec_option = "device_spec"
    log = logging.getLogger("client.DeviceClient")

    def __init__(self):
        self.api_url = DEPLOYMENTS_BASE_URL.format(
            pytest_config.getoption("host"), "devices"
        )
        super().__init__()

    def get_next_deployment(self, token="", artifact_name="", device_type=""):
        """Obtain next deployment"""
        return devices_v1_client(jwt=token).check_update(
            artifact_name=artifact_name, device_type=device_type
        )

    def report_status(self, token="", devdepid=None, status=None):
        """Report device deployment status"""
        res = devices_v1_client(jwt=token).update_deployment_status(
            id=devdepid, deployment_status=dv1.DeploymentStatus(status=status)
        )
        return res

    def upload_logs(self, token="", devdepid=None, logs=[]):
        levels = ["info", "debug", "warn", "error", "other"]
        dl = dv1.DeploymentLog(
            messages=[
                {
                    "timestamp": pytz.utc.localize(datetime.now()),
                    "level": random.choice(levels),
                    "message": l,
                }
                for l in logs
            ]
        )
        res = devices_v1_client(jwt=token).report_deployment_log(
            id=devdepid, deployment_log=dl
        )
        return res


class SimpleDeviceClient(DeviceClient):
    """Simple device API client, cannot be used as Pytest tests base class"""

    def __init__(self):
        super().__init__()


class InventoryClientError(Exception):
    pass


class InventoryClient(BaseApiClient, RequestsApiClient):
    def __init__(self):
        self.api_url = "http://%s/api/0.1.0/" % (
            pytest_config.getoption("inventory_host")
        )
        super().__init__()

    def report_attributes(self, devtoken, attributes):
        """Send device attributes to inventory service. Device is identified using
        authorization token passed in `devtoken`. Attributes can be a dict, a
        list, or anything else that can be serialized to JSON. Will raise
        InventoryClientError if request fails.

        """
        rsp = requests.patch(
            self.make_api_url("/attributes"),
            headers={"Authorization": "Bearer " + devtoken},
            json=attributes,
        )
        if rsp.status_code != 200:
            raise InventoryClientError(
                "request failed with status code {}".format(rsp.status_code)
            )


class CliClient:
    exec_path = "/usr/bin/deployments"

    def __init__(self, service="deployments"):
        self.docker = docker.from_env()
        _self = self.docker.containers.list(filters={"id": socket.gethostname()})[0]

        project = _self.labels.get("com.docker.compose.project")
        self.container = self.docker.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    f"com.docker.compose.service={service}",
                ]
            },
            limit=1,
        )[0]

    def migrate(self, tenant=None, **kwargs):
        cmd = [self.exec_path, "migrate"]

        if tenant is not None:
            cmd.extend(["--tenant", tenant])

        code, (stdout, stderr) = self.container.exec_run(cmd, demux=True, **kwargs)
        return code, stdout, stderr


class InternalApiClient(BaseApiClient):
    spec_option = "internal_spec"
    log = logging.getLogger("client.InternalApiClient")

    def __init__(self):
        self.api_url = DEPLOYMENTS_BASE_URL.format(
            pytest_config.getoption("host"), "internal"
        )
        super().__init__()

    def create_tenant(self, tenant_id):
        r = internal_v1_client().create_tenant_with_http_info(
            new_tenant=iv1.NewTenant(tenant_id=tenant_id)
        )
        return r[1]

    def get_jwt(self):
        return self._jwt

    def add_artifact(
        self, tenant_id, description="", size=0, data=None, artifact_id=None
    ):
        """Create new artifact with provided upload data. Data must be a file like
        object.

        Returns artifact ID or raises ArtifactsClientError if response checks
        failed
        """
        # prepare upload data for multipart/form-data
        files = ArtifactsClient.make_upload_meta(
            {
                "artifact_id": artifact_id,
                "description": (None, description),
                "size": (None, str(size)),
                "artifact": ("firmware", data, "application/octet-stream", {}),
            }
        )
        url = self.make_api_url("/tenants/{}/artifacts".format(tenant_id))
        rsp = requests.post(url, files=files, verify=False)
        # should have been created
        try:
            assert rsp.status_code == 201
            loc = rsp.headers.get("Location", None)
            assert loc
        except AssertionError:
            raise ArtifactsClientError("add failed", rsp)
        # return the artifact id
        loc = rsp.headers.get("Location", None)
        artid = os.path.basename(loc)
        return artid

    def set_settings(self, tenant_id, data, status_code=204):
        url = self.make_api_url("/tenants/{}/storage/settings".format(tenant_id))
        resp = requests.put(url, json=data)
        assert resp.status_code == status_code

    def get_settings(self, tenant_id, status_code=200):
        url = self.make_api_url("/tenants/{}/storage/settings".format(tenant_id))
        resp = requests.get(url)
        assert resp.status_code == status_code
        if resp.json() is None:
            return {}
        return resp.json()

    def get_last_device_deployment_status(self, devices_ids, tenant_id):
        return internal_v1_client().get_last_device_deployment_status(
            last_device_deployment_req=iv1.LastDeviceDeploymentReq(
                device_ids=devices_ids
            ),
            tenant_id=tenant_id,
        )


def generate_jwt(tenant_id: str = "", subject: str = "", is_user: bool = True) -> str:
    if len(subject) == 0:
        subject = str(uuid.uuid4())

    payload = {
        "sub": subject,
        "exp": datetime.utcnow()
        + timedelta(hours=1),  # PyJWT will handle the time conversion
        "mender.user": is_user,
        "mender.device": not is_user,
        "mender.tenant": tenant_id,
    }

    return jwt.encode(payload=payload, key="secretJWTkey", algorithm="HS256")


def management_v2_client(tenant_id=None, user_id=None, host=None, jwt=None):
    return management_client(mv2, "v2", tenant_id, user_id, host, jwt)


def management_v1_client(tenant_id=None, user_id=None, host=None, jwt=None):
    return management_client(mv1, "v1", tenant_id, user_id, host, jwt)


def management_client(spec, v, tenant_id=None, user_id=None, host=None, jwt=None):
    api_conf = spec.configuration.Configuration()
    # api_conf=spec.Configuration.get_default()
    if tenant_id is None:
        tenant_id = str(ObjectId())
    if not user_id:
        user_id = str(uuid.uuid4())
    if not jwt:
        jwt = generate_jwt(tenant_id, user_id, is_user=True)
    api_conf.access_token = jwt
    api_conf.api_key = {"Authorization": "Bearer " + jwt}
    if not host:
        host = pytest_config.getoption("host")
    api_conf.host = "http://" + host + "/api/management/" + v + "/deployments"
    return spec.ManagementAPIClient(spec.ApiClient(configuration=api_conf))


def devices_v1_client(tenant_id=None, device_id=None, host=None, jwt=None):
    return device_client(dv1, "v1", tenant_id, device_id, host, jwt)


def device_client(spec, v, tenant_id=None, device_id=None, host=None, jwt=None):
    api_conf = spec.configuration.Configuration()
    if tenant_id is None:
        tenant_id = str(ObjectId())
    if not device_id:
        device_id = str(uuid.uuid4())
    if not jwt:
        jwt = generate_jwt(tenant_id, device_id, is_user=False)
    api_conf.access_token = jwt
    api_conf.api_key = {"Authorization": "Bearer " + jwt}
    if not host:
        host = pytest_config.getoption("host")
    api_conf.host = "http://" + host + "/api/devices/" + v + "/deployments"
    return spec.DeviceAPIClient(spec.ApiClient(configuration=api_conf))


def internal_v1_client(host=None):
    return internal_client(iv1, "v1", host)


def internal_client(spec, v, host=None):
    api_conf = spec.configuration.Configuration()
    if not host:
        host = pytest_config.getoption("host")
    api_conf.host = "http://" + host + "/api/internal/" + v + "/deployments"
    return spec.InternalAPIClient(spec.ApiClient(configuration=api_conf))
