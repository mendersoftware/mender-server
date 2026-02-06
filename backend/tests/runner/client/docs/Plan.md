# Plan

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Unique name of the plan.  | 
**DisplayName** | **string** | Short information about the plan.  | 
**Features** | [**Features**](Features.md) |  | 

## Methods

### NewPlan

`func NewPlan(name string, displayName string, features Features, ) *Plan`

NewPlan instantiates a new Plan object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPlanWithDefaults

`func NewPlanWithDefaults() *Plan`

NewPlanWithDefaults instantiates a new Plan object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Plan) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Plan) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Plan) SetName(v string)`

SetName sets Name field to given value.


### GetDisplayName

`func (o *Plan) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *Plan) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *Plan) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.


### GetFeatures

`func (o *Plan) GetFeatures() Features`

GetFeatures returns the Features field if non-nil, zero value otherwise.

### GetFeaturesOk

`func (o *Plan) GetFeaturesOk() (*Features, bool)`

GetFeaturesOk returns a tuple with the Features field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFeatures

`func (o *Plan) SetFeatures(v Features)`

SetFeatures sets Features field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


