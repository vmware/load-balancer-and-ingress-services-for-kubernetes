package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteRuntime gslb site runtime
// swagger:model GslbSiteRuntime
type GslbSiteRuntime struct {

	// This field shadows glb_cfg.clear_on_max_retries. Field introduced in 17.2.5.
	ClearOnMaxRetries *int32 `json:"clear_on_max_retries,omitempty"`

	// This field tracks the glb-uuid. Field introduced in 17.2.5.
	GlbUUID *string `json:"glb_uuid,omitempty"`

	// Placeholder for description of property rxed_site_hs of obj type GslbSiteRuntime field type str  type object
	RxedSiteHs *GslbSiteHealthStatus `json:"rxed_site_hs,omitempty"`

	// Frequency with which group members communicate. This field shadows glb_cfg.send_interval. Field introduced in 17.2.5.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// Placeholder for description of property site_cfg of obj type GslbSiteRuntime field type str  type object
	SiteCfg *GslbSiteRuntimeCfg `json:"site_cfg,omitempty"`

	// Placeholder for description of property site_info of obj type GslbSiteRuntime field type str  type object
	SiteInfo *GslbSiteRuntimeInfo `json:"site_info,omitempty"`

	// Placeholder for description of property site_stats of obj type GslbSiteRuntime field type str  type object
	SiteStats *GslbSiteRuntimeStats `json:"site_stats,omitempty"`

	// Remap the tenant_uuid to its tenant-name so that we can use the tenant_name directly in remote-site ops. . Field introduced in 17.2.5.
	TenantName *string `json:"tenant_name,omitempty"`

	// This field shadows the glb_cfg.view_id.  . Field introduced in 17.2.5.
	ViewID *int64 `json:"view_id,omitempty"`
}
