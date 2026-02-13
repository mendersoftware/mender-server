# DeploymentInstructionsArtifact

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Source** | [**DeploymentInstructionsArtifactSource**](DeploymentInstructionsArtifactSource.md) |  | 
**DeviceTypesCompatible** | **[]string** | Compatible device types | 
**ArtifactName** | **string** |  | 

## Methods

### NewDeploymentInstructionsArtifact

`func NewDeploymentInstructionsArtifact(source DeploymentInstructionsArtifactSource, deviceTypesCompatible []string, artifactName string, ) *DeploymentInstructionsArtifact`

NewDeploymentInstructionsArtifact instantiates a new DeploymentInstructionsArtifact object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentInstructionsArtifactWithDefaults

`func NewDeploymentInstructionsArtifactWithDefaults() *DeploymentInstructionsArtifact`

NewDeploymentInstructionsArtifactWithDefaults instantiates a new DeploymentInstructionsArtifact object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeploymentInstructionsArtifact) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeploymentInstructionsArtifact) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeploymentInstructionsArtifact) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeploymentInstructionsArtifact) HasId() bool`

HasId returns a boolean if a field has been set.

### GetSource

`func (o *DeploymentInstructionsArtifact) GetSource() DeploymentInstructionsArtifactSource`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *DeploymentInstructionsArtifact) GetSourceOk() (*DeploymentInstructionsArtifactSource, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *DeploymentInstructionsArtifact) SetSource(v DeploymentInstructionsArtifactSource)`

SetSource sets Source field to given value.


### GetDeviceTypesCompatible

`func (o *DeploymentInstructionsArtifact) GetDeviceTypesCompatible() []string`

GetDeviceTypesCompatible returns the DeviceTypesCompatible field if non-nil, zero value otherwise.

### GetDeviceTypesCompatibleOk

`func (o *DeploymentInstructionsArtifact) GetDeviceTypesCompatibleOk() (*[]string, bool)`

GetDeviceTypesCompatibleOk returns a tuple with the DeviceTypesCompatible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceTypesCompatible

`func (o *DeploymentInstructionsArtifact) SetDeviceTypesCompatible(v []string)`

SetDeviceTypesCompatible sets DeviceTypesCompatible field to given value.


### GetArtifactName

`func (o *DeploymentInstructionsArtifact) GetArtifactName() string`

GetArtifactName returns the ArtifactName field if non-nil, zero value otherwise.

### GetArtifactNameOk

`func (o *DeploymentInstructionsArtifact) GetArtifactNameOk() (*string, bool)`

GetArtifactNameOk returns a tuple with the ArtifactName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactName

`func (o *DeploymentInstructionsArtifact) SetArtifactName(v string)`

SetArtifactName sets ArtifactName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


