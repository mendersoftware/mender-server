# \UserAdministrationManagementAPIAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePersonalAccessToken**](UserAdministrationManagementAPIAPI.md#CreatePersonalAccessToken) | **Post** /api/management/v1/useradm/settings/tokens | Create new Personal Access Token
[**CreateUserManagement**](UserAdministrationManagementAPIAPI.md#CreateUserManagement) | **Post** /api/management/v1/useradm/users | Create a new user under the tenant owning the JWT. 
[**ListPlans**](UserAdministrationManagementAPIAPI.md#ListPlans) | **Get** /api/management/v1/useradm/plans | Get list of available plans
[**ListUserPersonalAccessTokens**](UserAdministrationManagementAPIAPI.md#ListUserPersonalAccessTokens) | **Get** /api/management/v1/useradm/settings/tokens | Get user Personal Access Tokens
[**ListUsersManagement**](UserAdministrationManagementAPIAPI.md#ListUsersManagement) | **Get** /api/management/v1/useradm/users | List all users registered under the tenant owning the JWT. 
[**Login**](UserAdministrationManagementAPIAPI.md#Login) | **Post** /api/management/v1/useradm/auth/login | Log in to Mender
[**Logout**](UserAdministrationManagementAPIAPI.md#Logout) | **Post** /api/management/v1/useradm/auth/logout | Log out from Mender
[**RemoveUser**](UserAdministrationManagementAPIAPI.md#RemoveUser) | **Delete** /api/management/v1/useradm/users/{id} | Remove user from the system
[**RevokePersonalAccessToken**](UserAdministrationManagementAPIAPI.md#RevokePersonalAccessToken) | **Delete** /api/management/v1/useradm/settings/tokens/{id} | Revoke Personal Access Token
[**ShowMyUserSettings**](UserAdministrationManagementAPIAPI.md#ShowMyUserSettings) | **Get** /api/management/v1/useradm/settings/me | Get user settings for the current user
[**ShowOwnUserData**](UserAdministrationManagementAPIAPI.md#ShowOwnUserData) | **Get** /api/management/v1/useradm/users/me | Get user information
[**ShowPlanAndLimits**](UserAdministrationManagementAPIAPI.md#ShowPlanAndLimits) | **Get** /api/management/v1/useradm/plan_binding | Get plan and limits information for current tenant
[**ShowUser**](UserAdministrationManagementAPIAPI.md#ShowUser) | **Get** /api/management/v1/useradm/users/{id} | Get user information
[**ShowUserSettings**](UserAdministrationManagementAPIAPI.md#ShowUserSettings) | **Get** /api/management/v1/useradm/settings | Get global user settings
[**UpdateMyUserSettings**](UserAdministrationManagementAPIAPI.md#UpdateMyUserSettings) | **Post** /api/management/v1/useradm/settings/me | Set user settings for the current user
[**UpdateOwnUserData**](UserAdministrationManagementAPIAPI.md#UpdateOwnUserData) | **Put** /api/management/v1/useradm/users/me | Update own user information
[**UpdateUser**](UserAdministrationManagementAPIAPI.md#UpdateUser) | **Put** /api/management/v1/useradm/users/{id} | Update user information
[**UpdateUserSettings**](UserAdministrationManagementAPIAPI.md#UpdateUserSettings) | **Post** /api/management/v1/useradm/settings | Set global user settings



## CreatePersonalAccessToken

> string CreatePersonalAccessToken(ctx).PersonalAccessTokenRequest(personalAccessTokenRequest).Execute()

Create new Personal Access Token



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
	personalAccessTokenRequest := *openapiclient.NewPersonalAccessTokenRequest("Name_example") // PersonalAccessTokenRequest | The token object.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.CreatePersonalAccessToken(context.Background()).PersonalAccessTokenRequest(personalAccessTokenRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.CreatePersonalAccessToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreatePersonalAccessToken`: string
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.CreatePersonalAccessToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePersonalAccessTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **personalAccessTokenRequest** | [**PersonalAccessTokenRequest**](PersonalAccessTokenRequest.md) | The token object. | 

### Return type

**string**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateUserManagement

> CreateUserManagement(ctx).UserNew(userNew).Execute()

Create a new user under the tenant owning the JWT. 

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
	userNew := *openapiclient.NewUserNew("Email_example", "Password_example") // UserNew | New user data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.CreateUserManagement(context.Background()).UserNew(userNew).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.CreateUserManagement``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateUserManagementRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userNew** | [**UserNew**](UserNew.md) | New user data. | 

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


## ListPlans

> []Plan ListPlans(ctx).Page(page).PerPage(perPage).Execute()

Get list of available plans

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
	perPage := int32(56) // int32 | Maximum number of results per page. (optional) (default to 20)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ListPlans(context.Background()).Page(page).PerPage(perPage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ListPlans``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPlans`: []Plan
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ListPlans`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPlansRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** | Starting page. | [default to 1]
 **perPage** | **int32** | Maximum number of results per page. | [default to 20]

### Return type

[**[]Plan**](Plan.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListUserPersonalAccessTokens

> []PersonalAccessToken ListUserPersonalAccessTokens(ctx).Execute()

Get user Personal Access Tokens

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
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ListUserPersonalAccessTokens(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ListUserPersonalAccessTokens``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListUserPersonalAccessTokens`: []PersonalAccessToken
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ListUserPersonalAccessTokens`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListUserPersonalAccessTokensRequest struct via the builder pattern


### Return type

[**[]PersonalAccessToken**](PersonalAccessToken.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListUsersManagement

> []User ListUsersManagement(ctx).Id(id).Email(email).CreatedAfter(createdAfter).CreatedBefore(createdBefore).UpdatedAfter(updatedAfter).UpdatedBefore(updatedBefore).Execute()

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
	id := "id_example" // string | Limit result by user ID, can be repeated to include multiple users in the query.  (optional)
	email := "email_example" // string | Limit result by user email, can be repeated to include multiple users in the query.  (optional)
	createdAfter := int32(56) // int32 | Filter users created after timestamp (UNIX timestamp).  (optional)
	createdBefore := int32(56) // int32 | Filter users created before timestamp (UNIX timestamp).  (optional)
	updatedAfter := int32(56) // int32 | Filter users updated after timestamp (UNIX timestamp).  (optional)
	updatedBefore := int32(56) // int32 | Filter users updated before timestamp (UNIX timestamp).  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ListUsersManagement(context.Background()).Id(id).Email(email).CreatedAfter(createdAfter).CreatedBefore(createdBefore).UpdatedAfter(updatedAfter).UpdatedBefore(updatedBefore).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ListUsersManagement``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListUsersManagement`: []User
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ListUsersManagement`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListUsersManagementRequest struct via the builder pattern


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

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Login

> string Login(ctx).LoginOptions(loginOptions).Execute()

Log in to Mender



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
	loginOptions := *openapiclient.NewLoginOptions() // LoginOptions | Log in options (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.Login(context.Background()).LoginOptions(loginOptions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.Login``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Login`: string
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.Login`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiLoginRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginOptions** | [**LoginOptions**](LoginOptions.md) | Log in options | 

### Return type

**string**

### Authorization

[Login](../README.md#Login)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/jwt, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Logout

> Logout(ctx).Execute()

Log out from Mender



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
	r, err := apiClient.UserAdministrationManagementAPIAPI.Logout(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.Logout``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiLogoutRequest struct via the builder pattern


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


## RemoveUser

> RemoveUser(ctx, id).Execute()

Remove user from the system

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
	id := "id_example" // string | User id.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.RemoveUser(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.RemoveUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | User id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveUserRequest struct via the builder pattern


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


## RevokePersonalAccessToken

> RevokePersonalAccessToken(ctx, id).Execute()

Revoke Personal Access Token

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
	id := "id_example" // string | Token identifier.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.RevokePersonalAccessToken(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.RevokePersonalAccessToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Token identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRevokePersonalAccessTokenRequest struct via the builder pattern


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


## ShowMyUserSettings

> map[string]interface{} ShowMyUserSettings(ctx).Execute()

Get user settings for the current user

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
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ShowMyUserSettings(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ShowMyUserSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowMyUserSettings`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ShowMyUserSettings`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiShowMyUserSettingsRequest struct via the builder pattern


### Return type

**map[string]interface{}**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShowOwnUserData

> User ShowOwnUserData(ctx).Execute()

Get user information

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
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ShowOwnUserData(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ShowOwnUserData``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowOwnUserData`: User
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ShowOwnUserData`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiShowOwnUserDataRequest struct via the builder pattern


### Return type

[**User**](User.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShowPlanAndLimits

> PlanBindingDetails ShowPlanAndLimits(ctx).Execute()

Get plan and limits information for current tenant

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
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ShowPlanAndLimits(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ShowPlanAndLimits``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowPlanAndLimits`: PlanBindingDetails
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ShowPlanAndLimits`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiShowPlanAndLimitsRequest struct via the builder pattern


### Return type

[**PlanBindingDetails**](PlanBindingDetails.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShowUser

> User ShowUser(ctx, id).Execute()

Get user information

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
	id := "id_example" // string | User id.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ShowUser(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ShowUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowUser`: User
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ShowUser`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | User id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiShowUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ShowUserSettings

> map[string]interface{} ShowUserSettings(ctx).Execute()

Get global user settings

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
	resp, r, err := apiClient.UserAdministrationManagementAPIAPI.ShowUserSettings(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.ShowUserSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ShowUserSettings`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `UserAdministrationManagementAPIAPI.ShowUserSettings`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiShowUserSettingsRequest struct via the builder pattern


### Return type

**map[string]interface{}**

### Authorization

[ManagementJWT](../README.md#ManagementJWT)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateMyUserSettings

> UpdateMyUserSettings(ctx).Body(body).IfMatch(ifMatch).Execute()

Set user settings for the current user



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
	body := map[string]interface{}{ ... } // map[string]interface{} | New user settings.
	ifMatch := "ifMatch_example" // string | Contains the settings' current ETag, and performs the update only if it matches the one stored in the database. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.UpdateMyUserSettings(context.Background()).Body(body).IfMatch(ifMatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.UpdateMyUserSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateMyUserSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** | New user settings. | 
 **ifMatch** | **string** | Contains the settings&#39; current ETag, and performs the update only if it matches the one stored in the database. | 

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


## UpdateOwnUserData

> UpdateOwnUserData(ctx).UserUpdate(userUpdate).Execute()

Update own user information

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
	userUpdate := *openapiclient.NewUserUpdate() // UserUpdate | Updated user data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.UpdateOwnUserData(context.Background()).UserUpdate(userUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.UpdateOwnUserData``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateOwnUserDataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userUpdate** | [**UserUpdate**](UserUpdate.md) | Updated user data. | 

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


## UpdateUser

> UpdateUser(ctx, id).UserUpdate(userUpdate).Execute()

Update user information

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
	id := "id_example" // string | User id.
	userUpdate := *openapiclient.NewUserUpdate() // UserUpdate | Updated user data.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.UpdateUser(context.Background(), id).UserUpdate(userUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.UpdateUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | User id. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **userUpdate** | [**UserUpdate**](UserUpdate.md) | Updated user data. | 

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


## UpdateUserSettings

> UpdateUserSettings(ctx).Body(body).IfMatch(ifMatch).Execute()

Set global user settings



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
	body := map[string]interface{}{ ... } // map[string]interface{} | New user settings.
	ifMatch := "ifMatch_example" // string | Contains the settings' current ETag, and performs the update only if it matches the one stored in the database. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UserAdministrationManagementAPIAPI.UpdateUserSettings(context.Background()).Body(body).IfMatch(ifMatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAdministrationManagementAPIAPI.UpdateUserSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateUserSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** | New user settings. | 
 **ifMatch** | **string** | Contains the settings&#39; current ETag, and performs the update only if it matches the one stored in the database. | 

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

