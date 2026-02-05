# DeviceUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Device identifier. | 
**Revision** | **int32** | Device object revision. | 

## Methods

### NewDeviceUpdate

`func NewDeviceUpdate(id string, revision int32, ) *DeviceUpdate`

NewDeviceUpdate instantiates a new DeviceUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceUpdateWithDefaults

`func NewDeviceUpdateWithDefaults() *DeviceUpdate`

NewDeviceUpdateWithDefaults instantiates a new DeviceUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceUpdate) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceUpdate) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceUpdate) SetId(v string)`

SetId sets Id field to given value.


### GetRevision

`func (o *DeviceUpdate) GetRevision() int32`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *DeviceUpdate) GetRevisionOk() (*int32, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *DeviceUpdate) SetRevision(v int32)`

SetRevision sets Revision field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


