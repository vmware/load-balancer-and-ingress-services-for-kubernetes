// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Server server
// swagger:model Server
type Server struct {

	// Name of autoscaling group this server belongs to. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	AutoscalingGroupName *string `json:"autoscaling_group_name,omitempty"`

	// Availability-zone of the server VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// A description of the Server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// (internal-use) Discovered networks providing reachability for server IP. This field is used internally by Avi, not editable by the user. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiscoveredNetworks []*DiscoveredNetwork `json:"discovered_networks,omitempty"`

	// Enable, Disable or Graceful Disable determine if new or existing connections to the server are allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// UID of server in external orchestration systems. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExternalOrchestrationID *string `json:"external_orchestration_id,omitempty"`

	// UUID identifying VM in OpenStack and other external compute. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExternalUUID *string `json:"external_uuid,omitempty"`

	// DNS resolvable name of the server.  May be used in place of the IP address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hostname *string `json:"hostname,omitempty"`

	// IP Address of the server.  Required if there is no resolvable host name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IP *IPAddr `json:"ip"`

	// (internal-use) Geographic location of the server.Currently only for internal usage. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Location *GeoLocation `json:"location,omitempty"`

	// MAC address of server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MacAddress *string `json:"mac_address,omitempty"`

	// (internal-use) This field is used internally by Avi, not editable by the user. It is a reference to an object of type VIMgrNWRuntime. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NwRef *string `json:"nw_ref,omitempty"`

	// Optionally specify the servers port number.  This will override the pool's default server port attribute. Allowed values are 1-65535. Special values are 0- use backend port in pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *int32 `json:"port,omitempty"`

	// Preference order of this member in the group. The DNS Service chooses the member with the lowest preference that is operationally up. Allowed values are 1-128. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	PreferenceOrder *uint32 `json:"preference_order,omitempty"`

	// Header value for custom header persistence. . Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PrstHdrVal *string `json:"prst_hdr_val,omitempty"`

	// Ratio of selecting eligible servers in the pool. Allowed values are 1-20. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ratio *int32 `json:"ratio,omitempty"`

	// Auto resolve server's IP using DNS name. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	ResolveServerByDNS *bool `json:"resolve_server_by_dns,omitempty"`

	// Rewrite incoming Host Header to server name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RewriteHostHeader *bool `json:"rewrite_host_header,omitempty"`

	// Hostname of the node where the server VM or container resides. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerNode *string `json:"server_node,omitempty"`

	// If statically learned. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Static *bool `json:"static,omitempty"`

	// Verify server belongs to a discovered network or reachable via a discovered network. Verify reachable network isn't the OpenStack management network. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VerifyNetwork *bool `json:"verify_network,omitempty"`

	// (internal-use) This field is used internally by Avi, not editable by the user. It is a reference to an object of type VIMgrVMRuntime. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMRef *string `json:"vm_ref,omitempty"`
}
