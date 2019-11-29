package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPPolicies HTTP policies
// swagger:model HTTPPolicies
type HTTPPolicies struct {

	// UUID of the virtual service HTTP policy collection. It is a reference to an object of type HTTPPolicySet.
	// Required: true
	HTTPPolicySetRef *string `json:"http_policy_set_ref"`

	// Index of the virtual service HTTP policy collection.
	// Required: true
	Index *int32 `json:"index"`
}
