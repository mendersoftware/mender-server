# UpdateFile

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Checksum** | Pointer to **string** |  | [optional] 
**Size** | Pointer to **int32** |  | [optional] 
**Date** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewUpdateFile

`func NewUpdateFile() *UpdateFile`

NewUpdateFile instantiates a new UpdateFile object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateFileWithDefaults

`func NewUpdateFileWithDefaults() *UpdateFile`

NewUpdateFileWithDefaults instantiates a new UpdateFile object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *UpdateFile) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateFile) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateFile) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UpdateFile) HasName() bool`

HasName returns a boolean if a field has been set.

### GetChecksum

`func (o *UpdateFile) GetChecksum() string`

GetChecksum returns the Checksum field if non-nil, zero value otherwise.

### GetChecksumOk

`func (o *UpdateFile) GetChecksumOk() (*string, bool)`

GetChecksumOk returns a tuple with the Checksum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChecksum

`func (o *UpdateFile) SetChecksum(v string)`

SetChecksum sets Checksum field to given value.

### HasChecksum

`func (o *UpdateFile) HasChecksum() bool`

HasChecksum returns a boolean if a field has been set.

### GetSize

`func (o *UpdateFile) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *UpdateFile) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *UpdateFile) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *UpdateFile) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetDate

`func (o *UpdateFile) GetDate() time.Time`

GetDate returns the Date field if non-nil, zero value otherwise.

### GetDateOk

`func (o *UpdateFile) GetDateOk() (*time.Time, bool)`

GetDateOk returns a tuple with the Date field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDate

`func (o *UpdateFile) SetDate(v time.Time)`

SetDate sets Date field to given value.

### HasDate

`func (o *UpdateFile) HasDate() bool`

HasDate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


