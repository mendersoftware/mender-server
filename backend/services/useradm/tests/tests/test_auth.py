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
import json
from common import (
    init_users,
    init_users_f,
    cli,
    api_client_mgmt,
    api_client_int,
    mongo,
    clean_db,
    clean_db_f,
    clean_migrated_db,
    clean_migrated_db_f,
    migrate,
    make_auth,
    user_tokens,
    explode_jwt,
)
from mockserver import run_fake
import mender_client
import pytest
from base64 import urlsafe_b64encode


class TestAuthLogin:
    @pytest.mark.parametrize(
        "email,password",
        [
            ("foo@bar.com", "asdf1234zxcv"),
            ("user-1@foo.com", "asdf1234zxcv"),
            ("user-1@foo.com", ""),
        ],
    )
    def test_bad_user(self, api_client_mgmt, init_users, email, password):
        try:
            r = api_client_mgmt.login(email, password)
        except mender_client.ApiException as herr:
            assert herr.status == 401

    def test_ok(self, api_client_mgmt, init_users):
        email = "user-1@foo.com"
        password = "correcthorsebatterystaple"

        r = api_client_mgmt.login(email, password)
        assert r.status_code == 200

        token = r.text
        assert len(token)
        _, claims, _ = explode_jwt(token)
        assert "mender.user" in claims and claims["mender.user"] == True


class TestAuthLogout:
    def test_ok(self, api_client_int, api_client_mgmt, init_users):
        email = "user-1@foo.com"
        password = "correcthorsebatterystaple"

        # log in
        r = api_client_mgmt.login(email, password)
        assert r.status_code == 200
        token = r.text

        # token is valid
        r = api_client_int.verify(token)
        assert r.status_code == 200

        # log out
        r = api_client_mgmt.logout(auth={"Authorization": "Bearer {}".format(token)})
        assert r.status_code == 202

        # token is not valid anymore
        try:
            r = api_client_int.verify(token)
        except mender_client.ApiException as herr:
            assert herr.status == 401

    def test_internal_error(self, api_client_mgmt, init_users):
        try:
            r = api_client_mgmt.logout()
        except mender_client.ApiException as herr:
            assert herr.status == 500


class TestAuthVerify:
    @pytest.mark.parametrize(
        "token", ["garbage", "", make_auth("user-1@foo.com")["Authorization"],],
    )
    def test_fail(self, api_client_int, init_users, token):
        try:
            r = api_client_int.verify(token)
        except mender_client.ApiException as herr:
            assert herr.status == 401

    def test_ok(self, api_client_int, init_users, user_tokens):
        for user, token in zip(init_users, user_tokens):
            r = api_client_int.verify(token)

            assert r.status_code == 200

    def test_tamper_claims(self, api_client_int, init_users, user_tokens):
        for user, token in zip(init_users, user_tokens):
            hdr, claims, sign = explode_jwt(token)
            claims["mender.tenant"] = "foobar"

            tampered = ".".join(
                [
                    urlsafe_b64encode(json.dumps(hdr).encode()).decode(),
                    urlsafe_b64encode(json.dumps(claims).encode()).decode(),
                    urlsafe_b64encode(sign).decode(),
                ]
            )
            try:
                r = api_client_int.verify(tampered)
            except mender_client.ApiException as herr:
                assert herr.status == 401

    def test_bad_x_original(self, api_client_int, init_users, user_tokens):
        user, token = init_users[0], user_tokens[0]
        try:
            r = api_client_int.verify(token, uri="/foobar")
        except mender_client.ApiException as herr:
            assert herr.status == 500
