# NewDevice

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | ID of the new device. | 
**Status** | Pointer to **string** | Authorization status for the device. | [optional] 
**AuthSets** | Pointer to [**[]AuthSet**](AuthSet.md) |  | [optional] 
**CreatedTs** | Pointer to **time.Time** | The creation timestamp of the device. | [optional] 

## Methods

### NewNewDevice

`func NewNewDevice(id string, ) *NewDevice`

NewNewDevice instantiates a new NewDevice object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewDeviceWithDefaults

`func NewNewDeviceWithDefaults() *NewDevice`

NewNewDeviceWithDefaults instantiates a new NewDevice object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NewDevice) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NewDevice) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NewDevice) SetId(v string)`

SetId sets Id field to given value.


### GetStatus

`func (o *NewDevice) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NewDevice) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NewDevice) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NewDevice) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetAuthSets

`func (o *NewDevice) GetAuthSets() []AuthSet`

GetAuthSets returns the AuthSets field if non-nil, zero value otherwise.

### GetAuthSetsOk

`func (o *NewDevice) GetAuthSetsOk() (*[]AuthSet, bool)`

GetAuthSetsOk returns a tuple with the AuthSets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthSets

`func (o *NewDevice) SetAuthSets(v []AuthSet)`

SetAuthSets sets AuthSets field to given value.

### HasAuthSets

`func (o *NewDevice) HasAuthSets() bool`

HasAuthSets returns a boolean if a field has been set.

### GetCreatedTs

`func (o *NewDevice) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *NewDevice) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *NewDevice) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.

### HasCreatedTs

`func (o *NewDevice) HasCreatedTs() bool`

HasCreatedTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


