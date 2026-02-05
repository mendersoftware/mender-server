# NewDeploymentForGroup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the deployment | 
**ArtifactName** | **string** | Name of the artifact to deploy | 
**ForceInstallation** | Pointer to **bool** | Force the installation of the Artifact disabling the &#x60;already-installed&#x60; check. | [optional] 

## Methods

### NewNewDeploymentForGroup

`func NewNewDeploymentForGroup(name string, artifactName string, ) *NewDeploymentForGroup`

NewNewDeploymentForGroup instantiates a new NewDeploymentForGroup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewDeploymentForGroupWithDefaults

`func NewNewDeploymentForGroupWithDefaults() *NewDeploymentForGroup`

NewNewDeploymentForGroupWithDefaults instantiates a new NewDeploymentForGroup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NewDeploymentForGroup) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NewDeploymentForGroup) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NewDeploymentForGroup) SetName(v string)`

SetName sets Name field to given value.


### GetArtifactName

`func (o *NewDeploymentForGroup) GetArtifactName() string`

GetArtifactName returns the ArtifactName field if non-nil, zero value otherwise.

### GetArtifactNameOk

`func (o *NewDeploymentForGroup) GetArtifactNameOk() (*string, bool)`

GetArtifactNameOk returns a tuple with the ArtifactName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactName

`func (o *NewDeploymentForGroup) SetArtifactName(v string)`

SetArtifactName sets ArtifactName field to given value.


### GetForceInstallation

`func (o *NewDeploymentForGroup) GetForceInstallation() bool`

GetForceInstallation returns the ForceInstallation field if non-nil, zero value otherwise.

### GetForceInstallationOk

`func (o *NewDeploymentForGroup) GetForceInstallationOk() (*bool, bool)`

GetForceInstallationOk returns a tuple with the ForceInstallation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForceInstallation

`func (o *NewDeploymentForGroup) SetForceInstallation(v bool)`

SetForceInstallation sets ForceInstallation field to given value.

### HasForceInstallation

`func (o *NewDeploymentForGroup) HasForceInstallation() bool`

HasForceInstallation returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


