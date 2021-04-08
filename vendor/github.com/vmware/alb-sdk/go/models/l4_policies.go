package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// L4Policies l4 policies
// swagger:model L4Policies
type L4Policies struct {

	// Index of the virtual service L4 policy set. Field introduced in 17.2.7.
	// Required: true
	Index *int32 `json:"index"`

	// ID of the virtual service L4 policy set. It is a reference to an object of type L4PolicySet. Field introduced in 17.2.7.
	// Required: true
	L4PolicySetRef *string `json:"l4_policy_set_ref"`
}
