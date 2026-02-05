# Attribute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | A human readable, unique attribute ID, e.g. &#39;device_type&#39;, &#39;ip_addr&#39;, &#39;cpu_load&#39;, etc.  | 
**Description** | Pointer to **string** | Attribute description. | [optional] 
**Value** | **string** | The current value of the attribute.  Attribute type is implicit, inferred from the JSON type.  Supported types: number, string, array of numbers, array of strings. Mixed type arrays are not allowed.  | 

## Methods

### NewAttribute

`func NewAttribute(name string, value string, ) *Attribute`

NewAttribute instantiates a new Attribute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAttributeWithDefaults

`func NewAttributeWithDefaults() *Attribute`

NewAttributeWithDefaults instantiates a new Attribute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Attribute) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Attribute) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Attribute) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Attribute) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Attribute) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Attribute) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Attribute) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetValue

`func (o *Attribute) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Attribute) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Attribute) SetValue(v string)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


