package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScaleOutInfo server auto scale out info
// swagger:model ServerAutoScaleOutInfo
type ServerAutoScaleOutInfo struct {

	// alertconfig_name of ServerAutoScaleOutInfo.
	AlertconfigName *string `json:"alertconfig_name,omitempty"`

	//  It is a reference to an object of type AlertConfig.
	AlertconfigRef *string `json:"alertconfig_ref,omitempty"`

	// Placeholder for description of property available_capacity of obj type ServerAutoScaleOutInfo field type str  type number
	AvailableCapacity *float64 `json:"available_capacity,omitempty"`

	// Placeholder for description of property load of obj type ServerAutoScaleOutInfo field type str  type number
	Load *float64 `json:"load,omitempty"`

	// Number of num_scaleout_servers.
	// Required: true
	NumScaleoutServers *int32 `json:"num_scaleout_servers"`

	// Number of num_servers_up.
	// Required: true
	NumServersUp *int32 `json:"num_servers_up"`

	// UUID of the Pool. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// reason of ServerAutoScaleOutInfo.
	Reason *string `json:"reason,omitempty"`

	//  Enum options - SYSERR_SUCCESS. SYSERR_FAILURE. SYSERR_OUT_OF_MEMORY. SYSERR_NO_ENT. SYSERR_INVAL. SYSERR_ACCESS. SYSERR_FAULT. SYSERR_IO. SYSERR_TIMEOUT. SYSERR_NOT_SUPPORTED. SYSERR_NOT_READY. SYSERR_UPGRADE_IN_PROGRESS. SYSERR_WARM_START_IN_PROGRESS. SYSERR_TRY_AGAIN. SYSERR_NOT_UPGRADING. SYSERR_PENDING. SYSERR_EVENT_GEN_FAILURE. SYSERR_CONFIG_PARAM_MISSING. SYSERR_BAD_REQUEST. SYSERR_TEST1...
	ReasonCode *string `json:"reason_code,omitempty"`
}
