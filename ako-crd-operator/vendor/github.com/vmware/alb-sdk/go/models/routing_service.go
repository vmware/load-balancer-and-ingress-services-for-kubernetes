// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RoutingService routing service
// swagger:model RoutingService
type RoutingService struct {

	// Advertise reachability of backend server networks via ADC through BGP for default gateway feature. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdvertiseBackendNetworks *bool `json:"advertise_backend_networks,omitempty"`

	// Enable auto gateway to save and use the same L2 path to send the return traffic. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAutoGateway *bool `json:"enable_auto_gateway,omitempty"`

	// Service Engine acts as Default Gateway for this service. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRouting *bool `json:"enable_routing,omitempty"`

	// Enable VIP on all interfaces of this service. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use Virtual MAC address for interfaces on which floating interface IPs are placed. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableVMAC *bool `json:"enable_vmac,omitempty"`

	// Floating Interface IPs for the RoutingService. Field introduced in 18.2.5. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FloatingIntfIP []*IPAddr `json:"floating_intf_ip,omitempty"`

	// IPv6 Floating Interface IPs for the RoutingService. Field introduced in 22.1.6, 30.2.1. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingIntfIp6Addresses []*IPAddr `json:"floating_intf_ip6_addresses,omitempty"`

	// If ServiceEngineGroup is configured for Legacy 1+1 Active Standby HA Mode, IPv6 Floating IP's will be advertised only by the Active SE in the Pair. Virtual Services in this group must be disabled/enabled for any changes to the IPv6 Floating IP's to take effect. Only active SE hosting VS tagged with Active Standby SE 2 Tag will advertise this floating IP when manual load distribution is enabled. Field introduced in 22.1.6, 30.2.1. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingIntfIp6Se2Addresses []*IPAddr `json:"floating_intf_ip6_se_2_addresses,omitempty"`

	// If ServiceEngineGroup is configured for Legacy 1+1 Active Standby HA Mode, Floating IP's will be advertised only by the Active SE in the Pair. Virtual Services in this group must be disabled/enabled for any changes to the Floating IP's to take effect. Only active SE hosting VS tagged with Active Standby SE 2 Tag will advertise this floating IP when manual load distribution is enabled. Field introduced in 18.2.5. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FloatingIntfIPSe2 []*IPAddr `json:"floating_intf_ip_se_2,omitempty"`

	// Routing Service related Flow profile information. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowtableProfile *FlowtableProfile `json:"flowtable_profile,omitempty"`

	// Enable graceful restart feature in routing service. For example, BGP. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GracefulRestart *bool `json:"graceful_restart,omitempty"`

	// NAT policy for outbound NAT functionality. This is done in post-routing. It is a reference to an object of type NatPolicy. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NatPolicyRef *string `json:"nat_policy_ref,omitempty"`

	// For IP Routing feature, enabling this knob will fallback to routing through Linux, by default routing is done via Service Engine data-path. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoutingByLinuxIpstack *bool `json:"routing_by_linux_ipstack,omitempty"`
}
