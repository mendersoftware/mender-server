# DeploymentLogMessagesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Timestamp** | **time.Time** |  | 
**Level** | **string** |  | 
**Message** | **string** |  | 

## Methods

### NewDeploymentLogMessagesInner

`func NewDeploymentLogMessagesInner(timestamp time.Time, level string, message string, ) *DeploymentLogMessagesInner`

NewDeploymentLogMessagesInner instantiates a new DeploymentLogMessagesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentLogMessagesInnerWithDefaults

`func NewDeploymentLogMessagesInnerWithDefaults() *DeploymentLogMessagesInner`

NewDeploymentLogMessagesInnerWithDefaults instantiates a new DeploymentLogMessagesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTimestamp

`func (o *DeploymentLogMessagesInner) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *DeploymentLogMessagesInner) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *DeploymentLogMessagesInner) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.


### GetLevel

`func (o *DeploymentLogMessagesInner) GetLevel() string`

GetLevel returns the Level field if non-nil, zero value otherwise.

### GetLevelOk

`func (o *DeploymentLogMessagesInner) GetLevelOk() (*string, bool)`

GetLevelOk returns a tuple with the Level field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevel

`func (o *DeploymentLogMessagesInner) SetLevel(v string)`

SetLevel sets Level field to given value.


### GetMessage

`func (o *DeploymentLogMessagesInner) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *DeploymentLogMessagesInner) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *DeploymentLogMessagesInner) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


