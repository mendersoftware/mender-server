# DeviceNew

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Mender-assigned unique ID. | 
**UpdatedTs** | Pointer to **string** | Timestamp of the most recent attribute update. | [optional] 
**Attributes** | Pointer to [**[]Attribute**](Attribute.md) | A list of attribute descriptors. | [optional] 

## Methods

### NewDeviceNew

`func NewDeviceNew(id string, ) *DeviceNew`

NewDeviceNew instantiates a new DeviceNew object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceNewWithDefaults

`func NewDeviceNewWithDefaults() *DeviceNew`

NewDeviceNewWithDefaults instantiates a new DeviceNew object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceNew) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceNew) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceNew) SetId(v string)`

SetId sets Id field to given value.


### GetUpdatedTs

`func (o *DeviceNew) GetUpdatedTs() string`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *DeviceNew) GetUpdatedTsOk() (*string, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *DeviceNew) SetUpdatedTs(v string)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *DeviceNew) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetAttributes

`func (o *DeviceNew) GetAttributes() []Attribute`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *DeviceNew) GetAttributesOk() (*[]Attribute, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *DeviceNew) SetAttributes(v []Attribute)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *DeviceNew) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


