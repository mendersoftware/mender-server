# DeviceAuthEvent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Device unique ID. | 
**Status** | Pointer to **string** | The authentication status of the device. | [optional] 
**AuthSets** | Pointer to [**[]AuthSet**](AuthSet.md) |  | [optional] 
**CreatedTs** | Pointer to **time.Time** | The time the device was initialized in Mender. | [optional] 

## Methods

### NewDeviceAuthEvent

`func NewDeviceAuthEvent(id string, ) *DeviceAuthEvent`

NewDeviceAuthEvent instantiates a new DeviceAuthEvent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceAuthEventWithDefaults

`func NewDeviceAuthEventWithDefaults() *DeviceAuthEvent`

NewDeviceAuthEventWithDefaults instantiates a new DeviceAuthEvent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceAuthEvent) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceAuthEvent) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceAuthEvent) SetId(v string)`

SetId sets Id field to given value.


### GetStatus

`func (o *DeviceAuthEvent) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeviceAuthEvent) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeviceAuthEvent) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DeviceAuthEvent) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetAuthSets

`func (o *DeviceAuthEvent) GetAuthSets() []AuthSet`

GetAuthSets returns the AuthSets field if non-nil, zero value otherwise.

### GetAuthSetsOk

`func (o *DeviceAuthEvent) GetAuthSetsOk() (*[]AuthSet, bool)`

GetAuthSetsOk returns a tuple with the AuthSets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthSets

`func (o *DeviceAuthEvent) SetAuthSets(v []AuthSet)`

SetAuthSets sets AuthSets field to given value.

### HasAuthSets

`func (o *DeviceAuthEvent) HasAuthSets() bool`

HasAuthSets returns a boolean if a field has been set.

### GetCreatedTs

`func (o *DeviceAuthEvent) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *DeviceAuthEvent) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *DeviceAuthEvent) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.

### HasCreatedTs

`func (o *DeviceAuthEvent) HasCreatedTs() bool`

HasCreatedTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


