# \DeploymentsManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AbortDeployment**](DeploymentsManagementAPIAPI.md#AbortDeployment) | **Put** /api/management/v1/deployments/deployments/{deployment_id}/status | Abort the deployment
[**AbortDeploymentsForADevice**](DeploymentsManagementAPIAPI.md#AbortDeploymentsForADevice) | **Delete** /api/management/v1/deployments/deployments/devices/{id} | Abort all the active and pending Deployments for a Device
[**CompleteDirectUpload**](DeploymentsManagementAPIAPI.md#CompleteDirectUpload) | **Post** /api/management/v1/deployments/artifacts/directupload/{id}/complete | Notify the server that the direct upload is completed to make it available in the artifacts API. Optionally you can provide files metadata which will be absent otherwise if skip-verify flag is present in the deployments service. This is an on-prem endpoint only, not available on Hosted Mender.
[**CreateDeploymentForAGroupOfDevices**](DeploymentsManagementAPIAPI.md#CreateDeploymentForAGroupOfDevices) | **Post** /api/management/v1/deployments/deployments/group/{name} | Create a deployment for a group of devices
[**DeleteArtifact**](DeploymentsManagementAPIAPI.md#DeleteArtifact) | **Delete** /api/management/v1/deployments/artifacts/{id} | Delete the artifact
[**DeploymentStatusStatistics**](DeploymentsManagementAPIAPI.md#DeploymentStatusStatistics) | **Get** /api/management/v1/deployments/deployments/{deployment_id}/statistics | Get status count for all devices in a deployment. 
[**DeploymentStatusStatisticsList**](DeploymentsManagementAPIAPI.md#DeploymentStatusStatisticsList) | **Post** /api/management/v1/deployments/deployments/statistics/list | Get status count for all devices in the listed deployments (plural). 
[**DeploymentsCreateDeployment**](DeploymentsManagementAPIAPI.md#DeploymentsCreateDeployment) | **Post** /api/management/v1/deployments/deployments | Create a deployment
[**DeploymentsGetStorageUsage**](DeploymentsManagementAPIAPI.md#DeploymentsGetStorageUsage) | **Get** /api/management/v1/deployments/limits/storage | Get storage limit and current storage usage
[**DeploymentsListDeploymentsForADevice**](DeploymentsManagementAPIAPI.md#DeploymentsListDeploymentsForADevice) | **Get** /api/management/v1/deployments/deployments/devices/{id} | Return the Deployments history for a Device
[**DeploymentsV1ListArtifactsWithPagination**](DeploymentsManagementAPIAPI.md#DeploymentsV1ListArtifactsWithPagination) | **Get** /api/management/v1/deployments/artifacts/list | List known artifacts 
[**DeploymentsV1ListDeployments**](DeploymentsManagementAPIAPI.md#DeploymentsV1ListDeployments) | **Get** /api/management/v1/deployments/deployments | Find all deployments
[**DeploymentsV1ListReleasesWithPagination**](DeploymentsManagementAPIAPI.md#DeploymentsV1ListReleasesWithPagination) | **Get** /api/management/v1/deployments/deployments/releases/list | List releases with pagination 
[**DownloadArtifact**](DeploymentsManagementAPIAPI.md#DownloadArtifact) | **Get** /api/management/v1/deployments/artifacts/{id}/download | Get the download link of a selected artifact
[**GenerateArtifact**](DeploymentsManagementAPIAPI.md#GenerateArtifact) | **Post** /api/management/v1/deployments/artifacts/generate | Upload raw data to generate a new artifact
[**GetDeploymentLogForDevice**](DeploymentsManagementAPIAPI.md#GetDeploymentLogForDevice) | **Get** /api/management/v1/deployments/deployments/{deployment_id}/devices/{device_id}/log | Get the log of a selected device&#39;s deployment
[**ListAllDevicesInDeployment**](DeploymentsManagementAPIAPI.md#ListAllDevicesInDeployment) | **Get** /api/management/v1/deployments/deployments/{deployment_id}/devices | DEPRECATED: _since Wed May 19 2021_ this end-point is deprecated because it doesn&#39;t support pagination and will be removed in the future, please use the /deployments/{deployment_id}/devices/list end-point instead. 
[**ListArtifacts**](DeploymentsManagementAPIAPI.md#ListArtifacts) | **Get** /api/management/v1/deployments/artifacts | List all the artifacts 
[**ListDeviceIDsInDeployment**](DeploymentsManagementAPIAPI.md#ListDeviceIDsInDeployment) | **Get** /api/management/v1/deployments/deployments/{id}/device_list | Get the list of device IDs being part of the deployment.
[**ListDevicesInDeployment**](DeploymentsManagementAPIAPI.md#ListDevicesInDeployment) | **Get** /api/management/v1/deployments/deployments/{deployment_id}/devices/list | Get the list of devices and their respective status for the deployment with the given ID. The response includes devices as they get assigned to the deployment when checking for updates. Therefore, this endpoint will list all the devices only once each asks for updates and evaluates this deployment. 
[**ListReleases**](DeploymentsManagementAPIAPI.md#ListReleases) | **Get** /api/management/v1/deployments/deployments/releases | List releases 
[**RequestDirectUpload**](DeploymentsManagementAPIAPI.md#RequestDirectUpload) | **Post** /api/management/v1/deployments/artifacts/directupload | Request link for uploading artifact directly to the storage backend. This is an on-prem endpoint only, not available on Hosted Mender.
[**ResetDeviceDeploymentsHistory**](DeploymentsManagementAPIAPI.md#ResetDeviceDeploymentsHistory) | **Delete** /api/management/v1/deployments/deployments/devices/{id}/history | Reset the Device Deployments history
[**ShowArtifact**](DeploymentsManagementAPIAPI.md#ShowArtifact) | **Get** /api/management/v1/deployments/artifacts/{id} | Get the details of a selected artifact
[**ShowDeployment**](DeploymentsManagementAPIAPI.md#ShowDeployment) | **Get** /api/management/v1/deployments/deployments/{id} | Get the details of a selected deployment
[**UpdateArtifactInfo**](DeploymentsManagementAPIAPI.md#UpdateArtifactInfo) | **Put** /api/management/v1/deployments/artifacts/{id} | Update description of a selected artifact
[**UploadArtifact**](DeploymentsManagementAPIAPI.md#UploadArtifact) | **Post** /api/management/v1/deployments/artifacts | Upload mender artifact



## AbortDeployment

> AbortDeployment(ctx, deploymentId).AbortDeploymentRequest(abortDeploymentRequest).Execute()

Abort the deployment



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
	deploymentId := "deploymentId_example" // string | Deployment identifier.
	abortDeploymentRequest := *openapiclient.NewAbortDeploymentRequest("Status_example") // AbortDeploymentRequest | Deployment status.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.AbortDeployment(context.Background(), deploymentId).AbortDeploymentRequest(abortDeploymentRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.AbortDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAbortDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **abortDeploymentRequest** | [**AbortDeploymentRequest**](AbortDeploymentRequest.md) | Deployment status. | 

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


## AbortDeploymentsForADevice

> AbortDeploymentsForADevice(ctx, id).Execute()

Abort all the active and pending Deployments for a Device



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
	id := "id_example" // string | System wide device identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.AbortDeploymentsForADevice(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.AbortDeploymentsForADevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | System wide device identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiAbortDeploymentsForADeviceRequest struct via the builder pattern


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


## CompleteDirectUpload

> CompleteDirectUpload(ctx, id).DirectUploadMetadata(directUploadMetadata).Execute()

Notify the server that the direct upload is completed to make it available in the artifacts API. Optionally you can provide files metadata which will be absent otherwise if skip-verify flag is present in the deployments service. This is an on-prem endpoint only, not available on Hosted Mender.

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
	id := "id_example" // string | Artifact ID returned by \"Request Direct Upload\" API.
	directUploadMetadata := *openapiclient.NewDirectUploadMetadata() // DirectUploadMetadata | Metadata for contents of the artifact. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.CompleteDirectUpload(context.Background(), id).DirectUploadMetadata(directUploadMetadata).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.CompleteDirectUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Artifact ID returned by \&quot;Request Direct Upload\&quot; API. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCompleteDirectUploadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **directUploadMetadata** | [**DirectUploadMetadata**](DirectUploadMetadata.md) | Metadata for contents of the artifact. | 

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


## CreateDeploymentForAGroupOfDevices

> CreateDeploymentForAGroupOfDevices(ctx, name).NewDeploymentForGroup(newDeploymentForGroup).Execute()

Create a deployment for a group of devices



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
	name := "name_example" // string | Device group name.
	newDeploymentForGroup := *openapiclient.NewNewDeploymentForGroup("Name_example", "ArtifactName_example") // NewDeploymentForGroup | New deployment that needs to be created.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.CreateDeploymentForAGroupOfDevices(context.Background(), name).NewDeploymentForGroup(newDeploymentForGroup).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.CreateDeploymentForAGroupOfDevices``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Device group name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDeploymentForAGroupOfDevicesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **newDeploymentForGroup** | [**NewDeploymentForGroup**](NewDeploymentForGroup.md) | New deployment that needs to be created. | 

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


## DeleteArtifact

> DeleteArtifact(ctx, id).Execute()

Delete the artifact



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
	id := "id_example" // string | Artifact identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.DeleteArtifact(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeleteArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Artifact identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteArtifactRequest struct via the builder pattern


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


## DeploymentStatusStatistics

> Statistics DeploymentStatusStatistics(ctx, deploymentId).Execute()

Get status count for all devices in a deployment. 

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
	deploymentId := "deploymentId_example" // string | Deployment identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(context.Background(), deploymentId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentStatusStatistics``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentStatusStatistics`: Statistics
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentStatusStatistics`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentStatusStatisticsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Statistics**](Statistics.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentStatusStatisticsList

> []DeploymentStatusStatisticsList200ResponseInner DeploymentStatusStatisticsList(ctx).DeploymentIdentifier(deploymentIdentifier).Execute()

Get status count for all devices in the listed deployments (plural). 

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
	deploymentIdentifier := *openapiclient.NewDeploymentIdentifier() // DeploymentIdentifier | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentStatusStatisticsList(context.Background()).DeploymentIdentifier(deploymentIdentifier).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentStatusStatisticsList``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentStatusStatisticsList`: []DeploymentStatusStatisticsList200ResponseInner
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentStatusStatisticsList`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentStatusStatisticsListRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deploymentIdentifier** | [**DeploymentIdentifier**](DeploymentIdentifier.md) |  | 

### Return type

[**[]DeploymentStatusStatisticsList200ResponseInner**](DeploymentStatusStatisticsList200ResponseInner.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsCreateDeployment

> DeploymentsCreateDeployment(ctx).NewDeployment(newDeployment).Execute()

Create a deployment



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
	newDeployment := *openapiclient.NewNewDeployment("Name_example", "ArtifactName_example") // NewDeployment | New deployment that needs to be created.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(context.Background()).NewDeployment(newDeployment).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsCreateDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsCreateDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **newDeployment** | [**NewDeployment**](NewDeployment.md) | New deployment that needs to be created. | 

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


## DeploymentsGetStorageUsage

> StorageLimit DeploymentsGetStorageUsage(ctx).Execute()

Get storage limit and current storage usage



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
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsGetStorageUsage(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsGetStorageUsage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsGetStorageUsage`: StorageLimit
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentsGetStorageUsage`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsGetStorageUsageRequest struct via the builder pattern


### Return type

[**StorageLimit**](StorageLimit.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsListDeploymentsForADevice

> []DeviceDeploymentV1 DeploymentsListDeploymentsForADevice(ctx, id).Status(status).Page(page).PerPage(perPage).Execute()

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
	id := "id_example" // string | System wide device identifier
	status := "status_example" // string | Filter deployments by status for the given device. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsListDeploymentsForADevice(context.Background(), id).Status(status).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsListDeploymentsForADevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsListDeploymentsForADevice`: []DeviceDeploymentV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentsListDeploymentsForADevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | System wide device identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsListDeploymentsForADeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Filter deployments by status for the given device. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]DeviceDeploymentV1**](DeviceDeploymentV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsV1ListArtifactsWithPagination

> []ArtifactV1 DeploymentsV1ListArtifactsWithPagination(ctx).Name(name).Description(description).DeviceType(deviceType).Page(page).PerPage(perPage).Sort(sort).Execute()

List known artifacts 



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
	name := "name_example" // string | Artifact name filter. (optional)
	description := "description_example" // string | Artifact description filter. (optional)
	deviceType := "deviceType_example" // string | Artifact device type filter. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	sort := "sort_example" // string | Sort the artifact list by the specified field and direction.  (optional) (default to "name:asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsV1ListArtifactsWithPagination(context.Background()).Name(name).Description(description).DeviceType(deviceType).Page(page).PerPage(perPage).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsV1ListArtifactsWithPagination``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV1ListArtifactsWithPagination`: []ArtifactV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentsV1ListArtifactsWithPagination`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV1ListArtifactsWithPaginationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Artifact name filter. | 
 **description** | **string** | Artifact description filter. | 
 **deviceType** | **string** | Artifact device type filter. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **sort** | **string** | Sort the artifact list by the specified field and direction.  | [default to &quot;name:asc&quot;]

### Return type

[**[]ArtifactV1**](ArtifactV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsV1ListDeployments

> []DeploymentV1 DeploymentsV1ListDeployments(ctx).Status(status).Type_(type_).Search(search).Page(page).PerPage(perPage).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()

Find all deployments



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
	status := "status_example" // string | Deployment status filter. (optional)
	type_ := "type__example" // string | Deployment type filter.  (optional)
	search := "search_example" // string | Deployment name or description filter. (optional)
	page := int32(56) // int32 | Results page number (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	createdBefore := int32(56) // int32 | List only deployments created before and equal to Unix timestamp (UTC) (optional)
	createdAfter := int32(56) // int32 | List only deployments created after and equal to Unix timestamp (UTC) (optional)
	sort := "sort_example" // string | Supports sorting the deployments list by creation date.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsV1ListDeployments(context.Background()).Status(status).Type_(type_).Search(search).Page(page).PerPage(perPage).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsV1ListDeployments``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV1ListDeployments`: []DeploymentV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentsV1ListDeployments`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV1ListDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **status** | **string** | Deployment status filter. | 
 **type_** | **string** | Deployment type filter.  | 
 **search** | **string** | Deployment name or description filter. | 
 **page** | **int32** | Results page number | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **createdBefore** | **int32** | List only deployments created before and equal to Unix timestamp (UTC) | 
 **createdAfter** | **int32** | List only deployments created after and equal to Unix timestamp (UTC) | 
 **sort** | **string** | Supports sorting the deployments list by creation date.  | 

### Return type

[**[]DeploymentV1**](DeploymentV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsV1ListReleasesWithPagination

> []ReleaseV1 DeploymentsV1ListReleasesWithPagination(ctx).Name(name).Description(description).DeviceType(deviceType).UpdateType(updateType).Page(page).PerPage(perPage).Sort(sort).Execute()

List releases with pagination 



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
	name := "name_example" // string | Release name filter. (optional)
	description := "description_example" // string | Release description filter. (optional)
	deviceType := "deviceType_example" // string | Release device type filter. (optional)
	updateType := "updateType_example" // string | Update type filter. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	sort := "sort_example" // string | Sort the release list by the specified field and direction.  (optional) (default to "name:asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DeploymentsV1ListReleasesWithPagination(context.Background()).Name(name).Description(description).DeviceType(deviceType).UpdateType(updateType).Page(page).PerPage(perPage).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DeploymentsV1ListReleasesWithPagination``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV1ListReleasesWithPagination`: []ReleaseV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DeploymentsV1ListReleasesWithPagination`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV1ListReleasesWithPaginationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Release name filter. | 
 **description** | **string** | Release description filter. | 
 **deviceType** | **string** | Release device type filter. | 
 **updateType** | **string** | Update type filter. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **sort** | **string** | Sort the release list by the specified field and direction.  | [default to &quot;name:asc&quot;]

### Return type

[**[]ReleaseV1**](ReleaseV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DownloadArtifact

> ArtifactLink DownloadArtifact(ctx, id).Execute()

Get the download link of a selected artifact



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
	id := "id_example" // string | Artifact identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.DownloadArtifact(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.DownloadArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DownloadArtifact`: ArtifactLink
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.DownloadArtifact`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Artifact identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDownloadArtifactRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ArtifactLink**](ArtifactLink.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GenerateArtifact

> GenerateArtifact(ctx).Name(name).DeviceTypesCompatible(deviceTypesCompatible).Type_(type_).File(file).Description(description).Args(args).Execute()

Upload raw data to generate a new artifact



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
	name := "name_example" // string | Name of the artifact to generate.
	deviceTypesCompatible := []string{"Inner_example"} // []string | An array of compatible device types.
	type_ := "type__example" // string | Update Module used to generate the artifact.
	file := os.NewFile(1234, "some_file") // *os.File | Raw file to be used to generate the artifact. It has to be the last part of request.
	description := "description_example" // string | Description of the artifact to generate. (optional)
	args := "args_example" // string | String that represents a JSON document defining the arguments used to generate the artifact. The service won't parse the content of this parameter and pass it as it is to the create artifact worker. The available arguments and options depend on the Update Module implementation and are, therefore, Type-specific.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.GenerateArtifact(context.Background()).Name(name).DeviceTypesCompatible(deviceTypesCompatible).Type_(type_).File(file).Description(description).Args(args).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.GenerateArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGenerateArtifactRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Name of the artifact to generate. | 
 **deviceTypesCompatible** | **[]string** | An array of compatible device types. | 
 **type_** | **string** | Update Module used to generate the artifact. | 
 **file** | ***os.File** | Raw file to be used to generate the artifact. It has to be the last part of request. | 
 **description** | **string** | Description of the artifact to generate. | 
 **args** | **string** | String that represents a JSON document defining the arguments used to generate the artifact. The service won&#39;t parse the content of this parameter and pass it as it is to the create artifact worker. The available arguments and options depend on the Update Module implementation and are, therefore, Type-specific.  | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDeploymentLogForDevice

> string GetDeploymentLogForDevice(ctx, deploymentId, deviceId).Execute()

Get the log of a selected device's deployment



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
	deploymentId := "deploymentId_example" // string | Deployment identifier.
	deviceId := "deviceId_example" // string | Device identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.GetDeploymentLogForDevice(context.Background(), deploymentId, deviceId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.GetDeploymentLogForDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeploymentLogForDevice`: string
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.GetDeploymentLogForDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment identifier. | 
**deviceId** | **string** | Device identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDeploymentLogForDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

**string**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListAllDevicesInDeployment

> []DeviceWithImage ListAllDevicesInDeployment(ctx, deploymentId).Execute()

DEPRECATED: _since Wed May 19 2021_ this end-point is deprecated because it doesn't support pagination and will be removed in the future, please use the /deployments/{deployment_id}/devices/list end-point instead. 

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
	deploymentId := "deploymentId_example" // string | Deployment identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(context.Background(), deploymentId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ListAllDevicesInDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAllDevicesInDeployment`: []DeviceWithImage
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ListAllDevicesInDeployment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListAllDevicesInDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DeviceWithImage**](DeviceWithImage.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListArtifacts

> []ArtifactV1 ListArtifacts(ctx).Name(name).Description(description).DeviceType(deviceType).Execute()

List all the artifacts 



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
	name := "name_example" // string | Release name filter. (optional)
	description := "description_example" // string | Release description filter. (optional)
	deviceType := "deviceType_example" // string | Release device type filter. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ListArtifacts(context.Background()).Name(name).Description(description).DeviceType(deviceType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ListArtifacts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListArtifacts`: []ArtifactV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ListArtifacts`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListArtifactsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Release name filter. | 
 **description** | **string** | Release description filter. | 
 **deviceType** | **string** | Release device type filter. | 

### Return type

[**[]ArtifactV1**](ArtifactV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDeviceIDsInDeployment

> []string ListDeviceIDsInDeployment(ctx, id).Execute()

Get the list of device IDs being part of the deployment.

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
	id := "id_example" // string | Deployment identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ListDeviceIDsInDeployment(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ListDeviceIDsInDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDeviceIDsInDeployment`: []string
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ListDeviceIDsInDeployment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListDeviceIDsInDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## ListDevicesInDeployment

> []DeviceWithImage ListDevicesInDeployment(ctx, deploymentId).Status(status).Page(page).PerPage(perPage).Execute()

Get the list of devices and their respective status for the deployment with the given ID. The response includes devices as they get assigned to the deployment when checking for updates. Therefore, this endpoint will list all the devices only once each asks for updates and evaluates this deployment. 

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
	deploymentId := "deploymentId_example" // string | Deployment identifier.
	status := "status_example" // string | Filter devices by status within deployment. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ListDevicesInDeployment(context.Background(), deploymentId).Status(status).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ListDevicesInDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDevicesInDeployment`: []DeviceWithImage
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ListDevicesInDeployment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListDevicesInDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Filter devices by status within deployment. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]DeviceWithImage**](DeviceWithImage.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListReleases

> []ReleaseV1 ListReleases(ctx).Name(name).Description(description).DeviceType(deviceType).UpdateType(updateType).Execute()

List releases 



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
	name := "name_example" // string | Release name filter. (optional)
	description := "description_example" // string | Release description filter. (optional)
	deviceType := "deviceType_example" // string | Release device type filter. (optional)
	updateType := "updateType_example" // string | Update type filter. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ListReleases(context.Background()).Name(name).Description(description).DeviceType(deviceType).UpdateType(updateType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ListReleases``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListReleases`: []ReleaseV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ListReleases`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListReleasesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Release name filter. | 
 **description** | **string** | Release description filter. | 
 **deviceType** | **string** | Release device type filter. | 
 **updateType** | **string** | Update type filter. | 

### Return type

[**[]ReleaseV1**](ReleaseV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RequestDirectUpload

> ArtifactUploadLink RequestDirectUpload(ctx).Execute()

Request link for uploading artifact directly to the storage backend. This is an on-prem endpoint only, not available on Hosted Mender.

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
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.RequestDirectUpload(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.RequestDirectUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestDirectUpload`: ArtifactUploadLink
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.RequestDirectUpload`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiRequestDirectUploadRequest struct via the builder pattern


### Return type

[**ArtifactUploadLink**](ArtifactUploadLink.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ResetDeviceDeploymentsHistory

> ResetDeviceDeploymentsHistory(ctx, id).Execute()

Reset the Device Deployments history



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
	id := "id_example" // string | System wide device identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.ResetDeviceDeploymentsHistory(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ResetDeviceDeploymentsHistory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | System wide device identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiResetDeviceDeploymentsHistoryRequest struct via the builder pattern


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


## ShowArtifact

> ArtifactV1 ShowArtifact(ctx, id).Execute()

Get the details of a selected artifact



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
	id := "id_example" // string | Artifact identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ShowArtifact(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ShowArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowArtifact`: ArtifactV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ShowArtifact`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Artifact identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiShowArtifactRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ArtifactV1**](ArtifactV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShowDeployment

> DeploymentV1 ShowDeployment(ctx, id).Execute()

Get the details of a selected deployment



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
	id := "id_example" // string | Deployment identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsManagementAPIAPI.ShowDeployment(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.ShowDeployment``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowDeployment`: DeploymentV1
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsManagementAPIAPI.ShowDeployment`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiShowDeploymentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeploymentV1**](DeploymentV1.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateArtifactInfo

> UpdateArtifactInfo(ctx, id).ArtifactUpdateV1(artifactUpdateV1).Execute()

Update description of a selected artifact



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
	id := "id_example" // string | Artifact identifier.
	artifactUpdateV1 := *openapiclient.NewArtifactUpdateV1() // ArtifactUpdateV1 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.UpdateArtifactInfo(context.Background(), id).ArtifactUpdateV1(artifactUpdateV1).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.UpdateArtifactInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Artifact identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateArtifactInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **artifactUpdateV1** | [**ArtifactUpdateV1**](ArtifactUpdateV1.md) |  | 

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


## UploadArtifact

> UploadArtifact(ctx).Artifact(artifact).Size(size).Description(description).Execute()

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
	artifact := os.NewFile(1234, "some_file") // *os.File | Artifact. It has to be the last part of request.
	size := int32(56) // int32 | Size of the artifact file in bytes. DEPRECATED: _since Mon Apr 6 2020_ Size is determined from uploaded content.  (optional)
	description := "description_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsManagementAPIAPI.UploadArtifact(context.Background()).Artifact(artifact).Size(size).Description(description).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsManagementAPIAPI.UploadArtifact``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUploadArtifactRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **artifact** | ***os.File** | Artifact. It has to be the last part of request. | 
 **size** | **int32** | Size of the artifact file in bytes. DEPRECATED: _since Mon Apr 6 2020_ Size is determined from uploaded content.  | 
 **description** | **string** |  | 

### Return type

 (empty response body)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

