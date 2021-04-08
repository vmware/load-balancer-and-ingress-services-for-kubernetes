package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CfgState cfg state
// swagger:model CfgState
type CfgState struct {

	// cfg-version synced to follower. .
	CfgVersion *int32 `json:"cfg_version,omitempty"`

	// cfg-version in flight to follower. .
	CfgVersionInFlight *int32 `json:"cfg_version_in_flight,omitempty"`

	// Placeholder for description of property last_changed_time of obj type CfgState field type str  type object
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// reason of CfgState.
	Reason *string `json:"reason,omitempty"`

	// site_uuid to which the object was synced.
	SiteUUID *string `json:"site_uuid,omitempty"`

	// Status of the object. . Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST...
	Status *string `json:"status,omitempty"`

	// object-uuid that is being synced to follower. .
	UUID *string `json:"uuid,omitempty"`
}
