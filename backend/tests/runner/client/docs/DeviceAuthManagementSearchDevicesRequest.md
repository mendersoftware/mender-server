# DeviceAuthManagementSearchDevicesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to **[]string** | Device status filter. Can be an array for querying devices from multiple device statuses. | [optional] 
**Id** | Pointer to **[]string** | Device ID filter. Can be a string for querying for a single device. | [optional] 

## Methods

### NewDeviceAuthManagementSearchDevicesRequest

`func NewDeviceAuthManagementSearchDevicesRequest() *DeviceAuthManagementSearchDevicesRequest`

NewDeviceAuthManagementSearchDevicesRequest instantiates a new DeviceAuthManagementSearchDevicesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceAuthManagementSearchDevicesRequestWithDefaults

`func NewDeviceAuthManagementSearchDevicesRequestWithDefaults() *DeviceAuthManagementSearchDevicesRequest`

NewDeviceAuthManagementSearchDevicesRequestWithDefaults instantiates a new DeviceAuthManagementSearchDevicesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *DeviceAuthManagementSearchDevicesRequest) GetStatus() []string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeviceAuthManagementSearchDevicesRequest) GetStatusOk() (*[]string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeviceAuthManagementSearchDevicesRequest) SetStatus(v []string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DeviceAuthManagementSearchDevicesRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetId

`func (o *DeviceAuthManagementSearchDevicesRequest) GetId() []string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceAuthManagementSearchDevicesRequest) GetIdOk() (*[]string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceAuthManagementSearchDevicesRequest) SetId(v []string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceAuthManagementSearchDevicesRequest) HasId() bool`

HasId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


