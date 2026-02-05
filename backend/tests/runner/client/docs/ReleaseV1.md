# ReleaseV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | release name.  | [optional] 
**Modified** | Pointer to **time.Time** | Last modification time for the release.  | [optional] 
**Artifacts** | Pointer to [**[]ArtifactV1**](ArtifactV1.md) | List of artifacts for this release. | [optional] 
**ArtifactsCount** | Pointer to **int32** | Number of artifacts. | [optional] 
**Tags** | Pointer to **[]string** | Tags assigned to the release used for filtering releases. Each tag must be valid a ASCII string and contain only lowercase and uppercase letters, digits, underscores, periods and hyphens. | [optional] 
**Notes** | Pointer to **string** | Additional information describing a Release limited to 1024 characters. Please use the v2 API to set this field.  | [optional] 

## Methods

### NewReleaseV1

`func NewReleaseV1() *ReleaseV1`

NewReleaseV1 instantiates a new ReleaseV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseV1WithDefaults

`func NewReleaseV1WithDefaults() *ReleaseV1`

NewReleaseV1WithDefaults instantiates a new ReleaseV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ReleaseV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ReleaseV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ReleaseV1) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ReleaseV1) HasName() bool`

HasName returns a boolean if a field has been set.

### GetModified

`func (o *ReleaseV1) GetModified() time.Time`

GetModified returns the Modified field if non-nil, zero value otherwise.

### GetModifiedOk

`func (o *ReleaseV1) GetModifiedOk() (*time.Time, bool)`

GetModifiedOk returns a tuple with the Modified field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModified

`func (o *ReleaseV1) SetModified(v time.Time)`

SetModified sets Modified field to given value.

### HasModified

`func (o *ReleaseV1) HasModified() bool`

HasModified returns a boolean if a field has been set.

### GetArtifacts

`func (o *ReleaseV1) GetArtifacts() []ArtifactV1`

GetArtifacts returns the Artifacts field if non-nil, zero value otherwise.

### GetArtifactsOk

`func (o *ReleaseV1) GetArtifactsOk() (*[]ArtifactV1, bool)`

GetArtifactsOk returns a tuple with the Artifacts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifacts

`func (o *ReleaseV1) SetArtifacts(v []ArtifactV1)`

SetArtifacts sets Artifacts field to given value.

### HasArtifacts

`func (o *ReleaseV1) HasArtifacts() bool`

HasArtifacts returns a boolean if a field has been set.

### GetArtifactsCount

`func (o *ReleaseV1) GetArtifactsCount() int32`

GetArtifactsCount returns the ArtifactsCount field if non-nil, zero value otherwise.

### GetArtifactsCountOk

`func (o *ReleaseV1) GetArtifactsCountOk() (*int32, bool)`

GetArtifactsCountOk returns a tuple with the ArtifactsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactsCount

`func (o *ReleaseV1) SetArtifactsCount(v int32)`

SetArtifactsCount sets ArtifactsCount field to given value.

### HasArtifactsCount

`func (o *ReleaseV1) HasArtifactsCount() bool`

HasArtifactsCount returns a boolean if a field has been set.

### GetTags

`func (o *ReleaseV1) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *ReleaseV1) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *ReleaseV1) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *ReleaseV1) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetNotes

`func (o *ReleaseV1) GetNotes() string`

GetNotes returns the Notes field if non-nil, zero value otherwise.

### GetNotesOk

`func (o *ReleaseV1) GetNotesOk() (*string, bool)`

GetNotesOk returns a tuple with the Notes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotes

`func (o *ReleaseV1) SetNotes(v string)`

SetNotes sets Notes field to given value.

### HasNotes

`func (o *ReleaseV1) HasNotes() bool`

HasNotes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


