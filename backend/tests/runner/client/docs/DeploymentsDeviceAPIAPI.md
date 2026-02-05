# \DeploymentsDeviceAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CheckUpdate**](DeploymentsDeviceAPIAPI.md#CheckUpdate) | **Get** /api/devices/v1/deployments/device/deployments/next | Get next update
[**FetchConfiguration**](DeploymentsDeviceAPIAPI.md#FetchConfiguration) | **Get** /api/devices/v1/deployments/download/configuration/{deployment_id}/{device_type}/{device_id} | Internally generated download link for deploying device configurations. All parameters are generated internally when fetching a configuration deployment. 
[**ReportDeploymentLog**](DeploymentsDeviceAPIAPI.md#ReportDeploymentLog) | **Put** /api/devices/v1/deployments/device/deployments/{id}/log | Upload the device deployment log
[**UpdateDeploymentStatus**](DeploymentsDeviceAPIAPI.md#UpdateDeploymentStatus) | **Put** /api/devices/v1/deployments/device/deployments/{id}/status | Update the device deployment status



## CheckUpdate

> DeploymentInstructions CheckUpdate(ctx).ArtifactName(artifactName).DeviceType(deviceType).Execute()

Get next update



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
	artifactName := "artifactName_example" // string | currently installed artifact
	deviceType := "deviceType_example" // string | Device type of device

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsDeviceAPIAPI.CheckUpdate(context.Background()).ArtifactName(artifactName).DeviceType(deviceType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsDeviceAPIAPI.CheckUpdate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CheckUpdate`: DeploymentInstructions
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsDeviceAPIAPI.CheckUpdate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCheckUpdateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **artifactName** | **string** | currently installed artifact | 
 **deviceType** | **string** | Device type of device | 

### Return type

[**DeploymentInstructions**](DeploymentInstructions.md)

### Authorization

[DeviceJWT](../README.md#DeviceJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FetchConfiguration

> *os.File FetchConfiguration(ctx, deploymentId, deviceType, deviceId).XMenExpire(xMenExpire).XMenSignature(xMenSignature).TenantId(tenantId).Execute()

Internally generated download link for deploying device configurations. All parameters are generated internally when fetching a configuration deployment. 

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	deploymentId := "deploymentId_example" // string | Deployment UUID
	deviceType := "deviceType_example" // string | Device type of the calling device
	deviceId := "deviceId_example" // string | Device UUID
	xMenExpire := time.Now() // time.Time | Time of link expire
	xMenSignature := "xMenSignature_example" // string | Signature of the URL link
	tenantId := "tenantId_example" // string | Device tenant ID (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeploymentsDeviceAPIAPI.FetchConfiguration(context.Background(), deploymentId, deviceType, deviceId).XMenExpire(xMenExpire).XMenSignature(xMenSignature).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsDeviceAPIAPI.FetchConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `FetchConfiguration`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `DeploymentsDeviceAPIAPI.FetchConfiguration`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**deploymentId** | **string** | Deployment UUID | 
**deviceType** | **string** | Device type of the calling device | 
**deviceId** | **string** | Device UUID | 

### Other Parameters

Other parameters are passed through a pointer to a apiFetchConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xMenExpire** | **time.Time** | Time of link expire | 
 **xMenSignature** | **string** | Signature of the URL link | 
 **tenantId** | **string** | Device tenant ID | 

### Return type

[***os.File**](*os.File.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ReportDeploymentLog

> ReportDeploymentLog(ctx, id).DeploymentLog(deploymentLog).Execute()

Upload the device deployment log



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	id := "id_example" // string | Deployment identifier.
	deploymentLog := *openapiclient.NewDeploymentLog([]openapiclient.DeploymentLogMessagesInner{*openapiclient.NewDeploymentLogMessagesInner(time.Now(), "Level_example", "Message_example")}) // DeploymentLog | Deployment log

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsDeviceAPIAPI.ReportDeploymentLog(context.Background(), id).DeploymentLog(deploymentLog).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsDeviceAPIAPI.ReportDeploymentLog``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiReportDeploymentLogRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deploymentLog** | [**DeploymentLog**](DeploymentLog.md) | Deployment log | 

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


## UpdateDeploymentStatus

> UpdateDeploymentStatus(ctx, id).DeploymentStatus(deploymentStatus).Execute()

Update the device deployment status



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
	deploymentStatus := *openapiclient.NewDeploymentStatus("Status_example") // DeploymentStatus | Deployment status.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeploymentsDeviceAPIAPI.UpdateDeploymentStatus(context.Background(), id).DeploymentStatus(deploymentStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeploymentsDeviceAPIAPI.UpdateDeploymentStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Deployment identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeploymentStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deploymentStatus** | [**DeploymentStatus**](DeploymentStatus.md) | Deployment status. | 

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

