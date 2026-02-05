# FilterV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the saved filter.  | 
**Name** | **string** | Name of the saved filter.  | 
**Terms** | Pointer to [**[]AttributeFilterPredicate**](AttributeFilterPredicate.md) |  | [optional] 

## Methods

### NewFilterV2

`func NewFilterV2(id string, name string, ) *FilterV2`

NewFilterV2 instantiates a new FilterV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFilterV2WithDefaults

`func NewFilterV2WithDefaults() *FilterV2`

NewFilterV2WithDefaults instantiates a new FilterV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *FilterV2) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FilterV2) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FilterV2) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *FilterV2) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *FilterV2) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *FilterV2) SetName(v string)`

SetName sets Name field to given value.


### GetTerms

`func (o *FilterV2) GetTerms() []AttributeFilterPredicate`

GetTerms returns the Terms field if non-nil, zero value otherwise.

### GetTermsOk

`func (o *FilterV2) GetTermsOk() (*[]AttributeFilterPredicate, bool)`

GetTermsOk returns a tuple with the Terms field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerms

`func (o *FilterV2) SetTerms(v []AttributeFilterPredicate)`

SetTerms sets Terms field to given value.

### HasTerms

`func (o *FilterV2) HasTerms() bool`

HasTerms returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


