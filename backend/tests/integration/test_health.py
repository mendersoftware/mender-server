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

from mender_client import ApiClient, api


class TestHealthCheck:
    def test_health_check(self):
        api_client_deployments = ApiClient.get_default()
        api_client_deployments.configuration.host = "http://deployments:8080"
        api.DeploymentsInternalAPIInternalAPIApi(
            api_client_deployments
        ).deployments_internal_check_health()
        api_client_deviceauth = ApiClient.get_default()
        api_client_deviceauth.configuration.host = "http://deviceauth:8080"
        api.DeviceAuthenticationInternalAPIApi(
            api_client_deviceauth
        ).device_auth_internal_check_health()
        api_client_deviceconfig = ApiClient.get_default()
        api_client_deviceconfig.configuration.host = "http://deviceconfig:8080"
        api.DeviceConfigureInternalAPIApi(
            api_client_deviceconfig
        ).device_config_internal_check_health()
        api_client_deviceconnect = ApiClient.get_default()
        api_client_deviceconnect.configuration.host = "http://deviceconnect:8080"
        api.DeviceConnectInternalAPIApi(
            api_client_deviceconnect
        ).device_connect_internal_check_health()
        api_client_inventory = ApiClient.get_default()
        api_client_inventory.configuration.host = "http://inventory:8080"
        api.DeviceInventoryInternalAPIApi(
            api_client_inventory
        ).inventory_internal_check_health()
        api_client_iot_manager = ApiClient.get_default()
        api_client_iot_manager.configuration.host = "http://iot-manager:8080"
        api.IoTManagerInternalAPIApi(
            api_client_iot_manager
        ).io_t_manager_internal_check_health()
        api_client_useradm = ApiClient.get_default()
        api_client_useradm.configuration.host = "http://useradm:8080"
        api.UserAdministrationAndAuthenticationInternalAPIApi(
            api_client_useradm
        ).useradm_check_health()
        api_client_workflows = ApiClient.get_default()
        api_client_workflows.configuration.host = "http://workflows:8080"
        api.WorkflowsOtherApi(api_client_workflows).workflows_check_health()


class TestHealthCheckEnterprise(TestHealthCheck):
    def test_health_check(self):
        super().test_health_check()
        # FIXME: enterprise API specs are private
        api_client_auditlogs = ApiClient.get_default()
        api_client_auditlogs.configuration.host = "http://auditlogs:8080"
        api_client_auditlogs.call_api(
            *api_client_auditlogs.param_serialize(
                "GET", "/api/internal/v1/auditlogs/health"
            )
        )
        api_client_devicemonitor = ApiClient.get_default()
        api_client_devicemonitor.configuration.host = "http://devicemonitor:8080"
        api_client_devicemonitor.call_api(
            *api_client_devicemonitor.param_serialize(
                "GET", "/api/internal/v1/devicemonitor/health"
            )
        )
        api_client_tenantadm = ApiClient.get_default()
        api_client_tenantadm.configuration.host = "http://tenantadm:8080"
        api_client_tenantadm.call_api(
            *api_client_tenantadm.param_serialize(
                "GET", "/api/internal/v1/tenantadm/health"
            )
        )
