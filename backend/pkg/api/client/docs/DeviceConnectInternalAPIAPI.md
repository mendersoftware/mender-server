# \DeviceConnectInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteTenant**](DeviceConnectInternalAPIAPI.md#DeleteTenant) | **Delete** /api/internal/v1/deviceconnect/tenants/{tenantId} | Delete all the data for given tenant.
[**DeviceConnectInternalCheckHealth**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalCheckHealth) | **Get** /api/internal/v1/deviceconnect/health | Get health status of service
[**DeviceConnectInternalCheckLiveliness**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalCheckLiveliness) | **Get** /api/internal/v1/deviceconnect/alive | Get service liveliness status.
[**DeviceConnectInternalCheckUpdate**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalCheckUpdate) | **Post** /api/internal/v1/deviceconnect/tenants/{tenantId}/devices/{deviceId}/check-update | Trigger check-update for the Mender client running on the device
[**DeviceConnectInternalDecomissionDevice**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalDecomissionDevice) | **Delete** /api/internal/v1/deviceconnect/tenants/{tenantId}/devices/{deviceId} | Remove a device from the deviceconnect service.
[**DeviceConnectInternalProvisionDevice**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalProvisionDevice) | **Post** /api/internal/v1/deviceconnect/tenants/{tenantId}/devices | Register a new device with the deviceconnect service.
[**DeviceConnectInternalSendInventory**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalSendInventory) | **Post** /api/internal/v1/deviceconnect/tenants/{tenantId}/devices/{deviceId}/send-inventory | Trigger send-inventory for the Mender client running on the device
[**DeviceConnectInternalShutdown**](DeviceConnectInternalAPIAPI.md#DeviceConnectInternalShutdown) | **Get** /api/internal/v1/deviceconnect/shutdown | Shutdown the service.



## DeleteTenant

> DeleteTenant(ctx, tenantId).Execute()

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeleteTenant(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeleteTenant``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteTenantRequest struct via the builder pattern


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


## DeviceConnectInternalCheckHealth

> DeviceConnectInternalCheckHealth(ctx).Execute()

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectInternalCheckHealthRequest struct via the builder pattern


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


## DeviceConnectInternalCheckLiveliness

> DeviceConnectInternalCheckLiveliness(ctx).Execute()

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectInternalCheckLivelinessRequest struct via the builder pattern


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


## DeviceConnectInternalCheckUpdate

> DeviceConnectInternalCheckUpdate(ctx, tenantId, deviceId).Execute()

Trigger check-update for the Mender client running on the device

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
	deviceId := "deviceId_example" // string | ID for the target device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckUpdate(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalCheckUpdate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant the device belongs to. | 
**deviceId** | **string** | ID for the target device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectInternalCheckUpdateRequest struct via the builder pattern


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


## DeviceConnectInternalDecomissionDevice

> DeviceConnectInternalDecomissionDevice(ctx, tenantId, deviceId).Execute()

Remove a device from the deviceconnect service.

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalDecomissionDevice(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalDecomissionDevice``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeviceConnectInternalDecomissionDeviceRequest struct via the builder pattern


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


## DeviceConnectInternalProvisionDevice

> DeviceConnectInternalProvisionDevice(ctx, tenantId).ProvisionDevice(provisionDevice).Execute()

Register a new device with the deviceconnect service.

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalProvisionDevice(context.Background(), tenantId).ProvisionDevice(provisionDevice).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalProvisionDevice``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeviceConnectInternalProvisionDeviceRequest struct via the builder pattern


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


## DeviceConnectInternalSendInventory

> DeviceConnectInternalSendInventory(ctx, tenantId, deviceId).Execute()

Trigger send-inventory for the Mender client running on the device

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
	deviceId := "deviceId_example" // string | ID for the target device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalSendInventory(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalSendInventory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant the device belongs to. | 
**deviceId** | **string** | ID for the target device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectInternalSendInventoryRequest struct via the builder pattern


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


## DeviceConnectInternalShutdown

> DeviceConnectInternalShutdown(ctx).Execute()

Shutdown the service.

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
	r, err := apiClient.DeviceConnectInternalAPIAPI.DeviceConnectInternalShutdown(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectInternalAPIAPI.DeviceConnectInternalShutdown``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectInternalShutdownRequest struct via the builder pattern


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

