# AWSCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The credential type | 
**Aws** | [**AWSCredentialsAws**](AWSCredentialsAws.md) |  | 

## Methods

### NewAWSCredentials

`func NewAWSCredentials(type_ string, aws AWSCredentialsAws, ) *AWSCredentials`

NewAWSCredentials instantiates a new AWSCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAWSCredentialsWithDefaults

`func NewAWSCredentialsWithDefaults() *AWSCredentials`

NewAWSCredentialsWithDefaults instantiates a new AWSCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *AWSCredentials) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AWSCredentials) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AWSCredentials) SetType(v string)`

SetType sets Type field to given value.


### GetAws

`func (o *AWSCredentials) GetAws() AWSCredentialsAws`

GetAws returns the Aws field if non-nil, zero value otherwise.

### GetAwsOk

`func (o *AWSCredentials) GetAwsOk() (*AWSCredentialsAws, bool)`

GetAwsOk returns a tuple with the Aws field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAws

`func (o *AWSCredentials) SetAws(v AWSCredentialsAws)`

SetAws sets Aws field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


