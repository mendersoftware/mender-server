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

import logging
import os.path
import socket

import docker
import requests

import internal_v1
import management_v1
import common


class Response:
    """Simple response object compatible with requests.Response interface"""
    def __init__(self, status_code, text=None, data=None):
        self.status_code = status_code
        self.text = text if text is not None else ""
        self.data = data


class InternalApiClient:
    log = logging.getLogger("client.InternalClient")

    def __init__(self, host):
        self.api_url = "http://%s/api/internal/v1/useradm/" % host
        api_conf = internal_v1.Configuration.get_default_copy()
        self.client = internal_v1.InternalAPIClient(internal_v1.ApiClient(api_conf))

    def make_api_url(self, path):
        return os.path.join(
            self.api_url, path if not path.startswith("/") else path[1:]
        )

    def verify(self, token, uri="/api/management/1.0/auth/verify", method="POST"):
        if not token.startswith("Bearer "):
            token = "Bearer " + token
        r = self.client.verify_jwt_with_http_info(
            authorization=token,
            x_forwarded_uri=uri,
            x_forwarded_method=method,
        )
        return Response(status_code=r.status_code, data=r.data)

    def create_tenant(self, tenant_id):
        tenant = internal_v1.TenantNew(tenant_id=tenant_id)
        r = self.client.useradm_create_tenant_with_http_info(tenant_new=tenant)
        return Response(status_code=r.status_code, data=r.data)

    def create_user_for_tenant(self, tenant_id, user):
        r = self.client.create_user_internal_with_http_info(
            tenant_id=tenant_id, user_new=user
        )
        return Response(status_code=r.status_code, data=r.data)


class ManagementApiClient:
    log = logging.getLogger("client.ManagementClient")

    def __init__(self, host, auth):
        self.api_url = "http://%s/api/management/v1/useradm/" % host
        self.auth = auth
        api_conf = management_v1.Configuration.get_default_copy()
        # Extract token from Bearer header if present
        if auth and "Authorization" in auth:
            token = auth["Authorization"]
            if token.startswith("Bearer "):
                token = token[7:]
            api_conf.access_token = token
        self.client = management_v1.ManagementAPIClient(management_v1.ApiClient(api_conf))

    def make_api_url(self, path):
        return os.path.join(
            self.api_url, path if not path.startswith("/") else path[1:]
        )

    def _get_headers(self, auth=None):
        if auth is None:
            auth = self.auth
        headers = {}
        if auth and "Authorization" in auth:
            headers["Authorization"] = auth["Authorization"]
        return headers

    def get_users(self, auth=None):
        headers = self._get_headers(auth)
        # Update access token if auth provided
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        return self.client.list_users_management()

    def get_user(self, uid, auth=None):
        headers = self._get_headers(auth)
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        return self.client.show_user(id=uid)

    def create_user(self, user, auth=None):
        headers = self._get_headers(auth)
        # For tests sending malformed data, bypass the generated client to test server-side validation
        if isinstance(user, dict):
            if auth is None:
                auth = self.auth
            headers["Authorization"] = auth["Authorization"]
            rsp = requests.post(
                self.make_api_url("/users"),
                json=user,
                headers=headers
            )
            if rsp.status_code >= 400:
                raise management_v1.exceptions.ApiException(status=rsp.status_code, reason=rsp.text)
            return Response(status_code=rsp.status_code, data=rsp.json() if rsp.text else None)
        # For UserNew objects, use the generated client
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        r = self.client.create_user_management_with_http_info(user_new=user)
        return Response(status_code=r.status_code, data=r.data)

    def delete_user(self, user_id, auth=None, headers={}):
        if auth is None:
            auth = self.auth
        headers["Authorization"] = auth["Authorization"]
        rsp = requests.delete(
            self.make_api_url("/users/{}".format(user_id)), headers=headers
        )
        return rsp

    def update_user(self, uid, update, auth=None):
        headers = self._get_headers(auth)
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        r = self.client.update_user_with_http_info(id=uid, user_update=update)
        return Response(status_code=r.status_code, data=r.data)

    def login(self, username, password):
        auth = common.make_basic_auth(username, password)
        # For login, we need to use requests directly since it requires Basic auth
        rsp = requests.post(
            self.make_api_url("/auth/login"),
            headers={"Authorization": auth}
        )
        # Raise exception for non-2xx responses
        if rsp.status_code >= 400:
            raise management_v1.exceptions.ApiException(status=rsp.status_code, reason=rsp.text)
        return Response(status_code=rsp.status_code, text=rsp.text)

    def logout(self, auth=None):
        headers = self._get_headers(auth)
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        r = self.client.logout_with_http_info()
        return Response(status_code=r.status_code, data=r.data)

    def post_settings(self, settings, auth=None):
        if auth is None:
            auth = self.auth
        return requests.post(
            self.make_api_url("/settings"), json=settings, headers=auth
        )

    def get_settings(self, auth=None):
        if auth is None:
            auth = self.auth
        return requests.get(self.make_api_url("/settings"), headers=auth)

    def create_token(self, token_request, auth=None):
        if auth is None:
            auth = self.auth
        headers = {"Authorization": auth["Authorization"], "Content-Type": "application/json"}
        # Use requests directly since the response has application/jwt content type
        rsp = requests.post(
            self.make_api_url("/settings/tokens"),
            json=token_request.to_dict() if hasattr(token_request, 'to_dict') else token_request,
            headers=headers
        )
        if rsp.status_code >= 400:
            raise management_v1.exceptions.ApiException(status=rsp.status_code, reason=rsp.text)
        return Response(status_code=rsp.status_code, text=rsp.text)

    def list_tokens(self, auth=None):
        headers = self._get_headers(auth)
        if auth:
            token = headers.get("Authorization", "")
            if token.startswith("Bearer "):
                self.client.api_client.configuration.access_token = token[7:]
        r = self.client.list_user_personal_access_tokens_with_http_info()
        return Response(status_code=r.status_code, data=r.data)

    def delete_token(self, tid, auth=None, headers={}):
        if auth is None:
            auth = self.auth
        headers["Authorization"] = auth["Authorization"]
        rsp = requests.delete(
            self.make_api_url("/settings/tokens/{}".format(tid)), headers=headers
        )
        return rsp


class CliClient:
    cmd = "/usr/bin/useradm"

    def __init__(self, service="useradm"):
        self.client = docker.from_env()
        # Inspect the container we're running in
        hostname = socket.gethostname()
        res = self.client.containers.list(filters={"id": hostname}, limit=1)
        assert len(res) > 0, "Failed to resolve my own container!"
        _self = res[0]

        project = _self.labels.get("com.docker.compose.project")
        self.useradm = self.client.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    f"com.docker.compose.service={service}",
                ]
            },
            limit=1,
        )[0]

    def create_user(self, name, pwd, user_id=None, tenant_id=None):
        args = [self.cmd, "create-user", "--username", name, "--password", pwd]

        if user_id is not None:
            args.extend(["--user-id", user_id])

        if tenant_id is not None:
            args.extend(["--tenant-id", tenant_id])

        res = self.useradm.exec_run(args)
        assert res.exit_code == 0, res

    def set_password(self, name, pwd, tenant_id=None):
        args = [self.cmd, "set-password", "--username", name, "--password", pwd]

        if tenant_id is not None:
            args.extend(["--tenant-id", tenant_id])

        res = self.useradm.exec_run(args)
        assert res.exit_code == 0, res

    def migrate(self, tenant_id=None):
        args = [self.cmd, "migrate"]

        if tenant_id:
            args += ["--tenant", tenant_id]

        res = self.useradm.exec_run(args)
        assert res.exit_code == 0, res
