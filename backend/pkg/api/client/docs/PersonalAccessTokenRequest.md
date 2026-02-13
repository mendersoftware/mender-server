# PersonalAccessTokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of a token. | 
**ExpiresIn** | Pointer to **int32** | Expiration time in seconds (maximum one year - 31536000s). If you omit it or set it to zero, the Personal Access Token will never expire.  | [optional] 

## Methods

### NewPersonalAccessTokenRequest

`func NewPersonalAccessTokenRequest(name string, ) *PersonalAccessTokenRequest`

NewPersonalAccessTokenRequest instantiates a new PersonalAccessTokenRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPersonalAccessTokenRequestWithDefaults

`func NewPersonalAccessTokenRequestWithDefaults() *PersonalAccessTokenRequest`

NewPersonalAccessTokenRequestWithDefaults instantiates a new PersonalAccessTokenRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PersonalAccessTokenRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PersonalAccessTokenRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PersonalAccessTokenRequest) SetName(v string)`

SetName sets Name field to given value.


### GetExpiresIn

`func (o *PersonalAccessTokenRequest) GetExpiresIn() int32`

GetExpiresIn returns the ExpiresIn field if non-nil, zero value otherwise.

### GetExpiresInOk

`func (o *PersonalAccessTokenRequest) GetExpiresInOk() (*int32, bool)`

GetExpiresInOk returns a tuple with the ExpiresIn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresIn

`func (o *PersonalAccessTokenRequest) SetExpiresIn(v int32)`

SetExpiresIn sets ExpiresIn field to given value.

### HasExpiresIn

`func (o *PersonalAccessTokenRequest) HasExpiresIn() bool`

HasExpiresIn returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


