# DeploymentStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | **string** |  | 
**Substate** | Pointer to **string** | Additional state information | [optional] 

## Methods

### NewDeploymentStatus

`func NewDeploymentStatus(status string, ) *DeploymentStatus`

NewDeploymentStatus instantiates a new DeploymentStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentStatusWithDefaults

`func NewDeploymentStatusWithDefaults() *DeploymentStatus`

NewDeploymentStatusWithDefaults instantiates a new DeploymentStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *DeploymentStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeploymentStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeploymentStatus) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetSubstate

`func (o *DeploymentStatus) GetSubstate() string`

GetSubstate returns the Substate field if non-nil, zero value otherwise.

### GetSubstateOk

`func (o *DeploymentStatus) GetSubstateOk() (*string, bool)`

GetSubstateOk returns a tuple with the Substate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubstate

`func (o *DeploymentStatus) SetSubstate(v string)`

SetSubstate sets Substate field to given value.

### HasSubstate

`func (o *DeploymentStatus) HasSubstate() bool`

HasSubstate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


