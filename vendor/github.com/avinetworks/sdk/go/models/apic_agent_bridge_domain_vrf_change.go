package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApicAgentBridgeDomainVrfChange apic agent bridge domain vrf change
// swagger:model ApicAgentBridgeDomainVrfChange
type ApicAgentBridgeDomainVrfChange struct {

	// bridge_domain of ApicAgentBridgeDomainVrfChange.
	BridgeDomain *string `json:"bridge_domain,omitempty"`

	// new_vrf of ApicAgentBridgeDomainVrfChange.
	NewVrf *string `json:"new_vrf,omitempty"`

	// old_vrf of ApicAgentBridgeDomainVrfChange.
	OldVrf *string `json:"old_vrf,omitempty"`

	// pool_list of ApicAgentBridgeDomainVrfChange.
	PoolList []string `json:"pool_list,omitempty"`

	// vs_list of ApicAgentBridgeDomainVrfChange.
	VsList []string `json:"vs_list,omitempty"`
}
