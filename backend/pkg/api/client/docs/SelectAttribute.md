# SelectAttribute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attribute** | **string** | Attribute name. | 
**Scope** | [**Scope**](Scope.md) |  | 

## Methods

### NewSelectAttribute

`func NewSelectAttribute(attribute string, scope Scope, ) *SelectAttribute`

NewSelectAttribute instantiates a new SelectAttribute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSelectAttributeWithDefaults

`func NewSelectAttributeWithDefaults() *SelectAttribute`

NewSelectAttributeWithDefaults instantiates a new SelectAttribute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttribute

`func (o *SelectAttribute) GetAttribute() string`

GetAttribute returns the Attribute field if non-nil, zero value otherwise.

### GetAttributeOk

`func (o *SelectAttribute) GetAttributeOk() (*string, bool)`

GetAttributeOk returns a tuple with the Attribute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttribute

`func (o *SelectAttribute) SetAttribute(v string)`

SetAttribute sets Attribute field to given value.


### GetScope

`func (o *SelectAttribute) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SelectAttribute) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SelectAttribute) SetScope(v Scope)`

SetScope sets Scope field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


