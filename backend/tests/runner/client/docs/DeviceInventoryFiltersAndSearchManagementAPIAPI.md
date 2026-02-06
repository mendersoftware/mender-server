# \DeviceInventoryFiltersAndSearchManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetFilterableAttributes**](DeviceInventoryFiltersAndSearchManagementAPIAPI.md#GetFilterableAttributes) | **Get** /api/management/v2/inventory/filters/attributes | Get the list of filterable inventory attributes
[**InventoryV2SearchDeviceInventories**](DeviceInventoryFiltersAndSearchManagementAPIAPI.md#InventoryV2SearchDeviceInventories) | **Post** /api/management/v2/inventory/filters/search | Search devices based on inventory attributes



## GetFilterableAttributes

> []FilterAttribute GetFilterableAttributes(ctx).Execute()

Get the list of filterable inventory attributes



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
	resp, r, err := apiClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.GetFilterableAttributes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryFiltersAndSearchManagementAPIAPI.GetFilterableAttributes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFilterableAttributes`: []FilterAttribute
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryFiltersAndSearchManagementAPIAPI.GetFilterableAttributes`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetFilterableAttributesRequest struct via the builder pattern


### Return type

[**[]FilterAttribute**](FilterAttribute.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InventoryV2SearchDeviceInventories

> []DeviceInventory InventoryV2SearchDeviceInventories(ctx).InventoryV2SearchDeviceInventoriesRequest(inventoryV2SearchDeviceInventoriesRequest).Execute()

Search devices based on inventory attributes



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
	inventoryV2SearchDeviceInventoriesRequest := *openapiclient.NewInventoryV2SearchDeviceInventoriesRequest() // InventoryV2SearchDeviceInventoriesRequest | The search and sort parameters of the filter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.InventoryV2SearchDeviceInventories(context.Background()).InventoryV2SearchDeviceInventoriesRequest(inventoryV2SearchDeviceInventoriesRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryFiltersAndSearchManagementAPIAPI.InventoryV2SearchDeviceInventories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InventoryV2SearchDeviceInventories`: []DeviceInventory
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryFiltersAndSearchManagementAPIAPI.InventoryV2SearchDeviceInventories`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInventoryV2SearchDeviceInventoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **inventoryV2SearchDeviceInventoriesRequest** | [**InventoryV2SearchDeviceInventoriesRequest**](InventoryV2SearchDeviceInventoriesRequest.md) | The search and sort parameters of the filter | 

### Return type

[**[]DeviceInventory**](DeviceInventory.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

