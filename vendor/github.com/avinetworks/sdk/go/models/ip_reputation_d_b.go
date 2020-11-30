package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationDB IP reputation d b
// swagger:model IPReputationDB
type IPReputationDB struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// IP reputation DB base file. It is a reference to an object of type FileObject. Field introduced in 20.1.1.
	BaseFileRefs []string `json:"base_file_refs,omitempty"`

	// Description. Field introduced in 20.1.1.
	Description *string `json:"description,omitempty"`

	// IP reputation DB incremental update files. It is a reference to an object of type FileObject. Field introduced in 20.1.1.
	IncrementalFileRefs []string `json:"incremental_file_refs,omitempty"`

	// Key value pairs for granular object access control. Also allows for classification and tagging of similar objects. Field introduced in 20.1.3.
	Labels []*KeyValue `json:"labels,omitempty"`

	// IP reputation DB name. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// If this object is managed by the IP reputation service, this field contain the status of this syncronization. Field introduced in 20.1.1.
	ServiceStatus *IPReputationServiceStatus `json:"service_status,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of this object. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Organization providing IP reputation data. Enum options - IP_REPUTATION_VENDOR_WEBROOT. Field introduced in 20.1.1.
	// Required: true
	Vendor *string `json:"vendor"`

	// A version number for this database object. This is informal for the consumer of this API only, a tool which manages this object can store version information here. Field introduced in 20.1.1.
	Version *string `json:"version,omitempty"`
}
