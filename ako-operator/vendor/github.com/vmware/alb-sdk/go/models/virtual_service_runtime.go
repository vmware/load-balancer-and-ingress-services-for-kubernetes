// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VirtualServiceRuntime virtual service runtime
// swagger:model VirtualServiceRuntime
type VirtualServiceRuntime struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerIP *string `json:"controller_ip,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatapathDebug *DebugVirtualService `json:"datapath_debug,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EastWest *bool `json:"east_west,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GslbDNSUpdate *GslbDNSUpdate `json:"gslb_dns_update,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAMDNSRecords []*DNSRecord `json:"ipam_dns_records,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsDNSVs *bool `json:"is_dns_vs,omitempty"`

	// Number of times keys have been rotated. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeyRotationCount *uint64 `json:"key_rotation_count,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Timestamp of the last key rotation. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastKeyRotationTime *TimeStamp `json:"last_key_rotation_time,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Lif []string `json:"lif,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ManualPlacement *bool `json:"manual_placement,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarkedForDelete *bool `json:"marked_for_delete,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsMgrPort *string `json:"metrics_mgr_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumAdditionalSe *int32 `json:"num_additional_se,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OnePlusOneHa *bool `json:"one_plus_one_ha,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevControllerIP *string `json:"prev_controller_ip,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PrevMetricsMgrPort *string `json:"prev_metrics_mgr_port,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RedisDb *int32 `json:"redis_db,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RedisPort *int32 `json:"redis_port,omitempty"`

	// Runtime info from security_manager. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecMgrInfo *SecurityMgrRuntime `json:"sec_mgr_info,omitempty"`

	// Enable Service Engines to elect a primary amongst themselves in the absence of connectivity to controller. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SelfSeElection *bool `json:"self_se_election,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TLSTicketKey []*TLSTicket `json:"tls_ticket_key,omitempty"`

	//  Enum options - VS_TYPE_NORMAL, VS_TYPE_VH_PARENT, VS_TYPE_VH_CHILD. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Version number of the SE List update. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Version *uint64 `json:"version,omitempty"`

	//  It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VhChildVsRef []string `json:"vh_child_vs_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipRuntime []*VipRuntime `json:"vip_runtime,omitempty"`

	// VS update request received before warmstart finished. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUpdatePending *VirtualService `json:"vs_update_pending,omitempty"`
}
