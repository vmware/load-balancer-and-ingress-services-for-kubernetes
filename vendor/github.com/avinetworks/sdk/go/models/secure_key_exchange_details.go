package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecureKeyExchangeDetails secure key exchange details
// swagger:model SecureKeyExchangeDetails
type SecureKeyExchangeDetails struct {

	// Controller managememt IP for secure key exchange between controller and SE. Field introduced in 20.1.3.
	CtlrMgmtIP *string `json:"ctlr_mgmt_ip,omitempty"`

	// Controller public IP for secure key exchange between controller and SE. Field introduced in 20.1.3.
	CtlrPublicIP *string `json:"ctlr_public_ip,omitempty"`

	// Error message if secure key exchange failed. Field introduced in 20.1.3.
	Error *string `json:"error,omitempty"`

	// Follower IP for secure key exchange between controller and controller. Field introduced in 20.1.3.
	FollowerIP *string `json:"follower_ip,omitempty"`

	// Leader IP for secure key exchange between controller and controller. Field introduced in 20.1.3.
	LeaderIP *string `json:"leader_ip,omitempty"`

	// name of SE/controller who initiates the secure key exchange. Field introduced in 20.1.3.
	Name *string `json:"name,omitempty"`

	// SE IP for secure key exchange between controller and SE. Field introduced in 20.1.3.
	SeIP *string `json:"se_ip,omitempty"`

	// IP address of the client. Field introduced in 20.1.3.
	SourceIP *string `json:"source_ip,omitempty"`

	// Status. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST.... Field introduced in 20.1.3.
	Status *string `json:"status,omitempty"`

	// uuid of SE/controller who initiates the secure key exchange. Field introduced in 20.1.3.
	UUID *string `json:"uuid,omitempty"`
}
