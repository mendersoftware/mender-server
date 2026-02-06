# CLIParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **[]string** |  | [optional] 
**ExecutionTimeOut** | Pointer to **int32** |  | [optional] 

## Methods

### NewCLIParams

`func NewCLIParams() *CLIParams`

NewCLIParams instantiates a new CLIParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCLIParamsWithDefaults

`func NewCLIParamsWithDefaults() *CLIParams`

NewCLIParamsWithDefaults instantiates a new CLIParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *CLIParams) GetCommand() []string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *CLIParams) GetCommandOk() (*[]string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *CLIParams) SetCommand(v []string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *CLIParams) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetExecutionTimeOut

`func (o *CLIParams) GetExecutionTimeOut() int32`

GetExecutionTimeOut returns the ExecutionTimeOut field if non-nil, zero value otherwise.

### GetExecutionTimeOutOk

`func (o *CLIParams) GetExecutionTimeOutOk() (*int32, bool)`

GetExecutionTimeOutOk returns a tuple with the ExecutionTimeOut field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionTimeOut

`func (o *CLIParams) SetExecutionTimeOut(v int32)`

SetExecutionTimeOut sets ExecutionTimeOut field to given value.

### HasExecutionTimeOut

`func (o *CLIParams) HasExecutionTimeOut() bool`

HasExecutionTimeOut returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


