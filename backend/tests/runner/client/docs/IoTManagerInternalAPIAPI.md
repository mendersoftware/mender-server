# \IoTManagerInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**IoTManagerInternalCheckHealth**](IoTManagerInternalAPIAPI.md#IoTManagerInternalCheckHealth) | **Get** /api/internal/v1/iot-manager/health | Get health status of service
[**IoTManagerInternalCheckLiveliness**](IoTManagerInternalAPIAPI.md#IoTManagerInternalCheckLiveliness) | **Get** /api/internal/v1/iot-manager/alive | Get service liveliness status.
[**IoTManagerInternalDecommissionDevice**](IoTManagerInternalAPIAPI.md#IoTManagerInternalDecommissionDevice) | **Delete** /api/internal/v1/iot-manager/tenants/{tenantId}/devices/{deviceId} | Remove a device from Iot Hub.
[**IoTManagerInternalDeleteTenant**](IoTManagerInternalAPIAPI.md#IoTManagerInternalDeleteTenant) | **Delete** /api/internal/v1/iot-manager/tenants/{tenantId} | Delete all data belonging to a given tenant.
[**IoTManagerInternalProvisionDevice**](IoTManagerInternalAPIAPI.md#IoTManagerInternalProvisionDevice) | **Post** /api/internal/v1/iot-manager/tenants/{tenantId}/devices | Register a new device with the deviceconfig service.
[**IoTManagerInternalUpdateDeviceStatuses**](IoTManagerInternalAPIAPI.md#IoTManagerInternalUpdateDeviceStatuses) | **Put** /api/internal/v1/iot-manager/tenants/{tenantId}/bulk/devices/status/{status} | Update device statuses in bulk.



## IoTManagerInternalCheckHealth

> IoTManagerInternalCheckHealth(ctx).Execute()

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
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerInternalCheckHealthRequest struct via the builder pattern


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


## IoTManagerInternalCheckLiveliness

> IoTManagerInternalCheckLiveliness(ctx).Execute()

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
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerInternalCheckLivelinessRequest struct via the builder pattern


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


## IoTManagerInternalDecommissionDevice

> IoTManagerInternalDecommissionDevice(ctx, tenantId, deviceId).Execute()

Remove a device from Iot Hub.

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
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalDecommissionDevice(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalDecommissionDevice``: %v\n", err)
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

Other parameters are passed through a pointer to a apiIoTManagerInternalDecommissionDeviceRequest struct via the builder pattern


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


## IoTManagerInternalDeleteTenant

> IoTManagerInternalDeleteTenant(ctx, tenantId).Execute()

Delete all data belonging to a given tenant.

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
	tenantId := "tenantId_example" // string | ID of tenant to remove.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalDeleteTenant(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalDeleteTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant to remove. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerInternalDeleteTenantRequest struct via the builder pattern


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


## IoTManagerInternalProvisionDevice

> IoTManagerInternalProvisionDevice(ctx, tenantId).NewDevice(newDevice).Execute()

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
	newDevice := *openapiclient.NewNewDevice("Id_example") // NewDevice | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalProvisionDevice(context.Background(), tenantId).NewDevice(newDevice).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalProvisionDevice``: %v\n", err)
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

Other parameters are passed through a pointer to a apiIoTManagerInternalProvisionDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **newDevice** | [**NewDevice**](NewDevice.md) |  | 

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


## IoTManagerInternalUpdateDeviceStatuses

> IoTManagerInternalUpdateDeviceStatuses(ctx, tenantId, status).IoTManagerInternalUpdateDeviceStatusesRequestInner(ioTManagerInternalUpdateDeviceStatusesRequestInner).Execute()

Update device statuses in bulk.

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
	status := "status_example" // string | The status of the device
	ioTManagerInternalUpdateDeviceStatusesRequestInner := []openapiclient.IoTManagerInternalUpdateDeviceStatusesRequestInner{*openapiclient.NewIoTManagerInternalUpdateDeviceStatusesRequestInner("Id_example")} // []IoTManagerInternalUpdateDeviceStatusesRequestInner | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerInternalAPIAPI.IoTManagerInternalUpdateDeviceStatuses(context.Background(), tenantId, status).IoTManagerInternalUpdateDeviceStatusesRequestInner(ioTManagerInternalUpdateDeviceStatusesRequestInner).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerInternalAPIAPI.IoTManagerInternalUpdateDeviceStatuses``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of tenant the device belongs to. | 
**status** | **string** | The status of the device | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerInternalUpdateDeviceStatusesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **ioTManagerInternalUpdateDeviceStatusesRequestInner** | [**[]IoTManagerInternalUpdateDeviceStatusesRequestInner**](IoTManagerInternalUpdateDeviceStatusesRequestInner.md) |  | 

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

