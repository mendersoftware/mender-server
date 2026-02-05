# FilterAttribute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the attribute. | 
**Scope** | [**Scope**](Scope.md) |  | 
**Count** | **int32** | Number of occurrences of the attribute in the database. | 

## Methods

### NewFilterAttribute

`func NewFilterAttribute(name string, scope Scope, count int32, ) *FilterAttribute`

NewFilterAttribute instantiates a new FilterAttribute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFilterAttributeWithDefaults

`func NewFilterAttributeWithDefaults() *FilterAttribute`

NewFilterAttributeWithDefaults instantiates a new FilterAttribute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *FilterAttribute) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FilterAttribute) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FilterAttribute) SetName(v string)`

SetName sets Name field to given value.


### GetScope

`func (o *FilterAttribute) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *FilterAttribute) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *FilterAttribute) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetCount

`func (o *FilterAttribute) GetCount() int32`

GetCount returns the Count field if non-nil, zero value otherwise.

### GetCountOk

`func (o *FilterAttribute) GetCountOk() (*int32, bool)`

GetCountOk returns a tuple with the Count field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCount

`func (o *FilterAttribute) SetCount(v int32)`

SetCount sets Count field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


