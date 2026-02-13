# DeploymentPhase

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Phase identifier. | [optional] 
**BatchSize** | Pointer to **int32** | Percentage of devices to update in the phase.  | [optional] 
**StartTs** | Pointer to **time.Time** | Start date of a phase. May be undefined for the first phase of a deployment.  | [optional] 
**DeviceCount** | Pointer to **int32** | Number of devices which already requested an update within this phase.  | [optional] 

## Methods

### NewDeploymentPhase

`func NewDeploymentPhase() *DeploymentPhase`

NewDeploymentPhase instantiates a new DeploymentPhase object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentPhaseWithDefaults

`func NewDeploymentPhaseWithDefaults() *DeploymentPhase`

NewDeploymentPhaseWithDefaults instantiates a new DeploymentPhase object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeploymentPhase) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeploymentPhase) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeploymentPhase) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeploymentPhase) HasId() bool`

HasId returns a boolean if a field has been set.

### GetBatchSize

`func (o *DeploymentPhase) GetBatchSize() int32`

GetBatchSize returns the BatchSize field if non-nil, zero value otherwise.

### GetBatchSizeOk

`func (o *DeploymentPhase) GetBatchSizeOk() (*int32, bool)`

GetBatchSizeOk returns a tuple with the BatchSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBatchSize

`func (o *DeploymentPhase) SetBatchSize(v int32)`

SetBatchSize sets BatchSize field to given value.

### HasBatchSize

`func (o *DeploymentPhase) HasBatchSize() bool`

HasBatchSize returns a boolean if a field has been set.

### GetStartTs

`func (o *DeploymentPhase) GetStartTs() time.Time`

GetStartTs returns the StartTs field if non-nil, zero value otherwise.

### GetStartTsOk

`func (o *DeploymentPhase) GetStartTsOk() (*time.Time, bool)`

GetStartTsOk returns a tuple with the StartTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartTs

`func (o *DeploymentPhase) SetStartTs(v time.Time)`

SetStartTs sets StartTs field to given value.

### HasStartTs

`func (o *DeploymentPhase) HasStartTs() bool`

HasStartTs returns a boolean if a field has been set.

### GetDeviceCount

`func (o *DeploymentPhase) GetDeviceCount() int32`

GetDeviceCount returns the DeviceCount field if non-nil, zero value otherwise.

### GetDeviceCountOk

`func (o *DeploymentPhase) GetDeviceCountOk() (*int32, bool)`

GetDeviceCountOk returns a tuple with the DeviceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceCount

`func (o *DeploymentPhase) SetDeviceCount(v int32)`

SetDeviceCount sets DeviceCount field to given value.

### HasDeviceCount

`func (o *DeploymentPhase) HasDeviceCount() bool`

HasDeviceCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


