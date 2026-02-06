# PersonalAccessToken

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Token identifier. | 
**Name** | **string** | Name of a token. | 
**LastUsed** | Pointer to **time.Time** | Date of last usage of a token. The accuracy is 5 minutes.  | [optional] 
**ExpirationDate** | **time.Time** | Expiration date. | 
**CreatedTs** | **time.Time** | Server-side timestamp of the token creation.  | 

## Methods

### NewPersonalAccessToken

`func NewPersonalAccessToken(id string, name string, expirationDate time.Time, createdTs time.Time, ) *PersonalAccessToken`

NewPersonalAccessToken instantiates a new PersonalAccessToken object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPersonalAccessTokenWithDefaults

`func NewPersonalAccessTokenWithDefaults() *PersonalAccessToken`

NewPersonalAccessTokenWithDefaults instantiates a new PersonalAccessToken object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PersonalAccessToken) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PersonalAccessToken) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PersonalAccessToken) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *PersonalAccessToken) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PersonalAccessToken) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PersonalAccessToken) SetName(v string)`

SetName sets Name field to given value.


### GetLastUsed

`func (o *PersonalAccessToken) GetLastUsed() time.Time`

GetLastUsed returns the LastUsed field if non-nil, zero value otherwise.

### GetLastUsedOk

`func (o *PersonalAccessToken) GetLastUsedOk() (*time.Time, bool)`

GetLastUsedOk returns a tuple with the LastUsed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUsed

`func (o *PersonalAccessToken) SetLastUsed(v time.Time)`

SetLastUsed sets LastUsed field to given value.

### HasLastUsed

`func (o *PersonalAccessToken) HasLastUsed() bool`

HasLastUsed returns a boolean if a field has been set.

### GetExpirationDate

`func (o *PersonalAccessToken) GetExpirationDate() time.Time`

GetExpirationDate returns the ExpirationDate field if non-nil, zero value otherwise.

### GetExpirationDateOk

`func (o *PersonalAccessToken) GetExpirationDateOk() (*time.Time, bool)`

GetExpirationDateOk returns a tuple with the ExpirationDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpirationDate

`func (o *PersonalAccessToken) SetExpirationDate(v time.Time)`

SetExpirationDate sets ExpirationDate field to given value.


### GetCreatedTs

`func (o *PersonalAccessToken) GetCreatedTs() time.Time`

GetCreatedTs returns the CreatedTs field if non-nil, zero value otherwise.

### GetCreatedTsOk

`func (o *PersonalAccessToken) GetCreatedTsOk() (*time.Time, bool)`

GetCreatedTsOk returns a tuple with the CreatedTs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedTs

`func (o *PersonalAccessToken) SetCreatedTs(v time.Time)`

SetCreatedTs sets CreatedTs field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


