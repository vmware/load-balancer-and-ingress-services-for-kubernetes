// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPCommunity Ip community
// swagger:model IpCommunity
type IPCommunity struct {

	// Community *string either in aa nn format where aa, nn is within [1,65535] or local-AS|no-advertise|no-export|internet. Field introduced in 17.1.3. Minimum of 1 items required. Maximum of 16 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Community []string `json:"community,omitempty"`

	// Beginning of IP address range. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IPBegin *IPAddr `json:"ip_begin"`

	// End of IP address range. Optional if ip_begin is the only IP address in specified IP range. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPEnd *IPAddr `json:"ip_end,omitempty"`
}
