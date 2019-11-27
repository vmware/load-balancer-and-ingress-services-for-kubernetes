package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SEBandwidthLimit s e bandwidth limit
// swagger:model SEBandwidthLimit
type SEBandwidthLimit struct {

	// Total number of Service Engines for bandwidth based licenses. Field introduced in 17.2.5.
	Count *int32 `json:"count,omitempty"`

	// Maximum bandwidth allowed by each Service Engine. Enum options - SE_BANDWIDTH_UNLIMITED, SE_BANDWIDTH_25M, SE_BANDWIDTH_200M, SE_BANDWIDTH_1000M, SE_BANDWIDTH_10000M. Field introduced in 17.2.5.
	Type *string `json:"type,omitempty"`
}
