package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EmailConfiguration email configuration
// swagger:model EmailConfiguration
type EmailConfiguration struct {

	// Password for mail server.
	AuthPassword *string `json:"auth_password,omitempty"`

	// Username for mail server.
	AuthUsername *string `json:"auth_username,omitempty"`

	// When set, disables TLS on the connection to the mail server. Field introduced in 17.2.12, 18.1.3, 18.2.1.
	DisableTLS *bool `json:"disable_tls,omitempty"`

	// Email address in From field.
	FromEmail *string `json:"from_email,omitempty"`

	// Mail server host.
	MailServerName *string `json:"mail_server_name,omitempty"`

	// Mail server port.
	MailServerPort *int32 `json:"mail_server_port,omitempty"`

	// Type of SMTP Mail Service. Enum options - SMTP_NONE, SMTP_LOCAL_HOST, SMTP_SERVER, SMTP_ANONYMOUS_SERVER.
	// Required: true
	SMTPType *string `json:"smtp_type"`
}
