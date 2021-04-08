package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSProviderProfileAPIResponse ipam Dns provider profile Api response
// swagger:model IpamDnsProviderProfileApiResponse
type IPAMDNSProviderProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*IPAMDNSProviderProfile `json:"results,omitempty"`
}
