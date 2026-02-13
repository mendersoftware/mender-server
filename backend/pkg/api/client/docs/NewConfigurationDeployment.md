# NewConfigurationDeployment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Retries** | Pointer to **int32** | The number of times a device can retry the deployment in case of failure, defaults to 0 | [optional] [default to 0]
**UpdateControlMap** | Pointer to **map[string]interface{}** | A valid JSON object defining the update control map. *NOTE*: Available only in the Enterprise plan.  | [optional] 

## Methods

### NewNewConfigurationDeployment

`func NewNewConfigurationDeployment() *NewConfigurationDeployment`

NewNewConfigurationDeployment instantiates a new NewConfigurationDeployment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNewConfigurationDeploymentWithDefaults

`func NewNewConfigurationDeploymentWithDefaults() *NewConfigurationDeployment`

NewNewConfigurationDeploymentWithDefaults instantiates a new NewConfigurationDeployment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRetries

`func (o *NewConfigurationDeployment) GetRetries() int32`

GetRetries returns the Retries field if non-nil, zero value otherwise.

### GetRetriesOk

`func (o *NewConfigurationDeployment) GetRetriesOk() (*int32, bool)`

GetRetriesOk returns a tuple with the Retries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetries

`func (o *NewConfigurationDeployment) SetRetries(v int32)`

SetRetries sets Retries field to given value.

### HasRetries

`func (o *NewConfigurationDeployment) HasRetries() bool`

HasRetries returns a boolean if a field has been set.

### GetUpdateControlMap

`func (o *NewConfigurationDeployment) GetUpdateControlMap() map[string]interface{}`

GetUpdateControlMap returns the UpdateControlMap field if non-nil, zero value otherwise.

### GetUpdateControlMapOk

`func (o *NewConfigurationDeployment) GetUpdateControlMapOk() (*map[string]interface{}, bool)`

GetUpdateControlMapOk returns a tuple with the UpdateControlMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateControlMap

`func (o *NewConfigurationDeployment) SetUpdateControlMap(v map[string]interface{})`

SetUpdateControlMap sets UpdateControlMap field to given value.

### HasUpdateControlMap

`func (o *NewConfigurationDeployment) HasUpdateControlMap() bool`

HasUpdateControlMap returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


