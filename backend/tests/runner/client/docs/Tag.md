# Tag

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Tag is an attribute with &#39;tags&#39; scope.  A human readable, unique tag ID, e.g. &#39;location&#39;, &#39;environment&#39;, etc.  | 
**Description** | Pointer to **string** | Tag description. | [optional] 
**Value** | **string** | The current value of the tag. | 
**Timestamp** | Pointer to **time.Time** | The date and time of last tag update in RFC3339 format.  | [optional] 

## Methods

### NewTag

`func NewTag(name string, value string, ) *Tag`

NewTag instantiates a new Tag object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTagWithDefaults

`func NewTagWithDefaults() *Tag`

NewTagWithDefaults instantiates a new Tag object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Tag) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Tag) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Tag) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Tag) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Tag) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Tag) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Tag) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetValue

`func (o *Tag) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Tag) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Tag) SetValue(v string)`

SetValue sets Value field to given value.


### GetTimestamp

`func (o *Tag) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *Tag) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *Tag) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *Tag) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


