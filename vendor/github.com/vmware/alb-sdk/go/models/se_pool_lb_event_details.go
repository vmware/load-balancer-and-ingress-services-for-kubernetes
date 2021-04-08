package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SePoolLbEventDetails se pool lb event details
// swagger:model SePoolLbEventDetails
type SePoolLbEventDetails struct {

	// Reason code for load balancing failure. Enum options - PERSISTENT_SERVER_INVALID, PERSISTENT_SERVER_DOWN, SRVR_DOWN, ADD_PENDING, SLOW_START_MAX_CONN, MAX_CONN, NO_LPORT, SUSPECT_STATE, MAX_CONN_RATE, CAPEST_RAND_MAX_CONN, GET_NEXT.
	FailureCode *string `json:"failure_code,omitempty"`

	// Pool name. It is a reference to an object of type Pool.
	Pool *string `json:"pool,omitempty"`

	// Reason for Load balancing failure.
	Reason *string `json:"reason,omitempty"`

	// UUID of event generator.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}
