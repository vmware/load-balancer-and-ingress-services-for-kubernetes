package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SamlSettings saml settings
// swagger:model SamlSettings
type SamlSettings struct {

	// Configure remote Identity provider settings. Field introduced in 17.2.3.
	Idp *SamlIdentityProviderSettings `json:"idp,omitempty"`

	// Configure service provider settings for the Controller. Field introduced in 17.2.3.
	// Required: true
	Sp *SamlServiceProviderSettings `json:"sp"`
}
