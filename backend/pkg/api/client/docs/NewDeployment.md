# NewDeployment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the deployment | 
**ArtifactName** | **string** | Name of the artifact to deploy | 
**Devices** | Pointer to **[]string** | An array of devices&#39; identifiers. | [optional] 
**AllDevices** | Pointer to **bool** | When set, the deployment will be created for all currently accepted devices.  | [optional] 
**ForceInstallation** | Pointer to **bool** | Force the installation of the Artifact disabling the &#x60;already-installed&#x60; check. | [optional] 

## Methods

### NewNewDeployment

`func NewNewDeployment(name string, artifactName string, ) *NewDeployment`

NewNewDeployment instantiates a new NewDeployment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewDeploymentWithDefaults

`func NewNewDeploymentWithDefaults() *NewDeployment`

NewNewDeploymentWithDefaults instantiates a new NewDeployment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NewDeployment) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NewDeployment) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NewDeployment) SetName(v string)`

SetName sets Name field to given value.


### GetArtifactName

`func (o *NewDeployment) GetArtifactName() string`

GetArtifactName returns the ArtifactName field if non-nil, zero value otherwise.

### GetArtifactNameOk

`func (o *NewDeployment) GetArtifactNameOk() (*string, bool)`

GetArtifactNameOk returns a tuple with the ArtifactName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactName

`func (o *NewDeployment) SetArtifactName(v string)`

SetArtifactName sets ArtifactName field to given value.


### GetDevices

`func (o *NewDeployment) GetDevices() []string`

GetDevices returns the Devices field if non-nil, zero value otherwise.

### GetDevicesOk

`func (o *NewDeployment) GetDevicesOk() (*[]string, bool)`

GetDevicesOk returns a tuple with the Devices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevices

`func (o *NewDeployment) SetDevices(v []string)`

SetDevices sets Devices field to given value.

### HasDevices

`func (o *NewDeployment) HasDevices() bool`

HasDevices returns a boolean if a field has been set.

### GetAllDevices

`func (o *NewDeployment) GetAllDevices() bool`

GetAllDevices returns the AllDevices field if non-nil, zero value otherwise.

### GetAllDevicesOk

`func (o *NewDeployment) GetAllDevicesOk() (*bool, bool)`

GetAllDevicesOk returns a tuple with the AllDevices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllDevices

`func (o *NewDeployment) SetAllDevices(v bool)`

SetAllDevices sets AllDevices field to given value.

### HasAllDevices

`func (o *NewDeployment) HasAllDevices() bool`

HasAllDevices returns a boolean if a field has been set.

### GetForceInstallation

`func (o *NewDeployment) GetForceInstallation() bool`

GetForceInstallation returns the ForceInstallation field if non-nil, zero value otherwise.

### GetForceInstallationOk

`func (o *NewDeployment) GetForceInstallationOk() (*bool, bool)`

GetForceInstallationOk returns a tuple with the ForceInstallation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForceInstallation

`func (o *NewDeployment) SetForceInstallation(v bool)`

SetForceInstallation sets ForceInstallation field to given value.

### HasForceInstallation

`func (o *NewDeployment) HasForceInstallation() bool`

HasForceInstallation returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


