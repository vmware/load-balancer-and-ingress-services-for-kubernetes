// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapViolation icap violation
// swagger:model IcapViolation
type IcapViolation struct {

	// The file that ICAP server has identified as containing a violation. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FileName *string `json:"file_name,omitempty"`

	// Action taken by ICAP server in response to this threat. Enum options - ICAP_FILE_NOT_REPAIRED, ICAP_FILE_REPAIRED, ICAP_VIOLATING_SECTION_REMOVED. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Resolution *string `json:"resolution,omitempty"`

	// The name of the threat. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ThreatName *string `json:"threat_name,omitempty"`
}
