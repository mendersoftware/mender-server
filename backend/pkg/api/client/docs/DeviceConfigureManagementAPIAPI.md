# \DeviceConfigureManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceConfigManagementDeployDeviceConfiguration**](DeviceConfigureManagementAPIAPI.md#DeviceConfigManagementDeployDeviceConfiguration) | **Post** /api/management/v1/deviceconfig/configurations/device/{deviceId}/deploy | Deploy the device&#39;s configuration
[**DeviceConfigManagementGetDeviceConfiguration**](DeviceConfigureManagementAPIAPI.md#DeviceConfigManagementGetDeviceConfiguration) | **Get** /api/management/v1/deviceconfig/configurations/device/{deviceId} | Get the device&#39;s configuration
[**DeviceConfigManagementSetDeviceConfiguration**](DeviceConfigureManagementAPIAPI.md#DeviceConfigManagementSetDeviceConfiguration) | **Put** /api/management/v1/deviceconfig/configurations/device/{deviceId} | Set the device&#39;s configuration



## DeviceConfigManagementDeployDeviceConfiguration

> NewConfigurationDeploymentResponse DeviceConfigManagementDeployDeviceConfiguration(ctx, deviceId).NewConfigurationDeployment(newConfigurationDeployment).Execute()

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
	newConfigurationDeployment := *openapiclient.NewNewConfigurationDeployment() // NewConfigurationDeployment | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceConfigureManagementAPIAPI.DeviceConfigManagementDeployDeviceConfiguration(context.Background(), deviceId).NewConfigurationDeployment(newConfigurationDeployment).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureManagementAPIAPI.DeviceConfigManagementDeployDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConfigManagementDeployDeviceConfiguration`: NewConfigurationDeploymentResponse
	fmt.Fprintf(os.Stdout, "Response from `DeviceConfigureManagementAPIAPI.DeviceConfigManagementDeployDeviceConfiguration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigManagementDeployDeviceConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **newConfigurationDeployment** | [**NewConfigurationDeployment**](NewConfigurationDeployment.md) |  | 

### Return type

[**NewConfigurationDeploymentResponse**](NewConfigurationDeploymentResponse.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigManagementGetDeviceConfiguration

> DeviceConfiguration DeviceConfigManagementGetDeviceConfiguration(ctx, deviceId).Execute()

Get the device's configuration

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
	deviceId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device to query.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceConfigureManagementAPIAPI.DeviceConfigManagementGetDeviceConfiguration(context.Background(), deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureManagementAPIAPI.DeviceConfigManagementGetDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConfigManagementGetDeviceConfiguration`: DeviceConfiguration
	fmt.Fprintf(os.Stdout, "Response from `DeviceConfigureManagementAPIAPI.DeviceConfigManagementGetDeviceConfiguration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | ID of the device to query. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigManagementGetDeviceConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeviceConfiguration**](DeviceConfiguration.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigManagementSetDeviceConfiguration

> DeviceConfigManagementSetDeviceConfiguration(ctx, deviceId).RequestBody(requestBody).Execute()

Set the device's configuration

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
	deviceId := "deviceId_example" // string | ID of the device to query.
	requestBody := map[string]string{"key": "Inner_example"} // map[string]string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureManagementAPIAPI.DeviceConfigManagementSetDeviceConfiguration(context.Background(), deviceId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureManagementAPIAPI.DeviceConfigManagementSetDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deviceId** | **string** | ID of the device to query. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigManagementSetDeviceConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **requestBody** | **map[string]string** |  | 

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

