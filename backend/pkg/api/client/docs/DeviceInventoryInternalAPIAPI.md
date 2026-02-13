# \DeviceInventoryInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteDevice**](DeviceInventoryInternalAPIAPI.md#DeleteDevice) | **Delete** /api/internal/v1/inventory/tenants/{tenant_id}/devices/{device_id} | Remove a device from the inventory service
[**GetDeviceGroups**](DeviceInventoryInternalAPIAPI.md#GetDeviceGroups) | **Get** /api/internal/v1/inventory/tenants/{tenant_id}/devices/{device_id}/groups | Get a list of groups the device belongs to
[**InitializeDevice**](DeviceInventoryInternalAPIAPI.md#InitializeDevice) | **Post** /api/internal/v1/inventory/tenants/{tenant_id}/devices | Create a device resource with the supplied set of attributes
[**InventoryInternalCheckHealth**](DeviceInventoryInternalAPIAPI.md#InventoryInternalCheckHealth) | **Get** /api/internal/v1/inventory/health | Check the health of the service
[**InventoryInternalCheckLiveliness**](DeviceInventoryInternalAPIAPI.md#InventoryInternalCheckLiveliness) | **Get** /api/internal/v1/inventory/alive | Trivial endpoint that unconditionally returns an empty 200 response whenever the API handler is running correctly. 
[**InventoryInternalCreateTenant**](DeviceInventoryInternalAPIAPI.md#InventoryInternalCreateTenant) | **Post** /api/internal/v1/inventory/tenants | Create tenant
[**StartReIndexing**](DeviceInventoryInternalAPIAPI.md#StartReIndexing) | **Post** /api/internal/v1/inventory/tenants/{tenant_id}/devices/{device_id}/reindex | Start reindexing device attributes.
[**UpdateInventoryForADevice**](DeviceInventoryInternalAPIAPI.md#UpdateInventoryForADevice) | **Patch** /api/internal/v1/inventory/tenants/{tenant_id}/device/{device_id}/attribute/scope/{scope} | Update multiple inventory attributes in a single scope for a device
[**UpdateInventoryForADeviceScopeWise**](DeviceInventoryInternalAPIAPI.md#UpdateInventoryForADeviceScopeWise) | **Patch** /api/internal/v1/inventory/tenants/{tenant_id}/device/{device_id}/attributes | Update multiple inventory attributes for a device
[**UpdateStatusOfDevices**](DeviceInventoryInternalAPIAPI.md#UpdateStatusOfDevices) | **Post** /api/internal/v1/inventory/tenants/{tenant_id}/devices/status/{status} | Update the status of a list of devices



## DeleteDevice

> DeleteDevice(ctx, tenantId, deviceId).Execute()

Remove a device from the inventory service

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
	tenantId := "tenantId_example" // string | ID of given tenant.
	deviceId := "deviceId_example" // string | ID of given device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.DeleteDevice(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.DeleteDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 
**deviceId** | **string** | ID of given device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDeviceRequest struct via the builder pattern


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


## GetDeviceGroups

> Groups GetDeviceGroups(ctx, tenantId, deviceId).Execute()

Get a list of groups the device belongs to

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
	tenantId := "tenantId_example" // string | ID of given tenant.
	deviceId := "deviceId_example" // string | Device identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryInternalAPIAPI.GetDeviceGroups(context.Background(), tenantId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.GetDeviceGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeviceGroups`: Groups
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryInternalAPIAPI.GetDeviceGroups`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 
**deviceId** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Groups**](Groups.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InitializeDevice

> InitializeDevice(ctx, tenantId).DeviceNew(deviceNew).Execute()

Create a device resource with the supplied set of attributes

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
	tenantId := "tenantId_example" // string | ID of given tenant.
	deviceNew := *openapiclient.NewDeviceNew("Id_example") // DeviceNew | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.InitializeDevice(context.Background(), tenantId).DeviceNew(deviceNew).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.InitializeDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInitializeDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deviceNew** | [**DeviceNew**](DeviceNew.md) |  | 

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


## InventoryInternalCheckHealth

> InventoryInternalCheckHealth(ctx).Execute()

Check the health of the service

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
	r, err := apiClient.DeviceInventoryInternalAPIAPI.InventoryInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.InventoryInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInventoryInternalCheckHealthRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InventoryInternalCheckLiveliness

> InventoryInternalCheckLiveliness(ctx).Execute()

Trivial endpoint that unconditionally returns an empty 200 response whenever the API handler is running correctly. 

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
	r, err := apiClient.DeviceInventoryInternalAPIAPI.InventoryInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.InventoryInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInventoryInternalCheckLivelinessRequest struct via the builder pattern


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


## InventoryInternalCreateTenant

> InventoryInternalCreateTenant(ctx).TenantNew(tenantNew).Execute()

Create tenant



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
	tenantNew := *openapiclient.NewTenantNew() // TenantNew | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.InventoryInternalCreateTenant(context.Background()).TenantNew(tenantNew).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.InventoryInternalCreateTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInventoryInternalCreateTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantNew** | [**TenantNew**](TenantNew.md) |  | 

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


## StartReIndexing

> StartReIndexing(ctx, deviceId, tenantId).Service(service).Execute()

Start reindexing device attributes.

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
	deviceId := "deviceId_example" // string | ID of the device that needs reindexing.
	tenantId := "tenantId_example" // string | ID of tenant owning the device.
	service := "service_example" // string | The name of the calling service. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.StartReIndexing(context.Background(), deviceId, tenantId).Service(service).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.StartReIndexing``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | ID of the device that needs reindexing. | 
**tenantId** | **string** | ID of tenant owning the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartReIndexingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **service** | **string** | The name of the calling service. | 

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


## UpdateInventoryForADevice

> UpdateInventoryForADevice(ctx, tenantId, deviceId, scope).Attribute(attribute).IfUnmodifiedSince(ifUnmodifiedSince).Execute()

Update multiple inventory attributes in a single scope for a device



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
	tenantId := "tenantId_example" // string | ID of given tenant.
	deviceId := "deviceId_example" // string | ID of given device.
	scope := "scope_example" // string | Scope of the inventory attributes.
	attribute := []openapiclient.Attribute{*openapiclient.NewAttribute("Name_example", openapiclient.Attribute_value{ArrayOfFloat32: new([]float32)})} // []Attribute | List of inventory attributes to set.
	ifUnmodifiedSince := "ifUnmodifiedSince_example" // string | Skips updating the device if modified after the given RFC1123 timestamp. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.UpdateInventoryForADevice(context.Background(), tenantId, deviceId, scope).Attribute(attribute).IfUnmodifiedSince(ifUnmodifiedSince).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.UpdateInventoryForADevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 
**deviceId** | **string** | ID of given device. | 
**scope** | **string** | Scope of the inventory attributes. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateInventoryForADeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **attribute** | [**[]Attribute**](Attribute.md) | List of inventory attributes to set. | 
 **ifUnmodifiedSince** | **string** | Skips updating the device if modified after the given RFC1123 timestamp. | 

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


## UpdateInventoryForADeviceScopeWise

> UpdateInventoryForADeviceScopeWise(ctx, tenantId, deviceId).AttributeV2(attributeV2).IfUnmodifiedSince(ifUnmodifiedSince).Execute()

Update multiple inventory attributes for a device



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
	tenantId := "tenantId_example" // string | ID of given tenant.
	deviceId := "deviceId_example" // string | ID of given device.
	attributeV2 := []openapiclient.AttributeV2{*openapiclient.NewAttributeV2("Name_example", openapiclient.Scope("system"), openapiclient.AttributeV2_value{ArrayOfFloat32: new([]float32)})} // []AttributeV2 | List of inventory attributes to set.
	ifUnmodifiedSince := "ifUnmodifiedSince_example" // string | Skips updating the device if modified after the given RFC1123 timestamp. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.UpdateInventoryForADeviceScopeWise(context.Background(), tenantId, deviceId).AttributeV2(attributeV2).IfUnmodifiedSince(ifUnmodifiedSince).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.UpdateInventoryForADeviceScopeWise``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 
**deviceId** | **string** | ID of given device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateInventoryForADeviceScopeWiseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **attributeV2** | [**[]AttributeV2**](AttributeV2.md) | List of inventory attributes to set. | 
 **ifUnmodifiedSince** | **string** | Skips updating the device if modified after the given RFC1123 timestamp. | 

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


## UpdateStatusOfDevices

> UpdateStatusOfDevices(ctx, tenantId, status).DeviceUpdate(deviceUpdate).Execute()

Update the status of a list of devices



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
	tenantId := "tenantId_example" // string | ID of given tenant.
	status := "status_example" // string | New status to set for the specified devices.
	deviceUpdate := []openapiclient.DeviceUpdate{*openapiclient.NewDeviceUpdate("Id_example", int32(123))} // []DeviceUpdate | List of devices.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryInternalAPIAPI.UpdateStatusOfDevices(context.Background(), tenantId, status).DeviceUpdate(deviceUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryInternalAPIAPI.UpdateStatusOfDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of given tenant. | 
**status** | **string** | New status to set for the specified devices. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateStatusOfDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **deviceUpdate** | [**[]DeviceUpdate**](DeviceUpdate.md) | List of devices. | 

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

