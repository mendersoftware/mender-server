# StorageSettings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** | The storage provider type &#39;azure&#39; Blob storage or AWS &#39;s3&#39; (defaults to s3). | [optional] 
**Region** | Pointer to **string** | AWS region (S3 only: required). | [optional] 
**Bucket** | **string** | S3 Bucket (Azure: container) name. | 
**Uri** | Pointer to **string** | Bucket/container endpoint URI. | [optional] 
**ExternalUri** | Pointer to **string** | Public Endpoint URI for presigning URLs (S3 only). | [optional] 
**Key** | **string** | Access key identifier (Azure: account name). | 
**Secret** | **string** | Secret access key (Azure: access key). | 
**Token** | Pointer to **string** | AWS S3 session token (S3 only). | [optional] 
**ForcePathStyle** | Pointer to **bool** | Force S3 path-style instead of virtual-hosted style (S3 only). | [optional] 
**UseAccelerate** | Pointer to **bool** | Enable S3 Transfer acceleration (S3 only). | [optional] 
**ConnectionString** | Pointer to **string** | Shared access key connection string (Azure only). | [optional] 
**ContainerName** | Pointer to **string** | Alias for &#39;bucket&#39; (Azure only). | [optional] 
**AccountName** | Pointer to **string** | Alias for &#39;key&#39; (Azure only). | [optional] 
**AccountKey** | Pointer to **string** | Alias for &#39;secret&#39; (Azure only). | [optional] 

## Methods

### NewStorageSettings

`func NewStorageSettings(bucket string, key string, secret string, ) *StorageSettings`

NewStorageSettings instantiates a new StorageSettings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageSettingsWithDefaults

`func NewStorageSettingsWithDefaults() *StorageSettings`

NewStorageSettingsWithDefaults instantiates a new StorageSettings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *StorageSettings) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *StorageSettings) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *StorageSettings) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *StorageSettings) HasType() bool`

HasType returns a boolean if a field has been set.

### GetRegion

`func (o *StorageSettings) GetRegion() string`

GetRegion returns the Region field if non-nil, zero value otherwise.

### GetRegionOk

`func (o *StorageSettings) GetRegionOk() (*string, bool)`

GetRegionOk returns a tuple with the Region field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegion

`func (o *StorageSettings) SetRegion(v string)`

SetRegion sets Region field to given value.

### HasRegion

`func (o *StorageSettings) HasRegion() bool`

HasRegion returns a boolean if a field has been set.

### GetBucket

`func (o *StorageSettings) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *StorageSettings) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *StorageSettings) SetBucket(v string)`

SetBucket sets Bucket field to given value.


### GetUri

`func (o *StorageSettings) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *StorageSettings) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *StorageSettings) SetUri(v string)`

SetUri sets Uri field to given value.

### HasUri

`func (o *StorageSettings) HasUri() bool`

HasUri returns a boolean if a field has been set.

### GetExternalUri

`func (o *StorageSettings) GetExternalUri() string`

GetExternalUri returns the ExternalUri field if non-nil, zero value otherwise.

### GetExternalUriOk

`func (o *StorageSettings) GetExternalUriOk() (*string, bool)`

GetExternalUriOk returns a tuple with the ExternalUri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalUri

`func (o *StorageSettings) SetExternalUri(v string)`

SetExternalUri sets ExternalUri field to given value.

### HasExternalUri

`func (o *StorageSettings) HasExternalUri() bool`

HasExternalUri returns a boolean if a field has been set.

### GetKey

`func (o *StorageSettings) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *StorageSettings) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *StorageSettings) SetKey(v string)`

SetKey sets Key field to given value.


### GetSecret

`func (o *StorageSettings) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *StorageSettings) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *StorageSettings) SetSecret(v string)`

SetSecret sets Secret field to given value.


### GetToken

`func (o *StorageSettings) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *StorageSettings) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *StorageSettings) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *StorageSettings) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetForcePathStyle

`func (o *StorageSettings) GetForcePathStyle() bool`

GetForcePathStyle returns the ForcePathStyle field if non-nil, zero value otherwise.

### GetForcePathStyleOk

`func (o *StorageSettings) GetForcePathStyleOk() (*bool, bool)`

GetForcePathStyleOk returns a tuple with the ForcePathStyle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForcePathStyle

`func (o *StorageSettings) SetForcePathStyle(v bool)`

SetForcePathStyle sets ForcePathStyle field to given value.

### HasForcePathStyle

`func (o *StorageSettings) HasForcePathStyle() bool`

HasForcePathStyle returns a boolean if a field has been set.

### GetUseAccelerate

`func (o *StorageSettings) GetUseAccelerate() bool`

GetUseAccelerate returns the UseAccelerate field if non-nil, zero value otherwise.

### GetUseAccelerateOk

`func (o *StorageSettings) GetUseAccelerateOk() (*bool, bool)`

GetUseAccelerateOk returns a tuple with the UseAccelerate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseAccelerate

`func (o *StorageSettings) SetUseAccelerate(v bool)`

SetUseAccelerate sets UseAccelerate field to given value.

### HasUseAccelerate

`func (o *StorageSettings) HasUseAccelerate() bool`

HasUseAccelerate returns a boolean if a field has been set.

### GetConnectionString

`func (o *StorageSettings) GetConnectionString() string`

GetConnectionString returns the ConnectionString field if non-nil, zero value otherwise.

### GetConnectionStringOk

`func (o *StorageSettings) GetConnectionStringOk() (*string, bool)`

GetConnectionStringOk returns a tuple with the ConnectionString field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionString

`func (o *StorageSettings) SetConnectionString(v string)`

SetConnectionString sets ConnectionString field to given value.

### HasConnectionString

`func (o *StorageSettings) HasConnectionString() bool`

HasConnectionString returns a boolean if a field has been set.

### GetContainerName

`func (o *StorageSettings) GetContainerName() string`

GetContainerName returns the ContainerName field if non-nil, zero value otherwise.

### GetContainerNameOk

`func (o *StorageSettings) GetContainerNameOk() (*string, bool)`

GetContainerNameOk returns a tuple with the ContainerName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainerName

`func (o *StorageSettings) SetContainerName(v string)`

SetContainerName sets ContainerName field to given value.

### HasContainerName

`func (o *StorageSettings) HasContainerName() bool`

HasContainerName returns a boolean if a field has been set.

### GetAccountName

`func (o *StorageSettings) GetAccountName() string`

GetAccountName returns the AccountName field if non-nil, zero value otherwise.

### GetAccountNameOk

`func (o *StorageSettings) GetAccountNameOk() (*string, bool)`

GetAccountNameOk returns a tuple with the AccountName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccountName

`func (o *StorageSettings) SetAccountName(v string)`

SetAccountName sets AccountName field to given value.

### HasAccountName

`func (o *StorageSettings) HasAccountName() bool`

HasAccountName returns a boolean if a field has been set.

### GetAccountKey

`func (o *StorageSettings) GetAccountKey() string`

GetAccountKey returns the AccountKey field if non-nil, zero value otherwise.

### GetAccountKeyOk

`func (o *StorageSettings) GetAccountKeyOk() (*string, bool)`

GetAccountKeyOk returns a tuple with the AccountKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccountKey

`func (o *StorageSettings) SetAccountKey(v string)`

SetAccountKey sets AccountKey field to given value.

### HasAccountKey

`func (o *StorageSettings) HasAccountKey() bool`

HasAccountKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


