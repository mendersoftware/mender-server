# AddDevicesToGroup200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UpdatedCount** | **int32** | Number of devices listed that changed group.  | 
**MatchedCount** | **int32** | Number of devices listed that matched a valid device id internally.  | 

## Methods

### NewAddDevicesToGroup200Response

`func NewAddDevicesToGroup200Response(updatedCount int32, matchedCount int32, ) *AddDevicesToGroup200Response`

NewAddDevicesToGroup200Response instantiates a new AddDevicesToGroup200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddDevicesToGroup200ResponseWithDefaults

`func NewAddDevicesToGroup200ResponseWithDefaults() *AddDevicesToGroup200Response`

NewAddDevicesToGroup200ResponseWithDefaults instantiates a new AddDevicesToGroup200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUpdatedCount

`func (o *AddDevicesToGroup200Response) GetUpdatedCount() int32`

GetUpdatedCount returns the UpdatedCount field if non-nil, zero value otherwise.

### GetUpdatedCountOk

`func (o *AddDevicesToGroup200Response) GetUpdatedCountOk() (*int32, bool)`

GetUpdatedCountOk returns a tuple with the UpdatedCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedCount

`func (o *AddDevicesToGroup200Response) SetUpdatedCount(v int32)`

SetUpdatedCount sets UpdatedCount field to given value.


### GetMatchedCount

`func (o *AddDevicesToGroup200Response) GetMatchedCount() int32`

GetMatchedCount returns the MatchedCount field if non-nil, zero value otherwise.

### GetMatchedCountOk

`func (o *AddDevicesToGroup200Response) GetMatchedCountOk() (*int32, bool)`

GetMatchedCountOk returns a tuple with the MatchedCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMatchedCount

`func (o *AddDevicesToGroup200Response) SetMatchedCount(v int32)`

SetMatchedCount sets MatchedCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


