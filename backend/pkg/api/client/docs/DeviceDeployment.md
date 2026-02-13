# DeviceDeployment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Deployment** | [**DeploymentV1Internal**](DeploymentV1Internal.md) |  | 
**Device** | [**DeviceWithImage**](DeviceWithImage.md) |  | 

## Methods

### NewDeviceDeployment

`func NewDeviceDeployment(deployment DeploymentV1Internal, device DeviceWithImage, ) *DeviceDeployment`

NewDeviceDeployment instantiates a new DeviceDeployment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceDeploymentWithDefaults

`func NewDeviceDeploymentWithDefaults() *DeviceDeployment`

NewDeviceDeploymentWithDefaults instantiates a new DeviceDeployment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DeviceDeployment) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DeviceDeployment) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DeviceDeployment) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DeviceDeployment) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDeployment

`func (o *DeviceDeployment) GetDeployment() DeploymentV1Internal`

GetDeployment returns the Deployment field if non-nil, zero value otherwise.

### GetDeploymentOk

`func (o *DeviceDeployment) GetDeploymentOk() (*DeploymentV1Internal, bool)`

GetDeploymentOk returns a tuple with the Deployment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeployment

`func (o *DeviceDeployment) SetDeployment(v DeploymentV1Internal)`

SetDeployment sets Deployment field to given value.


### GetDevice

`func (o *DeviceDeployment) GetDevice() DeviceWithImage`

GetDevice returns the Device field if non-nil, zero value otherwise.

### GetDeviceOk

`func (o *DeviceDeployment) GetDeviceOk() (*DeviceWithImage, bool)`

GetDeviceOk returns a tuple with the Device field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevice

`func (o *DeviceDeployment) SetDevice(v DeviceWithImage)`

SetDevice sets Device field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


