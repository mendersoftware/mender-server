# DeploymentV1Internal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Created** | **time.Time** |  | 
**Name** | **string** |  | 
**ArtifactName** | **string** |  | 
**Id** | **string** |  | 
**Finished** | Pointer to **time.Time** |  | [optional] 
**Status** | **string** |  | 
**DeviceCount** | Pointer to **int32** |  | [optional] 
**Artifacts** | Pointer to **[]string** | An array of artifact&#39;s identifiers. | [optional] 
**Type** | Pointer to **string** |  | [optional] 

## Methods

### NewDeploymentV1Internal

`func NewDeploymentV1Internal(created time.Time, name string, artifactName string, id string, status string, ) *DeploymentV1Internal`

NewDeploymentV1Internal instantiates a new DeploymentV1Internal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentV1InternalWithDefaults

`func NewDeploymentV1InternalWithDefaults() *DeploymentV1Internal`

NewDeploymentV1InternalWithDefaults instantiates a new DeploymentV1Internal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreated

`func (o *DeploymentV1Internal) GetCreated() time.Time`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *DeploymentV1Internal) GetCreatedOk() (*time.Time, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *DeploymentV1Internal) SetCreated(v time.Time)`

SetCreated sets Created field to given value.


### GetName

`func (o *DeploymentV1Internal) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DeploymentV1Internal) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DeploymentV1Internal) SetName(v string)`

SetName sets Name field to given value.


### GetArtifactName

`func (o *DeploymentV1Internal) GetArtifactName() string`

GetArtifactName returns the ArtifactName field if non-nil, zero value otherwise.

### GetArtifactNameOk

`func (o *DeploymentV1Internal) GetArtifactNameOk() (*string, bool)`

GetArtifactNameOk returns a tuple with the ArtifactName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifactName

`func (o *DeploymentV1Internal) SetArtifactName(v string)`

SetArtifactName sets ArtifactName field to given value.


### GetId

`func (o *DeploymentV1Internal) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeploymentV1Internal) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeploymentV1Internal) SetId(v string)`

SetId sets Id field to given value.


### GetFinished

`func (o *DeploymentV1Internal) GetFinished() time.Time`

GetFinished returns the Finished field if non-nil, zero value otherwise.

### GetFinishedOk

`func (o *DeploymentV1Internal) GetFinishedOk() (*time.Time, bool)`

GetFinishedOk returns a tuple with the Finished field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFinished

`func (o *DeploymentV1Internal) SetFinished(v time.Time)`

SetFinished sets Finished field to given value.

### HasFinished

`func (o *DeploymentV1Internal) HasFinished() bool`

HasFinished returns a boolean if a field has been set.

### GetStatus

`func (o *DeploymentV1Internal) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeploymentV1Internal) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeploymentV1Internal) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetDeviceCount

`func (o *DeploymentV1Internal) GetDeviceCount() int32`

GetDeviceCount returns the DeviceCount field if non-nil, zero value otherwise.

### GetDeviceCountOk

`func (o *DeploymentV1Internal) GetDeviceCountOk() (*int32, bool)`

GetDeviceCountOk returns a tuple with the DeviceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceCount

`func (o *DeploymentV1Internal) SetDeviceCount(v int32)`

SetDeviceCount sets DeviceCount field to given value.

### HasDeviceCount

`func (o *DeploymentV1Internal) HasDeviceCount() bool`

HasDeviceCount returns a boolean if a field has been set.

### GetArtifacts

`func (o *DeploymentV1Internal) GetArtifacts() []string`

GetArtifacts returns the Artifacts field if non-nil, zero value otherwise.

### GetArtifactsOk

`func (o *DeploymentV1Internal) GetArtifactsOk() (*[]string, bool)`

GetArtifactsOk returns a tuple with the Artifacts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArtifacts

`func (o *DeploymentV1Internal) SetArtifacts(v []string)`

SetArtifacts sets Artifacts field to given value.

### HasArtifacts

`func (o *DeploymentV1Internal) HasArtifacts() bool`

HasArtifacts returns a boolean if a field has been set.

### GetType

`func (o *DeploymentV1Internal) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *DeploymentV1Internal) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *DeploymentV1Internal) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *DeploymentV1Internal) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


