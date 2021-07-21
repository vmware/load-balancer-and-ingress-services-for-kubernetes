// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMUsableNetwork ipam usable network
// swagger:model IpamUsableNetwork
type IPAMUsableNetwork struct {

	// Labels as key value pairs, used for selection of IPAM networks. Field introduced in 20.1.3. Maximum of 1 items allowed.
	Labels []*KeyValueTuple `json:"labels,omitempty"`

	// Network. It is a reference to an object of type Network. Field introduced in 20.1.3.
	// Required: true
	NwRef *string `json:"nw_ref"`
}
