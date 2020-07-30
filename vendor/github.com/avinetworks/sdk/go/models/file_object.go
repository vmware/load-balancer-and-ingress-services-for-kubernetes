package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FileObject file object
// swagger:model FileObject
type FileObject struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// SHA1 checksum of the file. Field introduced in 20.1.1.
	Checksum *string `json:"checksum,omitempty"`

	// This field indicates whether the file is gzip-compressed. Field introduced in 20.1.1.
	Compressed *bool `json:"compressed,omitempty"`

	// Timestamp of creation for the file. Field introduced in 20.1.1.
	Created *string `json:"created,omitempty"`

	// Description of the file. Field introduced in 20.1.1.
	Description *string `json:"description,omitempty"`

	// Timestamp when the file will be no longer needed and can be removed by the system. If this is set, a garbage collector process will try to remove the file after this time. Field introduced in 20.1.1.
	ExpiresAt *string `json:"expires_at,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines. If the field is set to true, then the object is replicated across the federation. Field introduced in 20.1.1.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Name of the file object. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// Path to the file. Field introduced in 20.1.1.
	Path *string `json:"path,omitempty"`

	// Enforce Read-Only on the file. Field introduced in 20.1.1.
	ReadOnly *bool `json:"read_only,omitempty"`

	// Flag to allow/restrict download of the file. Field introduced in 20.1.1.
	RestrictDownload *bool `json:"restrict_download,omitempty"`

	// Size of the file. Field introduced in 20.1.1.
	Size *int64 `json:"size,omitempty"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of the file. Enum options - OTHER_FILE_TYPES, IP_REPUTATION, GEO_DB, TECH_SUPPORT, HSMPACKAGES, IPAMDNSSCRIPTS, CONTROLLER_IMAGE. Field introduced in 20.1.1.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the file. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Version of the file. Field introduced in 20.1.1.
	Version *string `json:"version,omitempty"`
}
