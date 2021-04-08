package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AvailabilityZone availability zone
// swagger:model AvailabilityZone
type AvailabilityZone struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Availability zone belongs to cloud. It is a reference to an object of type Cloud. Field introduced in 20.1.1.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Availabilty zone where VCenter list belongs to. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// Availabilityzone belongs to tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Availability zone config UUID. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Group of VCenter list belong to availabilty zone. It is a reference to an object of type VCenterServer. Field introduced in 20.1.1. Minimum of 1 items required.
	VcenterRefs []string `json:"vcenter_refs,omitempty"`
}
