# \DeviceConnectDeviceAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceConnectConnect**](DeviceConnectDeviceAPIAPI.md#DeviceConnectConnect) | **Get** /api/devices/v1/deviceconnect/connect | Connect the device and make it available to the server.



## DeviceConnectConnect

> DeviceConnectConnect(ctx).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()

Connect the device and make it available to the server.



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
	connection := "connection_example" // string | Standard websocket request header. (optional)
	upgrade := "upgrade_example" // string | Standard websocket request header. (optional)
	secWebsocketKey := "secWebsocketKey_example" // string | Standard websocket request header. (optional)
	secWebsocketVersion := int32(56) // int32 | Standard websocket request header. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceConnectDeviceAPIAPI.DeviceConnectConnect(context.Background()).Connection(connection).Upgrade(upgrade).SecWebsocketKey(secWebsocketKey).SecWebsocketVersion(secWebsocketVersion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceConnectDeviceAPIAPI.DeviceConnectConnect``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceConnectConnectRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **connection** | **string** | Standard websocket request header. | 
 **upgrade** | **string** | Standard websocket request header. | 
 **secWebsocketKey** | **string** | Standard websocket request header. | 
 **secWebsocketVersion** | **int32** | Standard websocket request header. | 

### Return type

 (empty response body)

### Authorization

[DeviceJWT](../README.md#DeviceJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

