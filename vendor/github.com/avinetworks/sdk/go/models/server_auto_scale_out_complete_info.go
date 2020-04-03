package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScaleOutCompleteInfo server auto scale out complete info
// swagger:model ServerAutoScaleOutCompleteInfo
type ServerAutoScaleOutCompleteInfo struct {

	// Unique object identifier of launch_config.
	LaunchConfigUUID *string `json:"launch_config_uuid,omitempty"`

	// Number of nscaleout.
	// Required: true
	Nscaleout *int32 `json:"nscaleout"`

	// UUID of the Pool. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// reason of ServerAutoScaleOutCompleteInfo.
	Reason *string `json:"reason,omitempty"`

	//  Enum options - SYSERR_SUCCESS. SYSERR_FAILURE. SYSERR_OUT_OF_MEMORY. SYSERR_NO_ENT. SYSERR_INVAL. SYSERR_ACCESS. SYSERR_FAULT. SYSERR_IO. SYSERR_TIMEOUT. SYSERR_NOT_SUPPORTED. SYSERR_NOT_READY. SYSERR_UPGRADE_IN_PROGRESS. SYSERR_WARM_START_IN_PROGRESS. SYSERR_TRY_AGAIN. SYSERR_NOT_UPGRADING. SYSERR_PENDING. SYSERR_EVENT_GEN_FAILURE. SYSERR_CONFIG_PARAM_MISSING. SYSERR_BAD_REQUEST. SYSERR_TEST1...
	// Required: true
	ReasonCode *string `json:"reason_code"`

	// Placeholder for description of property scaled_out_servers of obj type ServerAutoScaleOutCompleteInfo field type str  type object
	ScaledOutServers []*ServerID `json:"scaled_out_servers,omitempty"`
}
