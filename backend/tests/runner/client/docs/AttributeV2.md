# AttributeV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | A human readable, unique attribute ID, e.g. &#39;device_type&#39;, &#39;ip_addr&#39;, &#39;cpu_load&#39;, etc.  | 
**Scope** | [**Scope**](Scope.md) |  | 
**Description** | Pointer to **string** | Attribute description. | [optional] 
**Value** | **string** | The current value of the attribute.  Attribute type is implicit, inferred from the JSON type.  Supported types: number, string, array of numbers, array of strings. Mixed arrays are not allowed.  | 

## Methods

### NewAttributeV2

`func NewAttributeV2(name string, scope Scope, value string, ) *AttributeV2`

NewAttributeV2 instantiates a new AttributeV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAttributeV2WithDefaults

`func NewAttributeV2WithDefaults() *AttributeV2`

NewAttributeV2WithDefaults instantiates a new AttributeV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AttributeV2) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AttributeV2) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AttributeV2) SetName(v string)`

SetName sets Name field to given value.


### GetScope

`func (o *AttributeV2) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AttributeV2) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AttributeV2) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetDescription

`func (o *AttributeV2) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AttributeV2) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AttributeV2) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AttributeV2) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetValue

`func (o *AttributeV2) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *AttributeV2) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *AttributeV2) SetValue(v string)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


