package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbApplicationPersistenceProfile gslb application persistence profile
// swagger:model GslbApplicationPersistenceProfile
type GslbApplicationPersistenceProfile struct {

	//  Field introduced in 17.1.1.
	Description string `json:"description,omitempty"`

	// A user-friendly name for the persistence profile. Field introduced in 17.1.1.
	// Required: true
	Name string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL string `json:"url,omitempty"`

	// UUID of the persistence profile. Field introduced in 17.1.1.
	UUID string `json:"uuid,omitempty"`
}
