// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlServiceProviderSettings saml service provider settings
// swagger:model SamlServiceProviderSettings
type SamlServiceProviderSettings struct {

	// FQDN if entity type is DNS_FQDN . Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// Service Provider Organization Display Name. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OrgDisplayName *string `json:"org_display_name,omitempty"`

	// Service Provider Organization Name. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OrgName *string `json:"org_name,omitempty"`

	// Service Provider Organization URL. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OrgURL *string `json:"org_url,omitempty"`

	// Type of SAML endpoint. Enum options - AUTH_SAML_CLUSTER_VIP, AUTH_SAML_DNS_FQDN, AUTH_SAML_APP_VS. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlEntityType *string `json:"saml_entity_type,omitempty"`

	// Service Provider node information. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SpNodes []*SamlServiceProviderNode `json:"sp_nodes,omitempty"`

	// Service Provider technical contact email. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TechContactEmail *string `json:"tech_contact_email,omitempty"`

	// Service Provider technical contact name. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TechContactName *string `json:"tech_contact_name,omitempty"`
}
