# DeviceWithImage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Device identifier. | 
**Status** | [**DeviceStatus**](DeviceStatus.md) |  | 
**Created** | Pointer to **time.Time** |  | [optional] 
**Started** | Pointer to **time.Time** |  | [optional] 
**Finished** | Pointer to **time.Time** |  | [optional] 
**Deleted** | Pointer to **time.Time** |  | [optional] 
**DeviceType** | Pointer to **string** |  | [optional] 
**Log** | **bool** | Availability of the device&#39;s deployment log. | 
**State** | Pointer to **string** | State reported by device | [optional] 
**Substate** | Pointer to **string** | Additional state information | [optional] 
**Image** | Pointer to [**DeviceWithImageImage**](DeviceWithImageImage.md) |  | [optional] 

## Methods

### NewDeviceWithImage

`func NewDeviceWithImage(id string, status DeviceStatus, log bool, ) *DeviceWithImage`

NewDeviceWithImage instantiates a new DeviceWithImage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceWithImageWithDefaults

`func NewDeviceWithImageWithDefaults() *DeviceWithImage`

NewDeviceWithImageWithDefaults instantiates a new DeviceWithImage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceWithImage) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceWithImage) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceWithImage) SetId(v string)`

SetId sets Id field to given value.


### GetStatus

`func (o *DeviceWithImage) GetStatus() DeviceStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DeviceWithImage) GetStatusOk() (*DeviceStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DeviceWithImage) SetStatus(v DeviceStatus)`

SetStatus sets Status field to given value.


### GetCreated

`func (o *DeviceWithImage) GetCreated() time.Time`

GetCreated returns the Created field if non-nil, zero value otherwise.

### GetCreatedOk

`func (o *DeviceWithImage) GetCreatedOk() (*time.Time, bool)`

GetCreatedOk returns a tuple with the Created field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreated

`func (o *DeviceWithImage) SetCreated(v time.Time)`

SetCreated sets Created field to given value.

### HasCreated

`func (o *DeviceWithImage) HasCreated() bool`

HasCreated returns a boolean if a field has been set.

### GetStarted

`func (o *DeviceWithImage) GetStarted() time.Time`

GetStarted returns the Started field if non-nil, zero value otherwise.

### GetStartedOk

`func (o *DeviceWithImage) GetStartedOk() (*time.Time, bool)`

GetStartedOk returns a tuple with the Started field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStarted

`func (o *DeviceWithImage) SetStarted(v time.Time)`

SetStarted sets Started field to given value.

### HasStarted

`func (o *DeviceWithImage) HasStarted() bool`

HasStarted returns a boolean if a field has been set.

### GetFinished

`func (o *DeviceWithImage) GetFinished() time.Time`

GetFinished returns the Finished field if non-nil, zero value otherwise.

### GetFinishedOk

`func (o *DeviceWithImage) GetFinishedOk() (*time.Time, bool)`

GetFinishedOk returns a tuple with the Finished field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFinished

`func (o *DeviceWithImage) SetFinished(v time.Time)`

SetFinished sets Finished field to given value.

### HasFinished

`func (o *DeviceWithImage) HasFinished() bool`

HasFinished returns a boolean if a field has been set.

### GetDeleted

`func (o *DeviceWithImage) GetDeleted() time.Time`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *DeviceWithImage) GetDeletedOk() (*time.Time, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *DeviceWithImage) SetDeleted(v time.Time)`

SetDeleted sets Deleted field to given value.

### HasDeleted

`func (o *DeviceWithImage) HasDeleted() bool`

HasDeleted returns a boolean if a field has been set.

### GetDeviceType

`func (o *DeviceWithImage) GetDeviceType() string`

GetDeviceType returns the DeviceType field if non-nil, zero value otherwise.

### GetDeviceTypeOk

`func (o *DeviceWithImage) GetDeviceTypeOk() (*string, bool)`

GetDeviceTypeOk returns a tuple with the DeviceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceType

`func (o *DeviceWithImage) SetDeviceType(v string)`

SetDeviceType sets DeviceType field to given value.

### HasDeviceType

`func (o *DeviceWithImage) HasDeviceType() bool`

HasDeviceType returns a boolean if a field has been set.

### GetLog

`func (o *DeviceWithImage) GetLog() bool`

GetLog returns the Log field if non-nil, zero value otherwise.

### GetLogOk

`func (o *DeviceWithImage) GetLogOk() (*bool, bool)`

GetLogOk returns a tuple with the Log field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLog

`func (o *DeviceWithImage) SetLog(v bool)`

SetLog sets Log field to given value.


### GetState

`func (o *DeviceWithImage) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *DeviceWithImage) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *DeviceWithImage) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *DeviceWithImage) HasState() bool`

HasState returns a boolean if a field has been set.

### GetSubstate

`func (o *DeviceWithImage) GetSubstate() string`

GetSubstate returns the Substate field if non-nil, zero value otherwise.

### GetSubstateOk

`func (o *DeviceWithImage) GetSubstateOk() (*string, bool)`

GetSubstateOk returns a tuple with the Substate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubstate

`func (o *DeviceWithImage) SetSubstate(v string)`

SetSubstate sets Substate field to given value.

### HasSubstate

`func (o *DeviceWithImage) HasSubstate() bool`

HasSubstate returns a boolean if a field has been set.

### GetImage

`func (o *DeviceWithImage) GetImage() DeviceWithImageImage`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *DeviceWithImage) GetImageOk() (*DeviceWithImageImage, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *DeviceWithImage) SetImage(v DeviceWithImageImage)`

SetImage sets Image field to given value.

### HasImage

`func (o *DeviceWithImage) HasImage() bool`

HasImage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


