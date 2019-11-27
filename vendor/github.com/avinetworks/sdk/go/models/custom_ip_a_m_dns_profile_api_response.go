package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CustomIPAMDNSProfileAPIResponse custom ipam Dns profile Api response
// swagger:model CustomIpamDnsProfileApiResponse
type CustomIPAMDNSProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*CustomIPAMDNSProfile `json:"results,omitempty"`
}
