# UserUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to **string** | A unique email address. | [optional] 
**Password** | Pointer to **string** | New password. | [optional] 
**CurrentPassword** | Pointer to **string** | Current password. | [optional] 

## Methods

### NewUserUpdate

`func NewUserUpdate() *UserUpdate`

NewUserUpdate instantiates a new UserUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserUpdateWithDefaults

`func NewUserUpdateWithDefaults() *UserUpdate`

NewUserUpdateWithDefaults instantiates a new UserUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserUpdate) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserUpdate) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserUpdate) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *UserUpdate) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetPassword

`func (o *UserUpdate) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserUpdate) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserUpdate) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *UserUpdate) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetCurrentPassword

`func (o *UserUpdate) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *UserUpdate) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *UserUpdate) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *UserUpdate) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


