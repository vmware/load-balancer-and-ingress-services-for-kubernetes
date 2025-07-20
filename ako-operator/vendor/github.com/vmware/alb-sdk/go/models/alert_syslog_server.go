// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertSyslogServer alert syslog server
// swagger:model AlertSyslogServer
type AlertSyslogServer struct {

	// Enable anonymous authentication of Syslog Serverwhich will disable server certificate authentication. Field introduced in 17.2.17, 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnonAuth *bool `json:"anon_auth,omitempty"`

	// Syslog output format - legacy, RFC 5424, JSON. Enum options - SYSLOG_LEGACY, SYSLOG_RFC5424, SYSLOG_JSON, SYSLOG_RFC5425_ENHANCED. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Format *string `json:"format,omitempty"`

	// Select the PKIProfile containing a CA or list of CA chainswhich will validate the certificate of the syslog server. It is a reference to an object of type PKIProfile. Field introduced in 17.2.17, 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PkiprofileRef *string `json:"pkiprofile_ref,omitempty"`

	// Select a certificate and key which will be used to authenticate to the syslog server. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 17.2.17, 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslKeyAndCertificateRef *string `json:"ssl_key_and_certificate_ref,omitempty"`

	// strict verificiation of certificate given by the server. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StrictCertVerify *bool `json:"strict_cert_verify,omitempty"`

	// The destination Syslog server IP(v4/v6) address or FQDN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SyslogServer *string `json:"syslog_server"`

	// The destination Syslog server's service port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyslogServerPort *uint32 `json:"syslog_server_port,omitempty"`

	// Enable TLS to the syslog server. Field introduced in 17.2.16, 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TLSEnable *bool `json:"tls_enable,omitempty"`

	// Network protocol to establish syslog session. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	UDP *bool `json:"udp"`
}
