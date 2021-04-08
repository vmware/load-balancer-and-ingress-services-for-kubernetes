package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IcapLog icap log
// swagger:model IcapLog
type IcapLog struct {

	// Denotes whether the content was processed by ICAP server and an action was taken. Enum options - ICAP_DISABLED, ICAP_PASSED, ICAP_MODIFIED, ICAP_BLOCKED, ICAP_FAILED. Field introduced in 20.1.1.
	Action *string `json:"action,omitempty"`

	// Logs for the HTTP request's content sent to the ICAP server. Field introduced in 20.1.1.
	RequestLogs []*IcapRequestLog `json:"request_logs,omitempty"`
}
