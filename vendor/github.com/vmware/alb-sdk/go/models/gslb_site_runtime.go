// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbSiteRuntime gslb site runtime
// swagger:model GslbSiteRuntime
type GslbSiteRuntime struct {

	// This field shadows glb_cfg.clear_on_max_retries. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClearOnMaxRetries *uint32 `json:"clear_on_max_retries,omitempty"`

	// This field tracks the glb-uuid. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlbUUID *string `json:"glb_uuid,omitempty"`

	// This field will provide information on origin(site name) of the health monitoring information. Field introduced in 22.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HealthMonitorInfo *string `json:"health_monitor_info,omitempty"`

	// Carries replication stats for a given site. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReplicationStats *GslbReplicationStats `json:"replication_stats,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RxedSiteHs *GslbSiteHealthStatus `json:"rxed_site_hs,omitempty"`

	// Frequency with which group members communicate. This field shadows glb_cfg.send_interval. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SendInterval *uint32 `json:"send_interval,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteCfg *GslbSiteRuntimeCfg `json:"site_cfg,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteInfo *GslbSiteRuntimeInfo `json:"site_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteStats *GslbSiteRuntimeStats `json:"site_stats,omitempty"`

	// Remap the tenant_uuid to its tenant-name so that we can use the tenant_name directly in remote-site ops. . Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantName *string `json:"tenant_name,omitempty"`

	// This field shadows the glb_cfg.view_id.  . Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ViewID *uint64 `json:"view_id,omitempty"`
}
