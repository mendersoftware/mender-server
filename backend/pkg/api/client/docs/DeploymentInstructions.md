# DeploymentInstructions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Deployment ID | 
**Artifact** | [**DeploymentInstructionsArtifact**](DeploymentInstructionsArtifact.md) |  | 

## Methods

### NewDeploymentInstructions

`func NewDeploymentInstructions(id string, artifact DeploymentInstructionsArtifact, ) *DeploymentInstructions`

NewDeploymentInstructions instantiates a new DeploymentInstructions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentInstructionsWithDefaults

`func NewDeploymentInstructionsWithDefaults() *DeploymentInstructions`

NewDeploymentInstructionsWithDefaults instantiates a new DeploymentInstructions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeploymentInstructions) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeploymentInstructions) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeploymentInstructions) SetId(v string)`

SetId sets Id field to given value.


### GetArtifact

`func (o *DeploymentInstructions) GetArtifact() DeploymentInstructionsArtifact`

GetArtifact returns the Artifact field if non-nil, zero value otherwise.

### GetArtifactOk

`func (o *DeploymentInstructions) GetArtifactOk() (*DeploymentInstructionsArtifact, bool)`

GetArtifactOk returns a tuple with the Artifact field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifact

`func (o *DeploymentInstructions) SetArtifact(v DeploymentInstructionsArtifact)`

SetArtifact sets Artifact field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


