# GetStatistics200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DevicesByStatus** | [**DeviceStatusStatistics**](DeviceStatusStatistics.md) |  | 

## Methods

### NewGetStatistics200Response

`func NewGetStatistics200Response(devicesByStatus DeviceStatusStatistics, ) *GetStatistics200Response`

NewGetStatistics200Response instantiates a new GetStatistics200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetStatistics200ResponseWithDefaults

`func NewGetStatistics200ResponseWithDefaults() *GetStatistics200Response`

NewGetStatistics200ResponseWithDefaults instantiates a new GetStatistics200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDevicesByStatus

`func (o *GetStatistics200Response) GetDevicesByStatus() DeviceStatusStatistics`

GetDevicesByStatus returns the DevicesByStatus field if non-nil, zero value otherwise.

### GetDevicesByStatusOk

`func (o *GetStatistics200Response) GetDevicesByStatusOk() (*DeviceStatusStatistics, bool)`

GetDevicesByStatusOk returns a tuple with the DevicesByStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevicesByStatus

`func (o *GetStatistics200Response) SetDevicesByStatus(v DeviceStatusStatistics)`

SetDevicesByStatus sets DevicesByStatus field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


