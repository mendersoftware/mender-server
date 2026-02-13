# Device

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Mender assigned Device ID. | [optional] 
**IdentityData** | Pointer to [**IdentityData**](IdentityData.md) |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**CreatedTs** | Pointer to **time.Time** | Created timestamp | [optional] 
**UpdatedTs** | Pointer to **time.Time** | Updated timestamp | [optional] 
**CheckInTime** | Pointer to **time.Time** | Time when accepted device contacted server for the last time. | [optional] 
**AuthSets** | Pointer to [**[]AuthSet**](AuthSet.md) |  | [optional] 
**Decommissioning** | Pointer to **bool** | Devices that are part of ongoing decomissioning process will return True | [optional] 

## Methods

### NewDevice

`func NewDevice() *Device`

NewDevice instantiates a new Device object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithDefaults

`func NewDeviceWithDefaults() *Device`

NewDeviceWithDefaults instantiates a new Device object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Device) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Device) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Device) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Device) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIdentityData

`func (o *Device) GetIdentityData() IdentityData`

GetIdentityData returns the IdentityData field if non-nil, zero value otherwise.

### GetIdentityDataOk

`func (o *Device) GetIdentityDataOk() (*IdentityData, bool)`

GetIdentityDataOk returns a tuple with the IdentityData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdentityData

`func (o *Device) SetIdentityData(v IdentityData)`

SetIdentityData sets IdentityData field to given value.

### HasIdentityData

`func (o *Device) HasIdentityData() bool`

HasIdentityData returns a boolean if a field has been set.

### GetStatus

`func (o *Device) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Device) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Device) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Device) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedTs

`func (o *Device) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *Device) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *Device) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.

### HasCreatedTs

`func (o *Device) HasCreatedTs() bool`

HasCreatedTs returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *Device) GetUpdatedTs() time.Time`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *Device) GetUpdatedTsOk() (*time.Time, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *Device) SetUpdatedTs(v time.Time)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *Device) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.

### GetCheckInTime

`func (o *Device) GetCheckInTime() time.Time`

GetCheckInTime returns the CheckInTime field if non-nil, zero value otherwise.

### GetCheckInTimeOk

`func (o *Device) GetCheckInTimeOk() (*time.Time, bool)`

GetCheckInTimeOk returns a tuple with the CheckInTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCheckInTime

`func (o *Device) SetCheckInTime(v time.Time)`

SetCheckInTime sets CheckInTime field to given value.

### HasCheckInTime

`func (o *Device) HasCheckInTime() bool`

HasCheckInTime returns a boolean if a field has been set.

### GetAuthSets

`func (o *Device) GetAuthSets() []AuthSet`

GetAuthSets returns the AuthSets field if non-nil, zero value otherwise.

### GetAuthSetsOk

`func (o *Device) GetAuthSetsOk() (*[]AuthSet, bool)`

GetAuthSetsOk returns a tuple with the AuthSets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthSets

`func (o *Device) SetAuthSets(v []AuthSet)`

SetAuthSets sets AuthSets field to given value.

### HasAuthSets

`func (o *Device) HasAuthSets() bool`

HasAuthSets returns a boolean if a field has been set.

### GetDecommissioning

`func (o *Device) GetDecommissioning() bool`

GetDecommissioning returns the Decommissioning field if non-nil, zero value otherwise.

### GetDecommissioningOk

`func (o *Device) GetDecommissioningOk() (*bool, bool)`

GetDecommissioningOk returns a tuple with the Decommissioning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDecommissioning

`func (o *Device) SetDecommissioning(v bool)`

SetDecommissioning sets Decommissioning field to given value.

### HasDecommissioning

`func (o *Device) HasDecommissioning() bool`

HasDecommissioning returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


