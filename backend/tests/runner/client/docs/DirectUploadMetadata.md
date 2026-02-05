# DirectUploadMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | Pointer to **int32** | wsize of the artifact file.  | [optional] 
**Updates** | Pointer to [**[]Update**](Update.md) | List of updates for this artifact. | [optional] 

## Methods

### NewDirectUploadMetadata

`func NewDirectUploadMetadata() *DirectUploadMetadata`

NewDirectUploadMetadata instantiates a new DirectUploadMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDirectUploadMetadataWithDefaults

`func NewDirectUploadMetadataWithDefaults() *DirectUploadMetadata`

NewDirectUploadMetadataWithDefaults instantiates a new DirectUploadMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *DirectUploadMetadata) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *DirectUploadMetadata) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *DirectUploadMetadata) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *DirectUploadMetadata) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetUpdates

`func (o *DirectUploadMetadata) GetUpdates() []Update`

GetUpdates returns the Updates field if non-nil, zero value otherwise.

### GetUpdatesOk

`func (o *DirectUploadMetadata) GetUpdatesOk() (*[]Update, bool)`

GetUpdatesOk returns a tuple with the Updates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdates

`func (o *DirectUploadMetadata) SetUpdates(v []Update)`

SetUpdates sets Updates field to given value.

### HasUpdates

`func (o *DirectUploadMetadata) HasUpdates() bool`

HasUpdates returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


