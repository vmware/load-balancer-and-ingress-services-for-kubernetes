// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TCPApplicationProfile TCP application profile
// swagger:model TCPApplicationProfile
type TCPApplicationProfile struct {

	// FTP profile configuration. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FtpProfile *FTPProfile `json:"ftp_profile,omitempty"`

	// Select the PKI profile to be associated with the Virtual Service. This profile defines the Certificate Authority and Revocation List. It is a reference to an object of type PKIProfile. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// Enable/Disable the usage of proxy protocol to convey client connection information to the back-end servers.  Valid only for L4 application profiles and TCP proxy. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ProxyProtocolEnabled *bool `json:"proxy_protocol_enabled,omitempty"`

	// Version of proxy protocol to be used to convey client connection information to the back-end servers. Enum options - PROXY_PROTOCOL_VERSION_1, PROXY_PROTOCOL_VERSION_2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- PROXY_PROTOCOL_VERSION_1), Basic edition(Allowed values- PROXY_PROTOCOL_VERSION_1), Enterprise with Cloud Services edition.
	ProxyProtocolVersion *string `json:"proxy_protocol_version,omitempty"`

	// Specifies whether the client side verification is set to none, request or require. Enum options - SSL_CLIENT_CERTIFICATE_NONE, SSL_CLIENT_CERTIFICATE_REQUEST, SSL_CLIENT_CERTIFICATE_REQUIRE. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- SSL_CLIENT_CERTIFICATE_NONE), Basic edition(Allowed values- SSL_CLIENT_CERTIFICATE_NONE), Enterprise with Cloud Services edition.
	SslClientCertificateMode *string `json:"ssl_client_certificate_mode,omitempty"`
}
