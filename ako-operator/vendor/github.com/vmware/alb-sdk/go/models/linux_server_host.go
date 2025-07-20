// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LinuxServerHost linux server host
// swagger:model LinuxServerHost
type LinuxServerHost struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostAttr []*HostAttributes `json:"host_attr,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HostIP *IPAddr `json:"host_ip"`

	// Node's availability zone. ServiceEngines belonging to the availability zone will be rebooted during a manual DR failover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeAvailabilityZone *string `json:"node_availability_zone,omitempty"`

	// The SE Group association for the SE. If None, then 'Default-Group' SEGroup is associated with the SE. It is a reference to an object of type ServiceEngineGroup. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupRef *string `json:"se_group_ref,omitempty"`
}
