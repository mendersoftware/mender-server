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
import bravado
import pytest
import tenantadm
import requests
import uuid
from base64 import urlsafe_b64encode


def verify_token(api_client_int, token, status_code):
    try:
        _, r = api_client_int.verify(token)
    except bravado.exception.HTTPError as herr:
        assert herr.response.status_code == status_code
    else:
        assert r.status_code == status_code


def verify_tokens(api_client_int, tokens, removed_tenant=None, removed_user=None):
    for t in tokens:
        if removed_tenant is None:
            verify_token(api_client_int, t, 200)
        else:
            _, claims, _ = explode_jwt(t)
            tenant = claims["mender.tenant"]
            user = claims["sub"]
            if (
                removed_user is None or user == removed_user
            ) and tenant == removed_tenant:
                verify_token(api_client_int, t, 401)
            else:
                verify_token(api_client_int, t, 200)
