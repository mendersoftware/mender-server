# DeploymentV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Deployment identifier | 
**Name** | **string** | Name of the deployment | 
**ArtifactName** | **string** | Name of the artifact to deploy | 
**Created** | **time.Time** | Deployment&#39;s creation date and time | 
**Finished** | Pointer to **time.Time** | Deployment&#39;s completion date and time | [optional] 
**Status** | **string** | Status of the deployment | 
**DeviceCount** | **int32** | Number of devices the deployment acted upon | 
**Artifacts** | Pointer to **[]string** | An array of artifact&#39;s identifiers. | [optional] 
**Groups** | Pointer to **[]string** | An array of groups the devices targeted by the deployment belong to. Available only if the user created the deployment for a group or a single device (if the device was in a static group).  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**Configuration** | Pointer to **string** | A string containing a configuration object provided with the deployment constructor.  | [optional] 
**Statistics** | Pointer to [**DeploymentStatistics**](DeploymentStatistics.md) |  | [optional] 
**Filter** | Pointer to [**FilterV1**](FilterV1.md) |  | [optional] 

## Methods

### NewDeploymentV1

`func NewDeploymentV1(id string, name string, artifactName string, created time.Time, status string, deviceCount int32, ) *DeploymentV1`

NewDeploymentV1 instantiates a new DeploymentV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentV1WithDefaults

`func NewDeploymentV1WithDefaults() *DeploymentV1`

NewDeploymentV1WithDefaults instantiates a new DeploymentV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeploymentV1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeploymentV1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeploymentV1) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *DeploymentV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DeploymentV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DeploymentV1) SetName(v string)`

SetName sets Name field to given value.


### GetArtifactName

`func (o *DeploymentV1) GetArtifactName() string`

GetArtifactName returns the ArtifactName field if non-nil, zero value otherwise.

### GetArtifactNameOk

`func (o *DeploymentV1) GetArtifactNameOk() (*string, bool)`

GetArtifactNameOk returns a tuple with the ArtifactName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactName

`func (o *DeploymentV1) SetArtifactName(v string)`

SetArtifactName sets ArtifactName field to given value.


### GetCreated

`func (o *DeploymentV1) GetCreated() time.Time`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *DeploymentV1) GetCreatedOk() (*time.Time, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *DeploymentV1) SetCreated(v time.Time)`

SetCreated sets Created field to given value.


### GetFinished

`func (o *DeploymentV1) GetFinished() time.Time`

GetFinished returns the Finished field if non-nil, zero value otherwise.

### GetFinishedOk

`func (o *DeploymentV1) GetFinishedOk() (*time.Time, bool)`

GetFinishedOk returns a tuple with the Finished field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFinished

`func (o *DeploymentV1) SetFinished(v time.Time)`

SetFinished sets Finished field to given value.

### HasFinished

`func (o *DeploymentV1) HasFinished() bool`

HasFinished returns a boolean if a field has been set.

### GetStatus

`func (o *DeploymentV1) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeploymentV1) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeploymentV1) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetDeviceCount

`func (o *DeploymentV1) GetDeviceCount() int32`

GetDeviceCount returns the DeviceCount field if non-nil, zero value otherwise.

### GetDeviceCountOk

`func (o *DeploymentV1) GetDeviceCountOk() (*int32, bool)`

GetDeviceCountOk returns a tuple with the DeviceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceCount

`func (o *DeploymentV1) SetDeviceCount(v int32)`

SetDeviceCount sets DeviceCount field to given value.


### GetArtifacts

`func (o *DeploymentV1) GetArtifacts() []string`

GetArtifacts returns the Artifacts field if non-nil, zero value otherwise.

### GetArtifactsOk

`func (o *DeploymentV1) GetArtifactsOk() (*[]string, bool)`

GetArtifactsOk returns a tuple with the Artifacts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifacts

`func (o *DeploymentV1) SetArtifacts(v []string)`

SetArtifacts sets Artifacts field to given value.

### HasArtifacts

`func (o *DeploymentV1) HasArtifacts() bool`

HasArtifacts returns a boolean if a field has been set.

### GetGroups

`func (o *DeploymentV1) GetGroups() []string`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *DeploymentV1) GetGroupsOk() (*[]string, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *DeploymentV1) SetGroups(v []string)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *DeploymentV1) HasGroups() bool`

HasGroups returns a boolean if a field has been set.

### GetType

`func (o *DeploymentV1) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *DeploymentV1) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *DeploymentV1) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *DeploymentV1) HasType() bool`

HasType returns a boolean if a field has been set.

### GetConfiguration

`func (o *DeploymentV1) GetConfiguration() string`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *DeploymentV1) GetConfigurationOk() (*string, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *DeploymentV1) SetConfiguration(v string)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *DeploymentV1) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.

### GetStatistics

`func (o *DeploymentV1) GetStatistics() DeploymentStatistics`

GetStatistics returns the Statistics field if non-nil, zero value otherwise.

### GetStatisticsOk

`func (o *DeploymentV1) GetStatisticsOk() (*DeploymentStatistics, bool)`

GetStatisticsOk returns a tuple with the Statistics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatistics

`func (o *DeploymentV1) SetStatistics(v DeploymentStatistics)`

SetStatistics sets Statistics field to given value.

### HasStatistics

`func (o *DeploymentV1) HasStatistics() bool`

HasStatistics returns a boolean if a field has been set.

### GetFilter

`func (o *DeploymentV1) GetFilter() FilterV1`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *DeploymentV1) GetFilterOk() (*FilterV1, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *DeploymentV1) SetFilter(v FilterV1)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *DeploymentV1) HasFilter() bool`

HasFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


