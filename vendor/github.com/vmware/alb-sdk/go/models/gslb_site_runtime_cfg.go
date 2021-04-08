package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbSiteRuntimeCfg gslb site runtime cfg
// swagger:model GslbSiteRuntimeCfg
type GslbSiteRuntimeCfg struct {

	// Gslb GeoDb files published for a site. Field introduced in 17.1.1.
	FdInfo *ConfigInfo `json:"fd_info,omitempty"`

	// Gslb Application Persistence info published for a site. Field introduced in 17.1.1.
	GapInfo *ConfigInfo `json:"gap_info,omitempty"`

	// Gslb GeoDb info published for a site. Field introduced in 17.1.1.
	GeoInfo *ConfigInfo `json:"geo_info,omitempty"`

	// GHM info published for a site.
	GhmInfo *ConfigInfo `json:"ghm_info,omitempty"`

	// Gslb info published for a site.
	GlbInfo *ConfigInfo `json:"glb_info,omitempty"`

	// Gslb PKI info published for a site. Field introduced in 17.1.3.
	GpkiInfo *ConfigInfo `json:"gpki_info,omitempty"`

	// GS info published for a site.
	GsInfo *ConfigInfo `json:"gs_info,omitempty"`

	// Maintenance mode info published for a site.
	MmInfo *ConfigInfo `json:"mm_info,omitempty"`

	// The replication queue for all object-types for a site. Field introduced in 17.2.7.
	ReplQueue *ConfigInfo `json:"repl_queue,omitempty"`

	// Configuration sync-info of the site .
	SyncInfo *GslbSiteCfgSyncInfo `json:"sync_info,omitempty"`
}
