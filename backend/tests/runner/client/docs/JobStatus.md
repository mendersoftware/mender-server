# JobStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**WorkflowName** | Pointer to **string** |  | [optional] 
**InputParameters** | Pointer to [**[]InputParameter**](InputParameter.md) |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**Results** | Pointer to [**[]TaskResult**](TaskResult.md) |  | [optional] 

## Methods

### NewJobStatus

`func NewJobStatus() *JobStatus`

NewJobStatus instantiates a new JobStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobStatusWithDefaults

`func NewJobStatusWithDefaults() *JobStatus`

NewJobStatusWithDefaults instantiates a new JobStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *JobStatus) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *JobStatus) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *JobStatus) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *JobStatus) HasId() bool`

HasId returns a boolean if a field has been set.

### GetWorkflowName

`func (o *JobStatus) GetWorkflowName() string`

GetWorkflowName returns the WorkflowName field if non-nil, zero value otherwise.

### GetWorkflowNameOk

`func (o *JobStatus) GetWorkflowNameOk() (*string, bool)`

GetWorkflowNameOk returns a tuple with the WorkflowName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkflowName

`func (o *JobStatus) SetWorkflowName(v string)`

SetWorkflowName sets WorkflowName field to given value.

### HasWorkflowName

`func (o *JobStatus) HasWorkflowName() bool`

HasWorkflowName returns a boolean if a field has been set.

### GetInputParameters

`func (o *JobStatus) GetInputParameters() []InputParameter`

GetInputParameters returns the InputParameters field if non-nil, zero value otherwise.

### GetInputParametersOk

`func (o *JobStatus) GetInputParametersOk() (*[]InputParameter, bool)`

GetInputParametersOk returns a tuple with the InputParameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputParameters

`func (o *JobStatus) SetInputParameters(v []InputParameter)`

SetInputParameters sets InputParameters field to given value.

### HasInputParameters

`func (o *JobStatus) HasInputParameters() bool`

HasInputParameters returns a boolean if a field has been set.

### GetStatus

`func (o *JobStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *JobStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *JobStatus) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *JobStatus) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetResults

`func (o *JobStatus) GetResults() []TaskResult`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *JobStatus) GetResultsOk() (*[]TaskResult, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *JobStatus) SetResults(v []TaskResult)`

SetResults sets Results field to given value.

### HasResults

`func (o *JobStatus) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


