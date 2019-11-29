package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HardwareSecurityModuleGroup hardware security module group
// swagger:model HardwareSecurityModuleGroup
type HardwareSecurityModuleGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Hardware Security Module configuration.
	// Required: true
	Hsm *HardwareSecurityModule `json:"hsm"`

	// Name of the HSM Group configuration object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the HSM Group configuration object.
	UUID *string `json:"uuid,omitempty"`
}
