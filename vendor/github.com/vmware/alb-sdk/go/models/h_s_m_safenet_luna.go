package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMSafenetLuna h s m safenet luna
// swagger:model HSMSafenetLuna
type HSMSafenetLuna struct {

	// Group Number of generated HA Group.
	HaGroupNum *int64 `json:"ha_group_num,omitempty"`

	// Set to indicate HA across more than one servers.
	// Required: true
	IsHa *bool `json:"is_ha"`

	// Node specific information.
	NodeInfo []*HSMSafenetClientInfo `json:"node_info,omitempty"`

	// SafeNet/Gemalto HSM Servers used for crypto operations.
	Server []*HSMSafenetLunaServer `json:"server,omitempty"`

	// Generated File - server.pem.
	ServerPem *string `json:"server_pem,omitempty"`

	// If enabled, dedicated network is used to communicate with HSM,else, the management network is used.
	UseDedicatedNetwork *bool `json:"use_dedicated_network,omitempty"`
}
