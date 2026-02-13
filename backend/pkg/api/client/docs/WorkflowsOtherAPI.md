# \WorkflowsOtherAPI

All URIs are relative to *https://hosted.mender.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**JobStructure**](WorkflowsOtherAPI.md#JobStructure) | **Get** /api/v1/jobs/{id} | Gets the job for the given id.
[**ListWorkflows**](WorkflowsOtherAPI.md#ListWorkflows) | **Get** /api/v1/metadata/workflows | Get all workflow definitions
[**RegisterWorkflow**](WorkflowsOtherAPI.md#RegisterWorkflow) | **Post** /api/v1/metadata/workflows | Register a new workflow
[**StartBatchWorkflows**](WorkflowsOtherAPI.md#StartBatchWorkflows) | **Post** /api/v1/workflow/{name}/batch | Start a batch of workflows
[**StartWorkflow**](WorkflowsOtherAPI.md#StartWorkflow) | **Post** /api/v1/workflow/{name} | Start a new workflow
[**WorkflowStatus**](WorkflowsOtherAPI.md#WorkflowStatus) | **Get** /api/v1/workflow/{name}/{id} | Gets the workflow status for the given id.
[**WorkflowsCheckHealth**](WorkflowsOtherAPI.md#WorkflowsCheckHealth) | **Get** /api/v1/health | Check if service and all operational dependencies are healthy.
[**WorkflowsCheckLiveliness**](WorkflowsOtherAPI.md#WorkflowsCheckLiveliness) | **Get** /status | Check if service API is alive and serving requests



## JobStructure

> JobObject JobStructure(ctx, id).Execute()

Gets the job for the given id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	id := "id_example" // string | Job identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.JobStructure(context.Background(), id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.JobStructure``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `JobStructure`: JobObject
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.JobStructure`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | Job identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiJobStructureRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**JobObject**](JobObject.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListWorkflows

> []Workflow ListWorkflows(ctx).Execute()

Get all workflow definitions

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.ListWorkflows(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.ListWorkflows``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListWorkflows`: []Workflow
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.ListWorkflows`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListWorkflowsRequest struct via the builder pattern


### Return type

[**[]Workflow**](Workflow.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RegisterWorkflow

> RegisterWorkflow(ctx).Workflow(workflow).Execute()

Register a new workflow

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	workflow := *openapiclient.NewWorkflow("Name_example", int32(123), []openapiclient.Task{*openapiclient.NewTask("Name_example", "Type_example")}) // Workflow | Workflow definition.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkflowsOtherAPI.RegisterWorkflow(context.Background()).Workflow(workflow).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.RegisterWorkflow``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRegisterWorkflowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workflow** | [**Workflow**](Workflow.md) | Workflow definition. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StartBatchWorkflows

> []StartBatchWorkflows201ResponseInner StartBatchWorkflows(ctx, name).InputParameter(inputParameter).Execute()

Start a batch of workflows



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	name := "name_example" // string | Workflow identifier.
	inputParameter := [][]InputParameter{[]openapiclient.InputParameter{*openapiclient.NewInputParameter("Name_example", "Value_example")}} // [][]InputParameter | Contains the definition of the job to be started.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.StartBatchWorkflows(context.Background(), name).InputParameter(inputParameter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.StartBatchWorkflows``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `StartBatchWorkflows`: []StartBatchWorkflows201ResponseInner
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.StartBatchWorkflows`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Workflow identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartBatchWorkflowsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **inputParameter** | [**[][]InputParameter**](array.md) | Contains the definition of the job to be started. | 

### Return type

[**[]StartBatchWorkflows201ResponseInner**](StartBatchWorkflows201ResponseInner.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StartWorkflow

> WorkflowsCheckLiveliness200Response StartWorkflow(ctx, name).InputParameter(inputParameter).Execute()

Start a new workflow



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	name := "name_example" // string | Workflow identifier.
	inputParameter := []openapiclient.InputParameter{*openapiclient.NewInputParameter("Name_example", "Value_example")} // []InputParameter | Contains the definition of the job to be started.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.StartWorkflow(context.Background(), name).InputParameter(inputParameter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.StartWorkflow``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `StartWorkflow`: WorkflowsCheckLiveliness200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.StartWorkflow`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Workflow identifier. | 

### Other Parameters

Other parameters are passed through a pointer to a apiStartWorkflowRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **inputParameter** | [**[]InputParameter**](InputParameter.md) | Contains the definition of the job to be started. | 

### Return type

[**WorkflowsCheckLiveliness200Response**](WorkflowsCheckLiveliness200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WorkflowStatus

> JobStatus WorkflowStatus(ctx, name, id).Execute()

Gets the workflow status for the given id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	name := "name_example" // string | Workflow identifier.
	id := "id_example" // string | Job identifier

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.WorkflowStatus(context.Background(), name, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.WorkflowStatus``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WorkflowStatus`: JobStatus
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.WorkflowStatus`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Workflow identifier. | 
**id** | **string** | Job identifier | 

### Other Parameters

Other parameters are passed through a pointer to a apiWorkflowStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**JobStatus**](JobStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WorkflowsCheckHealth

> WorkflowsCheckHealth(ctx).Execute()

Check if service and all operational dependencies are healthy.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WorkflowsOtherAPI.WorkflowsCheckHealth(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.WorkflowsCheckHealth``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiWorkflowsCheckHealthRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WorkflowsCheckLiveliness

> WorkflowsCheckLiveliness200Response WorkflowsCheckLiveliness(ctx).Execute()

Check if service API is alive and serving requests

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WorkflowsOtherAPI.WorkflowsCheckLiveliness(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WorkflowsOtherAPI.WorkflowsCheckLiveliness``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WorkflowsCheckLiveliness`: WorkflowsCheckLiveliness200Response
	fmt.Fprintf(os.Stdout, "Response from `WorkflowsOtherAPI.WorkflowsCheckLiveliness`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiWorkflowsCheckLivelinessRequest struct via the builder pattern


### Return type

[**WorkflowsCheckLiveliness200Response**](WorkflowsCheckLiveliness200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

