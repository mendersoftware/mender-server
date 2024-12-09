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
import os
import json
import requests
import random
import binascii
from base64 import b64encode, urlsafe_b64decode, urlsafe_b64encode

from itertools import count
from client import CliClient
from pymongo import MongoClient

import pytest

from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.backends import default_backend

from client import SimpleInternalClient, SimpleManagementClient, BaseDevicesApiClient

import mockserver
import orchestrator
import os


def get_keypair():
    private = rsa.generate_private_key(65537, 1024, default_backend())
    private_pem = private.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    ).decode("ascii")
    public_pem = (
        private.public_key()
        .public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo,
        )
        .decode("ascii")
    )
    return private_pem, public_pem


def sign_data(data, privateKey):
    rsakey = serialization.load_pem_private_key(
        data=privateKey if isinstance(privateKey, bytes) else privateKey.encode(),
        password=None,
        backend=default_backend(),
    )
    sign = rsakey.sign(
        data if isinstance(data, bytes) else data.encode(),
        padding.PKCS1v15(),
        hashes.SHA256(),
    )
    return b64encode(sign)


class Device(object):
    def __init__(self, id_data=None):
        if id_data is None:
            mac = ":".join(
                ["{:02x}".format(random.randint(0x00, 0xFF), "x") for i in range(6)]
            )
            self.identity = json.dumps({"mac": mac}).replace(" ", "")
        else:
            self.identity = id_data

        self.private_key, self.public_key = get_keypair()
        self.token = ""

    def rotate_key(self):
        self.private_key, self.public_key = get_keypair()


class DevAuthorizer(object):
    def __init__(self, tenant_token=""):
        self.tenant_token = tenant_token

    def make_req_payload(self, dev):
        """Make auth request for given device. Returns a tuple (payload, signature)"""
        payload = json.dumps(
            {
                "id_data": dev.identity,
                "tenant_token": self.tenant_token,
                "pubkey": dev.public_key,
            }
        )
        signature = sign_data(payload, dev.private_key)
        return payload, signature

    def parse_rsp_payload(self, dev, data):
        """Parse authorization payload and apply whatever settings to the device"""
        # data is supposed to be plain text token
        dev.token = data


def device_auth_req(url, dauth, dev):
    """Run authorization request to the backend and return requests.Response"""
    data, sign = dauth.make_req_payload(dev)
    headers = {
        "Content-type": "application/json",
        "X-MEN-Signature": sign,
    }
    rsp = requests.post(url, headers=headers, data=data, verify=False)
    return rsp


def make_devid(identity):
    """
    Generate device ID from device identity data trying to follow the same
    logic as devauth does. Returns a string containing device ID.
    """
    d = SHA256.new()
    # convert to binary as needed
    bid = identity if type(identity) is bytes else identity.encode()
    d.update(bid)
    return binascii.b2a_hex(d.digest()).decode()


def b64pad(b64data):
    """Pad base64 string with '=' to achieve a length that is a multiple of 4"""
    return b64data + "=" * (4 - (len(b64data) % 4))


def explode_jwt(token):
    parts = token.split(".")
    assert len(parts) == 3

    # JWT fields are passed in a header and use URL safe encoding, which
    # substitutes - instead of + and _ instead of /
    hdr_raw = urlsafe_b64decode(b64pad(parts[0]))
    claims_raw = urlsafe_b64decode(b64pad(parts[1]))
    sign = urlsafe_b64decode(b64pad(parts[2]))

    # unpack json data
    hdr = json.loads(hdr_raw.decode())
    claims = json.loads(claims_raw.decode())

    return hdr, claims, sign


@pytest.fixture(scope="session")
def cli(request):
    service = request.config.getoption("host").split(":")[0]
    return CliClient(service)


@pytest.fixture(scope="session")
def mongo(request):
    return MongoClient(request.config.getoption("mongo_url"))


def mongo_cleanup(mongo):
    dbs = mongo.list_database_names()
    dbs = [d for d in dbs if d not in ["local", "admin", "config"]]
    for d in dbs:
        mongo.drop_database(d)


@pytest.fixture(scope="function")
def clean_db(mongo):
    """Fixture setting up a clean (i.e. empty database). Yields
    pymongo.MongoClient connected to the DB."""
    mongo_cleanup(mongo)
    yield mongo
    mongo_cleanup(mongo)


@pytest.fixture(scope="function")
def clean_migrated_db(clean_db, cli):
    """Clean database with migrations applied. Yields pymongo.MongoClient connected to the DB."""
    cli.migrate()
    yield clean_db


@pytest.fixture(scope="session")
def management_api(request):
    yield SimpleManagementClient(
        request.config.getoption("--host"),
        request.config.getoption("--management-spec"),
    )


@pytest.fixture(scope="session")
def internal_api(request):
    yield SimpleInternalClient(
        request.config.getoption("--host"),
        request.config.getoption("--spec"),
    )


@pytest.fixture(scope="session")
def device_api(request):
    yield BaseDevicesApiClient(request.config.getoption("--host"))


def make_fake_tenant_token(tenant):
    """make_fake_tenant_token will generate a JWT-like tenant token which looks
    like this: 'fake.<base64 JSON encoded claims>.fake-sig'. The claims are:
    issuer (Mender), subject (fake-tenant), mender.tenant (foobar)
    """
    claims = {
        "iss": "Mender",
        "sub": "fake-tenant",
        "mender.tenant": tenant,
    }

    # serialize claims to JSON, encode as base64 and strip padding to be
    # compatible with JWT
    enc = urlsafe_b64encode(json.dumps(claims).encode()).decode().strip("==")

    return "fake." + enc + ".fake-sig"


def make_devices(device_api, devcount=1, tenant_token=""):
    url = device_api.auth_requests_url

    out_devices = []

    with orchestrator.run_fake_for_device_id(1) as server:
        for _ in range(devcount):
            dev = Device()
            da = DevAuthorizer(tenant_token=tenant_token)
            # poke devauth so that device appears
            rsp = device_auth_req(url, da, dev)
            assert rsp.status_code == 401
            out_devices.append((dev, da))

    return out_devices


@pytest.fixture(scope="function")
def devices(device_api, clean_migrated_db, request):
    """Make unauthorized devices. The fixture can be parametrized a number of
    devices to make. Yields a list of tuples:
    (instance of Device, instance of DevAuthorizer)"""
    if not hasattr(request, "param"):
        devcount = 1
    else:
        devcount = int(request.param)

    yield make_devices(device_api, devcount)


def get_fake_tenantadm_addr():
    return os.environ.get("FAKE_TENANTADM_ADDR", "0.0.0.0:9999")


def get_fake_workflows_addr():
    return os.environ.get("FAKE_ORCHESTRATOR_ADDR", "0.0.0.0:9998")
