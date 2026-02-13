# EventDeliveryStatusesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IntegrationId** | **string** | The ID of the integration the status belongs. | 
**Success** | **bool** | Whether the event hook was executed successfully. | 
**StatusCode** | Pointer to **int32** | The (HTTP) status code of the hook. | [optional] 
**Error** | Pointer to **string** | An error message if the hook failed. | [optional] 

## Methods

### NewEventDeliveryStatusesInner

`func NewEventDeliveryStatusesInner(integrationId string, success bool, ) *EventDeliveryStatusesInner`

NewEventDeliveryStatusesInner instantiates a new EventDeliveryStatusesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventDeliveryStatusesInnerWithDefaults

`func NewEventDeliveryStatusesInnerWithDefaults() *EventDeliveryStatusesInner`

NewEventDeliveryStatusesInnerWithDefaults instantiates a new EventDeliveryStatusesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIntegrationId

`func (o *EventDeliveryStatusesInner) GetIntegrationId() string`

GetIntegrationId returns the IntegrationId field if non-nil, zero value otherwise.

### GetIntegrationIdOk

`func (o *EventDeliveryStatusesInner) GetIntegrationIdOk() (*string, bool)`

GetIntegrationIdOk returns a tuple with the IntegrationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIntegrationId

`func (o *EventDeliveryStatusesInner) SetIntegrationId(v string)`

SetIntegrationId sets IntegrationId field to given value.


### GetSuccess

`func (o *EventDeliveryStatusesInner) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *EventDeliveryStatusesInner) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *EventDeliveryStatusesInner) SetSuccess(v bool)`

SetSuccess sets Success field to given value.


### GetStatusCode

`func (o *EventDeliveryStatusesInner) GetStatusCode() int32`

GetStatusCode returns the StatusCode field if non-nil, zero value otherwise.

### GetStatusCodeOk

`func (o *EventDeliveryStatusesInner) GetStatusCodeOk() (*int32, bool)`

GetStatusCodeOk returns a tuple with the StatusCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusCode

`func (o *EventDeliveryStatusesInner) SetStatusCode(v int32)`

SetStatusCode sets StatusCode field to given value.

### HasStatusCode

`func (o *EventDeliveryStatusesInner) HasStatusCode() bool`

HasStatusCode returns a boolean if a field has been set.

### GetError

`func (o *EventDeliveryStatusesInner) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *EventDeliveryStatusesInner) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *EventDeliveryStatusesInner) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *EventDeliveryStatusesInner) HasError() bool`

HasError returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


