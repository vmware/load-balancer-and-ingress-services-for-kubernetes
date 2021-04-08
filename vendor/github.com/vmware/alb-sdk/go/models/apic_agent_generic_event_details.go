package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApicAgentGenericEventDetails apic agent generic event details
// swagger:model ApicAgentGenericEventDetails
type ApicAgentGenericEventDetails struct {

	// contract_graphs of ApicAgentGenericEventDetails.
	ContractGraphs []string `json:"contract_graphs,omitempty"`

	// lif_cif_attachment of ApicAgentGenericEventDetails.
	LifCifAttachment []string `json:"lif_cif_attachment,omitempty"`

	// lifs of ApicAgentGenericEventDetails.
	Lifs []string `json:"lifs,omitempty"`

	// networks of ApicAgentGenericEventDetails.
	Networks []string `json:"networks,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// service_engine_vnics of ApicAgentGenericEventDetails.
	ServiceEngineVnics []string `json:"service_engine_vnics,omitempty"`

	// tenant_name of ApicAgentGenericEventDetails.
	TenantName *string `json:"tenant_name,omitempty"`

	// Unique object identifier of tenant.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// vnic_network_attachment of ApicAgentGenericEventDetails.
	VnicNetworkAttachment []string `json:"vnic_network_attachment,omitempty"`

	// vs_name of ApicAgentGenericEventDetails.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
