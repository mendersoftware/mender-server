# User

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | A unique email address. | 
**Id** | **string** | User Id. | 
**CreatedTs** | Pointer to **time.Time** | Server-side timestamp of the user creation.  | [optional] 
**UpdatedTs** | Pointer to **time.Time** | Server-side timestamp of the last user information update.  | [optional] 
**LoginTs** | Pointer to **time.Time** | Timestamp of last successful login. | [optional] 

## Methods

### NewUser

`func NewUser(email string, id string, ) *User`

NewUser instantiates a new User object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserWithDefaults

`func NewUserWithDefaults() *User`

NewUserWithDefaults instantiates a new User object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *User) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *User) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *User) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetId

`func (o *User) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *User) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *User) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedTs

`func (o *User) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *User) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *User) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.

### HasCreatedTs

`func (o *User) HasCreatedTs() bool`

HasCreatedTs returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *User) GetUpdatedTs() time.Time`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *User) GetUpdatedTsOk() (*time.Time, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *User) SetUpdatedTs(v time.Time)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *User) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetLoginTs

`func (o *User) GetLoginTs() time.Time`

GetLoginTs returns the LoginTs field if non-nil, zero value otherwise.

### GetLoginTsOk

`func (o *User) GetLoginTsOk() (*time.Time, bool)`

GetLoginTsOk returns a tuple with the LoginTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLoginTs

`func (o *User) SetLoginTs(v time.Time)`

SetLoginTs sets LoginTs field to given value.

### HasLoginTs

`func (o *User) HasLoginTs() bool`

HasLoginTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


