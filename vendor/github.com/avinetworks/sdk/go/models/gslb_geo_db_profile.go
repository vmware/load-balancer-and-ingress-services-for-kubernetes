package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbGeoDbProfile gslb geo db profile
// swagger:model GslbGeoDbProfile
type GslbGeoDbProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.1.1.
	Description *string `json:"description,omitempty"`

	// List of Geodb entries. An entry can either be a geodb file or an ip address group with geo properties. . Field introduced in 17.1.1. Minimum of 1 items required.
	Entries []*GslbGeoDbEntry `json:"entries,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 17.1.3.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Key value pairs for granular object access control. Also allows for classification and tagging of similar objects. Field introduced in 20.1.2. Maximum of 4 items allowed.
	Labels []*KeyValue `json:"labels,omitempty"`

	// A user-friendly name for the geodb profile. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the geodb profile. Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
