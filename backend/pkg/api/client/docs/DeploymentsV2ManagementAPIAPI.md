# \DeploymentsV2ManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AssignReleaseTags**](DeploymentsV2ManagementAPIAPI.md#AssignReleaseTags) | **Put** /api/management/v2/deployments/deployments/releases/{release_name}/tags | Update and replace the tags of a release. 
[**DeleteReleases**](DeploymentsV2ManagementAPIAPI.md#DeleteReleases) | **Delete** /api/management/v2/deployments/deployments/releases | Delete the releases with given names
[**DeploymentsV2ListArtifactsWithPagination**](DeploymentsV2ManagementAPIAPI.md#DeploymentsV2ListArtifactsWithPagination) | **Get** /api/management/v2/deployments/artifacts | Lists known artifacts. 
[**DeploymentsV2ListDeployments**](DeploymentsV2ManagementAPIAPI.md#DeploymentsV2ListDeployments) | **Get** /api/management/v2/deployments/deployments | List all the deployments matching the specified filter parameters
[**DeploymentsV2ListReleasesWithPagination**](DeploymentsV2ManagementAPIAPI.md#DeploymentsV2ListReleasesWithPagination) | **Get** /api/management/v2/deployments/deployments/releases | List releases 
[**GetReleaseWithGivenName**](DeploymentsV2ManagementAPIAPI.md#GetReleaseWithGivenName) | **Get** /api/management/v2/deployments/deployments/releases/{release_name} | Get release 
[**ListReleaseTags**](DeploymentsV2ManagementAPIAPI.md#ListReleaseTags) | **Get** /api/management/v2/deployments/releases/all/tags | Lists all available tags for releases. 
[**ListReleaseTypes**](DeploymentsV2ManagementAPIAPI.md#ListReleaseTypes) | **Get** /api/management/v2/deployments/releases/all/types | Lists all release update types. 
[**UpdateReleaseInformation**](DeploymentsV2ManagementAPIAPI.md#UpdateReleaseInformation) | **Patch** /api/management/v2/deployments/deployments/releases/{release_name} | Update selected fields of the Release object. 



## AssignReleaseTags

> AssignReleaseTags(ctx, releaseName).RequestBody(requestBody).Execute()

Update and replace the tags of a release. 



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
	releaseName := "releaseName_example" // string | Name of the release
	requestBody := []string{"Property_example"} // []string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsV2ManagementAPIAPI.AssignReleaseTags(context.Background(), releaseName).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.AssignReleaseTags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**releaseName** | **string** | Name of the release | 

### Other Parameters

Other parameters are passed through a pointer to a apiAssignReleaseTagsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **requestBody** | **[]string** |  | 

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


## DeleteReleases

> DeleteReleases(ctx).Name(name).Execute()

Delete the releases with given names



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
	name := "name_example" // string | Name of the release to be deleted

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsV2ManagementAPIAPI.DeleteReleases(context.Background()).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.DeleteReleases``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteReleasesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Name of the release to be deleted | 

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


## DeploymentsV2ListArtifactsWithPagination

> []ArtifactV2 DeploymentsV2ListArtifactsWithPagination(ctx).Name(name).Description(description).DeviceType(deviceType).Page(page).PerPage(perPage).Execute()

Lists known artifacts. 



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
	name := []string{"Inner_example"} // []string | Artifact(s) name(s) filter. Multiple names can be provided (e.g., `?name=foo&name=bar`). Supports exact matching or prefix matching by adding `*` to the end (e.g., `foo*`). Note: when using prefix matching you may pass only a single value and you cannot combine prefix matching and exact matching in the same request.  (optional)
	description := "description_example" // string | Artifact description filter. Supports exact matching or prefix matching by adding `*` to the end (e.g., `foo*`).  (optional)
	deviceType := "deviceType_example" // string | Artifact device type filter. Supports exact matching or prefix matching by adding `*` to the end (e.g., `foo*`).  (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.DeploymentsV2ListArtifactsWithPagination(context.Background()).Name(name).Description(description).DeviceType(deviceType).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListArtifactsWithPagination``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV2ListArtifactsWithPagination`: []ArtifactV2
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListArtifactsWithPagination`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV2ListArtifactsWithPaginationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **[]string** | Artifact(s) name(s) filter. Multiple names can be provided (e.g., &#x60;?name&#x3D;foo&amp;name&#x3D;bar&#x60;). Supports exact matching or prefix matching by adding &#x60;*&#x60; to the end (e.g., &#x60;foo*&#x60;). Note: when using prefix matching you may pass only a single value and you cannot combine prefix matching and exact matching in the same request.  | 
 **description** | **string** | Artifact description filter. Supports exact matching or prefix matching by adding &#x60;*&#x60; to the end (e.g., &#x60;foo*&#x60;).  | 
 **deviceType** | **string** | Artifact device type filter. Supports exact matching or prefix matching by adding &#x60;*&#x60; to the end (e.g., &#x60;foo*&#x60;).  | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]ArtifactV2**](ArtifactV2.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsV2ListDeployments

> []DeploymentV2 DeploymentsV2ListDeployments(ctx).Id(id).Name(name).Status(status).Type_(type_).Page(page).PerPage(perPage).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()

List all the deployments matching the specified filter parameters



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
	id := "id_example" // string | Deployment identifier. You can provide it multiple times to query a set of deployments.  (optional)
	name := "name_example" // string | Deployment name. You can provide it multiple times to query a set of deployments.  (optional)
	status := "status_example" // string | Deployment status filter. (optional)
	type_ := "type__example" // string | Deployment type filter.  (optional)
	page := int32(56) // int32 | Results page number (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	createdBefore := int32(56) // int32 | List only deployments created before and equal to Unix timestamp (UTC) (optional)
	createdAfter := int32(56) // int32 | List only deployments created after and equal to Unix timestamp (UTC) (optional)
	sort := "sort_example" // string | Supports sorting the deployments list by creation date.  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.DeploymentsV2ListDeployments(context.Background()).Id(id).Name(name).Status(status).Type_(type_).Page(page).PerPage(perPage).CreatedBefore(createdBefore).CreatedAfter(createdAfter).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListDeployments``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV2ListDeployments`: []DeploymentV2
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListDeployments`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV2ListDeploymentsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Deployment identifier. You can provide it multiple times to query a set of deployments.  | 
 **name** | **string** | Deployment name. You can provide it multiple times to query a set of deployments.  | 
 **status** | **string** | Deployment status filter. | 
 **type_** | **string** | Deployment type filter.  | 
 **page** | **int32** | Results page number | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **createdBefore** | **int32** | List only deployments created before and equal to Unix timestamp (UTC) | 
 **createdAfter** | **int32** | List only deployments created after and equal to Unix timestamp (UTC) | 
 **sort** | **string** | Supports sorting the deployments list by creation date.  | 

### Return type

[**[]DeploymentV2**](DeploymentV2.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeploymentsV2ListReleasesWithPagination

> []ReleaseV2 DeploymentsV2ListReleasesWithPagination(ctx).Name(name).Tag(tag).UpdateType(updateType).Page(page).PerPage(perPage).Sort(sort).Execute()

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
	tag := []string{"Inner_example"} // []string | Tag filter. (optional)
	updateType := "updateType_example" // string | Update type filter. (optional)
	page := int32(56) // int32 | Starting page. (optional) (default to 1)
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)
	sort := "sort_example" // string | Sort the release list by the specified field and direction.  (optional) (default to "name:asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.DeploymentsV2ListReleasesWithPagination(context.Background()).Name(name).Tag(tag).UpdateType(updateType).Page(page).PerPage(perPage).Sort(sort).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListReleasesWithPagination``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeploymentsV2ListReleasesWithPagination`: []ReleaseV2
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.DeploymentsV2ListReleasesWithPagination`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeploymentsV2ListReleasesWithPaginationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Release name filter. | 
 **tag** | **[]string** | Tag filter. | 
 **updateType** | **string** | Update type filter. | 
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]
 **sort** | **string** | Sort the release list by the specified field and direction.  | [default to &quot;name:asc&quot;]

### Return type

[**[]ReleaseV2**](ReleaseV2.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetReleaseWithGivenName

> ReleaseV2 GetReleaseWithGivenName(ctx, releaseName).Execute()

Get release 



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
	releaseName := "releaseName_example" // string | Name of the release

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.GetReleaseWithGivenName(context.Background(), releaseName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.GetReleaseWithGivenName``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetReleaseWithGivenName`: ReleaseV2
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.GetReleaseWithGivenName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**releaseName** | **string** | Name of the release | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetReleaseWithGivenNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ReleaseV2**](ReleaseV2.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListReleaseTags

> []string ListReleaseTags(ctx).Execute()

Lists all available tags for releases. 

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
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.ListReleaseTags(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.ListReleaseTags``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListReleaseTags`: []string
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.ListReleaseTags`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListReleaseTagsRequest struct via the builder pattern


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


## ListReleaseTypes

> []string ListReleaseTypes(ctx).Execute()

Lists all release update types. 

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
	resp, r, err := apiClient.DeploymentsV2ManagementAPIAPI.ListReleaseTypes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.ListReleaseTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListReleaseTypes`: []string
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsV2ManagementAPIAPI.ListReleaseTypes`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListReleaseTypesRequest struct via the builder pattern


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


## UpdateReleaseInformation

> UpdateReleaseInformation(ctx, releaseName).ReleaseUpdate(releaseUpdate).Execute()

Update selected fields of the Release object. 



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
	releaseName := "releaseName_example" // string | Name of the release
	releaseUpdate := *openapiclient.NewReleaseUpdate() // ReleaseUpdate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsV2ManagementAPIAPI.UpdateReleaseInformation(context.Background(), releaseName).ReleaseUpdate(releaseUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsV2ManagementAPIAPI.UpdateReleaseInformation``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**releaseName** | **string** | Name of the release | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateReleaseInformationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **releaseUpdate** | [**ReleaseUpdate**](ReleaseUpdate.md) |  | 

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

