# ArtifactUploadLink

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the artifact upload intent. | 
**Uri** | **string** |  | 
**Expire** | **time.Time** |  | 

## Methods

### NewArtifactUploadLink

`func NewArtifactUploadLink(id string, uri string, expire time.Time, ) *ArtifactUploadLink`

NewArtifactUploadLink instantiates a new ArtifactUploadLink object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewArtifactUploadLinkWithDefaults

`func NewArtifactUploadLinkWithDefaults() *ArtifactUploadLink`

NewArtifactUploadLinkWithDefaults instantiates a new ArtifactUploadLink object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ArtifactUploadLink) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ArtifactUploadLink) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ArtifactUploadLink) SetId(v string)`

SetId sets Id field to given value.


### GetUri

`func (o *ArtifactUploadLink) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *ArtifactUploadLink) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *ArtifactUploadLink) SetUri(v string)`

SetUri sets Uri field to given value.


### GetExpire

`func (o *ArtifactUploadLink) GetExpire() time.Time`

GetExpire returns the Expire field if non-nil, zero value otherwise.

### GetExpireOk

`func (o *ArtifactUploadLink) GetExpireOk() (*time.Time, bool)`

GetExpireOk returns a tuple with the Expire field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpire

`func (o *ArtifactUploadLink) SetExpire(v time.Time)`

SetExpire sets Expire field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


