# \IoTManagerManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**IoTManagerManagementGetDeviceState**](IoTManagerManagementAPIAPI.md#IoTManagerManagementGetDeviceState) | **Get** /api/management/v1/iot-manager/devices/{deviceId}/state/{integrationId} | Gets the desired and reported state of a device from an integration
[**IoTManagerManagementGetDeviceStates**](IoTManagerManagementAPIAPI.md#IoTManagerManagementGetDeviceStates) | **Get** /api/management/v1/iot-manager/devices/{deviceId}/state | Gets the desired and reported state of a device
[**IoTManagerManagementListEvents**](IoTManagerManagementAPIAPI.md#IoTManagerManagementListEvents) | **Get** /api/management/v1/iot-manager/events | List all stored events
[**IoTManagerManagementListIntegrations**](IoTManagerManagementAPIAPI.md#IoTManagerManagementListIntegrations) | **Get** /api/management/v1/iot-manager/integrations | List all configured integrations
[**IoTManagerManagementRegisterIntegration**](IoTManagerManagementAPIAPI.md#IoTManagerManagementRegisterIntegration) | **Post** /api/management/v1/iot-manager/integrations | Register a new cloud integration
[**IoTManagerManagementRemoveIntegration**](IoTManagerManagementAPIAPI.md#IoTManagerManagementRemoveIntegration) | **Delete** /api/management/v1/iot-manager/integrations/{id} | Remove a cloud integration
[**IoTManagerManagementReplaceState**](IoTManagerManagementAPIAPI.md#IoTManagerManagementReplaceState) | **Put** /api/management/v1/iot-manager/devices/{deviceId}/state/{integrationId} | Replaces the (desired) cloud state of the device for the given integration
[**IoTManagerManagementSetIntegrationCredentials**](IoTManagerManagementAPIAPI.md#IoTManagerManagementSetIntegrationCredentials) | **Put** /api/management/v1/iot-manager/integrations/{id}/credentials | Replace the credentials associated with the integration.
[**IoTManagerManagementUnregisterDeviceIntegrations**](IoTManagerManagementAPIAPI.md#IoTManagerManagementUnregisterDeviceIntegrations) | **Delete** /api/management/v1/iot-manager/devices/{deviceId} | Removes all associated cloud integrations for the device.



## IoTManagerManagementGetDeviceState

> DeviceState IoTManagerManagementGetDeviceState(ctx, deviceId, integrationId).Execute()

Gets the desired and reported state of a device from an integration

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
	deviceId := "deviceId_example" // string | The unique ID of the device.
	integrationId := "integrationId_example" // string | The unique ID of the integration.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceState(context.Background(), deviceId, integrationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementGetDeviceState`: DeviceState
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceState`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | The unique ID of the device. | 
**integrationId** | **string** | The unique ID of the integration. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementGetDeviceStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**DeviceState**](DeviceState.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IoTManagerManagementGetDeviceStates

> map[string]DeviceState IoTManagerManagementGetDeviceStates(ctx, deviceId).Execute()

Gets the desired and reported state of a device

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
	deviceId := "deviceId_example" // string | The unique ID of the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceStates(context.Background(), deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceStates``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementGetDeviceStates`: map[string]DeviceState
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementGetDeviceStates`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | The unique ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementGetDeviceStatesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**map[string]DeviceState**](DeviceState.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IoTManagerManagementListEvents

> []Event IoTManagerManagementListEvents(ctx).Page(page).PerPage(perPage).IntegrationId(integrationId).Execute()

List all stored events

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
	page := int32(56) // int32 | Page number. (optional) (default to 1)
	perPage := int32(56) // int32 | Number of results per page. (optional) (default to 20)
	integrationId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique ID of the integration to get the events from. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementListEvents(context.Background()).Page(page).PerPage(perPage).IntegrationId(integrationId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementListEvents``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementListEvents`: []Event
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementListEvents`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementListEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** | Page number. | [default to 1]
 **perPage** | **int32** | Number of results per page. | [default to 20]
 **integrationId** | **string** | The unique ID of the integration to get the events from. | 

### Return type

[**[]Event**](Event.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IoTManagerManagementListIntegrations

> []Integration IoTManagerManagementListIntegrations(ctx).Page(page).PerPage(perPage).Execute()

List all configured integrations

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
	page := int32(56) // int32 | Page number. (optional) (default to 1)
	perPage := int32(56) // int32 | Number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementListIntegrations(context.Background()).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementListIntegrations``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementListIntegrations`: []Integration
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementListIntegrations`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementListIntegrationsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** | Page number. | [default to 1]
 **perPage** | **int32** | Number of results per page. | [default to 20]

### Return type

[**[]Integration**](Integration.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IoTManagerManagementRegisterIntegration

> IoTManagerManagementRegisterIntegration(ctx).Integration(integration).Execute()

Register a new cloud integration

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
	integration := *openapiclient.NewIntegration("Provider_example", openapiclient.Credentials{AWSCredentials: openapiclient.NewAWSCredentials("Type_example", *openapiclient.NewAWSCredentialsAws("AccessKeyId_example", "SecretAccessKey_example", "Region_example", "DevicePolicyName_example"))}) // Integration | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementRegisterIntegration(context.Background()).Integration(integration).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementRegisterIntegration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementRegisterIntegrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **integration** | [**Integration**](Integration.md) |  | 

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


## IoTManagerManagementRemoveIntegration

> IoTManagerManagementRemoveIntegration(ctx, id).Execute()

Remove a cloud integration

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
	id := "id_example" // string | Integration identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementRemoveIntegration(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementRemoveIntegration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Integration identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementRemoveIntegrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## IoTManagerManagementReplaceState

> DeviceState IoTManagerManagementReplaceState(ctx, deviceId, integrationId).DeviceState(deviceState).Execute()

Replaces the (desired) cloud state of the device for the given integration

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
	deviceId := "deviceId_example" // string | The unique ID of the device.
	integrationId := "integrationId_example" // string | The unique ID of the integration.
	deviceState := *openapiclient.NewDeviceState() // DeviceState | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementReplaceState(context.Background(), deviceId, integrationId).DeviceState(deviceState).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementReplaceState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementReplaceState`: DeviceState
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementReplaceState`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | The unique ID of the device. | 
**integrationId** | **string** | The unique ID of the integration. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementReplaceStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **deviceState** | [**DeviceState**](DeviceState.md) |  | 

### Return type

[**DeviceState**](DeviceState.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## IoTManagerManagementSetIntegrationCredentials

> IoTManagerManagementSetIntegrationCredentials(ctx, id).Credentials(credentials).Execute()

Replace the credentials associated with the integration.

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
	id := "id_example" // string | Integration identifier.
	credentials := openapiclient.Credentials{AWSCredentials: openapiclient.NewAWSCredentials("Type_example", *openapiclient.NewAWSCredentialsAws("AccessKeyId_example", "SecretAccessKey_example", "Region_example", "DevicePolicyName_example"))} // Credentials | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementSetIntegrationCredentials(context.Background(), id).Credentials(credentials).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementSetIntegrationCredentials``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Integration identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementSetIntegrationCredentialsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **credentials** | [**Credentials**](Credentials.md) |  | 

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


## IoTManagerManagementUnregisterDeviceIntegrations

> DeviceState IoTManagerManagementUnregisterDeviceIntegrations(ctx, deviceId).Execute()

Removes all associated cloud integrations for the device.



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
	deviceId := "deviceId_example" // string | The unique ID of the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.IoTManagerManagementAPIAPI.IoTManagerManagementUnregisterDeviceIntegrations(context.Background(), deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IoTManagerManagementAPIAPI.IoTManagerManagementUnregisterDeviceIntegrations``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IoTManagerManagementUnregisterDeviceIntegrations`: DeviceState
	fmt.Fprintf(os.Stdout, "Response from `IoTManagerManagementAPIAPI.IoTManagerManagementUnregisterDeviceIntegrations`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | The unique ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiIoTManagerManagementUnregisterDeviceIntegrationsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeviceState**](DeviceState.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

