package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CertificateManagementProfile certificate management profile
// swagger:model CertificateManagementProfile
type CertificateManagementProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the PKI Profile.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property script_params of obj type CertificateManagementProfile field type str  type object
	ScriptParams []*CustomParams `json:"script_params,omitempty"`

	// script_path of CertificateManagementProfile.
	// Required: true
	ScriptPath *string `json:"script_path"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
