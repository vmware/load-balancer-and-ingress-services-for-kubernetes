// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VipSeAssigned vip se assigned
// swagger:model VipSeAssigned
type VipSeAssigned struct {

	// Vip is Active on Cloud. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ActiveOnCloud *bool `json:"active_on_cloud,omitempty"`

	// Vip is Active on this ServiceEngine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ActiveOnSe *bool `json:"active_on_se,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminDownRequested *bool `json:"admin_down_requested,omitempty"`

	// Attach IP is in progress. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AttachIPInProgress *bool `json:"attach_ip_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Connected *bool `json:"connected,omitempty"`

	// Detach IP is in progress. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPInProgress *bool `json:"detach_ip_in_progress,omitempty"`

	// Management IPv4 address of SE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIP *IPAddr `json:"mgmt_ip,omitempty"`

	// Management IPv6 address of SE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIp6 *IPAddr `json:"mgmt_ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Primary *bool `json:"primary,omitempty"`

	//  It is a reference to an object of type ServiceEngine. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleinInProgress *bool `json:"scalein_in_progress,omitempty"`

	// Vip is awaiting scaleout response from this ServiceEngine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScaleoutInProgress *bool `json:"scaleout_in_progress,omitempty"`

	// Vip is awaiting response from this ServiceEngine. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeReadyInProgress *bool `json:"se_ready_in_progress,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SnatIP *IPAddr `json:"snat_ip,omitempty"`

	// IPV6 address for SE snat. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SnatIp6Address *IPAddr `json:"snat_ip6_address,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Standby *bool `json:"standby,omitempty"`
}
