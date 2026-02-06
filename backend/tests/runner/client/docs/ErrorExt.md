# ErrorExt

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | **string** | Description of the error. | 
**RequestId** | Pointer to **string** | Request ID (same as in X-MEN-RequestID header). | [optional] 
**Metadata** | Pointer to **map[string]interface{}** |  | [optional] 

## Methods

### NewErrorExt

`func NewErrorExt(error_ string, ) *ErrorExt`

NewErrorExt instantiates a new ErrorExt object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewErrorExtWithDefaults

`func NewErrorExtWithDefaults() *ErrorExt`

NewErrorExtWithDefaults instantiates a new ErrorExt object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *ErrorExt) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *ErrorExt) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *ErrorExt) SetError(v string)`

SetError sets Error field to given value.


### GetRequestId

`func (o *ErrorExt) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *ErrorExt) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *ErrorExt) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.

### HasRequestId

`func (o *ErrorExt) HasRequestId() bool`

HasRequestId returns a boolean if a field has been set.

### GetMetadata

`func (o *ErrorExt) GetMetadata() map[string]interface{}`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ErrorExt) GetMetadataOk() (*map[string]interface{}, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ErrorExt) SetMetadata(v map[string]interface{})`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ErrorExt) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


