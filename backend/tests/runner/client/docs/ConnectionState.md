# ConnectionState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DeviceId** | Pointer to **string** | Device ID. | [optional] 
**Status** | Pointer to **string** | Device status. | [optional] 
**UpdatedTs** | Pointer to **time.Time** | Server-side timestamp of the last device information update. | [optional] 
**CreatedTs** | Pointer to **time.Time** | Server-side timestamp of the device creation. | [optional] 

## Methods

### NewConnectionState

`func NewConnectionState() *ConnectionState`

NewConnectionState instantiates a new ConnectionState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectionStateWithDefaults

`func NewConnectionStateWithDefaults() *ConnectionState`

NewConnectionStateWithDefaults instantiates a new ConnectionState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeviceId

`func (o *ConnectionState) GetDeviceId() string`

GetDeviceId returns the DeviceId field if non-nil, zero value otherwise.

### GetDeviceIdOk

`func (o *ConnectionState) GetDeviceIdOk() (*string, bool)`

GetDeviceIdOk returns a tuple with the DeviceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceId

`func (o *ConnectionState) SetDeviceId(v string)`

SetDeviceId sets DeviceId field to given value.

### HasDeviceId

`func (o *ConnectionState) HasDeviceId() bool`

HasDeviceId returns a boolean if a field has been set.

### GetStatus

`func (o *ConnectionState) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ConnectionState) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ConnectionState) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ConnectionState) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *ConnectionState) GetUpdatedTs() time.Time`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *ConnectionState) GetUpdatedTsOk() (*time.Time, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *ConnectionState) SetUpdatedTs(v time.Time)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *ConnectionState) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetCreatedTs

`func (o *ConnectionState) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *ConnectionState) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *ConnectionState) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.

### HasCreatedTs

`func (o *ConnectionState) HasCreatedTs() bool`

HasCreatedTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


