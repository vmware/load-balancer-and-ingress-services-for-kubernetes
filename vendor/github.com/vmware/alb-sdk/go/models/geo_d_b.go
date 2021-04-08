package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GeoDB geo d b
// swagger:model GeoDB
type GeoDB struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Description. Field introduced in 21.1.1.
	Description *string `json:"description,omitempty"`

	// Geo Database files. Field introduced in 21.1.1.
	// Required: true
	Files []*GeoDBFile `json:"files,omitempty"`

	// This field indicates that this object is replicated across GSLB federation. Field introduced in 21.1.1.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Custom mappings of geo values. All mappings which start with the prefix 'System-' (any case) are reserved for system default objects and may be overwritten. Field introduced in 21.1.1.
	Mappings []*GeoDBMapping `json:"mappings,omitempty"`

	// Geo Database name. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 21.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of this object. Field introduced in 21.1.1.
	UUID *string `json:"uuid,omitempty"`
}
