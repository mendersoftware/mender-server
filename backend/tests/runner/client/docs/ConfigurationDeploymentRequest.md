# ConfigurationDeploymentRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the deployment | 
**Configuration** | **string** | A string containing a configuration object. The deployments service will use it to generate configuration artifact for the device. The artifact will be generated when the device will ask for an update.  | 

## Methods

### NewConfigurationDeploymentRequest

`func NewConfigurationDeploymentRequest(name string, configuration string, ) *ConfigurationDeploymentRequest`

NewConfigurationDeploymentRequest instantiates a new ConfigurationDeploymentRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigurationDeploymentRequestWithDefaults

`func NewConfigurationDeploymentRequestWithDefaults() *ConfigurationDeploymentRequest`

NewConfigurationDeploymentRequestWithDefaults instantiates a new ConfigurationDeploymentRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ConfigurationDeploymentRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConfigurationDeploymentRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConfigurationDeploymentRequest) SetName(v string)`

SetName sets Name field to given value.


### GetConfiguration

`func (o *ConfigurationDeploymentRequest) GetConfiguration() string`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *ConfigurationDeploymentRequest) GetConfigurationOk() (*string, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *ConfigurationDeploymentRequest) SetConfiguration(v string)`

SetConfiguration sets Configuration field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


