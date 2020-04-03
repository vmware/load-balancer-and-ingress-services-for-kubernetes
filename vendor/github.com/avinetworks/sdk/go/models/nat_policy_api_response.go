package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NatPolicyAPIResponse nat policy Api response
// swagger:model NatPolicyApiResponse
type NatPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*NatPolicy `json:"results,omitempty"`
}
