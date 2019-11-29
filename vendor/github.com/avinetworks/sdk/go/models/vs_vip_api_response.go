package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsVipAPIResponse vs vip Api response
// swagger:model VsVipApiResponse
type VsVipAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*VsVip `json:"results,omitempty"`
}
