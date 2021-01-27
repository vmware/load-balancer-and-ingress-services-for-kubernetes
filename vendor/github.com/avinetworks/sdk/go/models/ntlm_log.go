package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NtlmLog ntlm log
// swagger:model NtlmLog
type NtlmLog struct {

	// Set to true, if request is detected to be NTLM. Field introduced in 20.1.3.
	NtlmDetected *bool `json:"ntlm_detected,omitempty"`

	// Set the NTLM Status. Enum options - NTLM_AUTHENTICATION_SUCCESS, NTLM_UNAUTHORIZED, NTLM_NEGOTIATION, NTLM_AUTHENTICATION_FAILURE, NTLM_AUTHENTICATED_REQUESTS. Field introduced in 20.1.3.
	NtlmStatus *string `json:"ntlm_status,omitempty"`
}
