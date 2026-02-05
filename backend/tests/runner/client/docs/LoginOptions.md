# LoginOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NoExpiry** | Pointer to **bool** | Generate a JWT token with no expiration date. | [optional] 

## Methods

### NewLoginOptions

`func NewLoginOptions() *LoginOptions`

NewLoginOptions instantiates a new LoginOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoginOptionsWithDefaults

`func NewLoginOptionsWithDefaults() *LoginOptions`

NewLoginOptionsWithDefaults instantiates a new LoginOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNoExpiry

`func (o *LoginOptions) GetNoExpiry() bool`

GetNoExpiry returns the NoExpiry field if non-nil, zero value otherwise.

### GetNoExpiryOk

`func (o *LoginOptions) GetNoExpiryOk() (*bool, bool)`

GetNoExpiryOk returns a tuple with the NoExpiry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoExpiry

`func (o *LoginOptions) SetNoExpiry(v bool)`

SetNoExpiry sets NoExpiry field to given value.

### HasNoExpiry

`func (o *LoginOptions) HasNoExpiry() bool`

HasNoExpiry returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


