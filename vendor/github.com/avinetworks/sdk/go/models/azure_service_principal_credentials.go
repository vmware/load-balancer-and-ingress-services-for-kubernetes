package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureServicePrincipalCredentials azure service principal credentials
// swagger:model AzureServicePrincipalCredentials
type AzureServicePrincipalCredentials struct {

	// Application Id created for Avi Controller. Required for application id based authentication only. Field introduced in 17.2.1.
	ApplicationID *string `json:"application_id,omitempty"`

	// Authentication token created for the Avi Controller application. Required for application id based authentication only. Field introduced in 17.2.1.
	AuthenticationToken *string `json:"authentication_token,omitempty"`

	// Tenant ID for the subscription. Required for application id based authentication only. Field introduced in 17.2.1.
	TenantID *string `json:"tenant_id,omitempty"`
}
