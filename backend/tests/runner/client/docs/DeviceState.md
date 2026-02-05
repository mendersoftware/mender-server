# DeviceState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Desired** | Pointer to **map[string]interface{}** | The desired state for the device, as reported by the cloud/user. | [optional] 
**Reported** | Pointer to **map[string]interface{}** | State reported by the device, this cannot be changed from the cloud. | [optional] 

## Methods

### NewDeviceState

`func NewDeviceState() *DeviceState`

NewDeviceState instantiates a new DeviceState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceStateWithDefaults

`func NewDeviceStateWithDefaults() *DeviceState`

NewDeviceStateWithDefaults instantiates a new DeviceState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDesired

`func (o *DeviceState) GetDesired() map[string]interface{}`

GetDesired returns the Desired field if non-nil, zero value otherwise.

### GetDesiredOk

`func (o *DeviceState) GetDesiredOk() (*map[string]interface{}, bool)`

GetDesiredOk returns a tuple with the Desired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesired

`func (o *DeviceState) SetDesired(v map[string]interface{})`

SetDesired sets Desired field to given value.

### HasDesired

`func (o *DeviceState) HasDesired() bool`

HasDesired returns a boolean if a field has been set.

### GetReported

`func (o *DeviceState) GetReported() map[string]interface{}`

GetReported returns the Reported field if non-nil, zero value otherwise.

### GetReportedOk

`func (o *DeviceState) GetReportedOk() (*map[string]interface{}, bool)`

GetReportedOk returns a tuple with the Reported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReported

`func (o *DeviceState) SetReported(v map[string]interface{})`

SetReported sets Reported field to given value.

### HasReported

`func (o *DeviceState) HasReported() bool`

HasReported returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


