# DeviceWithImageImage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Image ID | [optional] 
**Meta** | Pointer to [**DeviceWithImageImageMeta**](DeviceWithImageImageMeta.md) |  | [optional] 
**MetaArtifact** | Pointer to [**DeviceWithImageImageMetaArtifact**](DeviceWithImageImageMetaArtifact.md) |  | [optional] 
**Size** | Pointer to **int32** | Image size in bytes | [optional] 
**Modified** | Pointer to **time.Time** | Creation / last edition of any of the artifact properties | [optional] 

## Methods

### NewDeviceWithImageImage

`func NewDeviceWithImageImage() *DeviceWithImageImage`

NewDeviceWithImageImage instantiates a new DeviceWithImageImage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithImageImageWithDefaults

`func NewDeviceWithImageImageWithDefaults() *DeviceWithImageImage`

NewDeviceWithImageImageWithDefaults instantiates a new DeviceWithImageImage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceWithImageImage) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceWithImageImage) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceWithImageImage) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceWithImageImage) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMeta

`func (o *DeviceWithImageImage) GetMeta() DeviceWithImageImageMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DeviceWithImageImage) GetMetaOk() (*DeviceWithImageImageMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DeviceWithImageImage) SetMeta(v DeviceWithImageImageMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *DeviceWithImageImage) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetMetaArtifact

`func (o *DeviceWithImageImage) GetMetaArtifact() DeviceWithImageImageMetaArtifact`

GetMetaArtifact returns the MetaArtifact field if non-nil, zero value otherwise.

### GetMetaArtifactOk

`func (o *DeviceWithImageImage) GetMetaArtifactOk() (*DeviceWithImageImageMetaArtifact, bool)`

GetMetaArtifactOk returns a tuple with the MetaArtifact field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetaArtifact

`func (o *DeviceWithImageImage) SetMetaArtifact(v DeviceWithImageImageMetaArtifact)`

SetMetaArtifact sets MetaArtifact field to given value.

### HasMetaArtifact

`func (o *DeviceWithImageImage) HasMetaArtifact() bool`

HasMetaArtifact returns a boolean if a field has been set.

### GetSize

`func (o *DeviceWithImageImage) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *DeviceWithImageImage) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *DeviceWithImageImage) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *DeviceWithImageImage) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetModified

`func (o *DeviceWithImageImage) GetModified() time.Time`

GetModified returns the Modified field if non-nil, zero value otherwise.

### GetModifiedOk

`func (o *DeviceWithImageImage) GetModifiedOk() (*time.Time, bool)`

GetModifiedOk returns a tuple with the Modified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModified

`func (o *DeviceWithImageImage) SetModified(v time.Time)`

SetModified sets Modified field to given value.

### HasModified

`func (o *DeviceWithImageImage) HasModified() bool`

HasModified returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


