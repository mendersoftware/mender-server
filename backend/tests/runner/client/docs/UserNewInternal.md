# UserNewInternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | A unique email address. Non-ascii characters are invalid. | 
**Password** | **string** | Password. | 
**Propagate** | Pointer to **bool** | This paramter is deprecated _since Thu Jul 6 2023_, the propagation of user information to tenantadm is disabled permanently.  | [optional] 

## Methods

### NewUserNewInternal

`func NewUserNewInternal(email string, password string, ) *UserNewInternal`

NewUserNewInternal instantiates a new UserNewInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserNewInternalWithDefaults

`func NewUserNewInternalWithDefaults() *UserNewInternal`

NewUserNewInternalWithDefaults instantiates a new UserNewInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserNewInternal) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserNewInternal) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserNewInternal) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetPassword

`func (o *UserNewInternal) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserNewInternal) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserNewInternal) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetPropagate

`func (o *UserNewInternal) GetPropagate() bool`

GetPropagate returns the Propagate field if non-nil, zero value otherwise.

### GetPropagateOk

`func (o *UserNewInternal) GetPropagateOk() (*bool, bool)`

GetPropagateOk returns a tuple with the Propagate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPropagate

`func (o *UserNewInternal) SetPropagate(v bool)`

SetPropagate sets Propagate field to given value.

### HasPropagate

`func (o *UserNewInternal) HasPropagate() bool`

HasPropagate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


