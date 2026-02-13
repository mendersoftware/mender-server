# AttributeV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | A human readable, unique attribute ID, e.g. &#39;device_type&#39;, &#39;ip_addr&#39;, &#39;cpu_load&#39;, etc.  | 
**Scope** | **string** | The scope of the attribute.  Scope is a string and acts as namespace for the attribute name.  | 
**Description** | Pointer to **string** | Attribute description. | [optional] 
**Value** | [**AttributeValue**](AttributeValue.md) |  | 
**Timestamp** | Pointer to **time.Time** | The date and time of last tag update in RFC3339 format.  | [optional] 

## Methods

### NewAttributeV1

`func NewAttributeV1(name string, scope string, value AttributeValue, ) *AttributeV1`

NewAttributeV1 instantiates a new AttributeV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAttributeV1WithDefaults

`func NewAttributeV1WithDefaults() *AttributeV1`

NewAttributeV1WithDefaults instantiates a new AttributeV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AttributeV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AttributeV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AttributeV1) SetName(v string)`

SetName sets Name field to given value.


### GetScope

`func (o *AttributeV1) GetScope() string`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AttributeV1) GetScopeOk() (*string, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AttributeV1) SetScope(v string)`

SetScope sets Scope field to given value.


### GetDescription

`func (o *AttributeV1) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AttributeV1) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AttributeV1) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AttributeV1) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetValue

`func (o *AttributeV1) GetValue() AttributeValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *AttributeV1) GetValueOk() (*AttributeValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *AttributeV1) SetValue(v AttributeValue)`

SetValue sets Value field to given value.


### GetTimestamp

`func (o *AttributeV1) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *AttributeV1) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *AttributeV1) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *AttributeV1) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


