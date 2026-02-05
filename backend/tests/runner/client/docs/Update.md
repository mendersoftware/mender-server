# Update

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TypeInfo** | Pointer to [**ArtifactTypeInfo**](ArtifactTypeInfo.md) |  | [optional] 
**Files** | Pointer to [**[]UpdateFile**](UpdateFile.md) |  | [optional] 
**Metadata** | Pointer to **map[string]string** | metadata is an object of unknown structure as this is dependent of update type (also custom defined by user) | [optional] 
**MetaData** | Pointer to **[]map[string]interface{}** | Deprecated: Please use &#x60;metadata&#x60; instead. A list of objects of unknown structure as this is dependent of update type (also custom defined by user)  | [optional] 

## Methods

### NewUpdate

`func NewUpdate() *Update`

NewUpdate instantiates a new Update object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateWithDefaults

`func NewUpdateWithDefaults() *Update`

NewUpdateWithDefaults instantiates a new Update object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTypeInfo

`func (o *Update) GetTypeInfo() ArtifactTypeInfo`

GetTypeInfo returns the TypeInfo field if non-nil, zero value otherwise.

### GetTypeInfoOk

`func (o *Update) GetTypeInfoOk() (*ArtifactTypeInfo, bool)`

GetTypeInfoOk returns a tuple with the TypeInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeInfo

`func (o *Update) SetTypeInfo(v ArtifactTypeInfo)`

SetTypeInfo sets TypeInfo field to given value.

### HasTypeInfo

`func (o *Update) HasTypeInfo() bool`

HasTypeInfo returns a boolean if a field has been set.

### GetFiles

`func (o *Update) GetFiles() []UpdateFile`

GetFiles returns the Files field if non-nil, zero value otherwise.

### GetFilesOk

`func (o *Update) GetFilesOk() (*[]UpdateFile, bool)`

GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFiles

`func (o *Update) SetFiles(v []UpdateFile)`

SetFiles sets Files field to given value.

### HasFiles

`func (o *Update) HasFiles() bool`

HasFiles returns a boolean if a field has been set.

### GetMetadata

`func (o *Update) GetMetadata() map[string]string`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Update) GetMetadataOk() (*map[string]string, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Update) SetMetadata(v map[string]string)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *Update) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetMetaData

`func (o *Update) GetMetaData() []map[string]interface{}`

GetMetaData returns the MetaData field if non-nil, zero value otherwise.

### GetMetaDataOk

`func (o *Update) GetMetaDataOk() (*[]map[string]interface{}, bool)`

GetMetaDataOk returns a tuple with the MetaData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetaData

`func (o *Update) SetMetaData(v []map[string]interface{})`

SetMetaData sets MetaData field to given value.

### HasMetaData

`func (o *Update) HasMetaData() bool`

HasMetaData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


