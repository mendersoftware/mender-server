# DeviceWithImageImageMetaArtifact

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**DeviceTypesCompatible** | Pointer to **[]string** | An array of compatible device types. | [optional] 
**Info** | Pointer to [**ArtifactInfo**](ArtifactInfo.md) |  | [optional] 
**Signed** | Pointer to **bool** | Idicates if artifact is signed or not. | [optional] 
**Updates** | Pointer to [**[]Update**](Update.md) |  | [optional] 
**ArtifactProvides** | Pointer to **map[string]string** | List of Artifact provides.  Map of key/value pairs, where both keys and values are strings.  | [optional] 
**ArtifactDepends** | Pointer to **map[string][]string** | List of Artifact depends.  Map of key/value pairs, where keys are strings and values are lists of strings.  | [optional] 
**ClearsArtifactProvides** | Pointer to **[]string** | List of Clear Artifact provides. | [optional] 

## Methods

### NewDeviceWithImageImageMetaArtifact

`func NewDeviceWithImageImageMetaArtifact() *DeviceWithImageImageMetaArtifact`

NewDeviceWithImageImageMetaArtifact instantiates a new DeviceWithImageImageMetaArtifact object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithImageImageMetaArtifactWithDefaults

`func NewDeviceWithImageImageMetaArtifactWithDefaults() *DeviceWithImageImageMetaArtifact`

NewDeviceWithImageImageMetaArtifactWithDefaults instantiates a new DeviceWithImageImageMetaArtifact object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DeviceWithImageImageMetaArtifact) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DeviceWithImageImageMetaArtifact) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DeviceWithImageImageMetaArtifact) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DeviceWithImageImageMetaArtifact) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeviceTypesCompatible

`func (o *DeviceWithImageImageMetaArtifact) GetDeviceTypesCompatible() []string`

GetDeviceTypesCompatible returns the DeviceTypesCompatible field if non-nil, zero value otherwise.

### GetDeviceTypesCompatibleOk

`func (o *DeviceWithImageImageMetaArtifact) GetDeviceTypesCompatibleOk() (*[]string, bool)`

GetDeviceTypesCompatibleOk returns a tuple with the DeviceTypesCompatible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceTypesCompatible

`func (o *DeviceWithImageImageMetaArtifact) SetDeviceTypesCompatible(v []string)`

SetDeviceTypesCompatible sets DeviceTypesCompatible field to given value.

### HasDeviceTypesCompatible

`func (o *DeviceWithImageImageMetaArtifact) HasDeviceTypesCompatible() bool`

HasDeviceTypesCompatible returns a boolean if a field has been set.

### GetInfo

`func (o *DeviceWithImageImageMetaArtifact) GetInfo() ArtifactInfo`

GetInfo returns the Info field if non-nil, zero value otherwise.

### GetInfoOk

`func (o *DeviceWithImageImageMetaArtifact) GetInfoOk() (*ArtifactInfo, bool)`

GetInfoOk returns a tuple with the Info field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInfo

`func (o *DeviceWithImageImageMetaArtifact) SetInfo(v ArtifactInfo)`

SetInfo sets Info field to given value.

### HasInfo

`func (o *DeviceWithImageImageMetaArtifact) HasInfo() bool`

HasInfo returns a boolean if a field has been set.

### GetSigned

`func (o *DeviceWithImageImageMetaArtifact) GetSigned() bool`

GetSigned returns the Signed field if non-nil, zero value otherwise.

### GetSignedOk

`func (o *DeviceWithImageImageMetaArtifact) GetSignedOk() (*bool, bool)`

GetSignedOk returns a tuple with the Signed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSigned

`func (o *DeviceWithImageImageMetaArtifact) SetSigned(v bool)`

SetSigned sets Signed field to given value.

### HasSigned

`func (o *DeviceWithImageImageMetaArtifact) HasSigned() bool`

HasSigned returns a boolean if a field has been set.

### GetUpdates

`func (o *DeviceWithImageImageMetaArtifact) GetUpdates() []Update`

GetUpdates returns the Updates field if non-nil, zero value otherwise.

### GetUpdatesOk

`func (o *DeviceWithImageImageMetaArtifact) GetUpdatesOk() (*[]Update, bool)`

GetUpdatesOk returns a tuple with the Updates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdates

`func (o *DeviceWithImageImageMetaArtifact) SetUpdates(v []Update)`

SetUpdates sets Updates field to given value.

### HasUpdates

`func (o *DeviceWithImageImageMetaArtifact) HasUpdates() bool`

HasUpdates returns a boolean if a field has been set.

### GetArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) GetArtifactProvides() map[string]string`

GetArtifactProvides returns the ArtifactProvides field if non-nil, zero value otherwise.

### GetArtifactProvidesOk

`func (o *DeviceWithImageImageMetaArtifact) GetArtifactProvidesOk() (*map[string]string, bool)`

GetArtifactProvidesOk returns a tuple with the ArtifactProvides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) SetArtifactProvides(v map[string]string)`

SetArtifactProvides sets ArtifactProvides field to given value.

### HasArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) HasArtifactProvides() bool`

HasArtifactProvides returns a boolean if a field has been set.

### GetArtifactDepends

`func (o *DeviceWithImageImageMetaArtifact) GetArtifactDepends() map[string][]string`

GetArtifactDepends returns the ArtifactDepends field if non-nil, zero value otherwise.

### GetArtifactDependsOk

`func (o *DeviceWithImageImageMetaArtifact) GetArtifactDependsOk() (*map[string][]string, bool)`

GetArtifactDependsOk returns a tuple with the ArtifactDepends field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactDepends

`func (o *DeviceWithImageImageMetaArtifact) SetArtifactDepends(v map[string][]string)`

SetArtifactDepends sets ArtifactDepends field to given value.

### HasArtifactDepends

`func (o *DeviceWithImageImageMetaArtifact) HasArtifactDepends() bool`

HasArtifactDepends returns a boolean if a field has been set.

### GetClearsArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) GetClearsArtifactProvides() []string`

GetClearsArtifactProvides returns the ClearsArtifactProvides field if non-nil, zero value otherwise.

### GetClearsArtifactProvidesOk

`func (o *DeviceWithImageImageMetaArtifact) GetClearsArtifactProvidesOk() (*[]string, bool)`

GetClearsArtifactProvidesOk returns a tuple with the ClearsArtifactProvides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClearsArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) SetClearsArtifactProvides(v []string)`

SetClearsArtifactProvides sets ClearsArtifactProvides field to given value.

### HasClearsArtifactProvides

`func (o *DeviceWithImageImageMetaArtifact) HasClearsArtifactProvides() bool`

HasClearsArtifactProvides returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


