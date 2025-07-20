// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorSSlattributes health monitor s slattributes
// swagger:model HealthMonitorSSLAttributes
type HealthMonitorSSlattributes struct {

	// PKI profile used to validate the SSL certificate presented by a server. It is a reference to an object of type PKIProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// Fully qualified DNS hostname which will be used in the TLS SNI extension in server connections indicating SNI is enabled. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerName *string `json:"server_name,omitempty"`

	// Service engines will present this SSL certificate to the server. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslKeyAndCertificateRef *string `json:"ssl_key_and_certificate_ref,omitempty"`

	// SSL profile defines ciphers and SSL versions to be used for healthmonitor traffic to the back-end servers. It is a reference to an object of type SSLProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SslProfileRef *string `json:"ssl_profile_ref"`
}
