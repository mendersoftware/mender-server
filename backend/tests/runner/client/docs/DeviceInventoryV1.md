# DeviceInventoryV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Mender-assigned unique device ID. | [optional] 
**UpdatedTs** | Pointer to **string** | Timestamp of the most recent attribute update. | [optional] 
**Attributes** | Pointer to [**[]AttributeV1**](AttributeV1.md) | A list of attribute descriptors. | [optional] 

## Methods

### NewDeviceInventoryV1

`func NewDeviceInventoryV1() *DeviceInventoryV1`

NewDeviceInventoryV1 instantiates a new DeviceInventoryV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceInventoryV1WithDefaults

`func NewDeviceInventoryV1WithDefaults() *DeviceInventoryV1`

NewDeviceInventoryV1WithDefaults instantiates a new DeviceInventoryV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceInventoryV1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceInventoryV1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceInventoryV1) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceInventoryV1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *DeviceInventoryV1) GetUpdatedTs() string`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *DeviceInventoryV1) GetUpdatedTsOk() (*string, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *DeviceInventoryV1) SetUpdatedTs(v string)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *DeviceInventoryV1) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetAttributes

`func (o *DeviceInventoryV1) GetAttributes() []AttributeV1`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *DeviceInventoryV1) GetAttributesOk() (*[]AttributeV1, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *DeviceInventoryV1) SetAttributes(v []AttributeV1)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *DeviceInventoryV1) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


