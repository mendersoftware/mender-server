# InventoryInternalV2SearchDeviceInventoriesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **int32** | Starting page. | [optional] 
**PerPage** | Pointer to **int32** | Number of results per page. | [optional] 
**DeviceIds** | Pointer to **[]string** | List of device IDs | [optional] 
**Text** | Pointer to **string** | Free-text search query | [optional] 
**Filters** | Pointer to [**[]FilterPredicate**](FilterPredicate.md) | List of filter predicates, chained with boolean AND operators to build the search condition definition. | [optional] 
**Sort** | Pointer to [**[]SortCriteria**](SortCriteria.md) | List of ordered sort criterias | [optional] 
**Attributes** | Pointer to [**[]SelectAttribute**](SelectAttribute.md) | List of attributes to select and return | [optional] 

## Methods

### NewInventoryInternalV2SearchDeviceInventoriesRequest

`func NewInventoryInternalV2SearchDeviceInventoriesRequest() *InventoryInternalV2SearchDeviceInventoriesRequest`

NewInventoryInternalV2SearchDeviceInventoriesRequest instantiates a new InventoryInternalV2SearchDeviceInventoriesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInventoryInternalV2SearchDeviceInventoriesRequestWithDefaults

`func NewInventoryInternalV2SearchDeviceInventoriesRequestWithDefaults() *InventoryInternalV2SearchDeviceInventoriesRequest`

NewInventoryInternalV2SearchDeviceInventoriesRequestWithDefaults instantiates a new InventoryInternalV2SearchDeviceInventoriesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetPerPage() int32`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetPerPageOk() (*int32, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetPerPage(v int32)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetDeviceIds

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetDeviceIds() []string`

GetDeviceIds returns the DeviceIds field if non-nil, zero value otherwise.

### GetDeviceIdsOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetDeviceIdsOk() (*[]string, bool)`

GetDeviceIdsOk returns a tuple with the DeviceIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceIds

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetDeviceIds(v []string)`

SetDeviceIds sets DeviceIds field to given value.

### HasDeviceIds

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasDeviceIds() bool`

HasDeviceIds returns a boolean if a field has been set.

### GetText

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetText(v string)`

SetText sets Text field to given value.

### HasText

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasText() bool`

HasText returns a boolean if a field has been set.

### GetFilters

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetFilters() []FilterPredicate`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetFiltersOk() (*[]FilterPredicate, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetFilters(v []FilterPredicate)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasFilters() bool`

HasFilters returns a boolean if a field has been set.

### GetSort

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetSort() []SortCriteria`

GetSort returns the Sort field if non-nil, zero value otherwise.

### GetSortOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetSortOk() (*[]SortCriteria, bool)`

GetSortOk returns a tuple with the Sort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSort

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetSort(v []SortCriteria)`

SetSort sets Sort field to given value.

### HasSort

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasSort() bool`

HasSort returns a boolean if a field has been set.

### GetAttributes

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetAttributes() []SelectAttribute`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) GetAttributesOk() (*[]SelectAttribute, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) SetAttributes(v []SelectAttribute)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *InventoryInternalV2SearchDeviceInventoriesRequest) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


