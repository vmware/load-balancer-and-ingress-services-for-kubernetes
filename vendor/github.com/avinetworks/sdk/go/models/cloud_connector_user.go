package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudConnectorUser cloud connector user
// swagger:model CloudConnectorUser
type CloudConnectorUser struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.2.1.
	AzureServiceprincipal *AzureServicePrincipalCredentials `json:"azure_serviceprincipal,omitempty"`

	//  Field introduced in 17.2.1.
	AzureUserpass *AzureUserPassCredentials `json:"azure_userpass,omitempty"`

	// Credentials for Google Cloud Platform. Field introduced in 18.2.1.
	GcpCredentials *GCPCredentials `json:"gcp_credentials,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Credentials for Oracle Cloud Infrastructure. Field introduced in 18.2.1,18.1.3.
	OciCredentials *OCICredentials `json:"oci_credentials,omitempty"`

	// password of CloudConnectorUser.
	Password *string `json:"password,omitempty"`

	// private_key of CloudConnectorUser.
	PrivateKey *string `json:"private_key,omitempty"`

	// public_key of CloudConnectorUser.
	PublicKey *string `json:"public_key,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Credentials for Tencent Cloud. Field introduced in 18.2.3.
	TencentCredentials *TencentCredentials `json:"tencent_credentials,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
