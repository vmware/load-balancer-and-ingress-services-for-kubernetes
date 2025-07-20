// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMUsableNetwork ipam usable network
// swagger:model IpamUsableNetwork
type IPAMUsableNetwork struct {

	// Labels as key value pairs, used for selection of IPAM networks. Field introduced in 20.1.3. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Labels []*KeyValueTuple `json:"labels,omitempty"`

	// Network. It is a reference to an object of type Network. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	NwRef *string `json:"nw_ref"`
}
