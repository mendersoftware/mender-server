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

const indexDeploymentsTemplate = `{
	"index_patterns": ["%s*"],
	"priority": 1,
	"template": {
		"settings": {
			"number_of_shards": %d,
			"number_of_replicas": %d
		},
		"mappings": {
			"dynamic": false,
			"date_detection": false,
			"numeric_detection": false,
			"_source": {
				"enabled": true
			},
			"properties": {
				"id": {
					"type": "keyword"
				},
				"tenant_id": {
					"type": "keyword"
				},
				"device_id": {
					"type": "keyword"
				},
				"deployment_id": {
					"type": "keyword"
				},
				"deployment_name": {
					"type": "keyword"
				},
				"deployment_artifact_name": {
					"type": "keyword"
				},
				"deployment_created": {
					"type": "date"
				},
				"deployment_filter_id": {
					"type": "keyword"
				},
				"deployment_all_devices": {
					"type": "boolean"
				},
				"deployment_force_installation": {
					"type": "boolean"
				},
				"deployment_groups": {
					"type": "keyword"
				},
				"deployment_phased": {
					"type": "boolean"
				},
				"deployment_phase_id": {
					"type": "keyword"
				},
				"deployment_retries": {
					"type": "integer"
				},
				"deployment_max_devices": {
					"type": "integer"
				},
				"deployment_autogenerate_deta": {
					"type": "boolean"
				},
				"device_created": {
					"type": "date"
				},
				"device_finished": {
					"type": "date"
				},
				"device_elapsed_seconds": {
					"type": "integer"
				},
				"device_deleted": {
					"type": "date"
				},
				"device_status": {
					"type": "keyword"
				},
				"device_is_log_available": {
					"type": "boolean"
				},
				"device_retries": {
					"type": "integer"
				},
				"device_attempts": {
					"type": "integer"
				},
				"image_id": {
					"type": "keyword"
				},
				"image_description": {
					"type": "keyword"
				},
				"image_artifact_name": {
					"type": "keyword"
				},
				"image_device_types": {
					"type": "keyword"
				},
				"image_signed": {
					"type": "boolean"
				},
				"image_artifact_info_format": {
					"type": "keyword"
				},
				"image_artifact_info_version": {
					"type": "integer"
				},
				"image_provides": {
					"type": "object"
				},
				"image_depends": {
					"type": "object"
				},
				"image_clear_provides": {
					"type": "keyword"
				},
				"image_size": {
					"type": "integer"
				}
			}
		}
	}
}`
