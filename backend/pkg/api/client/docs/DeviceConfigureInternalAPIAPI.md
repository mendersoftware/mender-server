# \DeviceConfigureInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceConfigInternalCheckHealth**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalCheckHealth) | **Get** /api/internal/v1/deviceconfig/health | Get health status of service
[**DeviceConfigInternalCheckLiveliness**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalCheckLiveliness) | **Get** /api/internal/v1/deviceconfig/alive | Get service liveliness status.
[**DeviceConfigInternalDecommissionDevice**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalDecommissionDevice) | **Delete** /api/internal/v1/deviceconfig/tenants/{tenantId}/devices/{deviceId} | Remove a device from the deviceconfig service.
[**DeviceConfigInternalDeleteTenant**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalDeleteTenant) | **Delete** /api/internal/v1/deviceconfig/tenants/{tenantId} | Delete all the data for given tenant.
[**DeviceConfigInternalDeployDeviceConfiguration**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalDeployDeviceConfiguration) | **Post** /api/internal/v1/deviceconfig/tenants/{tenantId}/configurations/device/{deviceId}/deploy | Deploy the device&#39;s configuration
[**DeviceConfigInternalProvisionDevice**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalProvisionDevice) | **Post** /api/internal/v1/deviceconfig/tenants/{tenantId}/devices | Register a new device with the deviceconfig service.
[**DeviceConfigInternalProvisionTenant**](DeviceConfigureInternalAPIAPI.md#DeviceConfigInternalProvisionTenant) | **Post** /api/internal/v1/deviceconfig/tenants | Initialize internal state for a new tenant



## DeviceConfigInternalCheckHealth

> DeviceConfigInternalCheckHealth(ctx).Execute()

Get health status of service

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalCheckHealthRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalCheckLiveliness

> DeviceConfigInternalCheckLiveliness(ctx).Execute()

Get service liveliness status.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalCheckLivelinessRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalDecommissionDevice

> DeviceConfigInternalDecommissionDevice(ctx, tenantId, deviceId).Execute()

Remove a device from the deviceconfig service.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of tenant the device belongs to.
	deviceId := "deviceId_example" // string | ID of the target device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalDecommissionDevice(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalDecommissionDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant the device belongs to. | 
**deviceId** | **string** | ID of the target device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalDecommissionDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalDeleteTenant

> DeviceConfigInternalDeleteTenant(ctx, tenantId).Execute()

Delete all the data for given tenant.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of tenant.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalDeleteTenant(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalDeleteTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalDeleteTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalDeployDeviceConfiguration

> NewConfigurationDeploymentResponse DeviceConfigInternalDeployDeviceConfiguration(ctx, deviceId, tenantId).NewConfigurationDeployment(newConfigurationDeployment).Execute()

Deploy the device's configuration

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	deviceId := "deviceId_example" // string | ID of the device.
	tenantId := "tenantId_example" // string | ID of the tenant.
	newConfigurationDeployment := *openapiclient.NewNewConfigurationDeployment() // NewConfigurationDeployment |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalDeployDeviceConfiguration(context.Background(), deviceId, tenantId).NewConfigurationDeployment(newConfigurationDeployment).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalDeployDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConfigInternalDeployDeviceConfiguration`: NewConfigurationDeploymentResponse
	fmt.Fprintf(os.Stdout, "Response from `DeviceConfigureInternalAPIAPI.DeviceConfigInternalDeployDeviceConfiguration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | ID of the device. | 
**tenantId** | **string** | ID of the tenant. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalDeployDeviceConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **newConfigurationDeployment** | [**NewConfigurationDeployment**](NewConfigurationDeployment.md) |  | 

### Return type

[**NewConfigurationDeploymentResponse**](NewConfigurationDeploymentResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalProvisionDevice

> DeviceConfigInternalProvisionDevice(ctx, tenantId).ProvisionDevice(provisionDevice).Execute()

Register a new device with the deviceconfig service.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of tenant the device belongs to.
	provisionDevice := *openapiclient.NewProvisionDevice("DeviceId_example") // ProvisionDevice |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalProvisionDevice(context.Background(), tenantId).ProvisionDevice(provisionDevice).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalProvisionDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant the device belongs to. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalProvisionDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **provisionDevice** | [**ProvisionDevice**](ProvisionDevice.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigInternalProvisionTenant

> DeviceConfigInternalProvisionTenant(ctx).NewTenant(newTenant).Execute()

Initialize internal state for a new tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	newTenant := *openapiclient.NewNewTenant("TenantId_example") // NewTenant |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureInternalAPIAPI.DeviceConfigInternalProvisionTenant(context.Background()).NewTenant(newTenant).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureInternalAPIAPI.DeviceConfigInternalProvisionTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigInternalProvisionTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **newTenant** | [**NewTenant**](NewTenant.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

