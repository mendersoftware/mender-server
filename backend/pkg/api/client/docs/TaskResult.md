# TaskResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Success** | Pointer to **bool** |  | [optional] 
**Cli** | Pointer to [**TaskResultCLI**](TaskResultCLI.md) |  | [optional] 
**Request** | Pointer to [**TaskResultHTTPRequest**](TaskResultHTTPRequest.md) |  | [optional] 
**Response** | Pointer to [**TaskResultHTTPResponse**](TaskResultHTTPResponse.md) |  | [optional] 

## Methods

### NewTaskResult

`func NewTaskResult() *TaskResult`

NewTaskResult instantiates a new TaskResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskResultWithDefaults

`func NewTaskResultWithDefaults() *TaskResult`

NewTaskResultWithDefaults instantiates a new TaskResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuccess

`func (o *TaskResult) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *TaskResult) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *TaskResult) SetSuccess(v bool)`

SetSuccess sets Success field to given value.

### HasSuccess

`func (o *TaskResult) HasSuccess() bool`

HasSuccess returns a boolean if a field has been set.

### GetCli

`func (o *TaskResult) GetCli() TaskResultCLI`

GetCli returns the Cli field if non-nil, zero value otherwise.

### GetCliOk

`func (o *TaskResult) GetCliOk() (*TaskResultCLI, bool)`

GetCliOk returns a tuple with the Cli field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCli

`func (o *TaskResult) SetCli(v TaskResultCLI)`

SetCli sets Cli field to given value.

### HasCli

`func (o *TaskResult) HasCli() bool`

HasCli returns a boolean if a field has been set.

### GetRequest

`func (o *TaskResult) GetRequest() TaskResultHTTPRequest`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *TaskResult) GetRequestOk() (*TaskResultHTTPRequest, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *TaskResult) SetRequest(v TaskResultHTTPRequest)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *TaskResult) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetResponse

`func (o *TaskResult) GetResponse() TaskResultHTTPResponse`

GetResponse returns the Response field if non-nil, zero value otherwise.

### GetResponseOk

`func (o *TaskResult) GetResponseOk() (*TaskResultHTTPResponse, bool)`

GetResponseOk returns a tuple with the Response field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponse

`func (o *TaskResult) SetResponse(v TaskResultHTTPResponse)`

SetResponse sets Response field to given value.

### HasResponse

`func (o *TaskResult) HasResponse() bool`

HasResponse returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


