# DeviceDeploymentV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Deployment** | [**DeploymentV1**](DeploymentV1.md) |  | 
**Device** | [**DeviceWithImage**](DeviceWithImage.md) |  | 

## Methods

### NewDeviceDeploymentV1

`func NewDeviceDeploymentV1(deployment DeploymentV1, device DeviceWithImage, ) *DeviceDeploymentV1`

NewDeviceDeploymentV1 instantiates a new DeviceDeploymentV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceDeploymentV1WithDefaults

`func NewDeviceDeploymentV1WithDefaults() *DeviceDeploymentV1`

NewDeviceDeploymentV1WithDefaults instantiates a new DeviceDeploymentV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceDeploymentV1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceDeploymentV1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceDeploymentV1) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceDeploymentV1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDeployment

`func (o *DeviceDeploymentV1) GetDeployment() DeploymentV1`

GetDeployment returns the Deployment field if non-nil, zero value otherwise.

### GetDeploymentOk

`func (o *DeviceDeploymentV1) GetDeploymentOk() (*DeploymentV1, bool)`

GetDeploymentOk returns a tuple with the Deployment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeployment

`func (o *DeviceDeploymentV1) SetDeployment(v DeploymentV1)`

SetDeployment sets Deployment field to given value.


### GetDevice

`func (o *DeviceDeploymentV1) GetDevice() DeviceWithImage`

GetDevice returns the Device field if non-nil, zero value otherwise.

### GetDeviceOk

`func (o *DeviceDeploymentV1) GetDeviceOk() (*DeviceWithImage, bool)`

GetDeviceOk returns a tuple with the Device field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevice

`func (o *DeviceDeploymentV1) SetDevice(v DeviceWithImage)`

SetDevice sets Device field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


