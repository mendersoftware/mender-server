# Statistics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Success** | **int32** | Number of successful deployments. | 
**Pending** | **int32** | Number of pending deployments. | 
**Downloading** | **int32** | Number of deployments being downloaded. | 
**Rebooting** | **int32** | Number of deployments devices are rebooting into. | 
**Installing** | **int32** | Number of deployments devices being installed. | 
**Failure** | **int32** | Number of failed deployments. | 
**Noartifact** | **int32** | Do not have appropriate artifact for device type. | 
**AlreadyInstalled** | **int32** | Number of devices unaffected by upgrade, since they are already running the specified software version. | 
**Aborted** | **int32** | Number of deployments aborted by user. | 
**PauseBeforeInstalling** | **int32** | Number of deployments paused before install state. | 
**PauseBeforeRebooting** | **int32** | Number of deployments paused before reboot phase. | 
**PauseBeforeCommitting** | **int32** | Number of deployments paused before commit phase. | 

## Methods

### NewStatistics

`func NewStatistics(success int32, pending int32, downloading int32, rebooting int32, installing int32, failure int32, noartifact int32, alreadyInstalled int32, aborted int32, pauseBeforeInstalling int32, pauseBeforeRebooting int32, pauseBeforeCommitting int32, ) *Statistics`

NewStatistics instantiates a new Statistics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStatisticsWithDefaults

`func NewStatisticsWithDefaults() *Statistics`

NewStatisticsWithDefaults instantiates a new Statistics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuccess

`func (o *Statistics) GetSuccess() int32`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *Statistics) GetSuccessOk() (*int32, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *Statistics) SetSuccess(v int32)`

SetSuccess sets Success field to given value.


### GetPending

`func (o *Statistics) GetPending() int32`

GetPending returns the Pending field if non-nil, zero value otherwise.

### GetPendingOk

`func (o *Statistics) GetPendingOk() (*int32, bool)`

GetPendingOk returns a tuple with the Pending field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPending

`func (o *Statistics) SetPending(v int32)`

SetPending sets Pending field to given value.


### GetDownloading

`func (o *Statistics) GetDownloading() int32`

GetDownloading returns the Downloading field if non-nil, zero value otherwise.

### GetDownloadingOk

`func (o *Statistics) GetDownloadingOk() (*int32, bool)`

GetDownloadingOk returns a tuple with the Downloading field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloading

`func (o *Statistics) SetDownloading(v int32)`

SetDownloading sets Downloading field to given value.


### GetRebooting

`func (o *Statistics) GetRebooting() int32`

GetRebooting returns the Rebooting field if non-nil, zero value otherwise.

### GetRebootingOk

`func (o *Statistics) GetRebootingOk() (*int32, bool)`

GetRebootingOk returns a tuple with the Rebooting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRebooting

`func (o *Statistics) SetRebooting(v int32)`

SetRebooting sets Rebooting field to given value.


### GetInstalling

`func (o *Statistics) GetInstalling() int32`

GetInstalling returns the Installing field if non-nil, zero value otherwise.

### GetInstallingOk

`func (o *Statistics) GetInstallingOk() (*int32, bool)`

GetInstallingOk returns a tuple with the Installing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstalling

`func (o *Statistics) SetInstalling(v int32)`

SetInstalling sets Installing field to given value.


### GetFailure

`func (o *Statistics) GetFailure() int32`

GetFailure returns the Failure field if non-nil, zero value otherwise.

### GetFailureOk

`func (o *Statistics) GetFailureOk() (*int32, bool)`

GetFailureOk returns a tuple with the Failure field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailure

`func (o *Statistics) SetFailure(v int32)`

SetFailure sets Failure field to given value.


### GetNoartifact

`func (o *Statistics) GetNoartifact() int32`

GetNoartifact returns the Noartifact field if non-nil, zero value otherwise.

### GetNoartifactOk

`func (o *Statistics) GetNoartifactOk() (*int32, bool)`

GetNoartifactOk returns a tuple with the Noartifact field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoartifact

`func (o *Statistics) SetNoartifact(v int32)`

SetNoartifact sets Noartifact field to given value.


### GetAlreadyInstalled

`func (o *Statistics) GetAlreadyInstalled() int32`

GetAlreadyInstalled returns the AlreadyInstalled field if non-nil, zero value otherwise.

### GetAlreadyInstalledOk

`func (o *Statistics) GetAlreadyInstalledOk() (*int32, bool)`

GetAlreadyInstalledOk returns a tuple with the AlreadyInstalled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlreadyInstalled

`func (o *Statistics) SetAlreadyInstalled(v int32)`

SetAlreadyInstalled sets AlreadyInstalled field to given value.


### GetAborted

`func (o *Statistics) GetAborted() int32`

GetAborted returns the Aborted field if non-nil, zero value otherwise.

### GetAbortedOk

`func (o *Statistics) GetAbortedOk() (*int32, bool)`

GetAbortedOk returns a tuple with the Aborted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAborted

`func (o *Statistics) SetAborted(v int32)`

SetAborted sets Aborted field to given value.


### GetPauseBeforeInstalling

`func (o *Statistics) GetPauseBeforeInstalling() int32`

GetPauseBeforeInstalling returns the PauseBeforeInstalling field if non-nil, zero value otherwise.

### GetPauseBeforeInstallingOk

`func (o *Statistics) GetPauseBeforeInstallingOk() (*int32, bool)`

GetPauseBeforeInstallingOk returns a tuple with the PauseBeforeInstalling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPauseBeforeInstalling

`func (o *Statistics) SetPauseBeforeInstalling(v int32)`

SetPauseBeforeInstalling sets PauseBeforeInstalling field to given value.


### GetPauseBeforeRebooting

`func (o *Statistics) GetPauseBeforeRebooting() int32`

GetPauseBeforeRebooting returns the PauseBeforeRebooting field if non-nil, zero value otherwise.

### GetPauseBeforeRebootingOk

`func (o *Statistics) GetPauseBeforeRebootingOk() (*int32, bool)`

GetPauseBeforeRebootingOk returns a tuple with the PauseBeforeRebooting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPauseBeforeRebooting

`func (o *Statistics) SetPauseBeforeRebooting(v int32)`

SetPauseBeforeRebooting sets PauseBeforeRebooting field to given value.


### GetPauseBeforeCommitting

`func (o *Statistics) GetPauseBeforeCommitting() int32`

GetPauseBeforeCommitting returns the PauseBeforeCommitting field if non-nil, zero value otherwise.

### GetPauseBeforeCommittingOk

`func (o *Statistics) GetPauseBeforeCommittingOk() (*int32, bool)`

GetPauseBeforeCommittingOk returns a tuple with the PauseBeforeCommitting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPauseBeforeCommitting

`func (o *Statistics) SetPauseBeforeCommitting(v int32)`

SetPauseBeforeCommitting sets PauseBeforeCommitting field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


