package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbPoolRuntime gslb pool runtime
// swagger:model GslbPoolRuntime
type GslbPoolRuntime struct {

	// Placeholder for description of property members of obj type GslbPoolRuntime field type str  type object
	Members []*GslbPoolMemberRuntimeInfo `json:"members,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Gslb Pool's consolidated operational status . Field introduced in 18.2.3.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`
}
