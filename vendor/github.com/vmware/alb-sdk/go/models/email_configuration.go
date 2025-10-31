// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EmailConfiguration email configuration
// swagger:model EmailConfiguration
type EmailConfiguration struct {

	// Password for mail server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthPassword *string `json:"auth_password,omitempty"`

	// Username for mail server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AuthUsername *string `json:"auth_username,omitempty"`

	// When set, disables TLS on the connection to the mail server. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableTLS *bool `json:"disable_tls,omitempty"`

	// Timezone for timestamps in alert emails. Enum options - UTC, AFRICA_ABIDJAN, AFRICA_ACCRA, AFRICA_ADDIS_ABABA, AFRICA_ALGIERS, AFRICA_ASMARA, AFRICA_ASMERA, AFRICA_BAMAKO, AFRICA_BANGUI, AFRICA_BANJUL, AFRICA_BISSAU, AFRICA_BLANTYRE, AFRICA_BRAZZAVILLE, AFRICA_BUJUMBURA, AFRICA_CAIRO, AFRICA_CASABLANCA, AFRICA_CEUTA, AFRICA_CONAKRY, AFRICA_DAKAR, AFRICA_DAR_ES_SALAAM.... Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EmailTimezone *string `json:"email_timezone,omitempty"`

	// Email address in From field. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FromEmail *string `json:"from_email,omitempty"`

	// Friendly name in From field. Field introduced in 21.1.4, 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FromName *string `json:"from_name,omitempty"`

	// Mail server FQDN or IP(v4/v6) address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MailServerName *string `json:"mail_server_name,omitempty"`

	// Mail server port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MailServerPort *uint32 `json:"mail_server_port,omitempty"`

	// Type of SMTP Mail Service. Enum options - SMTP_NONE, SMTP_LOCAL_HOST, SMTP_SERVER, SMTP_ANONYMOUS_SERVER. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SMTPType *string `json:"smtp_type"`
}
