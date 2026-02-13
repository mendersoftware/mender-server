# DeploymentInstructionsArtifactSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uri** | Pointer to **string** | URL to fetch the artifact from | [optional] 
**Expire** | Pointer to **time.Time** | URL expiration time | [optional] 

## Methods

### NewDeploymentInstructionsArtifactSource

`func NewDeploymentInstructionsArtifactSource() *DeploymentInstructionsArtifactSource`

NewDeploymentInstructionsArtifactSource instantiates a new DeploymentInstructionsArtifactSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentInstructionsArtifactSourceWithDefaults

`func NewDeploymentInstructionsArtifactSourceWithDefaults() *DeploymentInstructionsArtifactSource`

NewDeploymentInstructionsArtifactSourceWithDefaults instantiates a new DeploymentInstructionsArtifactSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUri

`func (o *DeploymentInstructionsArtifactSource) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *DeploymentInstructionsArtifactSource) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *DeploymentInstructionsArtifactSource) SetUri(v string)`

SetUri sets Uri field to given value.

### HasUri

`func (o *DeploymentInstructionsArtifactSource) HasUri() bool`

HasUri returns a boolean if a field has been set.

### GetExpire

`func (o *DeploymentInstructionsArtifactSource) GetExpire() time.Time`

GetExpire returns the Expire field if non-nil, zero value otherwise.

### GetExpireOk

`func (o *DeploymentInstructionsArtifactSource) GetExpireOk() (*time.Time, bool)`

GetExpireOk returns a tuple with the Expire field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpire

`func (o *DeploymentInstructionsArtifactSource) SetExpire(v time.Time)`

SetExpire sets Expire field to given value.

### HasExpire

`func (o *DeploymentInstructionsArtifactSource) HasExpire() bool`

HasExpire returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


