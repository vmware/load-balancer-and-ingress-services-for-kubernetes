package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkProfile network profile
// swagger:model NetworkProfile
type NetworkProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// When enabled, Avi mirrors all TCP fastpath connections to standby. Applicable only in Legacy HA Mode. Field introduced in 18.1.3,18.2.1.
	ConnectionMirror *bool `json:"connection_mirror,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// The name of the network profile.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property profile of obj type NetworkProfile field type str  type object
	// Required: true
	Profile *NetworkProfileUnion `json:"profile"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the network profile.
	UUID *string `json:"uuid,omitempty"`
}
