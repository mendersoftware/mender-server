# TaskResultCLI

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **[]string** |  | [optional] 
**Output** | Pointer to **string** |  | [optional] 
**Error** | Pointer to **string** |  | [optional] 
**ExitCode** | Pointer to **int32** |  | [optional] 

## Methods

### NewTaskResultCLI

`func NewTaskResultCLI() *TaskResultCLI`

NewTaskResultCLI instantiates a new TaskResultCLI object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskResultCLIWithDefaults

`func NewTaskResultCLIWithDefaults() *TaskResultCLI`

NewTaskResultCLIWithDefaults instantiates a new TaskResultCLI object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *TaskResultCLI) GetCommand() []string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *TaskResultCLI) GetCommandOk() (*[]string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *TaskResultCLI) SetCommand(v []string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *TaskResultCLI) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetOutput

`func (o *TaskResultCLI) GetOutput() string`

GetOutput returns the Output field if non-nil, zero value otherwise.

### GetOutputOk

`func (o *TaskResultCLI) GetOutputOk() (*string, bool)`

GetOutputOk returns a tuple with the Output field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutput

`func (o *TaskResultCLI) SetOutput(v string)`

SetOutput sets Output field to given value.

### HasOutput

`func (o *TaskResultCLI) HasOutput() bool`

HasOutput returns a boolean if a field has been set.

### GetError

`func (o *TaskResultCLI) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *TaskResultCLI) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *TaskResultCLI) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *TaskResultCLI) HasError() bool`

HasError returns a boolean if a field has been set.

### GetExitCode

`func (o *TaskResultCLI) GetExitCode() int32`

GetExitCode returns the ExitCode field if non-nil, zero value otherwise.

### GetExitCodeOk

`func (o *TaskResultCLI) GetExitCodeOk() (*int32, bool)`

GetExitCodeOk returns a tuple with the ExitCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExitCode

`func (o *TaskResultCLI) SetExitCode(v int32)`

SetExitCode sets ExitCode field to given value.

### HasExitCode

`func (o *TaskResultCLI) HasExitCode() bool`

HasExitCode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


