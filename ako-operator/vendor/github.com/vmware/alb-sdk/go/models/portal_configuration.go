// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PortalConfiguration portal configuration
// swagger:model PortalConfiguration
type PortalConfiguration struct {

	// Enable/Disable HTTP basic authentication. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowBasicAuthentication *bool `json:"allow_basic_authentication,omitempty"`

	// Force API session timeout after the specified time (in hours). Allowed values are 1-24. Field introduced in 18.2.3. Unit is HOURS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIForceTimeout *uint32 `json:"api_force_timeout,omitempty"`

	// Disable Remote CLI Shell Client access. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableRemoteCliShell *bool `json:"disable_remote_cli_shell,omitempty"`

	// Disable Swagger access. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableSwagger *bool `json:"disable_swagger,omitempty"`

	// Enable/Disable Clickjacking protection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableClickjackingProtection *bool `json:"enable_clickjacking_protection,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableHTTP *bool `json:"enable_http,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableHTTPS *bool `json:"enable_https,omitempty"`

	// HTTP port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPPort *uint32 `json:"http_port,omitempty"`

	// HTTPS port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPSPort *uint32 `json:"https_port,omitempty"`

	// Minimum password length for user accounts. Allowed values are 6-32. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinimumPasswordLength *uint32 `json:"minimum_password_length,omitempty"`

	// Strict checking of password strength for user accounts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PasswordStrengthCheck *bool `json:"password_strength_check,omitempty"`

	// Reference to PKIProfile Config used for CRL validation. It is a reference to an object of type PKIProfile. Field introduced in 30.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PkiprofileRef *string `json:"pkiprofile_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RedirectToHTTPS *bool `json:"redirect_to_https,omitempty"`

	// Certificates for system portal. Maximum 2 allowed. Leave list empty to use system default certs. It is a reference to an object of type SSLKeyAndCertificate. Maximum of 2 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslkeyandcertificateRefs []string `json:"sslkeyandcertificate_refs,omitempty"`

	//  It is a reference to an object of type SSLProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslprofileRef *string `json:"sslprofile_ref,omitempty"`

	// Use UUID in POST object data as UUID of the new object, instead of a generated UUID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseUUIDFromInput *bool `json:"use_uuid_from_input,omitempty"`
}
