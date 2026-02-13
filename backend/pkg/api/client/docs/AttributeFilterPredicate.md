# AttributeFilterPredicate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | [**Scope**](Scope.md) |  | 
**Attribute** | **string** | Name of the attribute to be queried for filtering.  | 
**Type** | **string** | Type or operator of the filter predicate. | 
**Value** | **interface{}** | The value of the attribute to be used in filtering. Attribute type is implicit, inferred from the JSON type. Supported types: number, string, array of numbers, array of strings. Mixed arrays are not allowed.  | 

## Methods

### NewAttributeFilterPredicate

`func NewAttributeFilterPredicate(scope Scope, attribute string, type_ string, value interface{}, ) *AttributeFilterPredicate`

NewAttributeFilterPredicate instantiates a new AttributeFilterPredicate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAttributeFilterPredicateWithDefaults

`func NewAttributeFilterPredicateWithDefaults() *AttributeFilterPredicate`

NewAttributeFilterPredicateWithDefaults instantiates a new AttributeFilterPredicate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *AttributeFilterPredicate) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AttributeFilterPredicate) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AttributeFilterPredicate) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetAttribute

`func (o *AttributeFilterPredicate) GetAttribute() string`

GetAttribute returns the Attribute field if non-nil, zero value otherwise.

### GetAttributeOk

`func (o *AttributeFilterPredicate) GetAttributeOk() (*string, bool)`

GetAttributeOk returns a tuple with the Attribute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttribute

`func (o *AttributeFilterPredicate) SetAttribute(v string)`

SetAttribute sets Attribute field to given value.


### GetType

`func (o *AttributeFilterPredicate) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AttributeFilterPredicate) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AttributeFilterPredicate) SetType(v string)`

SetType sets Type field to given value.


### GetValue

`func (o *AttributeFilterPredicate) GetValue() interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *AttributeFilterPredicate) GetValueOk() (*interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *AttributeFilterPredicate) SetValue(v interface{})`

SetValue sets Value field to given value.


### SetValueNil

`func (o *AttributeFilterPredicate) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *AttributeFilterPredicate) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


