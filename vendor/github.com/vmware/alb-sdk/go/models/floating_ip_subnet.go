// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FloatingIPSubnet floating Ip subnet
// swagger:model FloatingIpSubnet
type FloatingIPSubnet struct {

	// FloatingIp subnet name if available, else uuid. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// FloatingIp subnet prefix. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Prefix *IPAddrPrefix `json:"prefix,omitempty"`

	// FloatingIp subnet uuid. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
