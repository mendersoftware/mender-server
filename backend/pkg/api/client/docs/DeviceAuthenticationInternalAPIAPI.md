# \DeviceAuthenticationInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceAuthInternalCheckHealth**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalCheckHealth) | **Get** /api/internal/v1/devauth/health | Check the health of the service
[**DeviceAuthInternalCheckLiveliness**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalCheckLiveliness) | **Get** /api/internal/v1/devauth/alive | Trivial endpoint that unconditionally returns an empty 204 response whenever the API handler is running correctly.
[**DeviceAuthInternalClearDeviceLimit**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalClearDeviceLimit) | **Delete** /api/internal/v1/devauth/tenant/{tenant_id}/limits/max_devices | Remove max device count limit
[**DeviceAuthInternalCountDevices**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalCountDevices) | **Get** /api/internal/v1/devauth/tenants/{tid}/devices/count | Count number of devices, optionally filtered by status.
[**DeviceAuthInternalCreateTenant**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalCreateTenant) | **Post** /api/internal/v1/devauth/tenants | Provision a new tenant
[**DeviceAuthInternalDeleteDevice**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalDeleteDevice) | **Delete** /api/internal/v1/devauth/tenants/{tid}/devices/{did} | Delete a device from deviceauth service.
[**DeviceAuthInternalDeviceStatus**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalDeviceStatus) | **Get** /api/internal/v1/devauth/tenants/{tid}/devices/{did}/status | Get the status of a tenant&#39;s device
[**DeviceAuthInternalGetDeviceLimit**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalGetDeviceLimit) | **Get** /api/internal/v1/devauth/tenant/{tenant_id}/limits/max_devices | Max device count limit
[**DeviceAuthInternalListDevices**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalListDevices) | **Get** /api/internal/v1/devauth/tenants/{tid}/devices | Get a list of tenant&#39;s devices.
[**DeviceAuthInternalRevokeDeviceTokens**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalRevokeDeviceTokens) | **Delete** /api/internal/v1/devauth/tokens | Delete device tokens
[**DeviceAuthInternalSetExternalIdentity**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalSetExternalIdentity) | **Put** /api/internal/v1/devauth/tenants/{tid}/devices/{did}/external | Replace the external identity of a device.
[**DeviceAuthInternalUpdateDeviceLimit**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalUpdateDeviceLimit) | **Put** /api/internal/v1/devauth/tenant/{tenant_id}/limits/max_devices | Update max device count limit
[**DeviceAuthInternalVerifyJWT**](DeviceAuthenticationInternalAPIAPI.md#DeviceAuthInternalVerifyJWT) | **Post** /api/internal/v1/devauth/tokens/verify | Check the validity of a token



## DeviceAuthInternalCheckHealth

> DeviceAuthInternalCheckHealth(ctx).Execute()

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
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalCheckHealthRequest struct via the builder pattern


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


## DeviceAuthInternalCheckLiveliness

> DeviceAuthInternalCheckLiveliness(ctx).Execute()

Trivial endpoint that unconditionally returns an empty 204 response whenever the API handler is running correctly.

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
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalCheckLivelinessRequest struct via the builder pattern


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


## DeviceAuthInternalClearDeviceLimit

> DeviceAuthInternalClearDeviceLimit(ctx, tenantId).Execute()

Remove max device count limit

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
	tenantId := "tenantId_example" // string | Tenant ID.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalClearDeviceLimit(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalClearDeviceLimit``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalClearDeviceLimitRequest struct via the builder pattern


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


## DeviceAuthInternalCountDevices

> Count DeviceAuthInternalCountDevices(ctx, tid).Status(status).Execute()

Count number of devices, optionally filtered by status.

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
	tid := "tid_example" // string | Tenant identifier.
	status := "status_example" // string | Device status filter, one of 'pending', 'accepted', 'rejected', 'noauth'. Default is 'all devices'. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCountDevices(context.Background(), tid).Status(status).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCountDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthInternalCountDevices`: Count
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCountDevices`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tid** | **string** | Tenant identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalCountDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Device status filter, one of &#39;pending&#39;, &#39;accepted&#39;, &#39;rejected&#39;, &#39;noauth&#39;. Default is &#39;all devices&#39;. | 

### Return type

[**Count**](Count.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthInternalCreateTenant

> DeviceAuthInternalCreateTenant(ctx).NewTenant(newTenant).Execute()

Provision a new tenant



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
	newTenant := *openapiclient.NewNewTenant("TenantId_example") // NewTenant | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCreateTenant(context.Background()).NewTenant(newTenant).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalCreateTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalCreateTenantRequest struct via the builder pattern


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


## DeviceAuthInternalDeleteDevice

> DeviceAuthInternalDeleteDevice(ctx, tid, did).Execute()

Delete a device from deviceauth service.

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
	tid := "tid_example" // string | 
	did := "did_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalDeleteDevice(context.Background(), tid, did).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalDeleteDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tid** | **string** |  | 
**did** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalDeleteDeviceRequest struct via the builder pattern


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


## DeviceAuthInternalDeviceStatus

> Status DeviceAuthInternalDeviceStatus(ctx, tid, did).Execute()

Get the status of a tenant's device



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
	tid := "tid_example" // string | Tenant identifier.
	did := "did_example" // string | Device identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalDeviceStatus(context.Background(), tid, did).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalDeviceStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthInternalDeviceStatus`: Status
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalDeviceStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tid** | **string** | Tenant identifier. | 
**did** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalDeviceStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Status**](Status.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthInternalGetDeviceLimit

> Limit DeviceAuthInternalGetDeviceLimit(ctx, tenantId).Execute()

Max device count limit

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
	tenantId := "tenantId_example" // string | Tenant ID.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalGetDeviceLimit(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalGetDeviceLimit``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthInternalGetDeviceLimit`: Limit
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalGetDeviceLimit`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalGetDeviceLimitRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Limit**](Limit.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthInternalListDevices

> []Device DeviceAuthInternalListDevices(ctx, tid).Status(status).Id(id).Page(page).PerPage(perPage).Execute()

Get a list of tenant's devices.



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
	tid := "tid_example" // string | Tenant identifier.
	status := "status_example" // string | Device status filter. If not specified, all devices are listed. (optional)
	id := []string{"Inner_example"} // []string | Device ID filter. Can be repeated to query a set of devices. (optional)
	page := int32(56) // int32 | Results page number. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalListDevices(context.Background(), tid).Status(status).Id(id).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalListDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthInternalListDevices`: []Device
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalListDevices`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tid** | **string** | Tenant identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalListDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Device status filter. If not specified, all devices are listed. | 
 **id** | **[]string** | Device ID filter. Can be repeated to query a set of devices. | 
 **page** | **int32** | Results page number. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]Device**](Device.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthInternalRevokeDeviceTokens

> DeviceAuthInternalRevokeDeviceTokens(ctx).TenantId(tenantId).DeviceId(deviceId).Execute()

Delete device tokens



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
	tenantId := "tenantId_example" // string | 
	deviceId := "deviceId_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalRevokeDeviceTokens(context.Background()).TenantId(tenantId).DeviceId(deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalRevokeDeviceTokens``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalRevokeDeviceTokensRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** |  | 
 **deviceId** | **string** |  | 

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


## DeviceAuthInternalSetExternalIdentity

> DeviceAuthInternalSetExternalIdentity(ctx, tid, did).ExternalDevice(externalDevice).Execute()

Replace the external identity of a device.

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
	tid := "tid_example" // string | Tenant identifier.
	did := "did_example" // string | Device identifier.
	externalDevice := *openapiclient.NewExternalDevice("Id_example", "Name_example", "Provider_example") // ExternalDevice | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalSetExternalIdentity(context.Background(), tid, did).ExternalDevice(externalDevice).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalSetExternalIdentity``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tid** | **string** | Tenant identifier. | 
**did** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalSetExternalIdentityRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **externalDevice** | [**ExternalDevice**](ExternalDevice.md) |  | 

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


## DeviceAuthInternalUpdateDeviceLimit

> DeviceAuthInternalUpdateDeviceLimit(ctx, tenantId).Limit(limit).Execute()

Update max device count limit

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
	tenantId := "tenantId_example" // string | Tenant ID.
	limit := *openapiclient.NewLimit(int32(123)) // Limit | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalUpdateDeviceLimit(context.Background(), tenantId).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalUpdateDeviceLimit``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalUpdateDeviceLimitRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **limit** | [**Limit**](Limit.md) |  | 

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


## DeviceAuthInternalVerifyJWT

> DeviceAuthInternalVerifyJWT(ctx).Authorization(authorization).Execute()

Check the validity of a token



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
	authorization := "authorization_example" // string | The token in base64-encoded form.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalVerifyJWT(context.Background()).Authorization(authorization).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationInternalAPIAPI.DeviceAuthInternalVerifyJWT``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthInternalVerifyJWTRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string** | The token in base64-encoded form. | 

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

