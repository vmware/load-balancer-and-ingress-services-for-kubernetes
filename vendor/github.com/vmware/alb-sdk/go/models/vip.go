// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Vip vip
// swagger:model Vip
type Vip struct {

	// Auto-allocate floating/elastic IP from the Cloud infrastructure. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AutoAllocateFloatingIP *bool `json:"auto_allocate_floating_ip,omitempty"`

	// Auto-allocate VIP from the provided subnet. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoAllocateIP *bool `json:"auto_allocate_ip,omitempty"`

	// Specifies whether to auto-allocate only a V4 address, only a V6 address, or one of each type. Enum options - V4_ONLY, V6_ONLY, V4_V6. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- V4_ONLY), Basic edition(Allowed values- V4_ONLY), Enterprise with Cloud Services edition.
	AutoAllocateIPType *string `json:"auto_allocate_ip_type,omitempty"`

	// Availability-zone to place the Virtual Service. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// (internal-use) FIP allocated by Avi in the Cloud infrastructure. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AviAllocatedFip *bool `json:"avi_allocated_fip,omitempty"`

	// (internal-use) VIP allocated by Avi in the Cloud infrastructure. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AviAllocatedVip *bool `json:"avi_allocated_vip,omitempty"`

	// Discovered networks providing reachability for client facing Vip IP. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiscoveredNetworks []*DiscoveredNetwork `json:"discovered_networks,omitempty"`

	// Enable or disable the Vip. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Floating IPv4 to associate with this Vip. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingIP *IPAddr `json:"floating_ip,omitempty"`

	// Floating IPv6 address to associate with this Vip. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingIp6 *IPAddr `json:"floating_ip6,omitempty"`

	// If auto_allocate_floating_ip is True and more than one floating-ip subnets exist, then the subnet for the floating IPv6 address allocation. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingSubnet6UUID *string `json:"floating_subnet6_uuid,omitempty"`

	// If auto_allocate_floating_ip is True and more than one floating-ip subnets exist, then the subnet for the floating IP address allocation. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingSubnetUUID *string `json:"floating_subnet_uuid,omitempty"`

	// IPv6 Address of the Vip. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ip6Address *IPAddr `json:"ip6_address,omitempty"`

	// IPv4 Address of the VIP. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// Subnet and/or Network for allocating VirtualService IP by IPAM Provider module. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAMNetworkSubnet *IPNetworkSubnet `json:"ipam_network_subnet,omitempty"`

	// Manually override the network on which the Vip is placed. It is a reference to an object of type Network. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkRef *string `json:"network_ref,omitempty"`

	// Placement networks/subnets to use for vip placement. Field introduced in 18.2.5. Maximum of 10 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PlacementNetworks []*VipPlacementNetwork `json:"placement_networks,omitempty"`

	// (internal-use) Network port assigned to the Vip IP address. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PortUUID *string `json:"port_uuid,omitempty"`

	// Mask applied for the Vip, non-default mask supported only for wildcard Vip. Allowed values are 0-32. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 32), Basic edition(Allowed values- 32), Enterprise with Cloud Services edition.
	PrefixLength *uint32 `json:"prefix_length,omitempty"`

	// Subnet providing reachability for client facing Vip IP. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// Subnet providing reachability for client facing Vip IPv6. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	Subnet6 *IPAddrPrefix `json:"subnet6,omitempty"`

	// If auto_allocate_ip is True, then the subnet for the Vip IPv6 address allocation. This field is applicable only if the VirtualService belongs to an Openstack or AWS cloud, in which case it is mandatory, if auto_allocate is selected. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	Subnet6UUID *string `json:"subnet6_uuid,omitempty"`

	// If auto_allocate_ip is True, then the subnet for the Vip IP address allocation. This field is applicable only if the VirtualService belongs to an Openstack or AWS cloud, in which case it is mandatory, if auto_allocate is selected. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`

	// Unique ID associated with the vip. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	VipID *string `json:"vip_id"`
}
