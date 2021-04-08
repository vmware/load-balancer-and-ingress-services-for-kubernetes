package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPCloudRouterUpdate g c p cloud router update
// swagger:model GCPCloudRouterUpdate
type GCPCloudRouterUpdate struct {

	// Action performed  Action can be either Route Added or Route Removed from Router. Field introduced in 18.2.5.
	Action *string `json:"action,omitempty"`

	// Cloud UUID. Field introduced in 18.2.5.
	CcID *string `json:"cc_id,omitempty"`

	// Reason for the failure. Field introduced in 18.2.5.
	ErrorString *string `json:"error_string,omitempty"`

	// Virtual Service Floating IP. Field introduced in 18.2.5.
	Fip *IPAddr `json:"fip,omitempty"`

	// GCP Cloud Router URL. Field introduced in 18.2.5.
	RouterURL *string `json:"router_url,omitempty"`

	// Virtual Service IP. Field introduced in 18.2.5.
	Vip *IPAddr `json:"vip,omitempty"`

	// Virtual Service UUID. Field introduced in 18.2.5.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
