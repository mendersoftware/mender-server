#!/usr/bin/python3
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
import json
import os
import logging

from contextlib import contextmanager

import mockserver


def get_fake_tenantadm_addr():
    return os.environ.get("FAKE_TENANTADM_ADDR", "0.0.0.0:9997")


def fake_create_user(user, status):
    def create_user(request):
        req_user = json.loads(request.body.decode())
        assert req_user["name"] == user["email"]
        return (status, {}, "")

    return create_user


def fake_update_user(update, status):
    def update_user(request):
        req_update = json.loads(request.body.decode())
        assert req_update["name"] == update["email"]
        return (status, {}, "")

    return update_user


def fake_get_tenants(tenant_id, status):
    def get_tenants(request):
        tenant = [{"id": tenant_id, "name": "foo"}]
        return (status, {}, json.dumps(tenant))

    return get_tenants


@contextmanager
def run_fake_create_user(user, status=201):
    handlers = [
        ("POST", "/api/internal/v1/tenantadm/users", fake_create_user(user, status))
    ]

    with mockserver.run_fake(get_fake_tenantadm_addr(), handlers=handlers) as server:
        yield server


def fake_delete_user(expected_tenant_id, expected_user_id):
    def delete_user(_request, tenant_id, user_id):
        if expected_tenant_id is not None:
            assert tenant_id == expected_tenant_id
        if expected_user_id is not None:
            assert user_id == expected_user_id
        return (204, {}, "")

    return delete_user


@contextmanager
def run_fake_delete_user(expected_tenant_id=None, expected_user_id=None):
    handlers = [
        (
            "DELETE",
            "/api/internal/v1/tenantadm/tenants/(.*)/users/(.*)",
            fake_delete_user(expected_tenant_id, expected_user_id),
        )
    ]

    with mockserver.run_fake(get_fake_tenantadm_addr(), handlers=handlers) as server:
        yield server


@contextmanager
def run_fake_update_user(tenant_id, id, update, status=204):
    """
    Runs the update endpoint *and* the GET /tenants endpoint - it's packaged
    together because some tests will verify login after an update.
    """
    handlers = [
        (
            "PUT",
            "/api/internal/v1/tenantadm/tenants/" + tenant_id + "/users/" + id,
            fake_update_user(update, status),
        ),
        ("GET", "/api/internal/v1/tenantadm/tenants", fake_get_tenants(tenant_id, 200)),
    ]

    with mockserver.run_fake(get_fake_tenantadm_addr(), handlers=handlers) as server:
        yield server


@contextmanager
def run_fake_get_tenants(tenant_id, status=200):
    """
    Runs just the GET /tenants endpoint for tests that don't need anything more.
    """
    handlers = [
        ("GET", "/api/internal/v1/tenantadm/tenants", fake_get_tenants(tenant_id, 200))
    ]

    with mockserver.run_fake(get_fake_tenantadm_addr(), handlers=handlers) as server:
        yield server


@contextmanager
def run_fake_user_tenants(tenant_users, status="active"):
    """
    Runs GET /tenants?username=<user>. Takes a dict of `{tenant: list of
    users}`, returns valid tenant with `tenant_id` for `user`, otherwise an
    empty list.
    Returned tenant will have the 'status' field set accordingly.

    """
    user_to_tenant = {}
    for tenant, users in tenant_users.items():
        for user in users:
            user_to_tenant[user] = tenant

    def fake_get_tenants_for_user(request):
        # extract username from query parameter
        user_args = request.arguments.get("username", "")
        req_user = user_args[0].decode() if len(user_args) > 0 else ""

        tenant = user_to_tenant.get(req_user, "")
        if tenant:
            return (
                200,
                {},
                json.dumps(
                    [
                        {
                            "id": tenant,
                            "status": status,
                            "name": "foo",
                            "tenant_token": "sometoken",
                        }
                    ]
                ),
            )
        else:
            return (200, {}, "[]")

    handlers = [
        ("GET", "/api/internal/v1/tenantadm/tenants", fake_get_tenants_for_user),
    ]

    with mockserver.run_fake(get_fake_tenantadm_addr(), handlers=handlers) as server:
        yield server
