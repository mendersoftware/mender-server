# DeviceInventory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Mender-assigned unique ID. | [optional] 
**UpdatedTs** | Pointer to **string** | Timestamp of the most recent attribute update. | [optional] 
**Attributes** | Pointer to [**[]AttributeV2**](AttributeV2.md) | A list of attribute descriptors. | [optional] 

## Methods

### NewDeviceInventory

`func NewDeviceInventory() *DeviceInventory`

NewDeviceInventory instantiates a new DeviceInventory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceInventoryWithDefaults

`func NewDeviceInventoryWithDefaults() *DeviceInventory`

NewDeviceInventoryWithDefaults instantiates a new DeviceInventory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceInventory) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceInventory) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceInventory) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceInventory) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *DeviceInventory) GetUpdatedTs() string`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *DeviceInventory) GetUpdatedTsOk() (*string, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *DeviceInventory) SetUpdatedTs(v string)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *DeviceInventory) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetAttributes

`func (o *DeviceInventory) GetAttributes() []AttributeV2`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *DeviceInventory) GetAttributesOk() (*[]AttributeV2, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *DeviceInventory) SetAttributes(v []AttributeV2)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *DeviceInventory) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


