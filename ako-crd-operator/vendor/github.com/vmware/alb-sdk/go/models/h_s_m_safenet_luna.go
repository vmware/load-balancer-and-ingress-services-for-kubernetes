// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HSMSafenetLuna h s m safenet luna
// swagger:model HSMSafenetLuna
type HSMSafenetLuna struct {

	// Group Number of generated HA Group. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HaGroupNum *uint64 `json:"ha_group_num,omitempty"`

	// Set to indicate HA across more than one servers. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IsHa *bool `json:"is_ha"`

	// Node specific information. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeInfo []*HSMSafenetClientInfo `json:"node_info,omitempty"`

	// SafeNet/Gemalto HSM Servers used for crypto operations. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server []*HSMSafenetLunaServer `json:"server,omitempty"`

	// Generated File - server.pem. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerPem *string `json:"server_pem,omitempty"`

	// If enabled, dedicated network is used to communicate with HSM,else, the management network is used. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseDedicatedNetwork *bool `json:"use_dedicated_network,omitempty"`
}
