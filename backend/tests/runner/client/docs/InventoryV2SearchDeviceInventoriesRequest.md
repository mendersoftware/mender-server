# InventoryV2SearchDeviceInventoriesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **int32** | Starting page. | [optional] 
**PerPage** | Pointer to **int32** | Maximum number of results per page. | [optional] 
**Text** | Pointer to **string** | Free-text search query | [optional] 
**Filters** | Pointer to [**[]FilterPredicate**](FilterPredicate.md) | List of filter predicates. | [optional] 
**Sort** | Pointer to [**[]SortCriteria**](SortCriteria.md) | List of ordered sort criterias | [optional] 
**Attributes** | Pointer to [**[]SelectAttribute**](SelectAttribute.md) | List of attributes to select and return | [optional] 

## Methods

### NewInventoryV2SearchDeviceInventoriesRequest

`func NewInventoryV2SearchDeviceInventoriesRequest() *InventoryV2SearchDeviceInventoriesRequest`

NewInventoryV2SearchDeviceInventoriesRequest instantiates a new InventoryV2SearchDeviceInventoriesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInventoryV2SearchDeviceInventoriesRequestWithDefaults

`func NewInventoryV2SearchDeviceInventoriesRequestWithDefaults() *InventoryV2SearchDeviceInventoriesRequest`

NewInventoryV2SearchDeviceInventoriesRequestWithDefaults instantiates a new InventoryV2SearchDeviceInventoriesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetPerPage() int32`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetPerPageOk() (*int32, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetPerPage(v int32)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetText

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetText(v string)`

SetText sets Text field to given value.

### HasText

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasText() bool`

HasText returns a boolean if a field has been set.

### GetFilters

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetFilters() []FilterPredicate`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetFiltersOk() (*[]FilterPredicate, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetFilters(v []FilterPredicate)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasFilters() bool`

HasFilters returns a boolean if a field has been set.

### GetSort

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetSort() []SortCriteria`

GetSort returns the Sort field if non-nil, zero value otherwise.

### GetSortOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetSortOk() (*[]SortCriteria, bool)`

GetSortOk returns a tuple with the Sort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSort

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetSort(v []SortCriteria)`

SetSort sets Sort field to given value.

### HasSort

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasSort() bool`

HasSort returns a boolean if a field has been set.

### GetAttributes

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetAttributes() []SelectAttribute`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *InventoryV2SearchDeviceInventoriesRequest) GetAttributesOk() (*[]SelectAttribute, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *InventoryV2SearchDeviceInventoriesRequest) SetAttributes(v []SelectAttribute)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *InventoryV2SearchDeviceInventoriesRequest) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


