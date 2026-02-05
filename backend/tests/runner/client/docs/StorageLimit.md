# StorageLimit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Limit** | **int32** | Storage limit in bytes. If set to 0 - there is no limit for storage.  | 
**Usage** | **int32** | Current storage usage in bytes.  | 

## Methods

### NewStorageLimit

`func NewStorageLimit(limit int32, usage int32, ) *StorageLimit`

NewStorageLimit instantiates a new StorageLimit object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageLimitWithDefaults

`func NewStorageLimitWithDefaults() *StorageLimit`

NewStorageLimitWithDefaults instantiates a new StorageLimit object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLimit

`func (o *StorageLimit) GetLimit() int32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *StorageLimit) GetLimitOk() (*int32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *StorageLimit) SetLimit(v int32)`

SetLimit sets Limit field to given value.


### GetUsage

`func (o *StorageLimit) GetUsage() int32`

GetUsage returns the Usage field if non-nil, zero value otherwise.

### GetUsageOk

`func (o *StorageLimit) GetUsageOk() (*int32, bool)`

GetUsageOk returns a tuple with the Usage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsage

`func (o *StorageLimit) SetUsage(v int32)`

SetUsage sets Usage field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


