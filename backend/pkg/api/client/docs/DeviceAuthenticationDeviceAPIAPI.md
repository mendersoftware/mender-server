# \DeviceAuthenticationDeviceAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeviceAuthAuthenticateDevice**](DeviceAuthenticationDeviceAPIAPI.md#DeviceAuthAuthenticateDevice) | **Post** /api/devices/v1/authentication/auth_requests | Submit an authentication request



## DeviceAuthAuthenticateDevice

> string DeviceAuthAuthenticateDevice(ctx).XMENSignature(xMENSignature).AuthRequest(authRequest).Execute()

Submit an authentication request



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
	xMENSignature := "xMENSignature_example" // string | Request signature. The request signature depends on the public key submitted in the AuthRequest. A summary of signature algorithms and format follows: | Type       | Digest              | Format                   | Algorithm    | |------------|---------------------|--------------------------|--------------| | RSA        | SHA256(AuthRequest) | Base64(Signature)        | [RFC2313]    | | ECDSA      | SHA256(AuthRequest) | Base64(ASN.1(SEQ{R, S})) | [ANSI x9.62] | | Ed25519    | AuthRequest         | Base64(Signature)        | [RFC8032]    | *Remark:* For ECDSA, the signature constitutes two integers (R and S) in which case the binary signature is taken as the ASN.1 sequence of the two numbers in the given order. 
	authRequest := *openapiclient.NewAuthRequest("IdData_example", "Pubkey_example") // AuthRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceAuthenticationDeviceAPIAPI.DeviceAuthAuthenticateDevice(context.Background()).XMENSignature(xMENSignature).AuthRequest(authRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceAuthenticationDeviceAPIAPI.DeviceAuthAuthenticateDevice``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeviceAuthAuthenticateDevice`: string
	fmt.Fprintf(os.Stdout, "Response from `DeviceAuthenticationDeviceAPIAPI.DeviceAuthAuthenticateDevice`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeviceAuthAuthenticateDeviceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xMENSignature** | **string** | Request signature. The request signature depends on the public key submitted in the AuthRequest. A summary of signature algorithms and format follows: | Type       | Digest              | Format                   | Algorithm    | |------------|---------------------|--------------------------|--------------| | RSA        | SHA256(AuthRequest) | Base64(Signature)        | [RFC2313]    | | ECDSA      | SHA256(AuthRequest) | Base64(ASN.1(SEQ{R, S})) | [ANSI x9.62] | | Ed25519    | AuthRequest         | Base64(Signature)        | [RFC8032]    | *Remark:* For ECDSA, the signature constitutes two integers (R and S) in which case the binary signature is taken as the ASN.1 sequence of the two numbers in the given order.  | 
 **authRequest** | [**AuthRequest**](AuthRequest.md) |  | 

### Return type

**string**

### Authorization

[DeviceJWT](../README.md#DeviceJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

