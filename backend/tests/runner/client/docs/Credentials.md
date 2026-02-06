# Credentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The credential type | 
**Aws** | [**AWSCredentialsAws**](AWSCredentialsAws.md) |  | 
**ConnectionString** | **string** |  | 
**Http** | [**HTTPHttp**](HTTPHttp.md) |  | 

## Methods

### NewCredentials

`func NewCredentials(type_ string, aws AWSCredentialsAws, connectionString string, http HTTPHttp, ) *Credentials`

NewCredentials instantiates a new Credentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsWithDefaults

`func NewCredentialsWithDefaults() *Credentials`

NewCredentialsWithDefaults instantiates a new Credentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *Credentials) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Credentials) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Credentials) SetType(v string)`

SetType sets Type field to given value.


### GetAws

`func (o *Credentials) GetAws() AWSCredentialsAws`

GetAws returns the Aws field if non-nil, zero value otherwise.

### GetAwsOk

`func (o *Credentials) GetAwsOk() (*AWSCredentialsAws, bool)`

GetAwsOk returns a tuple with the Aws field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAws

`func (o *Credentials) SetAws(v AWSCredentialsAws)`

SetAws sets Aws field to given value.


### GetConnectionString

`func (o *Credentials) GetConnectionString() string`

GetConnectionString returns the ConnectionString field if non-nil, zero value otherwise.

### GetConnectionStringOk

`func (o *Credentials) GetConnectionStringOk() (*string, bool)`

GetConnectionStringOk returns a tuple with the ConnectionString field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionString

`func (o *Credentials) SetConnectionString(v string)`

SetConnectionString sets ConnectionString field to given value.


### GetHttp

`func (o *Credentials) GetHttp() HTTPHttp`

GetHttp returns the Http field if non-nil, zero value otherwise.

### GetHttpOk

`func (o *Credentials) GetHttpOk() (*HTTPHttp, bool)`

GetHttpOk returns a tuple with the Http field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttp

`func (o *Credentials) SetHttp(v HTTPHttp)`

SetHttp sets Http field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


