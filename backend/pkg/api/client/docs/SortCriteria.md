# SortCriteria

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attribute** | **string** | Attribute name. | 
**Scope** | [**Scope**](Scope.md) |  | 
**Order** | **string** | Sort order. | 

## Methods

### NewSortCriteria

`func NewSortCriteria(attribute string, scope Scope, order string, ) *SortCriteria`

NewSortCriteria instantiates a new SortCriteria object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSortCriteriaWithDefaults

`func NewSortCriteriaWithDefaults() *SortCriteria`

NewSortCriteriaWithDefaults instantiates a new SortCriteria object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttribute

`func (o *SortCriteria) GetAttribute() string`

GetAttribute returns the Attribute field if non-nil, zero value otherwise.

### GetAttributeOk

`func (o *SortCriteria) GetAttributeOk() (*string, bool)`

GetAttributeOk returns a tuple with the Attribute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttribute

`func (o *SortCriteria) SetAttribute(v string)`

SetAttribute sets Attribute field to given value.


### GetScope

`func (o *SortCriteria) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SortCriteria) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SortCriteria) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetOrder

`func (o *SortCriteria) GetOrder() string`

GetOrder returns the Order field if non-nil, zero value otherwise.

### GetOrderOk

`func (o *SortCriteria) GetOrderOk() (*string, bool)`

GetOrderOk returns a tuple with the Order field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrder

`func (o *SortCriteria) SetOrder(v string)`

SetOrder sets Order field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


