package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TenantAPIResponse tenant Api response
// swagger:model TenantApiResponse
type TenantAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*Tenant `json:"results,omitempty"`
}
