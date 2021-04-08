package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuditComplianceEventInfo audit compliance event info
// swagger:model AuditComplianceEventInfo
type AuditComplianceEventInfo struct {

	// Detailed report of the audit event. Field introduced in 20.1.3.
	DetailedReason *string `json:"detailed_reason,omitempty"`

	// Information identifying physical location for audit event (e.g. Santa Clara (USA), Bengaluru (India)). Field introduced in 20.1.3.
	Location *string `json:"location,omitempty"`

	// Node on which crash is generated. Field introduced in 20.1.4.
	Node *string `json:"node,omitempty"`

	// Crashed core process name. Field introduced in 20.1.4.
	ProcessName *string `json:"process_name,omitempty"`

	// Protocol used for communication to the external entity. Enum options - SSH1_0, TLS1_2, HTTPS1_0, HTTP_PLAIN_TEXT, HTTPS_INSECURE, SSH2_0. Field introduced in 20.1.3.
	// Required: true
	Protocol *string `json:"protocol"`

	// Summarized failure of the transaction (e.g. Invalid request, expired certificate). Field introduced in 20.1.3.
	// Required: true
	Result *string `json:"result"`

	// Subjects of audit event. Field introduced in 20.1.3. Minimum of 1 items required.
	Subjects []*ACSubjectInfo `json:"subjects,omitempty"`

	// Type of audit event. Enum options - AUDIT_INVALID_CREDENTIALS, AUDIT_NAME_RESOLUTION_ERROR, AUDIT_DIAL_X509_ERROR, AUDIT_CORE_GENERATED, AUDIT_SECURE_KEY_EXCHANGE_BAD_REQUEST_FORMAT, AUDIT_SECURE_KEY_EXCHANGE_BAD_CLIENT_TYPE, AUDIT_SECURE_KEY_EXCHANGE_FIELD_NOT_FOUND, AUDIT_SECURE_KEY_EXCHANGE_BAD_FIELD_VALUE, AUDIT_SECURE_KEY_EXCHANGE_INVALID_AUTHORIZATION, AUDIT_SECURE_KEY_EXCHANGE_INTERNAL_ERROR, AUDIT_SECURE_KEY_EXCHANGE_CERTIFICATE_VERIFY_ERROR, AUDIT_SECURE_KEY_EXCHANGE_RESPONSE_ERROR. Field introduced in 20.1.3.
	// Required: true
	Type *string `json:"type"`

	// List of users (username etc) related to the audit event. Field introduced in 20.1.3. Minimum of 1 items required.
	UserIdentities []*ACUserIdentity `json:"user_identities,omitempty"`
}
