package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpswitchingAction httpswitching action
// swagger:model HTTPSwitchingAction
type HttpswitchingAction struct {

	// Content switching action type. Enum options - HTTP_SWITCHING_SELECT_POOL, HTTP_SWITCHING_SELECT_LOCAL, HTTP_SWITCHING_SELECT_POOLGROUP.
	// Required: true
	Action *string `json:"action"`

	// File from which to serve local response to the request.
	File *HTTPLocalFile `json:"file,omitempty"`

	// UUID of the pool group to serve the request. It is a reference to an object of type PoolGroup.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// UUID of the pool of servers to serve the request. It is a reference to an object of type Pool.
	PoolRef *string `json:"pool_ref,omitempty"`

	// Specific pool server to select.
	Server *PoolServer `json:"server,omitempty"`

	// HTTP status code to use when serving local response. Enum options - HTTP_LOCAL_RESPONSE_STATUS_CODE_200, HTTP_LOCAL_RESPONSE_STATUS_CODE_204, HTTP_LOCAL_RESPONSE_STATUS_CODE_403, HTTP_LOCAL_RESPONSE_STATUS_CODE_404, HTTP_LOCAL_RESPONSE_STATUS_CODE_429, HTTP_LOCAL_RESPONSE_STATUS_CODE_501.
	StatusCode *string `json:"status_code,omitempty"`
}
