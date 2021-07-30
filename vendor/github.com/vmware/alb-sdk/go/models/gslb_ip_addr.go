// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbIPAddr gslb Ip addr
// swagger:model GslbIpAddr
type GslbIPAddr struct {

	// Public IP address of the pool member. Field introduced in 17.1.2.
	IP *IPAddr `json:"ip,omitempty"`
}
