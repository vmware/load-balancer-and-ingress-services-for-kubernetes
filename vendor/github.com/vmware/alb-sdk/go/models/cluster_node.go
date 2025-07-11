// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNode cluster node
// swagger:model ClusterNode
type ClusterNode struct {

	// Optional service categories that a node can be assigned (e.g. SYSTEM, INFRASTRUCTURE or ANALYTICS). Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Categories []string `json:"categories,omitempty"`

	// Interface details of the controller node. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Interfaces []*ControllerInterface `json:"interfaces,omitempty"`

	// V4 IP address of controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// V6 IP address of controller VM. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6 *IPAddr `json:"ip6,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// The password we will use when authenticating with this node (Not persisted). Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	// Public IP address or hostname of the controller VM. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PublicIPOrName *IPAddr `json:"public_ip_or_name,omitempty"`

	// Static routes configured on the controller node. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StaticRoutes []*StaticRoute `json:"static_routes,omitempty"`

	// Hostname assigned to this controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMHostname *string `json:"vm_hostname,omitempty"`

	// Managed object reference of this controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMMor *string `json:"vm_mor,omitempty"`

	// Name of the controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMName *string `json:"vm_name,omitempty"`

	// UUID on the controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMUUID *string `json:"vm_uuid,omitempty"`
}
