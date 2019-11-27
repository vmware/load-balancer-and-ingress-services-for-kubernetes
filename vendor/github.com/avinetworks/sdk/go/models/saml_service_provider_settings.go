package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlServiceProviderSettings saml service provider settings
// swagger:model SamlServiceProviderSettings
type SamlServiceProviderSettings struct {

	// FQDN if entity type is DNS_FQDN . Field introduced in 17.2.3.
	Fqdn *string `json:"fqdn,omitempty"`

	// Service Provider Organization Display Name. Field introduced in 17.2.3.
	OrgDisplayName *string `json:"org_display_name,omitempty"`

	// Service Provider Organization Name. Field introduced in 17.2.3.
	OrgName *string `json:"org_name,omitempty"`

	// Service Provider Organization URL. Field introduced in 17.2.3.
	OrgURL *string `json:"org_url,omitempty"`

	// Type of SAML endpoint. Enum options - AUTH_SAML_CLUSTER_VIP, AUTH_SAML_DNS_FQDN, AUTH_SAML_APP_VS. Field introduced in 17.2.3.
	SamlEntityType *string `json:"saml_entity_type,omitempty"`

	// Service Provider node information. Field introduced in 17.2.3.
	SpNodes []*SamlServiceProviderNode `json:"sp_nodes,omitempty"`

	// Service Provider technical contact email. Field introduced in 17.2.3.
	TechContactEmail *string `json:"tech_contact_email,omitempty"`

	// Service Provider technical contact name. Field introduced in 17.2.3.
	TechContactName *string `json:"tech_contact_name,omitempty"`
}
