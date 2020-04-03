package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ObjectAccessPolicy object access policy
// swagger:model ObjectAccessPolicy
type ObjectAccessPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the Object Access Policy. Field introduced in 18.2.7.
	// Required: true
	Name *string `json:"name"`

	// Rules which grant access to specific objects. Field introduced in 18.2.7.
	// Required: true
	Rules []*ObjectAccessPolicyRule `json:"rules,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.7.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Uuid of the Object Access Policy. Field introduced in 18.2.7.
	UUID *string `json:"uuid,omitempty"`
}
