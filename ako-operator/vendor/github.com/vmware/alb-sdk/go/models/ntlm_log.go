// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NtlmLog ntlm log
// swagger:model NtlmLog
type NtlmLog struct {

	// Set to true, if request is detected to be NTLM. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NtlmDetected *bool `json:"ntlm_detected,omitempty"`

	// Set the NTLM Status. Enum options - NTLM_AUTHENTICATION_SUCCESS, NTLM_UNAUTHORIZED, NTLM_NEGOTIATION, NTLM_AUTHENTICATION_FAILURE, NTLM_AUTHENTICATED_REQUESTS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NtlmStatus *string `json:"ntlm_status,omitempty"`
}
