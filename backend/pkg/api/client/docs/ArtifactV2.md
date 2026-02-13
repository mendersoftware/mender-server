# ArtifactV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Description** | **string** |  | 
**DeviceTypesCompatible** | **[]string** | An array of compatible device types. | 
**Info** | Pointer to [**ArtifactInfo**](ArtifactInfo.md) |  | [optional] 
**Signed** | Pointer to **bool** | Idicates if artifact is signed or not. | [optional] 
**Updates** | Pointer to [**[]Update**](Update.md) |  | [optional] 
**ArtifactProvides** | Pointer to **map[string]string** | List of Artifact provides.  Map of key/value pairs, where both keys and values are strings.  | [optional] 
**ArtifactDepends** | Pointer to **map[string][]string** | List of Artifact depends.  Map of key/value pairs, where keys are strings and values are lists of strings.  | [optional] 
**ClearsArtifactProvides** | Pointer to **[]string** | List of Clear Artifact provides. | [optional] 
**Size** | Pointer to **int32** | Artifact total size in bytes - the size of the actual file that will be transferred to the device (compressed).  | [optional] 
**Modified** | **time.Time** | Represents creation / last edition of any of the artifact properties.  | 

## Methods

### NewArtifactV2

`func NewArtifactV2(id string, name string, description string, deviceTypesCompatible []string, modified time.Time, ) *ArtifactV2`

NewArtifactV2 instantiates a new ArtifactV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewArtifactV2WithDefaults

`func NewArtifactV2WithDefaults() *ArtifactV2`

NewArtifactV2WithDefaults instantiates a new ArtifactV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ArtifactV2) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ArtifactV2) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ArtifactV2) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ArtifactV2) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ArtifactV2) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ArtifactV2) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *ArtifactV2) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ArtifactV2) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ArtifactV2) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetDeviceTypesCompatible

`func (o *ArtifactV2) GetDeviceTypesCompatible() []string`

GetDeviceTypesCompatible returns the DeviceTypesCompatible field if non-nil, zero value otherwise.

### GetDeviceTypesCompatibleOk

`func (o *ArtifactV2) GetDeviceTypesCompatibleOk() (*[]string, bool)`

GetDeviceTypesCompatibleOk returns a tuple with the DeviceTypesCompatible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceTypesCompatible

`func (o *ArtifactV2) SetDeviceTypesCompatible(v []string)`

SetDeviceTypesCompatible sets DeviceTypesCompatible field to given value.


### GetInfo

`func (o *ArtifactV2) GetInfo() ArtifactInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *ArtifactV2) GetInfoOk() (*ArtifactInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *ArtifactV2) SetInfo(v ArtifactInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *ArtifactV2) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetSigned

`func (o *ArtifactV2) GetSigned() bool`

GetSigned returns the Signed field if non-nil, zero value otherwise.

### GetSignedOk

`func (o *ArtifactV2) GetSignedOk() (*bool, bool)`

GetSignedOk returns a tuple with the Signed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSigned

`func (o *ArtifactV2) SetSigned(v bool)`

SetSigned sets Signed field to given value.

### HasSigned

`func (o *ArtifactV2) HasSigned() bool`

HasSigned returns a boolean if a field has been set.

### GetUpdates

`func (o *ArtifactV2) GetUpdates() []Update`

GetUpdates returns the Updates field if non-nil, zero value otherwise.

### GetUpdatesOk

`func (o *ArtifactV2) GetUpdatesOk() (*[]Update, bool)`

GetUpdatesOk returns a tuple with the Updates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdates

`func (o *ArtifactV2) SetUpdates(v []Update)`

SetUpdates sets Updates field to given value.

### HasUpdates

`func (o *ArtifactV2) HasUpdates() bool`

HasUpdates returns a boolean if a field has been set.

### GetArtifactProvides

`func (o *ArtifactV2) GetArtifactProvides() map[string]string`

GetArtifactProvides returns the ArtifactProvides field if non-nil, zero value otherwise.

### GetArtifactProvidesOk

`func (o *ArtifactV2) GetArtifactProvidesOk() (*map[string]string, bool)`

GetArtifactProvidesOk returns a tuple with the ArtifactProvides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactProvides

`func (o *ArtifactV2) SetArtifactProvides(v map[string]string)`

SetArtifactProvides sets ArtifactProvides field to given value.

### HasArtifactProvides

`func (o *ArtifactV2) HasArtifactProvides() bool`

HasArtifactProvides returns a boolean if a field has been set.

### GetArtifactDepends

`func (o *ArtifactV2) GetArtifactDepends() map[string][]string`

GetArtifactDepends returns the ArtifactDepends field if non-nil, zero value otherwise.

### GetArtifactDependsOk

`func (o *ArtifactV2) GetArtifactDependsOk() (*map[string][]string, bool)`

GetArtifactDependsOk returns a tuple with the ArtifactDepends field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactDepends

`func (o *ArtifactV2) SetArtifactDepends(v map[string][]string)`

SetArtifactDepends sets ArtifactDepends field to given value.

### HasArtifactDepends

`func (o *ArtifactV2) HasArtifactDepends() bool`

HasArtifactDepends returns a boolean if a field has been set.

### GetClearsArtifactProvides

`func (o *ArtifactV2) GetClearsArtifactProvides() []string`

GetClearsArtifactProvides returns the ClearsArtifactProvides field if non-nil, zero value otherwise.

### GetClearsArtifactProvidesOk

`func (o *ArtifactV2) GetClearsArtifactProvidesOk() (*[]string, bool)`

GetClearsArtifactProvidesOk returns a tuple with the ClearsArtifactProvides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClearsArtifactProvides

`func (o *ArtifactV2) SetClearsArtifactProvides(v []string)`

SetClearsArtifactProvides sets ClearsArtifactProvides field to given value.

### HasClearsArtifactProvides

`func (o *ArtifactV2) HasClearsArtifactProvides() bool`

HasClearsArtifactProvides returns a boolean if a field has been set.

### GetSize

`func (o *ArtifactV2) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *ArtifactV2) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *ArtifactV2) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *ArtifactV2) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetModified

`func (o *ArtifactV2) GetModified() time.Time`

GetModified returns the Modified field if non-nil, zero value otherwise.

### GetModifiedOk

`func (o *ArtifactV2) GetModifiedOk() (*time.Time, bool)`

GetModifiedOk returns a tuple with the Modified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModified

`func (o *ArtifactV2) SetModified(v time.Time)`

SetModified sets Modified field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


