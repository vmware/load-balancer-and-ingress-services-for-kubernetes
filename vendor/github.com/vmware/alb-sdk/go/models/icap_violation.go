package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IcapViolation icap violation
// swagger:model IcapViolation
type IcapViolation struct {

	// The file that ICAP server has identified as containing a violation. Field introduced in 20.1.3.
	FileName *string `json:"file_name,omitempty"`

	// Action taken by ICAP server in response to this threat. Enum options - ICAP_FILE_NOT_REPAIRED, ICAP_FILE_REPAIRED, ICAP_VIOLATING_SECTION_REMOVED. Field introduced in 20.1.3.
	Resolution *string `json:"resolution,omitempty"`

	// The name of the threat. Field introduced in 20.1.3.
	ThreatName *string `json:"threat_name,omitempty"`
}
