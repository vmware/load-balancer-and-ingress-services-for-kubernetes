// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeList se list
// swagger:model SeList
type SeList struct {

	// Vip is Active on Cloud. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ActiveOnCloud *bool `json:"active_on_cloud,omitempty"`

	// Vip is Active on this ServiceEngine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ActiveOnSe *bool `json:"active_on_se,omitempty"`

	// This flag is set when scaling in an SE in admin down mode. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminDownRequested *bool `json:"admin_down_requested,omitempty"`

	// Attach IP is in progress. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AttachIPInProgress *bool `json:"attach_ip_in_progress,omitempty"`

	// All attempts to program the Vip on Cloud have been made. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudProgrammingDone *bool `json:"cloud_programming_done,omitempty"`

	// Status of Vip on the Cloud. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudProgrammingStatus *string `json:"cloud_programming_status,omitempty"`

	// This flag is set when an SE is admin down or scaling in. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DeleteInProgress *bool `json:"delete_in_progress,omitempty"`

	// Detach IP is in progress. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPInProgress *bool `json:"detach_ip_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FloatingIntfIP []*IPAddr `json:"floating_intf_ip,omitempty"`

	// IPv6 Floating Interface IPs for the RoutingService. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FloatingIntfIp6Addresses []*IPAddr `json:"floating_intf_ip6_addresses,omitempty"`

	// Updated whenever this entry is created. When the sees this has changed, it means that the SE should disrupt, since there was a delete then create, not an update. Field introduced in 18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Incarnation *string `json:"incarnation,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsPortchannel *bool `json:"is_portchannel,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsPrimary *bool `json:"is_primary,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsStandby *bool `json:"is_standby,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Memory *int32 `json:"memory,omitempty"`

	// Management IPv4 address of SE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIP *IPAddr `json:"mgmt_ip,omitempty"`

	// Management IPv6 address of SE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIp6 *IPAddr `json:"mgmt_ip6,omitempty"`

	// VIP Route is revoked as pool went down. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RouteRevokedPoolDown *bool `json:"route_revoked_pool_down,omitempty"`

	// This flag is set when a VS is actively scaling out to this SE. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleoutInProgress *bool `json:"scaleout_in_progress,omitempty"`

	// All attempts to program the Vip on this ServiceEngine have been made. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeProgrammingDone *bool `json:"se_programming_done,omitempty"`

	// Vip is awaiting response from this ServiceEngine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeReadyInProgress *bool `json:"se_ready_in_progress,omitempty"`

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SeRef *string `json:"se_ref"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecIdx *int32 `json:"sec_idx,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SnatIP *IPAddr `json:"snat_ip,omitempty"`

	// IPV6 address for SE snat. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SnatIp6Address *IPAddr `json:"snat_ip6_address,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vcpus *int32 `json:"vcpus,omitempty"`

	//  Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vip6SubnetMask *int32 `json:"vip6_subnet_mask,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipIntfIP *IPAddr `json:"vip_intf_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipIntfList []*SeVipInterfaceList `json:"vip_intf_list,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipIntfMac *string `json:"vip_intf_mac,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipSubnetMask *int32 `json:"vip_subnet_mask,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VlanID *int32 `json:"vlan_id,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vnic []*VsSeVnic `json:"vnic,omitempty"`
}
