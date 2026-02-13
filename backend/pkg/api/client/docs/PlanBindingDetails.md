# PlanBindingDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Plan** | Pointer to [**Plan**](Plan.md) |  | [optional] 
**Limits** | Pointer to [**PlanLimits**](PlanLimits.md) |  | [optional] 

## Methods

### NewPlanBindingDetails

`func NewPlanBindingDetails() *PlanBindingDetails`

NewPlanBindingDetails instantiates a new PlanBindingDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPlanBindingDetailsWithDefaults

`func NewPlanBindingDetailsWithDefaults() *PlanBindingDetails`

NewPlanBindingDetailsWithDefaults instantiates a new PlanBindingDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPlan

`func (o *PlanBindingDetails) GetPlan() Plan`

GetPlan returns the Plan field if non-nil, zero value otherwise.

### GetPlanOk

`func (o *PlanBindingDetails) GetPlanOk() (*Plan, bool)`

GetPlanOk returns a tuple with the Plan field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlan

`func (o *PlanBindingDetails) SetPlan(v Plan)`

SetPlan sets Plan field to given value.

### HasPlan

`func (o *PlanBindingDetails) HasPlan() bool`

HasPlan returns a boolean if a field has been set.

### GetLimits

`func (o *PlanBindingDetails) GetLimits() PlanLimits`

GetLimits returns the Limits field if non-nil, zero value otherwise.

### GetLimitsOk

`func (o *PlanBindingDetails) GetLimitsOk() (*PlanLimits, bool)`

GetLimitsOk returns a tuple with the Limits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimits

`func (o *PlanBindingDetails) SetLimits(v PlanLimits)`

SetLimits sets Limits field to given value.

### HasLimits

`func (o *PlanBindingDetails) HasLimits() bool`

HasLimits returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


