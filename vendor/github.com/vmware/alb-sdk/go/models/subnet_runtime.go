// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SubnetRuntime subnet runtime
// swagger:model SubnetRuntime
type SubnetRuntime struct {

	// Static IP range runtime. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IPRangeRuntimes []*StaticIPRangeRuntime `json:"ip_range_runtimes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`
}
