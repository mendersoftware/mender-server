# \DeploymentsInternalAPIInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeploymentsInternalCheckHealth**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalCheckHealth) | **Get** /api/internal/v1/deployments/health | Check the health of the service
[**DeploymentsInternalCheckLiveliness**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalCheckLiveliness) | **Get** /api/internal/v1/deployments/alive | Trivial endpoint that unconditionally returns an empty 204 response whenever the API handler is running correctly. 
[**DeploymentsInternalCreateDeployment**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalCreateDeployment) | **Post** /api/internal/v1/deployments/tenants/{tenant_id}/configuration/deployments/{deployment_id}/devices/{device_id} | Create a configuration deployment
[**DeploymentsInternalCreateTenant**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalCreateTenant) | **Post** /api/internal/v1/deployments/tenants | Provision a new tenant
[**DeploymentsInternalGetStorageUsage**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalGetStorageUsage) | **Get** /api/internal/v1/deployments/tenants/{id}/limits/storage | Get storage limit and current storage usage for given tenant
[**DeploymentsInternalListDeploymentsForADevice**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalListDeploymentsForADevice) | **Get** /api/internal/v1/deployments/tenants/{tenant_id}/deployments/devices/{id} | Return the Deployments history for a Device
[**DeploymentsInternalUploadArtifact**](DeploymentsInternalAPIInternalAPIAPI.md#DeploymentsInternalUploadArtifact) | **Post** /api/internal/v1/deployments/tenants/{id}/artifacts | Upload mender artifact
[**GetDeployments**](DeploymentsInternalAPIInternalAPIAPI.md#GetDeployments) | **Get** /api/internal/v1/deployments/tenants/{id}/deployments | Get all deployments for specific tenant
[**GetLastDeviceDeploymentStatus**](DeploymentsInternalAPIInternalAPIAPI.md#GetLastDeviceDeploymentStatus) | **Post** /api/internal/v1/deployments/tenants/{tenant_id}/devices/deployments/last | Get status of the last device devployment
[**GetStorageSettings**](DeploymentsInternalAPIInternalAPIAPI.md#GetStorageSettings) | **Get** /api/internal/v1/deployments/tenants/{id}/storage/settings | Get storage setting for a given tenant
[**ListDeviceDeploymentsEntries**](DeploymentsInternalAPIInternalAPIAPI.md#ListDeviceDeploymentsEntries) | **Get** /api/internal/v1/deployments/tenants/{tenant_id}/deployments/devices | Return the Deployments history entries for the specified IDs
[**RemoveDeviceFromDeployments**](DeploymentsInternalAPIInternalAPIAPI.md#RemoveDeviceFromDeployments) | **Delete** /api/internal/v1/deployments/tenants/{tenant_id}/deployments/devices/{id} | Remove device from all deployments
[**SetStorageLimit**](DeploymentsInternalAPIInternalAPIAPI.md#SetStorageLimit) | **Put** /api/internal/v1/deployments/tenants/{id}/limits/storage | Set storage limit for given tenant
[**SetStorageSettings**](DeploymentsInternalAPIInternalAPIAPI.md#SetStorageSettings) | **Put** /api/internal/v1/deployments/tenants/{id}/storage/settings | Set storage settings for a given tenant



## DeploymentsInternalCheckHealth

> DeploymentsInternalCheckHealth(ctx).Execute()

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
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalCheckHealthRequest struct via the builder pattern


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


## DeploymentsInternalCheckLiveliness

> DeploymentsInternalCheckLiveliness(ctx).Execute()

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
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalCheckLivelinessRequest struct via the builder pattern


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


## DeploymentsInternalCreateDeployment

> DeploymentsInternalCreateDeployment(ctx, tenantId, deviceId, deploymentId).ConfigurationDeploymentRequest(configurationDeploymentRequest).Execute()

Create a configuration deployment



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
	tenantId := "tenantId_example" // string | Tenant identifier.
	deviceId := "deviceId_example" // string | Device identifier.
	deploymentId := "deploymentId_example" // string | Deployment identifier.
	configurationDeploymentRequest := *openapiclient.NewConfigurationDeploymentRequest("Name_example", "Configuration_example") // ConfigurationDeploymentRequest | New deployment that needs to be created.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCreateDeployment(context.Background(), tenantId, deviceId, deploymentId).ConfigurationDeploymentRequest(configurationDeploymentRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCreateDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant identifier. | 
**deviceId** | **string** | Device identifier. | 
**deploymentId** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalCreateDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **configurationDeploymentRequest** | [**ConfigurationDeploymentRequest**](ConfigurationDeploymentRequest.md) | New deployment that needs to be created. | 

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


## DeploymentsInternalCreateTenant

> DeploymentsInternalCreateTenant(ctx).NewTenant(newTenant).Execute()

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
	newTenant := *openapiclient.NewNewTenant("TenantId_example") // NewTenant | New tenant descriptor.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCreateTenant(context.Background()).NewTenant(newTenant).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalCreateTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalCreateTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **newTenant** | [**NewTenant**](NewTenant.md) | New tenant descriptor. | 

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


## DeploymentsInternalGetStorageUsage

> StorageUsage DeploymentsInternalGetStorageUsage(ctx, id).Execute()

Get storage limit and current storage usage for given tenant



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
	id := "id_example" // string | Tenant ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalGetStorageUsage(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalGetStorageUsage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsInternalGetStorageUsage`: StorageUsage
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalGetStorageUsage`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalGetStorageUsageRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**StorageUsage**](StorageUsage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsInternalListDeploymentsForADevice

> []DeviceDeployment DeploymentsInternalListDeploymentsForADevice(ctx, tenantId, id).Status(status).Page(page).PerPage(perPage).Execute()

Return the Deployments history for a Device



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
	tenantId := "tenantId_example" // string | Tenant ID
	id := "id_example" // string | System wide device identifier
	status := "status_example" // string | Filter deployments by status for the given device. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalListDeploymentsForADevice(context.Background(), tenantId, id).Status(status).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalListDeploymentsForADevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsInternalListDeploymentsForADevice`: []DeviceDeployment
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalListDeploymentsForADevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID | 
**id** | **string** | System wide device identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalListDeploymentsForADeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **status** | **string** | Filter deployments by status for the given device. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]DeviceDeployment**](DeviceDeployment.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsInternalUploadArtifact

> DeploymentsInternalUploadArtifact(ctx, id).Artifact(artifact).ArtifactId(artifactId).Size(size).Description(description).Execute()

Upload mender artifact



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
	id := "id_example" // string | Tenant ID, or \"default\" if running in non-multitenant setup
	artifact := os.NewFile(1234, "some_file") // *os.File | Artifact. It has to be the last part of request.
	artifactId := "artifactId_example" // string | Artifact ID, optional; the server generates a randome one if not provided. (optional)
	size := int32(56) // int32 | Size of the artifact file in bytes. (optional)
	description := "description_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalUploadArtifact(context.Background(), id).Artifact(artifact).ArtifactId(artifactId).Size(size).Description(description).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.DeploymentsInternalUploadArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID, or \&quot;default\&quot; if running in non-multitenant setup | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsInternalUploadArtifactRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **artifact** | ***os.File** | Artifact. It has to be the last part of request. | 
 **artifactId** | **string** | Artifact ID, optional; the server generates a randome one if not provided. | 
 **size** | **int32** | Size of the artifact file in bytes. | 
 **description** | **string** |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeployments

> []DeploymentV1Internal GetDeployments(ctx, id).Status(status).Search(search).Page(page).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()

Get all deployments for specific tenant



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
	id := "id_example" // string | Tenant ID
	status := "status_example" // string | Deployment status filter. (optional)
	search := "search_example" // string | Deployment name or description filter. (optional)
	page := int32(56) // int32 | Results page number (optional) (default to 1)
	createdBefore := int32(56) // int32 | List only deployments created before and equal to Unix timestamp (UTC) (optional)
	createdAfter := int32(56) // int32 | List only deployments created after and equal to Unix timestamp (UTC) (optional)
	sort := "sort_example" // string | Supports sorting the deployments list by creation date.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.GetDeployments(context.Background(), id).Status(status).Search(search).Page(page).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.GetDeployments``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeployments`: []DeploymentV1Internal
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.GetDeployments`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Deployment status filter. | 
 **search** | **string** | Deployment name or description filter. | 
 **page** | **int32** | Results page number | [default to 1]
 **createdBefore** | **int32** | List only deployments created before and equal to Unix timestamp (UTC) | 
 **createdAfter** | **int32** | List only deployments created after and equal to Unix timestamp (UTC) | 
 **sort** | **string** | Supports sorting the deployments list by creation date.  | 

### Return type

[**[]DeploymentV1Internal**](DeploymentV1Internal.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetLastDeviceDeploymentStatus

> LastDeviceDeploymentsStatuses GetLastDeviceDeploymentStatus(ctx, tenantId).LastDeviceDeploymentReq(lastDeviceDeploymentReq).Execute()

Get status of the last device devployment



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
	tenantId := "tenantId_example" // string | Tenant identifier.
	lastDeviceDeploymentReq := *openapiclient.NewLastDeviceDeploymentReq([]string{"DeviceIds_example"}) // LastDeviceDeploymentReq | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.GetLastDeviceDeploymentStatus(context.Background(), tenantId).LastDeviceDeploymentReq(lastDeviceDeploymentReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.GetLastDeviceDeploymentStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetLastDeviceDeploymentStatus`: LastDeviceDeploymentsStatuses
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.GetLastDeviceDeploymentStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetLastDeviceDeploymentStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **lastDeviceDeploymentReq** | [**LastDeviceDeploymentReq**](LastDeviceDeploymentReq.md) |  | 

### Return type

[**LastDeviceDeploymentsStatuses**](LastDeviceDeploymentsStatuses.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStorageSettings

> StorageSettings GetStorageSettings(ctx, id).Execute()

Get storage setting for a given tenant



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
	id := "id_example" // string | Tenant ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.GetStorageSettings(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.GetStorageSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStorageSettings`: StorageSettings
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.GetStorageSettings`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetStorageSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**StorageSettings**](StorageSettings.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDeviceDeploymentsEntries

> []DeviceDeployment ListDeviceDeploymentsEntries(ctx, tenantId).Id(id).Execute()

Return the Deployments history entries for the specified IDs



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
	tenantId := "tenantId_example" // string | Tenant ID
	id := []string{"Inner_example"} // []string | Deployment Device ID filter. Can be repeated to query a set of entries. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.ListDeviceDeploymentsEntries(context.Background(), tenantId).Id(id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.ListDeviceDeploymentsEntries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDeviceDeploymentsEntries`: []DeviceDeployment
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsInternalAPIInternalAPIAPI.ListDeviceDeploymentsEntries`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiListDeviceDeploymentsEntriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **id** | **[]string** | Deployment Device ID filter. Can be repeated to query a set of entries. | 

### Return type

[**[]DeviceDeployment**](DeviceDeployment.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveDeviceFromDeployments

> RemoveDeviceFromDeployments(ctx, tenantId, id).Execute()

Remove device from all deployments



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
	tenantId := "tenantId_example" // string | Tenant ID
	id := "id_example" // string | System wide device identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.RemoveDeviceFromDeployments(context.Background(), tenantId, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.RemoveDeviceFromDeployments``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID | 
**id** | **string** | System wide device identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveDeviceFromDeploymentsRequest struct via the builder pattern


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


## SetStorageLimit

> SetStorageLimit(ctx, id).StorageLimit(storageLimit).Execute()

Set storage limit for given tenant



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
	id := "id_example" // string | Tenant ID
	storageLimit := *openapiclient.NewStorageLimit(int32(123), int32(123)) // StorageLimit | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.SetStorageLimit(context.Background(), id).StorageLimit(storageLimit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.SetStorageLimit``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetStorageLimitRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageLimit** | [**StorageLimit**](StorageLimit.md) |  | 

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


## SetStorageSettings

> SetStorageSettings(ctx, id).StorageSettings(storageSettings).Execute()

Set storage settings for a given tenant



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
	id := "id_example" // string | Tenant ID
	storageSettings := *openapiclient.NewStorageSettings("Bucket_example", "Key_example", "Secret_example") // StorageSettings | Settings to set. If set to null or an empty object, the tenant will use the default settings. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsInternalAPIInternalAPIAPI.SetStorageSettings(context.Background(), id).StorageSettings(storageSettings).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsInternalAPIInternalAPIAPI.SetStorageSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Tenant ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetStorageSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageSettings** | [**StorageSettings**](StorageSettings.md) | Settings to set. If set to null or an empty object, the tenant will use the default settings. | 

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

