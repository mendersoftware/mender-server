# \DeviceInventoryManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddDevicesToGroup**](DeviceInventoryManagementAPIAPI.md#AddDevicesToGroup) | **Patch** /api/management/v1/inventory/groups/{name}/devices | Add devices to group
[**AddTags**](DeviceInventoryManagementAPIAPI.md#AddTags) | **Patch** /api/management/v1/inventory/devices/{id}/tags | Adds a set of tags for a device
[**AssignGroup**](DeviceInventoryManagementAPIAPI.md#AssignGroup) | **Put** /api/management/v1/inventory/devices/{id}/group | Add a device to a group
[**AssignTags**](DeviceInventoryManagementAPIAPI.md#AssignTags) | **Put** /api/management/v1/inventory/devices/{id}/tags | Replace the set of tags for a device
[**ClearGroup**](DeviceInventoryManagementAPIAPI.md#ClearGroup) | **Delete** /api/management/v1/inventory/devices/{id}/group/{name} | Remove a device from a group
[**DeleteDeviceInventory**](DeviceInventoryManagementAPIAPI.md#DeleteDeviceInventory) | **Delete** /api/management/v1/inventory/devices/{id} | Remove selected device&#39;s inventory
[**GetDeviceGroup**](DeviceInventoryManagementAPIAPI.md#GetDeviceGroup) | **Get** /api/management/v1/inventory/devices/{id}/group | Get a selected device&#39;s group
[**GetDeviceInventory**](DeviceInventoryManagementAPIAPI.md#GetDeviceInventory) | **Get** /api/management/v1/inventory/devices/{id} | Get a selected device&#39;s inventory
[**GetDevicesInGroup**](DeviceInventoryManagementAPIAPI.md#GetDevicesInGroup) | **Get** /api/management/v1/inventory/groups/{name}/devices | List the devices belonging to a given group
[**ListDeviceInventories**](DeviceInventoryManagementAPIAPI.md#ListDeviceInventories) | **Get** /api/management/v1/inventory/devices | List devices inventories
[**ListGroups**](DeviceInventoryManagementAPIAPI.md#ListGroups) | **Get** /api/management/v1/inventory/groups | List all groups existing device groups
[**RemoveAGroup**](DeviceInventoryManagementAPIAPI.md#RemoveAGroup) | **Delete** /api/management/v1/inventory/groups/{name} | Remove a device group
[**RemoveDevicesFromGroup**](DeviceInventoryManagementAPIAPI.md#RemoveDevicesFromGroup) | **Delete** /api/management/v1/inventory/groups/{name}/devices | Clear devices&#39; group



## AddDevicesToGroup

> AddDevicesToGroup200Response AddDevicesToGroup(ctx, name).RequestBody(requestBody).Execute()

Add devices to group



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
	name := "name_example" // string | Group name.
	requestBody := []string{"Property_example"} // []string | JSON list of device IDs to append to the group.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.AddDevicesToGroup(context.Background(), name).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.AddDevicesToGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddDevicesToGroup`: AddDevicesToGroup200Response
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.AddDevicesToGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddDevicesToGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **requestBody** | **[]string** | JSON list of device IDs to append to the group. | 

### Return type

[**AddDevicesToGroup200Response**](AddDevicesToGroup200Response.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AddTags

> AddTags(ctx, id).Tag(tag).IfMatch(ifMatch).Execute()

Adds a set of tags for a device



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
	tag := []openapiclient.Tag{*openapiclient.NewTag("Name_example", "Value_example")} // []Tag | A list of tag descriptors.
	ifMatch := "ifMatch_example" // string | Contains the device object's current ETag, and performs the update only if it matches the one stored in the database. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryManagementAPIAPI.AddTags(context.Background(), id).Tag(tag).IfMatch(ifMatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.AddTags``: %v\n", err)
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

Other parameters are passed through a pointer to a apiAddTagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tag** | [**[]Tag**](Tag.md) | A list of tag descriptors. | 
 **ifMatch** | **string** | Contains the device object&#39;s current ETag, and performs the update only if it matches the one stored in the database. | 

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


## AssignGroup

> AssignGroup(ctx, id).Group(group).Execute()

Add a device to a group



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
	group := *openapiclient.NewGroup("Group_example") // Group | Group descriptor.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryManagementAPIAPI.AssignGroup(context.Background(), id).Group(group).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.AssignGroup``: %v\n", err)
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

Other parameters are passed through a pointer to a apiAssignGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **group** | [**Group**](Group.md) | Group descriptor. | 

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


## AssignTags

> AssignTags(ctx, id).Tag(tag).IfMatch(ifMatch).Execute()

Replace the set of tags for a device



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
	tag := []openapiclient.Tag{*openapiclient.NewTag("Name_example", "Value_example")} // []Tag | A list of tags descriptors.
	ifMatch := "ifMatch_example" // string | Contains the device object's current ETag, and performs the update only if it matches the one stored in the database. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryManagementAPIAPI.AssignTags(context.Background(), id).Tag(tag).IfMatch(ifMatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.AssignTags``: %v\n", err)
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

Other parameters are passed through a pointer to a apiAssignTagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tag** | [**[]Tag**](Tag.md) | A list of tags descriptors. | 
 **ifMatch** | **string** | Contains the device object&#39;s current ETag, and performs the update only if it matches the one stored in the database. | 

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


## ClearGroup

> ClearGroup(ctx, id, name).Execute()

Remove a device from a group



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
	name := "name_example" // string | Group name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryManagementAPIAPI.ClearGroup(context.Background(), id, name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.ClearGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 
**name** | **string** | Group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiClearGroupRequest struct via the builder pattern


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


## DeleteDeviceInventory

> DeleteDeviceInventory(ctx, id).Execute()

Remove selected device's inventory

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceInventoryManagementAPIAPI.DeleteDeviceInventory(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.DeleteDeviceInventory``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteDeviceInventoryRequest struct via the builder pattern


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


## GetDeviceGroup

> Group GetDeviceGroup(ctx, id).Execute()

Get a selected device's group

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.GetDeviceGroup(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.GetDeviceGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeviceGroup`: Group
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.GetDeviceGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Group**](Group.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeviceInventory

> DeviceInventoryV1 GetDeviceInventory(ctx, id).Execute()

Get a selected device's inventory

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.GetDeviceInventory(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.GetDeviceInventory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeviceInventory`: DeviceInventoryV1
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.GetDeviceInventory`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceInventoryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeviceInventoryV1**](DeviceInventoryV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDevicesInGroup

> []string GetDevicesInGroup(ctx, name).Page(page).PerPage(perPage).Execute()

List the devices belonging to a given group

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
	name := "name_example" // string | Group name.
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 10)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.GetDevicesInGroup(context.Background(), name).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.GetDevicesInGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDevicesInGroup`: []string
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.GetDevicesInGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDevicesInGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 10]

### Return type

**[]string**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDeviceInventories

> []DeviceInventoryV1 ListDeviceInventories(ctx).Page(page).PerPage(perPage).Sort(sort).HasGroup(hasGroup).Group(group).Execute()

List devices inventories



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
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 10)
	sort := "sort_example" // string | Sort devices by attribute. The parameter is formatted as a comma-separated list of attribute names and sort order.  The order direction (`ord`) must be either `asc` or `desc` for ascending and descending respectively. Defaults to `desc` if not specified.  For example: `?sort=attr1:asc,attr2:desc` will sort by 'attr1' ascending, and then by 'attr2' descending.  (optional)
	hasGroup := true // bool | Limit result to devices assigned to a group. (optional)
	group := "group_example" // string | Limits result to devices in the given group. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.ListDeviceInventories(context.Background()).Page(page).PerPage(perPage).Sort(sort).HasGroup(hasGroup).Group(group).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.ListDeviceInventories``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDeviceInventories`: []DeviceInventoryV1
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.ListDeviceInventories`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDeviceInventoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 10]
 **sort** | **string** | Sort devices by attribute. The parameter is formatted as a comma-separated list of attribute names and sort order.  The order direction (&#x60;ord&#x60;) must be either &#x60;asc&#x60; or &#x60;desc&#x60; for ascending and descending respectively. Defaults to &#x60;desc&#x60; if not specified.  For example: &#x60;?sort&#x3D;attr1:asc,attr2:desc&#x60; will sort by &#39;attr1&#39; ascending, and then by &#39;attr2&#39; descending.  | 
 **hasGroup** | **bool** | Limit result to devices assigned to a group. | 
 **group** | **string** | Limits result to devices in the given group. | 

### Return type

[**[]DeviceInventoryV1**](DeviceInventoryV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListGroups

> []string ListGroups(ctx).Status(status).Execute()

List all groups existing device groups

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
	status := "status_example" // string | Show groups for devices with the given auth set status. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.ListGroups(context.Background()).Status(status).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.ListGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGroups`: []string
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.ListGroups`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **status** | **string** | Show groups for devices with the given auth set status. | 

### Return type

**[]string**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveAGroup

> RemoveAGroup200Response RemoveAGroup(ctx, name).Execute()

Remove a device group



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
	name := "name_example" // string | Group name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.RemoveAGroup(context.Background(), name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.RemoveAGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RemoveAGroup`: RemoveAGroup200Response
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.RemoveAGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveAGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RemoveAGroup200Response**](RemoveAGroup200Response.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveDevicesFromGroup

> RemoveAGroup200Response RemoveDevicesFromGroup(ctx, name).RequestBody(requestBody).Execute()

Clear devices' group



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
	name := "name_example" // string | Group name.
	requestBody := []string{"Property_example"} // []string | JSON list of device IDs to remove from the group.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceInventoryManagementAPIAPI.RemoveDevicesFromGroup(context.Background(), name).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceInventoryManagementAPIAPI.RemoveDevicesFromGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RemoveDevicesFromGroup`: RemoveAGroup200Response
	fmt.Fprintf(os.Stdout, "Response from `DeviceInventoryManagementAPIAPI.RemoveDevicesFromGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveDevicesFromGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **requestBody** | **[]string** | JSON list of device IDs to remove from the group. | 

### Return type

[**RemoveAGroup200Response**](RemoveAGroup200Response.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

