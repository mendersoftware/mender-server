# ExternalDevice

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | ID assigned by external provider | 
**Name** | **string** | Name of the device | 
**Provider** | **string** | Name of the external provider | 
**IdData** | Pointer to **string** | Optional custom ID data | [optional] 

## Methods

### NewExternalDevice

`func NewExternalDevice(id string, name string, provider string, ) *ExternalDevice`

NewExternalDevice instantiates a new ExternalDevice object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExternalDeviceWithDefaults

`func NewExternalDeviceWithDefaults() *ExternalDevice`

NewExternalDeviceWithDefaults instantiates a new ExternalDevice object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ExternalDevice) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ExternalDevice) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ExternalDevice) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ExternalDevice) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ExternalDevice) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ExternalDevice) SetName(v string)`

SetName sets Name field to given value.


### GetProvider

`func (o *ExternalDevice) GetProvider() string`

GetProvider returns the Provider field if non-nil, zero value otherwise.

### GetProviderOk

`func (o *ExternalDevice) GetProviderOk() (*string, bool)`

GetProviderOk returns a tuple with the Provider field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvider

`func (o *ExternalDevice) SetProvider(v string)`

SetProvider sets Provider field to given value.


### GetIdData

`func (o *ExternalDevice) GetIdData() string`

GetIdData returns the IdData field if non-nil, zero value otherwise.

### GetIdDataOk

`func (o *ExternalDevice) GetIdDataOk() (*string, bool)`

GetIdDataOk returns a tuple with the IdData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdData

`func (o *ExternalDevice) SetIdData(v string)`

SetIdData sets IdData field to given value.

### HasIdData

`func (o *ExternalDevice) HasIdData() bool`

HasIdData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


