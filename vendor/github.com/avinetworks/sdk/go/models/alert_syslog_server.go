package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertSyslogServer alert syslog server
// swagger:model AlertSyslogServer
type AlertSyslogServer struct {

	// Enable anonymous authentication of Syslog Serverwhich will disable server certificate authentication. Field introduced in 17.2.17, 18.2.5.
	AnonAuth *bool `json:"anon_auth,omitempty"`

	// Syslog output format - legacy, RFC 5424, JSON. Enum options - SYSLOG_LEGACY, SYSLOG_RFC5424, SYSLOG_JSON. Field introduced in 17.2.8.
	Format *string `json:"format,omitempty"`

	// Select the PKIProfile containing a CA or list of CA chainswhich will validate the certificate of the syslog server. It is a reference to an object of type PKIProfile. Field introduced in 17.2.17, 18.2.5.
	PkiprofileRef *string `json:"pkiprofile_ref,omitempty"`

	// Select a certificate and key which will be used to authenticate to the syslog server. It is a reference to an object of type SSLKeyAndCertificate. Field introduced in 17.2.17, 18.2.5.
	SslKeyAndCertificateRef *string `json:"ssl_key_and_certificate_ref,omitempty"`

	// The destination Syslog server IP address or hostname.
	// Required: true
	SyslogServer *string `json:"syslog_server"`

	// The destination Syslog server's service port.
	SyslogServerPort *int32 `json:"syslog_server_port,omitempty"`

	// Enable TLS to the syslog server. Field introduced in 17.2.16, 18.2.3.
	TLSEnable *bool `json:"tls_enable,omitempty"`

	// Network protocol to establish syslog session.
	// Required: true
	UDP *bool `json:"udp"`
}
