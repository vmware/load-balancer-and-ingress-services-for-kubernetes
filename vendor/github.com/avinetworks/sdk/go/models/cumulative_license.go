package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CumulativeLicense cumulative license
// swagger:model CumulativeLicense
type CumulativeLicense struct {

	// Total number of Service Engine cores for burst core based licenses. Field introduced in 17.2.5.
	BurstCores *int32 `json:"burst_cores,omitempty"`

	// Total number of Service Engine cores for core based licenses. Field introduced in 17.2.5.
	Cores *int32 `json:"cores,omitempty"`

	// Total number of Service Engines for host based licenses. Field introduced in 17.2.5.
	MaxSes *int32 `json:"max_ses,omitempty"`

	// Service Engine bandwidth limits for bandwidth based licenses. Field introduced in 17.2.5.
	SeBandwidthLimits []*SEBandwidthLimit `json:"se_bandwidth_limits,omitempty"`

	// Total number of Service Engine sockets for socket based licenses. Field introduced in 17.2.5.
	Sockets *int32 `json:"sockets,omitempty"`

	// Specifies the licensed tier. Enum options - ENTERPRISE_16, ENTERPRISE_18. Field introduced in 17.2.5.
	TierType *string `json:"tier_type,omitempty"`
}
