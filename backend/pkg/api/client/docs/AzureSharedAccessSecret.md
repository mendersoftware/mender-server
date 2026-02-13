# AzureSharedAccessSecret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The credential type | 
**ConnectionString** | **string** |  | 

## Methods

### NewAzureSharedAccessSecret

`func NewAzureSharedAccessSecret(type_ string, connectionString string, ) *AzureSharedAccessSecret`

NewAzureSharedAccessSecret instantiates a new AzureSharedAccessSecret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAzureSharedAccessSecretWithDefaults

`func NewAzureSharedAccessSecretWithDefaults() *AzureSharedAccessSecret`

NewAzureSharedAccessSecretWithDefaults instantiates a new AzureSharedAccessSecret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *AzureSharedAccessSecret) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AzureSharedAccessSecret) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AzureSharedAccessSecret) SetType(v string)`

SetType sets Type field to given value.


### GetConnectionString

`func (o *AzureSharedAccessSecret) GetConnectionString() string`

GetConnectionString returns the ConnectionString field if non-nil, zero value otherwise.

### GetConnectionStringOk

`func (o *AzureSharedAccessSecret) GetConnectionStringOk() (*string, bool)`

GetConnectionStringOk returns a tuple with the ConnectionString field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionString

`func (o *AzureSharedAccessSecret) SetConnectionString(v string)`

SetConnectionString sets ConnectionString field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


