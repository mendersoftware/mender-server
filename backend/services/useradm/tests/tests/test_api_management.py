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
from common import (
    TENANTS,
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
)
import bravado
import pytest
import tenantadm


class TestManagementApiPostUsersBase:
    def _do_test_ok(self, api_client_mgmt, init_users, new_user, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        _, r = api_client_mgmt.create_user(new_user, auth)
        assert r.status_code == 201

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users) + 1

        found_user = [u for u in users if u.email == new_user["email"]]
        assert len(found_user) == 1
        found_user = found_user[0]

    def _do_test_fail_unprocessable_entity(
        self, api_client_mgmt, init_users, new_user, tenant_id=None
    ):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        try:
            api_client_mgmt.create_user(new_user, auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422


class TestManagementApiPostUsers(TestManagementApiPostUsersBase):
    def test_ok(self, api_client_mgmt, init_users):
        new_user = {"email": "foo@bar.com", "password": "asdf1234zxcv"}
        self._do_test_ok(api_client_mgmt, init_users, new_user)

    def test_fail_malformed_body(self, api_client_mgmt):
        new_user = {"foo": "bar"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_no_password(self, api_client_mgmt):
        new_user = {"email": "foobar"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_no_email(self, api_client_mgmt):
        new_user = {"password": "asdf1234zxcv"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_not_an_email(self, api_client_mgmt):
        new_user = {"email": "foobar", "password": "asdf1234zxcv"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def test_fail_pwd_too_short(self, api_client_mgmt):
        new_user = {"email": "foo@bar.com", "password": "asdf"}
        try:
            api_client_mgmt.create_user(new_user)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422

    def test_fail_duplicate_email(self, api_client_mgmt, init_users):
        new_user = {"email": "foo@bar.com", "password": "asdf"}
        self._do_test_fail_unprocessable_entity(api_client_mgmt, init_users, new_user)


class TestManagementApiGetUserBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        for u in init_users:
            found = api_client_mgmt.get_user(u.id, auth)
            assert found.id == u.id
            assert found.email == u.email
            assert found.created_ts == u.created_ts
            assert found.updated_ts == u.updated_ts

    def _do_test_fail_not_found(self, api_client_mgmt, init_users, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        try:
            not_found = api_client_mgmt.get_user("madeupid", auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 404


class TestManagementApiGetUser(TestManagementApiGetUserBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)

    def test_fail_not_found(self, api_client_mgmt, init_users):
        self._do_test_fail_not_found(api_client_mgmt, init_users)


class TestManagementApiGetUsersBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users)

    def _do_test_no_users(self, api_client_mgmt, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        users = api_client_mgmt.get_users(auth)
        assert len(users) == 0


class TestManagementApiGetUsersOk(TestManagementApiGetUsersBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)


class TestManagementApiGetUsersNoUsers(TestManagementApiGetUsersBase):
    def test_no_users(self, api_client_mgmt):
        self._do_test_no_users(api_client_mgmt)


class TestManagementApiDeleteUserBase:
    def _do_test_ok(self, api_client_mgmt, init_users, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        rsp = api_client_mgmt.delete_user(init_users[0]["id"], auth)
        assert rsp.status_code == 204

        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users) - 1

        found = [u for u in users if u.id == init_users[0]["id"]]
        assert len(found) == 0

    def _do_test_not_found(self, api_client_mgmt, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        rsp = api_client_mgmt.delete_user("nonexistent_id", auth)
        assert rsp.status_code == 204


class TestManagementApiDeleteUser(TestManagementApiDeleteUserBase):
    def test_ok(self, api_client_mgmt, init_users):
        self._do_test_ok(api_client_mgmt, init_users)

    def test_not_found(self, api_client_mgmt, init_users):
        self._do_test_not_found(api_client_mgmt)


class TestManagementApiPutUserBase:
    def _do_test_ok_email(
        self, api_client_mgmt, init_users, user, user_to_update, update, tenant_id=None
    ):
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token = r.text
        auth = {"Authorization": "Bearer " + token}

        if user.id == user_to_update.id:
            user_to_update.id = "me"

        # test update
        _, r = api_client_mgmt.update_user(user_to_update.id, update, auth)
        assert r.status_code == 204

        # get/verify users
        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users)

        found = api_client_mgmt.get_user(user_to_update.id, auth)
        assert found.email == update["email"]

    def _do_test_ok_email_or_pass(
        self, api_client_mgmt, init_users, user, user_to_update, update, tenant_id=None
    ):
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token = r.text
        auth = {"Authorization": "Bearer " + token}

        if user.id == user_to_update.id:
            user_to_update.id = "me"

        # test update
        _, r = api_client_mgmt.update_user(user_to_update.id, update, auth)
        assert r.status_code == 204

        # get/verify users
        users = api_client_mgmt.get_users(auth)
        assert len(users) == len(init_users)

        # find the user via (new?) email
        email = user_to_update.email
        new_email = update.get("email", None)
        if new_email != None and new_email != user.email:
            email = new_email

        found = api_client_mgmt.get_user(user_to_update.id, auth)
        assert found.email == email

        # try if login still works
        _, r = api_client_mgmt.login(email, update["password"])

        assert r.status_code == 200

    def _do_test_fail_not_found(
        self, api_client_mgmt, init_users, update, tenant_id=None
    ):
        _, r = api_client_mgmt.login(init_users[0].email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token = r.text
        auth = {"Authorization": "Bearer " + token}

        try:
            _, r = api_client_mgmt.update_user("madeupid", update, auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 404

    def _do_test_fail_bad_update(self, api_client_mgmt, init_users, tenant_id=None):
        try:
            _, r = api_client_mgmt.update_user(init_users[0].id, {"foo": "bar"})
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    def _do_test_fail_unprocessable_entity(
        self, api_client_mgmt, init_users, user, user_to_update, update, tenant_id=None
    ):
        _, r = api_client_mgmt.login(user.email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token = r.text
        auth = {"Authorization": "Bearer " + token}

        try:
            _, r = api_client_mgmt.update_user(user_to_update.id, update, auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 422


class TestManagementApiPutUser(TestManagementApiPutUserBase):
    def test_ok_email_me(self, api_client_mgmt, init_users_f):
        update = {
            "email": "unique1@foo.com",
            "current_password": "correcthorsebatterystaple",
        }
        self._do_test_ok_email(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[0], update
        )

    def test_ok_email_another_user(self, api_client_mgmt, init_users_f):
        update = {
            "email": "unique1@foo.com",
        }
        self._do_test_ok_email(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[1], update
        )

    def test_ok_pass_me(self, api_client_mgmt, init_users_f):
        update = {
            "current_password": "correcthorsebatterystaple",
            "password": "secretpassword123",
        }
        self._do_test_ok_email_or_pass(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[0], update
        )

    def test_ok_email_and_pass_me(self, api_client_mgmt, init_users_f):
        update = {
            "email": "definitelyunique@foo.com",
            "current_password": "correcthorsebatterystaple",
            "password": "secretpassword123",
        }
        self._do_test_ok_email_or_pass(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[0], update
        )

    def test_fail_pass_another_user(self, api_client_mgmt, init_users_f):
        update = {
            "password": "secretpassword123",
        }
        self._do_test_fail_unprocessable_entity(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[1], update
        )

    def test_fail_password_mismatch(self, api_client_mgmt, init_users_f):
        update = {"current_password": "dummy", "password": "secretpassword123"}
        self._do_test_fail_unprocessable_entity(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[0], update
        )

    def test_fail_not_found(self, api_client_mgmt, init_users_f):
        update = {"email": "foo@bar.com", "password": "secretpassword123"}
        self._do_test_fail_not_found(api_client_mgmt, init_users_f, update)

    def test_fail_bad_update(self, api_client_mgmt, init_users_f):
        self._do_test_fail_bad_update(api_client_mgmt, init_users_f)

    def test_fail_duplicate_email(self, api_client_mgmt, init_users_f):
        update = {"email": init_users_f[1].email, "password": "secretpassword123"}
        self._do_test_fail_unprocessable_entity(
            api_client_mgmt, init_users_f, init_users_f[0], init_users_f[0], update
        )

    def test_fail_invalidated_tokens_after_update(
        self, api_client_mgmt, api_client_int, init_users_f
    ):
        users = [init_users_f[0], init_users_f[1]]
        update = {
            "email": "unique1@foo.com",
            "current_password": "correcthorsebatterystaple",
        }
        _, r = api_client_mgmt.login(users[0].email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token_one = r.text
        auth = {"Authorization": "Bearer " + token_one}

        _, r = api_client_mgmt.login(users[1].email, "correcthorsebatterystaple")
        assert r.status_code == 200
        token_two = r.text
        _, r = api_client_int.verify(token_two)
        assert r.status_code == 200

        # test update
        _, r = api_client_mgmt.update_user(users[1].id, update, auth)
        assert r.status_code == 204

        # verify tokens
        _, r = api_client_int.verify(token_one)
        assert r.status_code == 200
        with pytest.raises(bravado.exception.HTTPError) as excinfo:
            _, r = api_client_int.verify(token_two)
            assert excinfo.value.response.status_code == 401


class TestManagementApiSettingsBase:
    def _do_test_ok(self, api_client_mgmt, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        # nonempty
        self._set_and_verify(
            {"foo": "foo-val", "bar": "bar-val"}, api_client_mgmt, auth,
        )

        # empty
        self._set_and_verify({}, api_client_mgmt, auth)

    def _do_test_no_settings(self, api_client_mgmt, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        found = api_client_mgmt.get_settings(auth)
        assert found.json() == {}

    def _set_and_verify(self, settings, api_client_mgmt, auth):
        r = api_client_mgmt.post_settings(settings, auth)
        assert r.status_code == 201

        found = api_client_mgmt.get_settings(auth)
        assert found.json() == settings

    def _do_test_fail_bad_request(self, api_client_mgmt, tenant_id=None):
        auth = None
        if tenant_id is not None:
            auth = make_auth("foo", tenant_id)

        try:
            r = api_client_mgmt.post_settings("asdf", auth)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400


class TestManagementApiSettings(TestManagementApiSettingsBase):
    def test_ok(self, api_client_mgmt):
        self._do_test_ok(api_client_mgmt)

    def test_no_settings(self, api_client_mgmt):
        self._do_test_no_settings(api_client_mgmt)

    def test_bad_request(self, api_client_mgmt):
        self._do_test_fail_bad_request(api_client_mgmt)
