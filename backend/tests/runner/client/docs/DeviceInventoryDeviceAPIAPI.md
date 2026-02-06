# \DeviceInventoryDeviceAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AssignAttributes**](DeviceInventoryDeviceAPIAPI.md#AssignAttributes) | **Patch** /api/devices/v1/inventory/device/attributes | Assign a set of attributes for a device
[**ReplaceAttributes**](DeviceInventoryDeviceAPIAPI.md#ReplaceAttributes) | **Put** /api/devices/v1/inventory/device/attributes | Replace the set of attributes for a device



## AssignAttributes

> AssignAttributes(ctx).Attribute(attribute).Execute()

Assign a set of attributes for a device



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
	attribute := []openapiclient.Attribute{*openapiclient.NewAttribute("Name_example", "Value_example")} // []Attribute | A list of attribute descriptors.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryDeviceAPIAPI.AssignAttributes(context.Background()).Attribute(attribute).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryDeviceAPIAPI.AssignAttributes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAssignAttributesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **attribute** | [**[]Attribute**](Attribute.md) | A list of attribute descriptors. | 

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


## ReplaceAttributes

> ReplaceAttributes(ctx).Attribute(attribute).Execute()

Replace the set of attributes for a device



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
	attribute := []openapiclient.Attribute{*openapiclient.NewAttribute("Name_example", "Value_example")} // []Attribute | A list of attribute descriptors.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryDeviceAPIAPI.ReplaceAttributes(context.Background()).Attribute(attribute).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryDeviceAPIAPI.ReplaceAttributes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiReplaceAttributesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **attribute** | [**[]Attribute**](Attribute.md) | A list of attribute descriptors. | 

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

