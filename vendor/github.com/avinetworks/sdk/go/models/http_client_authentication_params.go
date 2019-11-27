package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPClientAuthenticationParams HTTP client authentication params
// swagger:model HTTPClientAuthenticationParams
type HTTPClientAuthenticationParams struct {

	// Auth Profile to use for validating users. It is a reference to an object of type AuthProfile.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// Basic authentication realm to present to a user along with the prompt for credentials.
	Realm *string `json:"realm,omitempty"`

	// Rrequest URI path when the authentication applies.
	RequestURIPath *StringMatch `json:"request_uri_path,omitempty"`

	// type of client authentication. Enum options - HTTP_BASIC_AUTH.
	Type *string `json:"type,omitempty"`
}
