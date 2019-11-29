package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureChannelAvailableLocalIpsAPIResponse secure channel available local ips Api response
// swagger:model SecureChannelAvailableLocalIPsApiResponse
type SecureChannelAvailableLocalIpsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*SecureChannelAvailableLocalIps `json:"results,omitempty"`
}
