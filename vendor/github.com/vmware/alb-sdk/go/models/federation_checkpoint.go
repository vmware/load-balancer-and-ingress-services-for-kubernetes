package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FederationCheckpoint federation checkpoint
// swagger:model FederationCheckpoint
type FederationCheckpoint struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Date when the checkpoint was created. Field introduced in 20.1.1.
	Date *string `json:"date,omitempty"`

	// Description for this checkpoint. Field introduced in 20.1.1.
	Description *string `json:"description,omitempty"`

	// This field describes the object's replication scope. If the field is set to false, then the object is visible within the controller-cluster and its associated service-engines. If the field is set to true, then the object is replicated across the federation. Field introduced in 20.1.1.
	IsFederated *bool `json:"is_federated,omitempty"`

	// Name of the Checkpoint. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Checkpoint. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}
