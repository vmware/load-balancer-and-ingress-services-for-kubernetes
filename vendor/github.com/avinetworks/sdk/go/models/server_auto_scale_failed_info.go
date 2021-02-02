package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerAutoScaleFailedInfo server auto scale failed info
// swagger:model ServerAutoScaleFailedInfo
type ServerAutoScaleFailedInfo struct {

	// Number of num_scalein_servers.
	// Required: true
	NumScaleinServers *int32 `json:"num_scalein_servers"`

	// Number of num_servers_up.
	// Required: true
	NumServersUp *int32 `json:"num_servers_up"`

	// UUID of the Pool. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// reason of ServerAutoScaleFailedInfo.
	Reason *string `json:"reason,omitempty"`

	//  Enum options - SYSERR_SUCCESS. SYSERR_FAILURE. SYSERR_OUT_OF_MEMORY. SYSERR_NO_ENT. SYSERR_INVAL. SYSERR_ACCESS. SYSERR_FAULT. SYSERR_IO. SYSERR_TIMEOUT. SYSERR_NOT_SUPPORTED. SYSERR_NOT_READY. SYSERR_UPGRADE_IN_PROGRESS. SYSERR_WARM_START_IN_PROGRESS. SYSERR_TRY_AGAIN. SYSERR_NOT_UPGRADING. SYSERR_PENDING. SYSERR_EVENT_GEN_FAILURE. SYSERR_CONFIG_PARAM_MISSING. SYSERR_RANGE. SYSERR_BAD_REQUEST...
	// Required: true
	ReasonCode *string `json:"reason_code"`
}
