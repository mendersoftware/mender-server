# JobObject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**WorkflowName** | Pointer to **string** |  | [optional] 
**InputParameters** | Pointer to [**[]InputParameter**](InputParameter.md) |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**Results** | Pointer to [**[]TaskResult**](TaskResult.md) |  | [optional] 
**InsertTime** | Pointer to **time.Time** |  | [optional] 
**Version** | Pointer to **string** |  | [optional] 

## Methods

### NewJobObject

`func NewJobObject() *JobObject`

NewJobObject instantiates a new JobObject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobObjectWithDefaults

`func NewJobObjectWithDefaults() *JobObject`

NewJobObjectWithDefaults instantiates a new JobObject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *JobObject) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *JobObject) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *JobObject) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *JobObject) HasId() bool`

HasId returns a boolean if a field has been set.

### GetWorkflowName

`func (o *JobObject) GetWorkflowName() string`

GetWorkflowName returns the WorkflowName field if non-nil, zero value otherwise.

### GetWorkflowNameOk

`func (o *JobObject) GetWorkflowNameOk() (*string, bool)`

GetWorkflowNameOk returns a tuple with the WorkflowName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkflowName

`func (o *JobObject) SetWorkflowName(v string)`

SetWorkflowName sets WorkflowName field to given value.

### HasWorkflowName

`func (o *JobObject) HasWorkflowName() bool`

HasWorkflowName returns a boolean if a field has been set.

### GetInputParameters

`func (o *JobObject) GetInputParameters() []InputParameter`

GetInputParameters returns the InputParameters field if non-nil, zero value otherwise.

### GetInputParametersOk

`func (o *JobObject) GetInputParametersOk() (*[]InputParameter, bool)`

GetInputParametersOk returns a tuple with the InputParameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputParameters

`func (o *JobObject) SetInputParameters(v []InputParameter)`

SetInputParameters sets InputParameters field to given value.

### HasInputParameters

`func (o *JobObject) HasInputParameters() bool`

HasInputParameters returns a boolean if a field has been set.

### GetStatus

`func (o *JobObject) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *JobObject) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *JobObject) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *JobObject) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetResults

`func (o *JobObject) GetResults() []TaskResult`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *JobObject) GetResultsOk() (*[]TaskResult, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *JobObject) SetResults(v []TaskResult)`

SetResults sets Results field to given value.

### HasResults

`func (o *JobObject) HasResults() bool`

HasResults returns a boolean if a field has been set.

### GetInsertTime

`func (o *JobObject) GetInsertTime() time.Time`

GetInsertTime returns the InsertTime field if non-nil, zero value otherwise.

### GetInsertTimeOk

`func (o *JobObject) GetInsertTimeOk() (*time.Time, bool)`

GetInsertTimeOk returns a tuple with the InsertTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInsertTime

`func (o *JobObject) SetInsertTime(v time.Time)`

SetInsertTime sets InsertTime field to given value.

### HasInsertTime

`func (o *JobObject) HasInsertTime() bool`

HasInsertTime returns a boolean if a field has been set.

### GetVersion

`func (o *JobObject) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *JobObject) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *JobObject) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *JobObject) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


