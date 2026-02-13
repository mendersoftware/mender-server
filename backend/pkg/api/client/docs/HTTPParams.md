# HTTPParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uri** | Pointer to **string** |  | [optional] 
**Method** | Pointer to **string** |  | [optional] 
**ContentType** | Pointer to **string** |  | [optional] 
**Body** | Pointer to **string** |  | [optional] 
**Json** | Pointer to **map[string]interface{}** |  | [optional] 
**Headers** | Pointer to **map[string]interface{}** |  | [optional] 
**StatusCodes** | Pointer to **[]int32** |  | [optional] 
**ConnectionTimeOut** | Pointer to **int32** |  | [optional] 
**ReadTimeOut** | Pointer to **int32** |  | [optional] 

## Methods

### NewHTTPParams

`func NewHTTPParams() *HTTPParams`

NewHTTPParams instantiates a new HTTPParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHTTPParamsWithDefaults

`func NewHTTPParamsWithDefaults() *HTTPParams`

NewHTTPParamsWithDefaults instantiates a new HTTPParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUri

`func (o *HTTPParams) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *HTTPParams) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *HTTPParams) SetUri(v string)`

SetUri sets Uri field to given value.

### HasUri

`func (o *HTTPParams) HasUri() bool`

HasUri returns a boolean if a field has been set.

### GetMethod

`func (o *HTTPParams) GetMethod() string`

GetMethod returns the Method field if non-nil, zero value otherwise.

### GetMethodOk

`func (o *HTTPParams) GetMethodOk() (*string, bool)`

GetMethodOk returns a tuple with the Method field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethod

`func (o *HTTPParams) SetMethod(v string)`

SetMethod sets Method field to given value.

### HasMethod

`func (o *HTTPParams) HasMethod() bool`

HasMethod returns a boolean if a field has been set.

### GetContentType

`func (o *HTTPParams) GetContentType() string`

GetContentType returns the ContentType field if non-nil, zero value otherwise.

### GetContentTypeOk

`func (o *HTTPParams) GetContentTypeOk() (*string, bool)`

GetContentTypeOk returns a tuple with the ContentType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentType

`func (o *HTTPParams) SetContentType(v string)`

SetContentType sets ContentType field to given value.

### HasContentType

`func (o *HTTPParams) HasContentType() bool`

HasContentType returns a boolean if a field has been set.

### GetBody

`func (o *HTTPParams) GetBody() string`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *HTTPParams) GetBodyOk() (*string, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *HTTPParams) SetBody(v string)`

SetBody sets Body field to given value.

### HasBody

`func (o *HTTPParams) HasBody() bool`

HasBody returns a boolean if a field has been set.

### GetJson

`func (o *HTTPParams) GetJson() map[string]interface{}`

GetJson returns the Json field if non-nil, zero value otherwise.

### GetJsonOk

`func (o *HTTPParams) GetJsonOk() (*map[string]interface{}, bool)`

GetJsonOk returns a tuple with the Json field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJson

`func (o *HTTPParams) SetJson(v map[string]interface{})`

SetJson sets Json field to given value.

### HasJson

`func (o *HTTPParams) HasJson() bool`

HasJson returns a boolean if a field has been set.

### GetHeaders

`func (o *HTTPParams) GetHeaders() map[string]interface{}`

GetHeaders returns the Headers field if non-nil, zero value otherwise.

### GetHeadersOk

`func (o *HTTPParams) GetHeadersOk() (*map[string]interface{}, bool)`

GetHeadersOk returns a tuple with the Headers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeaders

`func (o *HTTPParams) SetHeaders(v map[string]interface{})`

SetHeaders sets Headers field to given value.

### HasHeaders

`func (o *HTTPParams) HasHeaders() bool`

HasHeaders returns a boolean if a field has been set.

### GetStatusCodes

`func (o *HTTPParams) GetStatusCodes() []int32`

GetStatusCodes returns the StatusCodes field if non-nil, zero value otherwise.

### GetStatusCodesOk

`func (o *HTTPParams) GetStatusCodesOk() (*[]int32, bool)`

GetStatusCodesOk returns a tuple with the StatusCodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusCodes

`func (o *HTTPParams) SetStatusCodes(v []int32)`

SetStatusCodes sets StatusCodes field to given value.

### HasStatusCodes

`func (o *HTTPParams) HasStatusCodes() bool`

HasStatusCodes returns a boolean if a field has been set.

### GetConnectionTimeOut

`func (o *HTTPParams) GetConnectionTimeOut() int32`

GetConnectionTimeOut returns the ConnectionTimeOut field if non-nil, zero value otherwise.

### GetConnectionTimeOutOk

`func (o *HTTPParams) GetConnectionTimeOutOk() (*int32, bool)`

GetConnectionTimeOutOk returns a tuple with the ConnectionTimeOut field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionTimeOut

`func (o *HTTPParams) SetConnectionTimeOut(v int32)`

SetConnectionTimeOut sets ConnectionTimeOut field to given value.

### HasConnectionTimeOut

`func (o *HTTPParams) HasConnectionTimeOut() bool`

HasConnectionTimeOut returns a boolean if a field has been set.

### GetReadTimeOut

`func (o *HTTPParams) GetReadTimeOut() int32`

GetReadTimeOut returns the ReadTimeOut field if non-nil, zero value otherwise.

### GetReadTimeOutOk

`func (o *HTTPParams) GetReadTimeOutOk() (*int32, bool)`

GetReadTimeOutOk returns a tuple with the ReadTimeOut field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadTimeOut

`func (o *HTTPParams) SetReadTimeOut(v int32)`

SetReadTimeOut sets ReadTimeOut field to given value.

### HasReadTimeOut

`func (o *HTTPParams) HasReadTimeOut() bool`

HasReadTimeOut returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


