package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortalConfiguration portal configuration
// swagger:model PortalConfiguration
type PortalConfiguration struct {

	// Enable/Disable HTTP basic authentication.
	AllowBasicAuthentication *bool `json:"allow_basic_authentication,omitempty"`

	// Force API session timeout after the specified time (in hours). Allowed values are 1-24. Field introduced in 18.2.3.
	APIForceTimeout *int32 `json:"api_force_timeout,omitempty"`

	// Disable Remote CLI Shell Client access.
	DisableRemoteCliShell *bool `json:"disable_remote_cli_shell,omitempty"`

	// Disable Swagger access. Field introduced in 18.2.3.
	DisableSwagger *bool `json:"disable_swagger,omitempty"`

	// Enable/Disable Clickjacking protection.
	EnableClickjackingProtection *bool `json:"enable_clickjacking_protection,omitempty"`

	// Placeholder for description of property enable_http of obj type PortalConfiguration field type str  type boolean
	EnableHTTP *bool `json:"enable_http,omitempty"`

	// Placeholder for description of property enable_https of obj type PortalConfiguration field type str  type boolean
	EnableHTTPS *bool `json:"enable_https,omitempty"`

	// HTTP port.
	HTTPPort *int32 `json:"http_port,omitempty"`

	// HTTPS port.
	HTTPSPort *int32 `json:"https_port,omitempty"`

	// Strict checking of password strength for user accounts.
	PasswordStrengthCheck *bool `json:"password_strength_check,omitempty"`

	// Placeholder for description of property redirect_to_https of obj type PortalConfiguration field type str  type boolean
	RedirectToHTTPS *bool `json:"redirect_to_https,omitempty"`

	// Certificates for system portal. Maximum 2 allowed. Leave list empty to use system default certs. It is a reference to an object of type SSLKeyAndCertificate.
	SslkeyandcertificateRefs []string `json:"sslkeyandcertificate_refs,omitempty"`

	//  It is a reference to an object of type SSLProfile.
	SslprofileRef *string `json:"sslprofile_ref,omitempty"`

	// Use UUID in POST object data as UUID of the new object, instead of a generated UUID.
	UseUUIDFromInput *bool `json:"use_uuid_from_input,omitempty"`
}
