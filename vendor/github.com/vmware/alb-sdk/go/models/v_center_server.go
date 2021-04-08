package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VCenterServer v center server
// swagger:model VCenterServer
type VCenterServer struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// VCenter belongs to cloud. It is a reference to an object of type Cloud. Field introduced in 20.1.1.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// VCenter template to create Service Engine. Field introduced in 20.1.1.
	// Required: true
	ContentLib *ContentLibConfig `json:"content_lib"`

	// Availabilty zone where VCenter list belongs to. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	// VCenter belongs to tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// VCenter config UUID. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`

	// Credentials to access VCenter. It is a reference to an object of type CloudConnectorUser. Field introduced in 20.1.1.
	// Required: true
	VcenterCredentialsRef *string `json:"vcenter_credentials_ref"`

	// VCenter hostname or IP address. Field introduced in 20.1.1.
	// Required: true
	VcenterURL *string `json:"vcenter_url"`
}
