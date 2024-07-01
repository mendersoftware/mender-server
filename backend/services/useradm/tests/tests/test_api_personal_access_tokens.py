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
from random import sample
import string

from common import (
    init_users,
    init_users_f,
    cli,
    api_client_int,
    api_client_mgmt,
    mongo,
    clean_db,
    clean_db_f,
    clean_migrated_db,
    clean_migrated_db_f,
    migrate,
    make_auth,
    TENANTS,
)
import bravado
import pytest
import tenantadm


class TestManagementApiPostTokenBase:
    def _do_test_ok(
        self,
        api_client_int,
        api_client_mgmt,
        init_users,
        token_request,
        tenant_id=None,
    ):
        user = init_users[0]
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        user_token = r.text

        auth = {"Authorization": "Bearer " + user_token}

        _, r = api_client_mgmt.create_token(token_request, auth)
        assert r.status_code == 200
        personal_access_token = r.text

        # check if the token is valid
        _, r = api_client_int.verify(personal_access_token)
        assert r.status_code == 200

        # get tokens
        tokens, r = api_client_mgmt.list_tokens(auth)
        assert r.status_code == 200
        assert len(tokens) == 1

        # revoke token
        r = api_client_mgmt.delete_token(tokens[0].id, auth)
        assert r.status_code == 204

        # verify token has been removed
        tokens, r = api_client_mgmt.list_tokens(auth)
        assert r.status_code == 200
        assert len(tokens) == 0
        with pytest.raises(bravado.exception.HTTPError) as e:
            _, r = api_client_int.verify(personal_access_token)
            assert e.response.status_code == 401

    def _test_pat_limit(self, api_client_mgmt, init_users):
        user = init_users[0]
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        user_token = r.text

        auth = {"Authorization": "Bearer " + user_token}

        # first, create maximum number of tokens per user (10 is default)
        for _ in range(10):
            token_request = {
                "name": f"personal_access_token_{''.join(sample(string.ascii_lowercase, 5))}",
                "expires_in": 3600,
            }
            _, r = api_client_mgmt.create_token(token_request, auth)
            assert r.status_code == 200

        # send one token request more
        token_request = {
            "name": f"personal_access_token_{''.join(sample(string.ascii_lowercase, 5))}",
            "expires_in": 3600,
        }
        with pytest.raises(bravado.exception.HTTPUnprocessableEntity):
            _, r = api_client_mgmt.create_token(token_request, auth)
            assert r.status_code == 422

    def _test_pat_name_collision_for_one_user(self, api_client_mgmt, init_users):
        user = init_users[1]
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        user_token = r.text

        auth = {"Authorization": "Bearer " + user_token}

        token_request = {
            "name": "conflicting_personal_access_token",
            "expires_in": 3600,
        }
        _, r = api_client_mgmt.create_token(token_request, auth)
        assert r.status_code == 200

        with pytest.raises(bravado.exception.HTTPConflict):
            _, r = api_client_mgmt.create_token(token_request, auth)
            assert r.status_code == 409

    def _test_pat_name_collision_for_multiple_users(self, api_client_mgmt, init_users):
        first_user = init_users[2]
        _, r = api_client_mgmt.login(first_user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        first_user_token = r.text

        auth = {"Authorization": "Bearer " + first_user_token}

        token_request = {
            "name": "conflicting_personal_access_token",
            "expires_in": 3600,
        }
        _, r = api_client_mgmt.create_token(token_request, auth)
        assert r.status_code == 200

        # two names with same name for one user cannot exist
        with pytest.raises(bravado.exception.HTTPConflict):
            _, r = api_client_mgmt.create_token(token_request, auth)
            assert r.status_code == 409

        second_user = init_users[3]
        _, r = api_client_mgmt.login(second_user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        second_user_token = r.text
        auth = {"Authorization": "Bearer " + second_user_token}

        # another user can create token with the same name
        _, r = api_client_mgmt.create_token(token_request, auth)
        assert r.status_code == 200


class TestManagementApiPostToken(TestManagementApiPostTokenBase):
    def test_ok(self, api_client_int, api_client_mgmt, init_users):
        token_request = {"name": "my_personal_access_token", "expires_in": 3600}
        self._do_test_ok(api_client_int, api_client_mgmt, init_users, token_request)

    def test_tokens_limit_for_single_user(self, api_client_mgmt, init_users):
        self._test_pat_limit(api_client_mgmt, init_users)

    def test_tokens_naming_collisions_one_user(self, api_client_mgmt, init_users):
        self._test_pat_name_collision_for_one_user(api_client_mgmt, init_users)

    def test_tokens_naming_collisions_multiple_users(self, api_client_mgmt, init_users):
        self._test_pat_name_collision_for_multiple_users(api_client_mgmt, init_users)
