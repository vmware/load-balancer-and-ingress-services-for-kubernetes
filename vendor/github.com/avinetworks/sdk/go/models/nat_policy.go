package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NatPolicy nat policy
// swagger:model NatPolicy
type NatPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Creator name. Field introduced in 18.2.3.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Name of the Nat policy. Field introduced in 18.2.3.
	Name *string `json:"name,omitempty"`

	// Nat policy Rules. Field introduced in 18.2.3.
	Rules []*NatRule `json:"rules,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 18.2.3.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Nat policy. Field introduced in 18.2.3.
	UUID *string `json:"uuid,omitempty"`
}
