# ArtifactInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Format** | Pointer to **string** |  | [optional] 
**Version** | Pointer to **int32** |  | [optional] 

## Methods

### NewArtifactInfo

`func NewArtifactInfo() *ArtifactInfo`

NewArtifactInfo instantiates a new ArtifactInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewArtifactInfoWithDefaults

`func NewArtifactInfoWithDefaults() *ArtifactInfo`

NewArtifactInfoWithDefaults instantiates a new ArtifactInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFormat

`func (o *ArtifactInfo) GetFormat() string`

GetFormat returns the Format field if non-nil, zero value otherwise.

### GetFormatOk

`func (o *ArtifactInfo) GetFormatOk() (*string, bool)`

GetFormatOk returns a tuple with the Format field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFormat

`func (o *ArtifactInfo) SetFormat(v string)`

SetFormat sets Format field to given value.

### HasFormat

`func (o *ArtifactInfo) HasFormat() bool`

HasFormat returns a boolean if a field has been set.

### GetVersion

`func (o *ArtifactInfo) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ArtifactInfo) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ArtifactInfo) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *ArtifactInfo) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


