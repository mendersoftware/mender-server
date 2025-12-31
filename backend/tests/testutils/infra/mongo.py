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

import re

from pymongo import MongoClient as PyMongoClient
from testutils.infra.container_manager.kubernetes_manager import isK8S


class MongoClient:
    def __init__(self, addr="mender-mongo:27017"):
        self.client = PyMongoClient(addr)

    def cleanup(self):
        if isK8S():
            return
        dbs = self.client.list_database_names()
        dbs = [d for d in dbs if d not in ["local", "admin", "config", "workflows"]]
        for d in dbs:
            if re.match(r"^(deployment_service|inventory)-[0-9a-f]{24}", d):
                self.client.drop_database(d)
            else:
                db = self.client[d]
                for coll in db.list_collection_names(
                    filter={
                        "name": {"$ne": "migration_info"},
                        "$or": [
                            {"options.capped": {"$exists": False}},
                            {"options.capped": False},
                        ],
                    }
                ):
                    db[coll].delete_many({})
