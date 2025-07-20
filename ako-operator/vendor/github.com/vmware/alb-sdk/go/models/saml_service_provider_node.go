// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlServiceProviderNode saml service provider node
// swagger:model SamlServiceProviderNode
type SamlServiceProviderNode struct {

	// Globally unique entityID for this node. Entity ID on the IDP should match this. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EntityID *string `json:"entity_id,omitempty"`

	// Refers to the Cluster name identifier (Virtual IP or FQDN). Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Service Engines will use this SSL certificate to sign assertions going to the IdP. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SigningSslKeyAndCertificateRef *string `json:"signing_ssl_key_and_certificate_ref,omitempty"`

	// Single Signon URL to be programmed on the IDP. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SingleSignonURL *string `json:"single_signon_url,omitempty"`
}
