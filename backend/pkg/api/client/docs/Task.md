# Task

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Type** | **string** |  | 
**Retries** | Pointer to **int32** |  | [optional] 
**RetryDelaySeconds** | Pointer to **int32** |  | [optional] 
**Requires** | Pointer to **[]string** |  | [optional] 
**Cli** | Pointer to [**CLIParams**](CLIParams.md) |  | [optional] 
**Http** | Pointer to [**HTTPParams**](HTTPParams.md) |  | [optional] 

## Methods

### NewTask

`func NewTask(name string, type_ string, ) *Task`

NewTask instantiates a new Task object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskWithDefaults

`func NewTaskWithDefaults() *Task`

NewTaskWithDefaults instantiates a new Task object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Task) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Task) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Task) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *Task) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Task) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Task) SetType(v string)`

SetType sets Type field to given value.


### GetRetries

`func (o *Task) GetRetries() int32`

GetRetries returns the Retries field if non-nil, zero value otherwise.

### GetRetriesOk

`func (o *Task) GetRetriesOk() (*int32, bool)`

GetRetriesOk returns a tuple with the Retries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetries

`func (o *Task) SetRetries(v int32)`

SetRetries sets Retries field to given value.

### HasRetries

`func (o *Task) HasRetries() bool`

HasRetries returns a boolean if a field has been set.

### GetRetryDelaySeconds

`func (o *Task) GetRetryDelaySeconds() int32`

GetRetryDelaySeconds returns the RetryDelaySeconds field if non-nil, zero value otherwise.

### GetRetryDelaySecondsOk

`func (o *Task) GetRetryDelaySecondsOk() (*int32, bool)`

GetRetryDelaySecondsOk returns a tuple with the RetryDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetryDelaySeconds

`func (o *Task) SetRetryDelaySeconds(v int32)`

SetRetryDelaySeconds sets RetryDelaySeconds field to given value.

### HasRetryDelaySeconds

`func (o *Task) HasRetryDelaySeconds() bool`

HasRetryDelaySeconds returns a boolean if a field has been set.

### GetRequires

`func (o *Task) GetRequires() []string`

GetRequires returns the Requires field if non-nil, zero value otherwise.

### GetRequiresOk

`func (o *Task) GetRequiresOk() (*[]string, bool)`

GetRequiresOk returns a tuple with the Requires field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequires

`func (o *Task) SetRequires(v []string)`

SetRequires sets Requires field to given value.

### HasRequires

`func (o *Task) HasRequires() bool`

HasRequires returns a boolean if a field has been set.

### GetCli

`func (o *Task) GetCli() CLIParams`

GetCli returns the Cli field if non-nil, zero value otherwise.

### GetCliOk

`func (o *Task) GetCliOk() (*CLIParams, bool)`

GetCliOk returns a tuple with the Cli field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCli

`func (o *Task) SetCli(v CLIParams)`

SetCli sets Cli field to given value.

### HasCli

`func (o *Task) HasCli() bool`

HasCli returns a boolean if a field has been set.

### GetHttp

`func (o *Task) GetHttp() HTTPParams`

GetHttp returns the Http field if non-nil, zero value otherwise.

### GetHttpOk

`func (o *Task) GetHttpOk() (*HTTPParams, bool)`

GetHttpOk returns a tuple with the Http field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttp

`func (o *Task) SetHttp(v HTTPParams)`

SetHttp sets Http field to given value.

### HasHttp

`func (o *Task) HasHttp() bool`

HasHttp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


