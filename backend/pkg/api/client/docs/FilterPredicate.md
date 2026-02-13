# FilterPredicate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | [**Scope**](Scope.md) |  | 
**Attribute** | **string** | Name of the attribute to be queried for filtering.  | 
**Type** | **string** | Type or operator of the filter predicate. | 
**Value** | **string** | The value of the attribute to be used in filtering.  Attribute type is implicit, inferred from the JSON type.  Supported types: number, string, array of numbers, array of strings. Mixed arrays are not allowed.  The $exists operator expects a boolean value: true means the specified attribute exists, false means the specified attribute doesn&#39;t exist.  The $regex operator expects a string as a Perl compatible regular expression (PCRE), automatically anchored by ^. If the regular expression is not valid, the filter will produce no results. If you need to specify options and flags, you can provide the full regex in the format of /regex/flags, for example &#x60;/[a-z]+/i&#x60;.  | 

## Methods

### NewFilterPredicate

`func NewFilterPredicate(scope Scope, attribute string, type_ string, value string, ) *FilterPredicate`

NewFilterPredicate instantiates a new FilterPredicate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFilterPredicateWithDefaults

`func NewFilterPredicateWithDefaults() *FilterPredicate`

NewFilterPredicateWithDefaults instantiates a new FilterPredicate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *FilterPredicate) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *FilterPredicate) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *FilterPredicate) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetAttribute

`func (o *FilterPredicate) GetAttribute() string`

GetAttribute returns the Attribute field if non-nil, zero value otherwise.

### GetAttributeOk

`func (o *FilterPredicate) GetAttributeOk() (*string, bool)`

GetAttributeOk returns a tuple with the Attribute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttribute

`func (o *FilterPredicate) SetAttribute(v string)`

SetAttribute sets Attribute field to given value.


### GetType

`func (o *FilterPredicate) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FilterPredicate) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FilterPredicate) SetType(v string)`

SetType sets Type field to given value.


### GetValue

`func (o *FilterPredicate) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *FilterPredicate) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *FilterPredicate) SetValue(v string)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


