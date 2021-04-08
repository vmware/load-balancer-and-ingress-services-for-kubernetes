package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OperationalStatus operational status
// swagger:model OperationalStatus
type OperationalStatus struct {

	// Placeholder for description of property last_changed_time of obj type OperationalStatus field type str  type object
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// reason of OperationalStatus.
	Reason []string `json:"reason,omitempty"`

	// Number of reason_code.
	ReasonCode *int64 `json:"reason_code,omitempty"`

	// reason_code_string of OperationalStatus.
	ReasonCodeString *string `json:"reason_code_string,omitempty"`

	//  Enum options - OPER_UP, OPER_DOWN, OPER_CREATING, OPER_RESOURCES, OPER_INACTIVE, OPER_DISABLED, OPER_UNUSED, OPER_UNKNOWN, OPER_PROCESSING, OPER_INITIALIZING, OPER_ERROR_DISABLED, OPER_AWAIT_MANUAL_PLACEMENT, OPER_UPGRADING, OPER_SE_PROCESSING, OPER_PARTITIONED, OPER_DISABLING, OPER_FAILED, OPER_UNAVAIL.
	State *string `json:"state,omitempty"`
}
