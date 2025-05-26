#!/usr/bin/python
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

import csv

import pytest

from pymongo import MongoClient

import openapi_client as oas

from client import CliClient


@pytest.fixture(scope="session")
def inventory_attributes():
    attributeList = []

    filename = "inventory_items"

    with open(filename) as inf:
        r = csv.reader(inf)
        for row in r:
            n, v, d = row[0], row[1], row[2] if len(row) == 3 else None
            # does it matter if you pass a field name = None?
            attr = oas.Attribute(
                name=n, value=oas.AttributeValue(v), scope="inventory", description=d
            )
            attributeList.append(attr)

    return attributeList


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
    mongo_cleanup(mongo)
    yield mongo
    mongo_cleanup(mongo)


@pytest.fixture(scope="session")
def cli(request):
    service = request.config.getoption("host").split(":")[0]
    return CliClient(service)
