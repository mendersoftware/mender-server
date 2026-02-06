# DeploymentStatistics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to [**Statistics**](Statistics.md) |  | [optional] 
**TotalSize** | Pointer to **int32** | Sum of sizes (in bytes) of all artifacts assigned to all device deployments, which are part of this deployment. If the same artifact is assigned to multiple device deployments, its size will be counted multiple times.  | [optional] 

## Methods

### NewDeploymentStatistics

`func NewDeploymentStatistics() *DeploymentStatistics`

NewDeploymentStatistics instantiates a new DeploymentStatistics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentStatisticsWithDefaults

`func NewDeploymentStatisticsWithDefaults() *DeploymentStatistics`

NewDeploymentStatisticsWithDefaults instantiates a new DeploymentStatistics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *DeploymentStatistics) GetStatus() Statistics`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeploymentStatistics) GetStatusOk() (*Statistics, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeploymentStatistics) SetStatus(v Statistics)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DeploymentStatistics) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetTotalSize

`func (o *DeploymentStatistics) GetTotalSize() int32`

GetTotalSize returns the TotalSize field if non-nil, zero value otherwise.

### GetTotalSizeOk

`func (o *DeploymentStatistics) GetTotalSizeOk() (*int32, bool)`

GetTotalSizeOk returns a tuple with the TotalSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalSize

`func (o *DeploymentStatistics) SetTotalSize(v int32)`

SetTotalSize sets TotalSize field to given value.

### HasTotalSize

`func (o *DeploymentStatistics) HasTotalSize() bool`

HasTotalSize returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


