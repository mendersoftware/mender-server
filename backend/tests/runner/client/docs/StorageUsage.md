# StorageUsage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Limit** | **int32** | Storage limit in bytes. If set to 0 - there is no limit for storage.  | 
**Usage** | **int32** | Current storage usage in bytes.  | 

## Methods

### NewStorageUsage

`func NewStorageUsage(limit int32, usage int32, ) *StorageUsage`

NewStorageUsage instantiates a new StorageUsage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageUsageWithDefaults

`func NewStorageUsageWithDefaults() *StorageUsage`

NewStorageUsageWithDefaults instantiates a new StorageUsage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLimit

`func (o *StorageUsage) GetLimit() int32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *StorageUsage) GetLimitOk() (*int32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *StorageUsage) SetLimit(v int32)`

SetLimit sets Limit field to given value.


### GetUsage

`func (o *StorageUsage) GetUsage() int32`

GetUsage returns the Usage field if non-nil, zero value otherwise.

### GetUsageOk

`func (o *StorageUsage) GetUsageOk() (*int32, bool)`

GetUsageOk returns a tuple with the Usage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsage

`func (o *StorageUsage) SetUsage(v int32)`

SetUsage sets Usage field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


