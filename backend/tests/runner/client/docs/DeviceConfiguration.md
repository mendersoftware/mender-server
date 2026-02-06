# DeviceConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Configured** | Pointer to **map[string]string** |  | [optional] 
**Reported** | Pointer to **map[string]string** |  | [optional] 
**DeploymentId** | Pointer to **string** | ID of the latest configuration deployment | [optional] 
**ReportedTs** | Pointer to **time.Time** |  | [optional] 
**UpdatedTs** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewDeviceConfiguration

`func NewDeviceConfiguration() *DeviceConfiguration`

NewDeviceConfiguration instantiates a new DeviceConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceConfigurationWithDefaults

`func NewDeviceConfigurationWithDefaults() *DeviceConfiguration`

NewDeviceConfigurationWithDefaults instantiates a new DeviceConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceConfiguration) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceConfiguration) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceConfiguration) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceConfiguration) HasId() bool`

HasId returns a boolean if a field has been set.

### GetConfigured

`func (o *DeviceConfiguration) GetConfigured() map[string]string`

GetConfigured returns the Configured field if non-nil, zero value otherwise.

### GetConfiguredOk

`func (o *DeviceConfiguration) GetConfiguredOk() (*map[string]string, bool)`

GetConfiguredOk returns a tuple with the Configured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigured

`func (o *DeviceConfiguration) SetConfigured(v map[string]string)`

SetConfigured sets Configured field to given value.

### HasConfigured

`func (o *DeviceConfiguration) HasConfigured() bool`

HasConfigured returns a boolean if a field has been set.

### GetReported

`func (o *DeviceConfiguration) GetReported() map[string]string`

GetReported returns the Reported field if non-nil, zero value otherwise.

### GetReportedOk

`func (o *DeviceConfiguration) GetReportedOk() (*map[string]string, bool)`

GetReportedOk returns a tuple with the Reported field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReported

`func (o *DeviceConfiguration) SetReported(v map[string]string)`

SetReported sets Reported field to given value.

### HasReported

`func (o *DeviceConfiguration) HasReported() bool`

HasReported returns a boolean if a field has been set.

### GetDeploymentId

`func (o *DeviceConfiguration) GetDeploymentId() string`

GetDeploymentId returns the DeploymentId field if non-nil, zero value otherwise.

### GetDeploymentIdOk

`func (o *DeviceConfiguration) GetDeploymentIdOk() (*string, bool)`

GetDeploymentIdOk returns a tuple with the DeploymentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeploymentId

`func (o *DeviceConfiguration) SetDeploymentId(v string)`

SetDeploymentId sets DeploymentId field to given value.

### HasDeploymentId

`func (o *DeviceConfiguration) HasDeploymentId() bool`

HasDeploymentId returns a boolean if a field has been set.

### GetReportedTs

`func (o *DeviceConfiguration) GetReportedTs() time.Time`

GetReportedTs returns the ReportedTs field if non-nil, zero value otherwise.

### GetReportedTsOk

`func (o *DeviceConfiguration) GetReportedTsOk() (*time.Time, bool)`

GetReportedTsOk returns a tuple with the ReportedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReportedTs

`func (o *DeviceConfiguration) SetReportedTs(v time.Time)`

SetReportedTs sets ReportedTs field to given value.

### HasReportedTs

`func (o *DeviceConfiguration) HasReportedTs() bool`

HasReportedTs returns a boolean if a field has been set.

### GetUpdatedTs

`func (o *DeviceConfiguration) GetUpdatedTs() time.Time`

GetUpdatedTs returns the UpdatedTs field if non-nil, zero value otherwise.

### GetUpdatedTsOk

`func (o *DeviceConfiguration) GetUpdatedTsOk() (*time.Time, bool)`

GetUpdatedTsOk returns a tuple with the UpdatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedTs

`func (o *DeviceConfiguration) SetUpdatedTs(v time.Time)`

SetUpdatedTs sets UpdatedTs field to given value.

### HasUpdatedTs

`func (o *DeviceConfiguration) HasUpdatedTs() bool`

HasUpdatedTs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


