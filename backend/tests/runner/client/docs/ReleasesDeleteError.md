# ReleasesDeleteError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to **string** | Description of the error. | [optional] 
**ActiveDeployments** | Pointer to **[]string** | List of IDs of active deployments which are using releases from the request. | [optional] 
**RequestId** | Pointer to **string** | Request ID (same as in X-MEN-RequestID header). | [optional] 

## Methods

### NewReleasesDeleteError

`func NewReleasesDeleteError() *ReleasesDeleteError`

NewReleasesDeleteError instantiates a new ReleasesDeleteError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleasesDeleteErrorWithDefaults

`func NewReleasesDeleteErrorWithDefaults() *ReleasesDeleteError`

NewReleasesDeleteErrorWithDefaults instantiates a new ReleasesDeleteError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *ReleasesDeleteError) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *ReleasesDeleteError) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *ReleasesDeleteError) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *ReleasesDeleteError) HasError() bool`

HasError returns a boolean if a field has been set.

### GetActiveDeployments

`func (o *ReleasesDeleteError) GetActiveDeployments() []string`

GetActiveDeployments returns the ActiveDeployments field if non-nil, zero value otherwise.

### GetActiveDeploymentsOk

`func (o *ReleasesDeleteError) GetActiveDeploymentsOk() (*[]string, bool)`

GetActiveDeploymentsOk returns a tuple with the ActiveDeployments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActiveDeployments

`func (o *ReleasesDeleteError) SetActiveDeployments(v []string)`

SetActiveDeployments sets ActiveDeployments field to given value.

### HasActiveDeployments

`func (o *ReleasesDeleteError) HasActiveDeployments() bool`

HasActiveDeployments returns a boolean if a field has been set.

### GetRequestId

`func (o *ReleasesDeleteError) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *ReleasesDeleteError) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *ReleasesDeleteError) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.

### HasRequestId

`func (o *ReleasesDeleteError) HasRequestId() bool`

HasRequestId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


