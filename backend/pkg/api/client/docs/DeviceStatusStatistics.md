# DeviceStatusStatistics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Accepted** | [**DeviceCountByTier**](DeviceCountByTier.md) |  | 
**Pending** | [**DeviceCountByTier**](DeviceCountByTier.md) |  | 

## Methods

### NewDeviceStatusStatistics

`func NewDeviceStatusStatistics(accepted DeviceCountByTier, pending DeviceCountByTier, ) *DeviceStatusStatistics`

NewDeviceStatusStatistics instantiates a new DeviceStatusStatistics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceStatusStatisticsWithDefaults

`func NewDeviceStatusStatisticsWithDefaults() *DeviceStatusStatistics`

NewDeviceStatusStatisticsWithDefaults instantiates a new DeviceStatusStatistics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccepted

`func (o *DeviceStatusStatistics) GetAccepted() DeviceCountByTier`

GetAccepted returns the Accepted field if non-nil, zero value otherwise.

### GetAcceptedOk

`func (o *DeviceStatusStatistics) GetAcceptedOk() (*DeviceCountByTier, bool)`

GetAcceptedOk returns a tuple with the Accepted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccepted

`func (o *DeviceStatusStatistics) SetAccepted(v DeviceCountByTier)`

SetAccepted sets Accepted field to given value.


### GetPending

`func (o *DeviceStatusStatistics) GetPending() DeviceCountByTier`

GetPending returns the Pending field if non-nil, zero value otherwise.

### GetPendingOk

`func (o *DeviceStatusStatistics) GetPendingOk() (*DeviceCountByTier, bool)`

GetPendingOk returns a tuple with the Pending field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPending

`func (o *DeviceStatusStatistics) SetPending(v DeviceCountByTier)`

SetPending sets Pending field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


