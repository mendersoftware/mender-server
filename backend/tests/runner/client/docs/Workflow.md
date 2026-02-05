# Workflow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**Version** | **int32** |  | 
**Schemaversion** | Pointer to **int32** |  | [optional] 
**Tasks** | [**[]Task**](Task.md) |  | 
**InputParameters** | Pointer to **[]string** |  | [optional] 

## Methods

### NewWorkflow

`func NewWorkflow(name string, version int32, tasks []Task, ) *Workflow`

NewWorkflow instantiates a new Workflow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowWithDefaults

`func NewWorkflowWithDefaults() *Workflow`

NewWorkflowWithDefaults instantiates a new Workflow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Workflow) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Workflow) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Workflow) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Workflow) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Workflow) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Workflow) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Workflow) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetVersion

`func (o *Workflow) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *Workflow) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *Workflow) SetVersion(v int32)`

SetVersion sets Version field to given value.


### GetSchemaversion

`func (o *Workflow) GetSchemaversion() int32`

GetSchemaversion returns the Schemaversion field if non-nil, zero value otherwise.

### GetSchemaversionOk

`func (o *Workflow) GetSchemaversionOk() (*int32, bool)`

GetSchemaversionOk returns a tuple with the Schemaversion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchemaversion

`func (o *Workflow) SetSchemaversion(v int32)`

SetSchemaversion sets Schemaversion field to given value.

### HasSchemaversion

`func (o *Workflow) HasSchemaversion() bool`

HasSchemaversion returns a boolean if a field has been set.

### GetTasks

`func (o *Workflow) GetTasks() []Task`

GetTasks returns the Tasks field if non-nil, zero value otherwise.

### GetTasksOk

`func (o *Workflow) GetTasksOk() (*[]Task, bool)`

GetTasksOk returns a tuple with the Tasks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTasks

`func (o *Workflow) SetTasks(v []Task)`

SetTasks sets Tasks field to given value.


### GetInputParameters

`func (o *Workflow) GetInputParameters() []string`

GetInputParameters returns the InputParameters field if non-nil, zero value otherwise.

### GetInputParametersOk

`func (o *Workflow) GetInputParametersOk() (*[]string, bool)`

GetInputParametersOk returns a tuple with the InputParameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputParameters

`func (o *Workflow) SetInputParameters(v []string)`

SetInputParameters sets InputParameters field to given value.

### HasInputParameters

`func (o *Workflow) HasInputParameters() bool`

HasInputParameters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


