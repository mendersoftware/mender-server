# \DeviceAuthenticationManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceAuthManagementCountDevices**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementCountDevices) | **Get** /api/management/v2/devauth/devices/count | Count number of devices, optionally filtered by status.
[**DeviceAuthManagementDecommissionDevice**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementDecommissionDevice) | **Delete** /api/management/v2/devauth/devices/{id} | Remove device and associated authentication set
[**DeviceAuthManagementGetAuthenticationStatus**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementGetAuthenticationStatus) | **Get** /api/management/v2/devauth/devices/{id}/auth/{aid}/status | Get the device authentication set status
[**DeviceAuthManagementGetDevice**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementGetDevice) | **Get** /api/management/v2/devauth/devices/{id} | Get a particular device.
[**DeviceAuthManagementGetDeviceLimit**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementGetDeviceLimit) | **Get** /api/management/v2/devauth/limits/max_devices | Obtain limit of accepted devices.
[**DeviceAuthManagementListDevices**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementListDevices) | **Get** /api/management/v2/devauth/devices | List devices sorted by age and optionally filter on device status.
[**DeviceAuthManagementPreauthorize**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementPreauthorize) | **Post** /api/management/v2/devauth/devices | Submit a preauthorized device.
[**DeviceAuthManagementRemoveAuthentication**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementRemoveAuthentication) | **Delete** /api/management/v2/devauth/devices/{id}/auth/{aid} | Remove (dismiss) the device authentication set
[**DeviceAuthManagementRevokeAPIToken**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementRevokeAPIToken) | **Delete** /api/management/v2/devauth/tokens/{id} | Revoke JWT with given id
[**DeviceAuthManagementSearchDevices**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementSearchDevices) | **Post** /api/management/v2/devauth/devices/search | Query for devices. Returns a list of matching devices with AuthSets sorted by age.
[**DeviceAuthManagementSetAuthenticationStatus**](DeviceAuthenticationManagementAPIAPI.md#DeviceAuthManagementSetAuthenticationStatus) | **Put** /api/management/v2/devauth/devices/{id}/auth/{aid}/status | Update the device authentication set status



## DeviceAuthManagementCountDevices

> Count DeviceAuthManagementCountDevices(ctx).Status(status).XMENRequestID(xMENRequestID).Execute()

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
	status := "status_example" // string | Device status filter, one of 'pending', 'accepted', 'rejected', 'noauth', 'preauthorized'. Default is 'all devices', meaning devices with any of these statuses will be counted. (optional)
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementCountDevices(context.Background()).Status(status).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementCountDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementCountDevices`: Count
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementCountDevices`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementCountDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **status** | **string** | Device status filter, one of &#39;pending&#39;, &#39;accepted&#39;, &#39;rejected&#39;, &#39;noauth&#39;, &#39;preauthorized&#39;. Default is &#39;all devices&#39;, meaning devices with any of these statuses will be counted. | 
 **xMENRequestID** | **string** | A request identification | 

### Return type

[**Count**](Count.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementDecommissionDevice

> DeviceAuthManagementDecommissionDevice(ctx, id).XMENRequestID(xMENRequestID).Execute()

Remove device and associated authentication set

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
	id := "id_example" // string | Device identifier.
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementDecommissionDevice(context.Background(), id).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementDecommissionDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementDecommissionDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xMENRequestID** | **string** | A request identification | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementGetAuthenticationStatus

> Status DeviceAuthManagementGetAuthenticationStatus(ctx, id, aid).XMENRequestID(xMENRequestID).Execute()

Get the device authentication set status

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
	id := "id_example" // string | Device identifier.
	aid := "aid_example" // string | Authentication data set identifier.
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetAuthenticationStatus(context.Background(), id, aid).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetAuthenticationStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementGetAuthenticationStatus`: Status
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetAuthenticationStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 
**aid** | **string** | Authentication data set identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementGetAuthenticationStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xMENRequestID** | **string** | A request identification | 

### Return type

[**Status**](Status.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementGetDevice

> Device DeviceAuthManagementGetDevice(ctx, id).XMENRequestID(xMENRequestID).Execute()

Get a particular device.

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
	id := "id_example" // string | Device identifier.
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDevice(context.Background(), id).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementGetDevice`: Device
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementGetDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xMENRequestID** | **string** | A request identification | 

### Return type

[**Device**](Device.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementGetDeviceLimit

> Limit DeviceAuthManagementGetDeviceLimit(ctx).XMENRequestID(xMENRequestID).Execute()

Obtain limit of accepted devices.

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
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDeviceLimit(context.Background()).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDeviceLimit``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementGetDeviceLimit`: Limit
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementGetDeviceLimit`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementGetDeviceLimitRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xMENRequestID** | **string** | A request identification | 

### Return type

[**Limit**](Limit.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementListDevices

> []Device DeviceAuthManagementListDevices(ctx).Status(status).Id(id).Page(page).PerPage(perPage).XMENRequestID(xMENRequestID).Execute()

List devices sorted by age and optionally filter on device status.

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
	status := "status_example" // string | Device status filter. If not specified, all devices are listed. (optional)
	id := []string{"Inner_example"} // []string | Device ID filter. Can be repeated to query a set of devices. (optional)
	page := int32(56) // int32 | Results page number (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementListDevices(context.Background()).Status(status).Id(id).Page(page).PerPage(perPage).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementListDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementListDevices`: []Device
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementListDevices`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementListDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **status** | **string** | Device status filter. If not specified, all devices are listed. | 
 **id** | **[]string** | Device ID filter. Can be repeated to query a set of devices. | 
 **page** | **int32** | Results page number | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **xMENRequestID** | **string** | A request identification | 

### Return type

[**[]Device**](Device.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementPreauthorize

> DeviceAuthManagementPreauthorize(ctx).PreAuthSet(preAuthSet).XMENRequestID(xMENRequestID).Execute()

Submit a preauthorized device.



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
	preAuthSet := *openapiclient.NewPreAuthSet(*openapiclient.NewIdentityData(), "Pubkey_example") // PreAuthSet | 
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementPreauthorize(context.Background()).PreAuthSet(preAuthSet).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementPreauthorize``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementPreauthorizeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **preAuthSet** | [**PreAuthSet**](PreAuthSet.md) |  | 
 **xMENRequestID** | **string** | A request identification | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementRemoveAuthentication

> DeviceAuthManagementRemoveAuthentication(ctx, id, aid).XMENRequestID(xMENRequestID).Execute()

Remove (dismiss) the device authentication set



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
	id := "id_example" // string | Device identifier.
	aid := "aid_example" // string | Authentication data set identifier.
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementRemoveAuthentication(context.Background(), id, aid).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementRemoveAuthentication``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 
**aid** | **string** | Authentication data set identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementRemoveAuthenticationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **xMENRequestID** | **string** | A request identification | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementRevokeAPIToken

> DeviceAuthManagementRevokeAPIToken(ctx, id).XMENRequestID(xMENRequestID).Execute()

Revoke JWT with given id



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
	id := "id_example" // string | Unique token identifier('jti').
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementRevokeAPIToken(context.Background(), id).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementRevokeAPIToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Unique token identifier(&#39;jti&#39;). | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementRevokeAPITokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xMENRequestID** | **string** | A request identification | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementSearchDevices

> []Device DeviceAuthManagementSearchDevices(ctx).DeviceAuthManagementSearchDevicesRequest(deviceAuthManagementSearchDevicesRequest).Page(page).PerPage(perPage).XMENRequestID(xMENRequestID).Execute()

Query for devices. Returns a list of matching devices with AuthSets sorted by age.

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
	deviceAuthManagementSearchDevicesRequest := *openapiclient.NewDeviceAuthManagementSearchDevicesRequest() // DeviceAuthManagementSearchDevicesRequest | Device status filter. All properties can be either a single string or an array of strings.
	page := int32(56) // int32 | Results page number (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementSearchDevices(context.Background()).DeviceAuthManagementSearchDevicesRequest(deviceAuthManagementSearchDevicesRequest).Page(page).PerPage(perPage).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementSearchDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthManagementSearchDevices`: []Device
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementSearchDevices`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementSearchDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deviceAuthManagementSearchDevicesRequest** | [**DeviceAuthManagementSearchDevicesRequest**](DeviceAuthManagementSearchDevicesRequest.md) | Device status filter. All properties can be either a single string or an array of strings. | 
 **page** | **int32** | Results page number | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **xMENRequestID** | **string** | A request identification | 

### Return type

[**[]Device**](Device.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceAuthManagementSetAuthenticationStatus

> DeviceAuthManagementSetAuthenticationStatus(ctx, id, aid).Status(status).XMENRequestID(xMENRequestID).Execute()

Update the device authentication set status



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
	id := "id_example" // string | Device identifier.
	aid := "aid_example" // string | Authentication data set identifier.
	status := *openapiclient.NewStatus("Status_example") // Status | 
	xMENRequestID := "xMENRequestID_example" // string | A request identification (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementSetAuthenticationStatus(context.Background(), id, aid).Status(status).XMENRequestID(xMENRequestID).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementSetAuthenticationStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 
**aid** | **string** | Authentication data set identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthManagementSetAuthenticationStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **status** | [**Status**](Status.md) |  | 
 **xMENRequestID** | **string** | A request identification | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

