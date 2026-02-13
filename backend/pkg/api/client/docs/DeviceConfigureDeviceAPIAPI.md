# \DeviceConfigureDeviceAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceConfigGetDeviceConfiguration**](DeviceConfigureDeviceAPIAPI.md#DeviceConfigGetDeviceConfiguration) | **Get** /api/devices/v1/deviceconfig/configuration | Query the configuration store; retrieve all key-value pairs
[**DeviceConfigReportDeviceConfiguration**](DeviceConfigureDeviceAPIAPI.md#DeviceConfigReportDeviceConfiguration) | **Put** /api/devices/v1/deviceconfig/configuration | Set a key-value pair store, updating if existing, removing if empty



## DeviceConfigGetDeviceConfiguration

> map[string]string DeviceConfigGetDeviceConfiguration(ctx).Execute()

Query the configuration store; retrieve all key-value pairs

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
	resp, r, err := apiClient.DeviceConfigureDeviceAPIAPI.DeviceConfigGetDeviceConfiguration(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureDeviceAPIAPI.DeviceConfigGetDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConfigGetDeviceConfiguration`: map[string]string
	fmt.Fprintf(os.Stdout, "Response from `DeviceConfigureDeviceAPIAPI.DeviceConfigGetDeviceConfiguration`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigGetDeviceConfigurationRequest struct via the builder pattern


### Return type

**map[string]string**

### Authorization

[DeviceJWT](../README.md#DeviceJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConfigReportDeviceConfiguration

> DeviceConfigReportDeviceConfiguration(ctx).RequestBody(requestBody).Execute()

Set a key-value pair store, updating if existing, removing if empty

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
	requestBody := map[string]string{"key": "Inner_example"} // map[string]string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConfigureDeviceAPIAPI.DeviceConfigReportDeviceConfiguration(context.Background()).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConfigureDeviceAPIAPI.DeviceConfigReportDeviceConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConfigReportDeviceConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestBody** | **map[string]string** |  | 

### Return type

 (empty response body)

### Authorization

[DeviceJWT](../README.md#DeviceJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

