# \UserAdministrationAndAuthenticationInternalAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateUserInternal**](UserAdministrationAndAuthenticationInternalAPIAPI.md#CreateUserInternal) | **Post** /api/internal/v1/useradm/tenants/{tenant_id}/users | Create user
[**DeleteUserInternal**](UserAdministrationAndAuthenticationInternalAPIAPI.md#DeleteUserInternal) | **Delete** /api/internal/v1/useradm/tenants/{tenant_id}/users/{user_id} | Delete a user
[**ListUsersInternal**](UserAdministrationAndAuthenticationInternalAPIAPI.md#ListUsersInternal) | **Get** /api/internal/v1/useradm/tenants/{tenant_id}/users | List all users registered under the tenant owning the JWT. 
[**RevokeUserTokens**](UserAdministrationAndAuthenticationInternalAPIAPI.md#RevokeUserTokens) | **Delete** /api/internal/v1/useradm/tokens | Delete all user tokens
[**UseradmCheckHealth**](UserAdministrationAndAuthenticationInternalAPIAPI.md#UseradmCheckHealth) | **Get** /api/internal/v1/useradm/health | Check the health of the service
[**UseradmCheckLiveliness**](UserAdministrationAndAuthenticationInternalAPIAPI.md#UseradmCheckLiveliness) | **Get** /api/internal/v1/useradm/alive | Trivial endpoint that unconditionally returns an empty 200 response whenever the API handler is running correctly. 
[**UseradmCreateTenant**](UserAdministrationAndAuthenticationInternalAPIAPI.md#UseradmCreateTenant) | **Post** /api/internal/v1/useradm/tenants | Create a tenant with provided configuration.
[**VerifyJWT**](UserAdministrationAndAuthenticationInternalAPIAPI.md#VerifyJWT) | **Post** /api/internal/v1/useradm/auth/verify | Check the validity of a token



## CreateUserInternal

> CreateUserInternal(ctx, tenantId).UserNewInternal(userNewInternal).Execute()

Create user

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
	userNewInternal := *openapiclient.NewUserNewInternal("Email_example", "Password_example") // UserNewInternal | New user data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.CreateUserInternal(context.Background(), tenantId).UserNewInternal(userNewInternal).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.CreateUserInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateUserInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **userNewInternal** | [**UserNewInternal**](UserNewInternal.md) | New user data. | 

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


## DeleteUserInternal

> DeleteUserInternal(ctx, tenantId, userId).Execute()

Delete a user



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
	userId := "userId_example" // string | User ID.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.DeleteUserInternal(context.Background(), tenantId, userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.DeleteUserInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 
**userId** | **string** | User ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteUserInternalRequest struct via the builder pattern


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


## ListUsersInternal

> []User ListUsersInternal(ctx, tenantId).Id(id).Email(email).CreatedAfter(createdAfter).CreatedBefore(createdBefore).UpdatedAfter(updatedAfter).UpdatedBefore(updatedBefore).Execute()

List all users registered under the tenant owning the JWT. 

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
	id := "id_example" // string | Limit result by user ID, can be repeated to include multiple users in the query.  (optional)
	email := "email_example" // string | Limit result by user email, can be repeated to include multiple users in the query.  (optional)
	createdAfter := int32(56) // int32 | Filter users created after timestamp (UNIX timestamp).  (optional)
	createdBefore := int32(56) // int32 | Filter users created before timestamp (UNIX timestamp).  (optional)
	updatedAfter := int32(56) // int32 | Filter users updated after timestamp (UNIX timestamp).  (optional)
	updatedBefore := int32(56) // int32 | Filter users updated before timestamp (UNIX timestamp).  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.ListUsersInternal(context.Background(), tenantId).Id(id).Email(email).CreatedAfter(createdAfter).CreatedBefore(createdBefore).UpdatedAfter(updatedAfter).UpdatedBefore(updatedBefore).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.ListUsersInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListUsersInternal`: []User
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationAndAuthenticationInternalAPIAPI.ListUsersInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | Tenant ID. | 

### Other Parameters

Other parameters are passed through a pointer to a apiListUsersInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **id** | **string** | Limit result by user ID, can be repeated to include multiple users in the query.  | 
 **email** | **string** | Limit result by user email, can be repeated to include multiple users in the query.  | 
 **createdAfter** | **int32** | Filter users created after timestamp (UNIX timestamp).  | 
 **createdBefore** | **int32** | Filter users created before timestamp (UNIX timestamp).  | 
 **updatedAfter** | **int32** | Filter users updated after timestamp (UNIX timestamp).  | 
 **updatedBefore** | **int32** | Filter users updated before timestamp (UNIX timestamp).  | 

### Return type

[**[]User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RevokeUserTokens

> RevokeUserTokens(ctx).TenantId(tenantId).UserId(userId).Execute()

Delete all user tokens



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
	userId := "userId_example" // string | User ID. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.RevokeUserTokens(context.Background()).TenantId(tenantId).UserId(userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.RevokeUserTokens``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRevokeUserTokensRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | Tenant ID. | 
 **userId** | **string** | User ID. | 

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


## UseradmCheckHealth

> UseradmCheckHealth(ctx).Execute()

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
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiUseradmCheckHealthRequest struct via the builder pattern


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


## UseradmCheckLiveliness

> UseradmCheckLiveliness(ctx).Execute()

Trivial endpoint that unconditionally returns an empty 200 response whenever the API handler is running correctly. 

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
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiUseradmCheckLivelinessRequest struct via the builder pattern


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


## UseradmCreateTenant

> UseradmCreateTenant(ctx).TenantNew(tenantNew).Execute()

Create a tenant with provided configuration.

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
	tenantNew := *openapiclient.NewTenantNew() // TenantNew | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCreateTenant(context.Background()).TenantNew(tenantNew).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.UseradmCreateTenant``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUseradmCreateTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantNew** | [**TenantNew**](TenantNew.md) |  | 

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


## VerifyJWT

> VerifyJWT(ctx).Authorization(authorization).XForwardedUri(xForwardedUri).XForwardedMethod(xForwardedMethod).Execute()

Check the validity of a token



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
	authorization := "authorization_example" // string | The token in base64-encoded form.
	xForwardedUri := "xForwardedUri_example" // string | URI the original request was sent to, the URI is expected to have at least 4 components, eg. /api/management/1.0/foo/bar 
	xForwardedMethod := "xForwardedMethod_example" // string | HTTP method used when accessing the original URI

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationAndAuthenticationInternalAPIAPI.VerifyJWT(context.Background()).Authorization(authorization).XForwardedUri(xForwardedUri).XForwardedMethod(xForwardedMethod).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationAndAuthenticationInternalAPIAPI.VerifyJWT``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiVerifyJWTRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string** | The token in base64-encoded form. | 
 **xForwardedUri** | **string** | URI the original request was sent to, the URI is expected to have at least 4 components, eg. /api/management/1.0/foo/bar  | 
 **xForwardedMethod** | **string** | HTTP method used when accessing the original URI | 

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

