# ArtifactLink

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uri** | **string** |  | 
**Expire** | **time.Time** |  | 

## Methods

### NewArtifactLink

`func NewArtifactLink(uri string, expire time.Time, ) *ArtifactLink`

NewArtifactLink instantiates a new ArtifactLink object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewArtifactLinkWithDefaults

`func NewArtifactLinkWithDefaults() *ArtifactLink`

NewArtifactLinkWithDefaults instantiates a new ArtifactLink object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUri

`func (o *ArtifactLink) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *ArtifactLink) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *ArtifactLink) SetUri(v string)`

SetUri sets Uri field to given value.


### GetExpire

`func (o *ArtifactLink) GetExpire() time.Time`

GetExpire returns the Expire field if non-nil, zero value otherwise.

### GetExpireOk

`func (o *ArtifactLink) GetExpireOk() (*time.Time, bool)`

GetExpireOk returns a tuple with the Expire field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpire

`func (o *ArtifactLink) SetExpire(v time.Time)`

SetExpire sets Expire field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


