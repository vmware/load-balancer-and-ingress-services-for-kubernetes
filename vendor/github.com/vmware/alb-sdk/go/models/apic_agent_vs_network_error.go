package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApicAgentVsNetworkError apic agent vs network error
// swagger:model ApicAgentVsNetworkError
type ApicAgentVsNetworkError struct {

	// pool_name of ApicAgentVsNetworkError.
	PoolName *string `json:"pool_name,omitempty"`

	// pool_network of ApicAgentVsNetworkError.
	PoolNetwork *string `json:"pool_network,omitempty"`

	// vs_name of ApicAgentVsNetworkError.
	VsName *string `json:"vs_name,omitempty"`

	// vs_network of ApicAgentVsNetworkError.
	VsNetwork *string `json:"vs_network,omitempty"`
}
