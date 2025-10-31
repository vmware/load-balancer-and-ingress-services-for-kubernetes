// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SAMLSPConfig s a m l s p config
// swagger:model SAMLSPConfig
type SAMLSPConfig struct {

	// Index to be used in the AssertionConsumerServiceIndex attribute of the Authentication request, if the authn_req_acs_type is set to Use AssertionConsumerServiceIndex. Allowed values are 0-64. Field introduced in 21.1.6, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AcsIndex *int32 `json:"acs_index,omitempty"`

	// Option to set the ACS attributes in the AuthnRequest . Enum options - SAML_AUTHN_REQ_ACS_TYPE_URL, SAML_AUTHN_REQ_ACS_TYPE_INDEX, SAML_AUTHN_REQ_ACS_TYPE_NONE. Field introduced in 21.1.6, 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AuthnReqAcsType *string `json:"authn_req_acs_type"`

	// HTTP cookie name for authenticated session. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CookieName *string `json:"cookie_name,omitempty"`

	// Cookie timeout in minutes. Allowed values are 1-1440. Field introduced in 18.2.3. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CookieTimeout *int32 `json:"cookie_timeout,omitempty"`

	// Globally unique SAML entityID for this node. The SAML application entity ID on the IDP should match this. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	EntityID *string `json:"entity_id"`

	// Key to generate the cookie. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Key []*HTTPCookiePersistenceKey `json:"key,omitempty"`

	// SP will use this SSL certificate to sign requests going to the IdP and decrypt the assertions coming from IdP. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SigningSslKeyAndCertificateRef *string `json:"signing_ssl_key_and_certificate_ref,omitempty"`

	// SAML Single Signon endpoint to receive the Authentication response. This also specifies the destination endpoint to be configured for this application on the IDP. If the authn_req_acs_type is set to 'Use AssertionConsumerServiceURL', this endpoint will be sent in the AssertionConsumerServiceURL attribute of the Authentication request. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SingleSignonURL *string `json:"single_signon_url"`

	// SAML SP metadata for this application. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	SpMetadata *string `json:"sp_metadata,omitempty"`

	// By enabling this field IdP can control how long the SP session can exist through the SessionNotOnOrAfter field in the AuthNStatement of SAML Response. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseIdpSessionTimeout *bool `json:"use_idp_session_timeout,omitempty"`
}
