# \DeviceInventoryFiltersAndSearchInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InventoryInternalV2SearchDeviceInventories**](DeviceInventoryFiltersAndSearchInternalAPIAPI.md#InventoryInternalV2SearchDeviceInventories) | **Post** /api/internal/v2/inventory/tenants/{tenant_id}/filters/search | Search device inventories based on attributes



## InventoryInternalV2SearchDeviceInventories

> []DeviceInventory InventoryInternalV2SearchDeviceInventories(ctx, tenantId).InventoryInternalV2SearchDeviceInventoriesRequest(inventoryInternalV2SearchDeviceInventoriesRequest).Execute()

Search device inventories based on attributes



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
	inventoryInternalV2SearchDeviceInventoriesRequest := *openapiclient.NewInventoryInternalV2SearchDeviceInventoriesRequest() // InventoryInternalV2SearchDeviceInventoriesRequest | The search and sort parameters of the filter (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryFiltersAndSearchInternalAPIAPI.InventoryInternalV2SearchDeviceInventories(context.Background(), tenantId).InventoryInternalV2SearchDeviceInventoriesRequest(inventoryInternalV2SearchDeviceInventoriesRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryFiltersAndSearchInternalAPIAPI.InventoryInternalV2SearchDeviceInventories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InventoryInternalV2SearchDeviceInventories`: []DeviceInventory
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryFiltersAndSearchInternalAPIAPI.InventoryInternalV2SearchDeviceInventories`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInventoryInternalV2SearchDeviceInventoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **inventoryInternalV2SearchDeviceInventoriesRequest** | [**InventoryInternalV2SearchDeviceInventoriesRequest**](InventoryInternalV2SearchDeviceInventoriesRequest.md) | The search and sort parameters of the filter | 

### Return type

[**[]DeviceInventory**](DeviceInventory.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

