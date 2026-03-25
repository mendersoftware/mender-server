# GetStatisticsInternal200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DevicesByStatus** | [**DeviceStatusStatistics**](DeviceStatusStatistics.md) |  | 

## Methods

### NewGetStatisticsInternal200Response

`func NewGetStatisticsInternal200Response(devicesByStatus DeviceStatusStatistics, ) *GetStatisticsInternal200Response`

NewGetStatisticsInternal200Response instantiates a new GetStatisticsInternal200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetStatisticsInternal200ResponseWithDefaults

`func NewGetStatisticsInternal200ResponseWithDefaults() *GetStatisticsInternal200Response`

NewGetStatisticsInternal200ResponseWithDefaults instantiates a new GetStatisticsInternal200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDevicesByStatus

`func (o *GetStatisticsInternal200Response) GetDevicesByStatus() DeviceStatusStatistics`

GetDevicesByStatus returns the DevicesByStatus field if non-nil, zero value otherwise.

### GetDevicesByStatusOk

`func (o *GetStatisticsInternal200Response) GetDevicesByStatusOk() (*DeviceStatusStatistics, bool)`

GetDevicesByStatusOk returns a tuple with the DevicesByStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevicesByStatus

`func (o *GetStatisticsInternal200Response) SetDevicesByStatus(v DeviceStatusStatistics)`

SetDevicesByStatus sets DevicesByStatus field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


