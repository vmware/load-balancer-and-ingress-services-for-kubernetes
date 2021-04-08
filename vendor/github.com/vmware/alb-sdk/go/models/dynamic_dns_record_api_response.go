package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DynamicDNSRecordAPIResponse dynamic Dns record Api response
// swagger:model DynamicDnsRecordApiResponse
type DynamicDNSRecordAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*DynamicDNSRecord `json:"results,omitempty"`
}
