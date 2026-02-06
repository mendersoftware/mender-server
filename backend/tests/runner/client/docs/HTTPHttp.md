# HTTPHttp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | **string** | The destination URL for the webhook. The webhook will send POST requests with event details to this target URL. | 
**Secret** | Pointer to **string** | An optional secret used to verify the integrity of the payload. The string must be in hexadecimal format. | [optional] 

## Methods

### NewHTTPHttp

`func NewHTTPHttp(url string, ) *HTTPHttp`

NewHTTPHttp instantiates a new HTTPHttp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHTTPHttpWithDefaults

`func NewHTTPHttpWithDefaults() *HTTPHttp`

NewHTTPHttpWithDefaults instantiates a new HTTPHttp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *HTTPHttp) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *HTTPHttp) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *HTTPHttp) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetSecret

`func (o *HTTPHttp) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *HTTPHttp) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *HTTPHttp) SetSecret(v string)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *HTTPHttp) HasSecret() bool`

HasSecret returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


