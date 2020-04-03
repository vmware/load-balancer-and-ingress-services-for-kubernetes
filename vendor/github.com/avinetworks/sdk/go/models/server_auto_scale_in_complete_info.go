package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScaleInCompleteInfo server auto scale in complete info
// swagger:model ServerAutoScaleInCompleteInfo
type ServerAutoScaleInCompleteInfo struct {

	// Number of nscalein.
	// Required: true
	Nscalein *int32 `json:"nscalein"`

	// UUID of the Pool. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// reason of ServerAutoScaleInCompleteInfo.
	Reason *string `json:"reason,omitempty"`

	//  Enum options - SYSERR_SUCCESS. SYSERR_FAILURE. SYSERR_OUT_OF_MEMORY. SYSERR_NO_ENT. SYSERR_INVAL. SYSERR_ACCESS. SYSERR_FAULT. SYSERR_IO. SYSERR_TIMEOUT. SYSERR_NOT_SUPPORTED. SYSERR_NOT_READY. SYSERR_UPGRADE_IN_PROGRESS. SYSERR_WARM_START_IN_PROGRESS. SYSERR_TRY_AGAIN. SYSERR_NOT_UPGRADING. SYSERR_PENDING. SYSERR_EVENT_GEN_FAILURE. SYSERR_CONFIG_PARAM_MISSING. SYSERR_BAD_REQUEST. SYSERR_TEST1...
	// Required: true
	ReasonCode *string `json:"reason_code"`

	// Placeholder for description of property scaled_in_servers of obj type ServerAutoScaleInCompleteInfo field type str  type object
	ScaledInServers []*ServerID `json:"scaled_in_servers,omitempty"`
}
