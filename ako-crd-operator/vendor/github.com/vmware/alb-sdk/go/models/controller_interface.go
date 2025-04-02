// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerInterface controller interface
// swagger:model ControllerInterface
type ControllerInterface struct {

	// IPv4 default gateway of the interface. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Gateway *IPAddr `json:"gateway,omitempty"`

	// IPv6 default gateway of the interface. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Gateway6 *IPAddr `json:"gateway6,omitempty"`

	// Interface name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IfName *string `json:"if_name,omitempty"`

	// IPv4 prefix of the interface. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IP *IPAddrPrefix `json:"ip,omitempty"`

	// IPv6 prefix of the interface. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6 *IPAddrPrefix `json:"ip6,omitempty"`

	// Interface label like mgmt, secure channel or HSM. Enum options - MGMT, SE_SECURE_CHANNEL, HSM. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Labels []string `json:"labels,omitempty"`

	// Mac address of interface. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MacAddress *string `json:"mac_address,omitempty"`

	// IPv4 address mode DHCP/STATIC. Enum options - DHCP, STATIC, VIP, DOCKER_HOST. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`

	// IPv6 address mode STATIC. Enum options - DHCP, STATIC, VIP, DOCKER_HOST. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Mode6 *string `json:"mode6,omitempty"`

	// Public IP of interface. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PublicIPOrName *IPAddr `json:"public_ip_or_name,omitempty"`

	// Enable V4 IP on this interface. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	V4Enabled *bool `json:"v4_enabled,omitempty"`

	// Enable V6 IP on this interface. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	V6Enabled *bool `json:"v6_enabled,omitempty"`
}
