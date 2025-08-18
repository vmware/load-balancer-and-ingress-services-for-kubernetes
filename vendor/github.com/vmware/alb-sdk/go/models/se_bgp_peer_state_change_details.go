// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeBgpPeerStateChangeDetails se bgp peer state change details
// swagger:model SeBgpPeerStateChangeDetails
type SeBgpPeerStateChangeDetails struct {

	// IP address of BGP peer. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PeerIP *string `json:"peer_ip"`

	// BGP peer state. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PeerState *string `json:"peer_state"`

	// Name of Virtual Routing Context in which BGP is configured. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfName *string `json:"vrf_name,omitempty"`
}
