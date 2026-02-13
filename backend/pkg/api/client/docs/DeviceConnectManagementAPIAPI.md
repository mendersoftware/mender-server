# \DeviceConnectManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceConnectManagementCheckUpdate**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementCheckUpdate) | **Post** /api/management/v1/deviceconnect/devices/{id}/check-update | Trigger check-update for the Mender client running on the device
[**DeviceConnectManagementConnect**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementConnect) | **Get** /api/management/v1/deviceconnect/devices/{id}/connect | Establish permanent connection with device
[**DeviceConnectManagementDownload**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementDownload) | **Get** /api/management/v1/deviceconnect/devices/{id}/download | Download a file from the device
[**DeviceConnectManagementGetDevice**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementGetDevice) | **Get** /api/management/v1/deviceconnect/devices/{id} | Fetch the state of a device.
[**DeviceConnectManagementPlayback**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementPlayback) | **Get** /api/management/v1/deviceconnect/sessions/{session_id}/playback | Establish a connection for playing back a session
[**DeviceConnectManagementSendInventory**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementSendInventory) | **Post** /api/management/v1/deviceconnect/devices/{id}/send-inventory | Trigger send-inventory for the Mender client running on the device
[**DeviceConnectManagementUpload**](DeviceConnectManagementAPIAPI.md#DeviceConnectManagementUpload) | **Put** /api/management/v1/deviceconnect/devices/{id}/upload | Upload a file to the device



## DeviceConnectManagementCheckUpdate

> DeviceConnectManagementCheckUpdate(ctx, id).Execute()

Trigger check-update for the Mender client running on the device

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementCheckUpdate(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementCheckUpdate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementCheckUpdateRequest struct via the builder pattern


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


## DeviceConnectManagementConnect

> DeviceConnectManagementConnect(ctx, id).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()

Establish permanent connection with device

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.
	connection := "connection_example" // string | Standard websocket request header. (optional)
	upgrade := "upgrade_example" // string | Standard websocket request header. (optional)
	secWebsocketKey := "secWebsocketKey_example" // string | Standard websocket request header. (optional)
	secWebsocketVersion := int32(56) // int32 | Standard websocket request header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementConnect(context.Background(), id).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementConnect``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementConnectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **connection** | **string** | Standard websocket request header. | 
 **upgrade** | **string** | Standard websocket request header. | 
 **secWebsocketKey** | **string** | Standard websocket request header. | 
 **secWebsocketVersion** | **int32** | Standard websocket request header. | 

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


## DeviceConnectManagementDownload

> *os.File DeviceConnectManagementDownload(ctx, id).Path(path).Execute()

Download a file from the device

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.
	path := "path_example" // string | Path of the file on the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementDownload(context.Background(), id).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementDownload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConnectManagementDownload`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `DeviceConnectManagementAPIAPI.DeviceConnectManagementDownload`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementDownloadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **path** | **string** | Path of the file on the device. | 

### Return type

[***os.File**](*os.File.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConnectManagementGetDevice

> ConnectionState DeviceConnectManagementGetDevice(ctx, id).Execute()

Fetch the state of a device.

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementGetDevice(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementGetDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceConnectManagementGetDevice`: ConnectionState
	fmt.Fprintf(os.Stdout, "Response from `DeviceConnectManagementAPIAPI.DeviceConnectManagementGetDevice`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementGetDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ConnectionState**](ConnectionState.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeviceConnectManagementPlayback

> DeviceConnectManagementPlayback(ctx, sessionId).SleepMs(sleepMs).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()

Establish a connection for playing back a session

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
	sessionId := "sessionId_example" // string | ID for the session to play back.
	sleepMs := int32(56) // int32 | Time in millisconds to sleep between the subsequent playback data writes. (optional)
	connection := "connection_example" // string | Standard websocket request header. (optional)
	upgrade := "upgrade_example" // string | Standard websocket request header. (optional)
	secWebsocketKey := "secWebsocketKey_example" // string | Standard websocket request header. (optional)
	secWebsocketVersion := int32(56) // int32 | Standard websocket request header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementPlayback(context.Background(), sessionId).SleepMs(sleepMs).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementPlayback``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sessionId** | **string** | ID for the session to play back. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementPlaybackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **sleepMs** | **int32** | Time in millisconds to sleep between the subsequent playback data writes. | 
 **connection** | **string** | Standard websocket request header. | 
 **upgrade** | **string** | Standard websocket request header. | 
 **secWebsocketKey** | **string** | Standard websocket request header. | 
 **secWebsocketVersion** | **int32** | Standard websocket request header. | 

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


## DeviceConnectManagementSendInventory

> DeviceConnectManagementSendInventory(ctx, id).Execute()

Trigger send-inventory for the Mender client running on the device

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementSendInventory(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementSendInventory``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementSendInventoryRequest struct via the builder pattern


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


## DeviceConnectManagementUpload

> DeviceConnectManagementUpload(ctx, id).Path(path).Uid(uid).Gid(gid).Mode(mode).File(file).Execute()

Upload a file to the device

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
	id := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | ID of the device.
	path := "path_example" // string | The destination path on the device
	uid := int32(56) // int32 | The numerical UID of the file on the device (optional)
	gid := int32(56) // int32 | The numerical GID of the file on the device (optional)
	mode := "mode_example" // string | The octal representation of the mode of the file on the device (optional)
	file := os.NewFile(1234, "some_file") // *os.File |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectManagementAPIAPI.DeviceConnectManagementUpload(context.Background(), id).Path(path).Uid(uid).Gid(gid).Mode(mode).File(file).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectManagementAPIAPI.DeviceConnectManagementUpload``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | ID of the device. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectManagementUploadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **path** | **string** | The destination path on the device | 
 **uid** | **int32** | The numerical UID of the file on the device | 
 **gid** | **int32** | The numerical GID of the file on the device | 
 **mode** | **string** | The octal representation of the mode of the file on the device | 
 **file** | ***os.File** |  | 

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

