# HTTP

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The credential type | 
**Http** | [**HTTPHttp**](HTTPHttp.md) |  | 

## Methods

### NewHTTP

`func NewHTTP(type_ string, http HTTPHttp, ) *HTTP`

NewHTTP instantiates a new HTTP object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHTTPWithDefaults

`func NewHTTPWithDefaults() *HTTP`

NewHTTPWithDefaults instantiates a new HTTP object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *HTTP) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *HTTP) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *HTTP) SetType(v string)`

SetType sets Type field to given value.


### GetHttp

`func (o *HTTP) GetHttp() HTTPHttp`

GetHttp returns the Http field if non-nil, zero value otherwise.

### GetHttpOk

`func (o *HTTP) GetHttpOk() (*HTTPHttp, bool)`

GetHttpOk returns a tuple with the Http field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHttp

`func (o *HTTP) SetHttp(v HTTPHttp)`

SetHttp sets Http field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


