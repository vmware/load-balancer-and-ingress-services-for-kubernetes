package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAddrGroupAPIResponse Ip addr group Api response
// swagger:model IpAddrGroupApiResponse
type IPAddrGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*IPAddrGroup `json:"results,omitempty"`
}
