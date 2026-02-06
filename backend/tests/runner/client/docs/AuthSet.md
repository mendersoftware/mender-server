# AuthSet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The unique ID of the authentication set. | [optional] 
**DeviceId** | Pointer to **string** | The unique ID of the device the authentication set belongs. | [optional] 
**IdentityData** | Pointer to **map[string]interface{}** | The device&#39;s identity data. | [optional] 
**Pubkey** | Pointer to **string** | PEM-encoded public key of the device authentication set. | [optional] 
**Status** | Pointer to **string** | Authorization status of the set. | [optional] 
**Ts** | Pointer to **time.Time** | The creation timestamp of the authentication set. | [optional] 

## Methods

### NewAuthSet

`func NewAuthSet() *AuthSet`

NewAuthSet instantiates a new AuthSet object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthSetWithDefaults

`func NewAuthSetWithDefaults() *AuthSet`

NewAuthSetWithDefaults instantiates a new AuthSet object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AuthSet) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AuthSet) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AuthSet) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AuthSet) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDeviceId

`func (o *AuthSet) GetDeviceId() string`

GetDeviceId returns the DeviceId field if non-nil, zero value otherwise.

### GetDeviceIdOk

`func (o *AuthSet) GetDeviceIdOk() (*string, bool)`

GetDeviceIdOk returns a tuple with the DeviceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceId

`func (o *AuthSet) SetDeviceId(v string)`

SetDeviceId sets DeviceId field to given value.

### HasDeviceId

`func (o *AuthSet) HasDeviceId() bool`

HasDeviceId returns a boolean if a field has been set.

### GetIdentityData

`func (o *AuthSet) GetIdentityData() map[string]interface{}`

GetIdentityData returns the IdentityData field if non-nil, zero value otherwise.

### GetIdentityDataOk

`func (o *AuthSet) GetIdentityDataOk() (*map[string]interface{}, bool)`

GetIdentityDataOk returns a tuple with the IdentityData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentityData

`func (o *AuthSet) SetIdentityData(v map[string]interface{})`

SetIdentityData sets IdentityData field to given value.

### HasIdentityData

`func (o *AuthSet) HasIdentityData() bool`

HasIdentityData returns a boolean if a field has been set.

### GetPubkey

`func (o *AuthSet) GetPubkey() string`

GetPubkey returns the Pubkey field if non-nil, zero value otherwise.

### GetPubkeyOk

`func (o *AuthSet) GetPubkeyOk() (*string, bool)`

GetPubkeyOk returns a tuple with the Pubkey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPubkey

`func (o *AuthSet) SetPubkey(v string)`

SetPubkey sets Pubkey field to given value.

### HasPubkey

`func (o *AuthSet) HasPubkey() bool`

HasPubkey returns a boolean if a field has been set.

### GetStatus

`func (o *AuthSet) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *AuthSet) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *AuthSet) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *AuthSet) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetTs

`func (o *AuthSet) GetTs() time.Time`

GetTs returns the Ts field if non-nil, zero value otherwise.

### GetTsOk

`func (o *AuthSet) GetTsOk() (*time.Time, bool)`

GetTsOk returns a tuple with the Ts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTs

`func (o *AuthSet) SetTs(v time.Time)`

SetTs sets Ts field to given value.

### HasTs

`func (o *AuthSet) HasTs() bool`

HasTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


