// Copyright 2023 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package opensearch

const indexDevicesTemplate = `{
	"index_patterns": ["%s*"],
	"priority": 1,
	"template": {
		"settings": {
			"number_of_shards": %d,
			"number_of_replicas": %d
		},
		"mappings": {
			"dynamic": true,
			"date_detection": false,
			"numeric_detection": false,
			"_source": {
				"enabled": true
			},
			"properties": {
				"id": {
					"type": "keyword"
				},
				"tenantID": {
					"type": "keyword"
				},
				"name": {
					"type": "keyword"
				},
				"location": {
					"type": "geo_point"
				}
			},
			"dynamic_templates": [
				{
					"versions": {
						"match": "*_version*",
						"mapping": {
							"type": "version"
						}
					}
				},
				{
					"nums": {
						"match": "*_num",
						"mapping": {
							"type": "double"
						}
					}
				},
				{
					"strings": {
						"match": "*_str",
						"mapping": {
							"type": "keyword"
						}
					}
				},
				{
					"bools": {
						"match": "*_bool",
						"mapping": {
							"type": "boolean"
						}
					}
				}
			]
		}
	}
}`
